// Copyright © 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package qsfeventsp is the interrupt handler for the Front QSFP port on MC board
// of CH1. It publishes to redis.

package qsfpeventsd

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/uiodevs"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/atsock"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
)

const (
	QSFP_RESET_BIT  = 1 << 2
	QSFP_LPMODE_BIT = 1 << 3
	MAX_IRQ_EVENTS  = 32
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

var (
	first  int
	VdevEp I2cDev // qsfp internal eeprom
	VdevIo I2cDev // qsfp io lines via pca9534

	Slotid        int // temporary
	portLpage0    lpage0
	portUpage3    upage3
	PortIsCopper  bool
	New_present_n uint8
	Old_present_n uint8
)

func (*Command) String() string { return "qsfpeventsd" }

func (*Command) Usage() string { return "qsfpeventsd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "qsfpeventsd monitoring daemon, publishes to redis",
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
	x, err := uiodevs.GetIndex("qsfpeventsd")
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

	// Dynamic data handler
	go qsfpioTicker(c)

	// Event loop
	data := make([]byte, 4)
	for {
		// Epoll blocks and receives number of events
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
		// log.Print("unmask interrupt: ", err)
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
		// log.Print("unmask interrupt: ", err)
		return err
	}
	return nil
}

func qsfpioTicker(c *Command) error {
	t := time.NewTicker(1 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err := c.updateDynamic(); err != nil {
				close(c.stop)
				return err
			}
		}
	}
}

