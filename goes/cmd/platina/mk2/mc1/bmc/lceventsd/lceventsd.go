// Copyright © 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package lceventsd is the interrupt handler for LC present signals
// on CH1. It publishes to redis.

package lceventsd

import (
	"fmt"
	"os"
	"strconv"
	"sync"
	"syscall"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/uiodevs"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/atsock"
	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
)

const MAX_IRQ_EVENTS = 8

var (
	VdevIo I2cDev // lcabs status via pca9539

	first         int
	maxLcNumber   int
	ChassisType   uint8
	BoardType     uint8
	New_present_n uint16
	Old_present_n uint16
)

type Command struct {
	Info
	Init func()
	init sync.Once
}

type Info struct {
	mutex sync.Mutex
	rpc   *atsock.RpcServer
	pub   *publisher.Publisher
	stop  chan struct{}
	last  map[string]uint16
	lasts map[string]string
}

type I2cDev struct {
	Bus       int
	Addr      int
	MuxBus    int
	MuxAddr   int
	MuxValue  int
	MuxBus2   int
	MuxAddr2  int
	MuxValue2 int
}

type uioDev struct {
	Name  string   // name of gpio or int in in dts file
	File  *os.File // uio device in linux dev directory
	Fd    int
	Count int
}

func (*Command) String() string { return "lceventsd" }

func (*Command) Usage() string { return "lceventsd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "lceventsd server daemon",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

func (c *Command) Main(...string) error {
	var si syscall.Sysinfo_t
	var err error
	var event syscall.EpollEvent
	var revents [MAX_IRQ_EVENTS]syscall.EpollEvent // received events

	if c.Init != nil {
		c.init.Do(c.Init)
	}

	err = redis.IsReady()
	if err != nil {
		return err
	}

	first = 1

	c.stop = make(chan struct{})
	c.last = make(map[string]uint16)
	c.lasts = make(map[string]string)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}
	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	// Setup UIO device
	x, err := uiodevs.GetIndex("lceventsd")
	if err != nil {
		log.Print("uio device not found")
		return err
	}
	dir := fmt.Sprintf("/dev/uio%d", x)
	file, err := os.OpenFile(dir, os.O_RDWR, 0)
	if e, ok := err.(*os.PathError); ok && e.Err == syscall.ENOENT {
		log.Print("opening uio: ", err)
		return err
	}
	defer file.Close()
	fd := int(file.Fd())
	if err = syscall.SetNonblock(fd, true); err != nil {
		log.Print("setnonblock: ", err)
	}
	dev := new(uioDev)
	dev.File = file
	dev.Fd = fd
	dev.Count = int(0)

	// Create Epoll file descriptor
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		log.Print("epoll create: ", err)
		return err
	}
	defer syscall.Close(epfd)

	// Add file descriptor of uio device to the Epoll facility
	event.Events = (syscall.EPOLLIN)
	event.Fd = int32(dev.Fd)
	if err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, dev.Fd, &event); err != nil {
		log.Print("epoll ctl: ", err)
		return err
	}

	// Call update() to initialize and clear hw interrupt
	if err := c.update(); err != nil {
		close(c.stop)
	}

	// Unmask device interrupt
	if err := dev.IrqEnable(); err != nil {
		log.Print("unmask interrupt: ", err)
		return err
	}

	// Event loop
	data := make([]byte, 4)
	for {
		// Epoll blocks and receives number of events when woken
		nevents, err := syscall.EpollWait(epfd, revents[:], -1)
		if err != nil {
			log.Print("epoll wait: ", err)
			break
		}
		log.Print("Irq event(s): ", nevents)
		for i := 0; i < nevents; i++ {
			// Error occurred
			if ((revents[i].Events & syscall.EPOLLERR) != 0) ||
				((revents[i].Events & syscall.EPOLLIN) == 0) {
				log.Print("epoll error")
				continue
			} else {
				// checks ownership
				fd := int(revents[i].Fd)
				if fd == dev.Fd {
					// Read file descriptor to clear event
					file := dev.File
					_, err := file.Read(data)
					if err != nil {
						log.Print("read descriptor: ", err)
						continue
					}
					dev.Count = int((data[3] << 24) | (data[2] << 16) | (data[1] << 8) | data[0])
					// log.Print("irq count: ", dev.Count)

					// Handle event
					if err := c.update(); err != nil {
						close(c.stop)
					}

					// Unmask interrupt
					if err := dev.IrqEnable(); err != nil {
						log.Print("unmask interrupt: ", err)
						continue
					}
				}
			}
		}
	}
	return nil
}

func (dev *uioDev) IrqEnable() error {
	mask := []byte{0x01, 0x00, 0x00, 0x00}

	// Unmask device interrupt
	file := dev.File
	_, err := file.Write(mask)
	if err != nil {
		return err
	}
	return nil
}

func (dev *uioDev) IrqDisable() error {
	mask := []byte{0x00, 0x00, 0x00, 0x00}

	// Mask device interrupt
	file := dev.File
	_, err := file.Write(mask)
	if err != nil {
		return err
	}
	return nil
}

func (c *Command) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}
	if first == 1 {
		const (
			TOR1    uint8 = 0x00
			CH1_4S  uint8 = 0x01
			CH1_8S  uint8 = 0x02
			CH1_16S uint8 = 0x03
			CH1MC   uint8 = 0x04
			CH1LC   uint8 = 0x05
		)
		d := eeprom.Device{
			BusIndex:   0,
			BusAddress: 0x55,
		}
		if err := d.GetInfo(); err != nil {
			return err
		}
		switch d.Fields.ChassisType {
		case CH1_4S:
			maxLcNumber = int(4)
		case CH1_8S:
			maxLcNumber = int(8)
		case CH1_16S:
			maxLcNumber = int(16)
		default:
		}

		// initialize pca9539
		if err := VdevIo.LcabsInit(0xff, 0xff, 0x00, 0x00, 0xff, 0xff); err != nil {
			return err
		}
		first = 0
	}

	New_present_n = VdevIo.ReadMuxInputReg()
	if Old_present_n != New_present_n {
		var v string
		for i := 1; i <= maxLcNumber; i++ {
			k := "LC-" + strconv.Itoa(i) + ".presence"
			if ((New_present_n >> uint8(i)) & 0x01) == 1 {
				v = "empty"
			} else {
				v = "installed"
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v

			}
		}
	}
	Old_present_n = New_present_n
	return nil
}

func (h *I2cDev) LcabsInit(out0 byte, out1 byte, pol0 byte, pol1 byte, conf0 byte, conf1 byte) error {
	//all ports default in reset
	r := getRegs()
	r.Output[0].set(h, out0)
	r.Output[1].set(h, out0)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	r.Polarity[0].set(h, pol0)
	r.Polarity[1].set(h, pol1)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	r.Config[0].set(h, conf0)
	r.Config[1].set(h, conf1)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) ReadMuxInputReg() uint16 {
	r := getRegs()
	r.Input[0].get(h)
	r.Input[1].get(h)
	closeMux(h)
	DoI2cRpc()
	p := uint16(s[1].D[0])
	p += uint16(s[3].D[0]) << 8
	return p
}