func (c *Command) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}

	if first == 1 {
		// initialize pca9534
		if err := VdevIo.QsfpInit(0xff, 0x00, 0x33); err != nil {
			return err
		}
		first = 0
	}

	// reads qsfp status register and updates
	// New_present_n variable
	k := "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.presence"
	v := VdevIo.QsfpStatus(uint8(Slotid))
	if v != c.lasts[k] {
		c.pub.Print(k, ": ", v)
		c.lasts[k] = v
	}

	//*** TBD
	//ready, err := redis.Hget(machine.Name, "vnet.ready")
	//if err != nil || ready == "false" {
	//	return nil
	//}

	//when qsfp is installed or removed from a port
	if Old_present_n != New_present_n {
		var typeString string

		// when qsfp is installed, publish static data
		if (New_present_n & 0x01) == 0 {
			k := "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.compliance"
			v := VdevEp.Compliance() // reads qsfp eeprom's field

			//identify copper vs optic and set media and speed
			var portConfig string

			//*** TBD
			/****
			media, err := redis.Hget(machine.Name, "vnet.eth-"+"M"+strconv.Itoa(Slotid)+".media")
			if err != nil {
				log.Print("qsfp hget error:", err)
			}
			speed, err := redis.Hget(machine.Name, "vnet.eth-"+"M"+strconv.Itoa(Slotid)+".speed")
			if err != nil {
				log.Print("qsfp hget error:", err)
			}
			****/

			if strings.Contains(v, "-CR") {
				PortIsCopper = true
				/****
				if media != "copper" {
					ret, err := redis.Hset(machine.Name,
						"vnet.eth-"+"M"+strconv.Itoa(Slotid)+".media", "copper")
					if err != nil || ret != 1 {
						log.Print("qsfp hset error:", err, " ", ret)
					} else {
						portConfig += "copper "
					}
				}
				****/
			} else if strings.Contains(v, "40G") {
				PortIsCopper = false
				/****
				if media != "fiber" {
					ret, err := redis.Hset(machine.Name,
						"vnet.eth-"+"M"+strconv.Itoa(Slotid)+".media", "fiber")
					if err != nil || ret != 1 {
						log.Print("qsfp hset error:", err, " ", ret)
					} else {
						portConfig += "fiber "
					}
				}
				if speed != "40g" {
					ret, err := redis.Hset(machine.Name,
						"vnet.eth-"+"M"+strconv.Itoa(Slotid)+".speed", "40g")
					if err != nil || ret != 1 {
						log.Print("qsfp hset error:", err, " ", ret)
					} else {
						portConfig += "40g fixed speed"
					}
				}
				****/
			} else {
				PortIsCopper = false
				/****
				if media != "fiber" {
					ret, err := redis.Hset(machine.Name,
						"vnet.eth-"+"M"+strconv.Itoa(Slotid)+".media", "fiber")
					if err != nil || ret != 1 {
						log.Print("qsfp hset error:", err, " ", ret)
					} else {
						portConfig += "fiber "
					}
				}
				****/
			}

			typeString += strings.Trim(v, " ") + ", "
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}

			k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.vendor"
			v = VdevEp.Vendor()
			typeString += strings.Trim(v, " ") + ", "
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
			k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.partnumber"
			v = VdevEp.PN()
			typeString += strings.Trim(v, " ") + ", "
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
			k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.serialnumber"
			v = VdevEp.SN()
			typeString += strings.Trim(v, " ")
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}

			// get monitoring thresholds if qsfp is not a cable
			if !PortIsCopper {
				VdevEp.StaticBlocks(Slotid)
				v = Temp(portUpage3.tempHighAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.temperature.highAlarmThreshold.units.C"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Temp(portUpage3.tempLowAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.temperature.lowAlarmThreshold.units.C"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Temp(portUpage3.tempHighWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.temperature.highWarnThreshold.units.C"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Temp(portUpage3.tempLowWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.temperature.lowWarnThreshold.units.C"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Voltage(portUpage3.vccHighAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.vcc.highAlarmThreshold.units.V"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Voltage(portUpage3.vccLowAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.vcc.lowAlarmThreshold.units.V"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Voltage(portUpage3.vccHighWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.vcc.highWarnThreshold.units.V"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Voltage(portUpage3.vccLowWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.vcc.lowWarnThreshold.units.V"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.rxPowerHighAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.rx.power.highAlarmThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.rxPowerLowAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.rx.power.lowAlarmThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.rxPowerHighWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.rx.power.highWarnThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.rxPowerLowWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.rx.power.lowWarnThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.txPowerHighAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.power.highAlarmThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.txPowerLowAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.power.lowAlarmThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.txPowerHighWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.power.highWarnThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = Power(portUpage3.txPowerLowWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.power.lowWarnThreshold.units.mW"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = TxBias(portUpage3.txBiasHighAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.biasHighAlarmThreshold.units.mA"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = TxBias(portUpage3.txBiasLowAlarm)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.biasLowAlarmThreshold.units.mA"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = TxBias(portUpage3.txBiasHighWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.biasHighWarnThreshold.units.mA"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				v = TxBias(portUpage3.txBiasLowWarning)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx.biasLowWarnThreshold.units.mA"
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
			}
			log.Printf("QSFP detected in MC%d port: %s", Slotid, typeString)
			if portConfig != "" {
				log.Print("MC Port ", Slotid, " setting changed to ", portConfig)
			}
		} else {
			//when qsfp is removed, delete associated fields
			for _, v := range redisFields {
				k := "port-" + "M" + strconv.Itoa(Slotid) + "." + v
				c.pub.Print("delete: ", k)
				c.lasts[k] = ""
			}
			log.Printf("QSFP removed from MC%d port", Slotid)
			PortIsCopper = true
		}
	}
	Old_present_n = New_present_n
	return nil
}

func (c *Command) updateDynamic() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}

	// TBD **
	//ready, err := redis.Hget(machine.Name, "vnet.ready")
	//if err != nil || ready == "false" {
	//        return nil
	//}

	// publish dynamic monitoring data
	// get monitoring data only if qsfp is present and not a cable
	if (New_present_n & 0x01) == 0 {
		if !PortIsCopper {
			if VdevEp.DataReady() {
				VdevEp.DynamicBlocks(Slotid)
				k := "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.temperature.units.C"
				v := Temp(portLpage0.freeMonTemp)
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.vcc.units.V"
				v = Voltage(portLpage0.freeMonVoltage)
				if v != c.lasts[k] {
					c.pub.Print(k, ": ", v)
					c.lasts[k] = v
				}
				va := LanePower(portLpage0.rxPower)
				for x := 0; x < 4; x++ {
					k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.rx" + strconv.Itoa(x+1) + ".power.units.mW"
					if va[x] != c.lasts[k] {
						c.pub.Print(k, ": ", va[x])
						c.lasts[k] = va[x]
					}
				}
				va = LanePower(portLpage0.txPower)
				for x := 0; x < 4; x++ {
					k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx" + strconv.Itoa(x+1) + ".power.units.mW"
					if va[x] != c.lasts[k] {
						c.pub.Print(k, ": ", va[x])
						c.lasts[k] = va[x]
					}
				}
				va = LanesTxBias(portLpage0.txBias)
				for x := 0; x < 4; x++ {
					k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx" + strconv.Itoa(x+1) + ".bias.units.mA"
					if va[x] != c.lasts[k] {
						c.pub.Print(k, ": ", va[x])
						c.lasts[k] = va[x]
					}
				}
				vs := ChannelAlarms(portLpage0.channelStatusInterrupt,
					portLpage0.channelMonitorInterruptFlags)
				for x := 0; x < 4; x++ {
					k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.rx" + strconv.Itoa(x+1) + ".alarms"
					if vs[x] != c.lasts[k] {
						c.pub.Print(k, ": ", vs[x])
						c.lasts[k] = vs[x]
					}
				}
				for x := 4; x < 8; x++ {
					k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.tx" + strconv.Itoa(x-3) + ".alarms"
					if vs[x] != c.lasts[k] {
						c.pub.Print(k, ": ", vs[x])
						c.lasts[k] = vs[x]
					}
				}
				vs[0] = FreeSideAlarms(portLpage0.freeMonitorInterruptFlags)
				k = "port-" + "M" + strconv.Itoa(Slotid) + ".qsfp.alarms"
				if vs[0] != c.lasts[k] {
					c.pub.Print(k, ": ", vs[0])
					c.lasts[k] = vs[0]
				}
			}
		}
	}
	return nil
}

func (h *I2cDev) DataReady() bool {
	var t bool
	r := getRegsLpage0()

	r.status.get(h)
	closeMux(h)
	DoI2cRpc()

	if (s[1].D[1] & 0x1) == 1 {
		t = false
	} else {
		t = true
	}

	return t
}
func (h *I2cDev) Compliance() string {
	r := getRegsUpage0()

	r.SpecCompliance.get(h)
	closeMux(h)
	DoI2cRpc()
	cp := s[1].D[0]

	r.ExtSpecCompliance.get(h)
	closeMux(h)
	DoI2cRpc()
	ecp := s[1].D[0]

	var t string
	if (cp & 0x80) != 0x80 {
		t = specComplianceValues[cp]
	} else {
		t = extSpecComplianceValues[ecp]
	}
	return t
}

func (h *I2cDev) Vendor() string {
	r := getRegsUpage0()
	r.VendorName.get(h, 16)
	closeMux(h)
	DoI2cRpc()
	t := string(s[1].D[1:16])

	return t
}

func (h *I2cDev) PN() string {
	r := getRegsUpage0()
	r.VendorPN.get(h, 16)
	closeMux(h)
	DoI2cRpc()
	t := string(s[1].D[1:16])

	return t
}

func (h *I2cDev) SN() string {
	r := getRegsUpage0()
	r.VendorSN.get(h, 16)
	closeMux(h)
	DoI2cRpc()
	t := string(s[1].D[1:16])

	return t
}

func LanePower(t [8]byte) [4]string {
	var v [4]string
	var u uint16
	for i := 0; i < 4; i++ {
		u = uint16(t[i*2])<<8 + uint16(t[i*2+1])
		v[i] = strconv.FormatFloat(float64(u)*0.0001, 'f', 3, 64)
	}
	return v
}

func Power(t uint16) string {
	v := strconv.FormatFloat(float64(t)*0.0001, 'f', 3, 64)
	return v
}

func LanesTxBias(t [8]byte) [4]string {
	var v [4]string
	var u uint16
	for i := 0; i < 4; i++ {
		u = uint16(t[i*2])<<8 + uint16(t[i*2+1])
		v[i] = strconv.FormatFloat(float64(u)*0.002, 'f', 3, 64)
	}
	return v
}

func TxBias(t uint16) string {
	v := strconv.FormatFloat(float64(t)*0.002, 'f', 3, 64)
	return v
}

func FreeSideAlarms(t [3]byte) string {
	var v string
	if ((1 << 4) & t[0]) != 0 {
		v += "TempLowWarn,"
	}
	if ((1 << 5) & t[0]) != 0 {
		v += "TempHighWarn,"
	}
	if ((1 << 6) & t[0]) != 0 {
		v += "TempLowAlarm,"
	}
	if ((1 << 7) & t[0]) != 0 {
		v += "TempHighAlarm,"
	}
	if ((1 << 4) & t[1]) != 0 {
		v += "VccLowWarn,"
	}
	if ((1 << 5) & t[1]) != 0 {
		v += "VccHighWarn,"
	}
	if ((1 << 6) & t[1]) != 0 {
		v += "VccLowAlarm,"
	}
	if ((1 << 7) & t[1]) != 0 {
		v += "VccHighAlarm,"
	}
	if v == "" {
		v = "none"
	}
	v = strings.Trim(v, ",")
	return v
}

func ChannelAlarms(t [3]byte, w [6]byte) [8]string {
	var v [8]string
	for i := uint(0); i < 4; i++ {
		if ((1 << i) & t[0]) != 0 {
			v[i] += "RxLos,"
		}
		if ((1 << (i + 4)) & t[0]) != 0 {
			v[i+4] += "TxLos,"
		}
		if ((1 << i) & t[1]) != 0 {
			v[i+4] += "TxFault,"
		}
		if ((1 << (i + 4)) & t[1]) != 0 {
			v[i+4] += "TxEqFault,"
		}
		if ((1 << i) & t[2]) != 0 {
			v[i] += "RxCdrLol,"
		}
		if ((1 << (i + 4)) & t[2]) != 0 {
			v[i+4] += "TxCdrLol,"
		}
		if ((1 << (4 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerLowWarn,"
		}
		if ((1 << (5 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerHighWarn,"
		}
		if ((1 << (6 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerLowAlarm,"
		}
		if ((1 << (7 - (i%2)*4)) & w[i/2]) != 0 {
			v[i] += "PowerHighAlarm,"
		}
		if ((1 << (4 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasLowWarn,"
		}
		if ((1 << (5 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasHighWarn,"
		}
		if ((1 << (6 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasLowAlarm,"
		}
		if ((1 << (7 - (i%2)*4)) & w[i/2+2]) != 0 {
			v[i+4] += "BiasHighAlarm,"
		}
		if ((1 << (4 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerLowWarn,"
		}
		if ((1 << (5 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerHighWarn,"
		}
		if ((1 << (6 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerLowAlarm,"
		}
		if ((1 << (7 - (i%2)*4)) & w[i/2+4]) != 0 {
			v[i+4] += "PowerHighAlarm,"
		}
		if v[i] == "" {
			v[i] = "none"
		}
		if v[i+4] == "" {
			v[i+4] = "none"
		}
		v[i] = strings.Trim(v[i], ",")
		v[i+4] = strings.Trim(v[i+4], ",")
	}
	return v
}

func Temp(t uint16) string {
	var u float64
	var v string
	if (t & 0x8000) != 0 {
		u = float64((t^0xffff)+1) / 256 * (-1)
	} else {
		u = float64(t) / 256
	}
	v = strconv.FormatFloat(u, 'f', 1, 64)
	return v
}

func Voltage(t uint16) string {
	var u float64
	var v string
	u = float64(t) * 0.0001
	v = strconv.FormatFloat(u, 'f', 3, 64)
	return v
}
func (h *I2cDev) DynamicBlocks(port int) {
	r := getRegsLpage0()
	r.pageSelect.set(h, 0)
	closeMux(h)
	DoI2cRpc()
	rb := getBlocks()
	rb.lpage0b.get(h, 32)
	closeMux(h)
	DoI2cRpc()

	portLpage0.id = s[1].D[1]
	portLpage0.status = uint16(s[1].D[2]) + uint16(s[1].D[3])<<8
	copy(portLpage0.channelStatusInterrupt[:], s[1].D[4:7])
	copy(portLpage0.freeMonitorInterruptFlags[:], s[1].D[7:10])
	copy(portLpage0.channelMonitorInterruptFlags[:], s[1].D[10:16])
	portLpage0.freeMonTemp = uint16(s[1].D[24]) + uint16(s[1].D[23])<<8
	portLpage0.freeMonVoltage = uint16(s[1].D[28]) + uint16(s[1].D[27])<<8

	rb.lpage1b.get(h, 32)
	closeMux(h)
	DoI2cRpc()

	copy(portLpage0.rxPower[:], s[1].D[3:11])
	copy(portLpage0.txBias[:], s[1].D[11:19])
	copy(portLpage0.txPower[:], s[1].D[19:27])
}

func (h *I2cDev) StaticBlocks(port int) {
	if !PortIsCopper {
		r := getRegsLpage0()
		r.pageSelect.set(h, 3)
		closeMux(h)
		DoI2cRpc()

		rb := getBlocks()
		rb.upage0b.get(h, 32)
		closeMux(h)
		DoI2cRpc()
		portUpage3.tempHighAlarm = (uint16(s[1].D[1]) << 8) + uint16(s[1].D[2])
		portUpage3.tempLowAlarm = (uint16(s[1].D[3]) << 8) + uint16(s[1].D[4])
		portUpage3.tempHighWarning = (uint16(s[1].D[5]) << 8) + uint16(s[1].D[6])
		portUpage3.tempLowWarning = (uint16(s[1].D[7]) << 8) + uint16(s[1].D[8])
		portUpage3.vccHighAlarm = (uint16(s[1].D[17]) << 8) + uint16(s[1].D[18])
		portUpage3.vccLowAlarm = (uint16(s[1].D[19]) << 8) + uint16(s[1].D[20])
		portUpage3.vccHighWarning = (uint16(s[1].D[21]) << 8) + uint16(s[1].D[22])
		portUpage3.vccLowWarning = (uint16(s[1].D[23]) << 8) + uint16(s[1].D[24])

		rb.upage1b.get(h, 32)
		closeMux(h)
		DoI2cRpc()
		portUpage3.rxPowerHighAlarm = (uint16(s[1].D[17]) << 8) + uint16(s[1].D[18])
		portUpage3.rxPowerLowAlarm = (uint16(s[1].D[19]) << 8) + uint16(s[1].D[20])
		portUpage3.rxPowerHighWarning = (uint16(s[1].D[21]) << 8) + uint16(s[1].D[22])
		portUpage3.rxPowerLowWarning = (uint16(s[1].D[23]) << 8) + uint16(s[1].D[24])
		portUpage3.txBiasHighAlarm = (uint16(s[1].D[25]) << 8) + uint16(s[1].D[26])
		portUpage3.txBiasLowAlarm = (uint16(s[1].D[27]) << 8) + uint16(s[1].D[28])
		portUpage3.txBiasHighWarning = (uint16(s[1].D[29]) << 8) + uint16(s[1].D[30])
		portUpage3.txBiasLowWarning = (uint16(s[1].D[31]) << 8) + uint16(s[1].D[32])

		rb.upage2b.get(h, 32)
		closeMux(h)
		DoI2cRpc()
		portUpage3.txPowerHighAlarm = (uint16(s[1].D[1]) << 8) + uint16(s[1].D[2])
		portUpage3.txPowerLowAlarm = (uint16(s[1].D[3]) << 8) + uint16(s[1].D[4])
		portUpage3.txPowerHighWarning = (uint16(s[1].D[5]) << 8) + uint16(s[1].D[6])
		portUpage3.txPowerLowWarning = (uint16(s[1].D[7]) << 8) + uint16(s[1].D[8])

		r = getRegsLpage0()
		r.pageSelect.set(h, 0)
		closeMux(h)
		DoI2cRpc()
	}
	return
}

func (h *I2cDev) QsfpStatus(port uint8) string {
	var present_n uint8

	p := h.ReadMuxInputReg()
	present_n = uint8(p & 0x01) // present_n bit

	//if module was removed or inserted, set reset and lpmode lines accordingly
	if Old_present_n != present_n { // sanity check
		r := getRegs()
		v := uint8((p & QSFP_RESET_BIT & QSFP_LPMODE_BIT))
		r.Output.set(h, v)
		closeMux(h)
		DoI2cRpc()

		New_present_n = present_n
	}

	if (present_n & 0x01) == 1 {
		return "empty"
	}
	return "installed"
}

func (h *I2cDev) QsfpInit(out0 byte, pol0 byte, conf0 byte) error {
	//all ports default in reset
	r := getRegs()
	r.Output.set(h, out0)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	r.Polarity.set(h, pol0)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	r.Config.set(h, conf0)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) ReadMuxInputReg() uint8 {
	r := getRegs()
	r.Input.get(h)
	closeMux(h)
	DoI2cRpc() // reads pca8534 register
	data := s[1].D[0]
	return data
}
