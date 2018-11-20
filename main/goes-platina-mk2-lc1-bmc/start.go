// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	//"github.com/platinasystems/go/goes/cmd/platina/mk1/bmc/upgrade"
	"github.com/platinasystems/gpio"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis/publisher"
)

const (
	MaxEpollEvents = 32
	KB             = 1024
)

func startConfGpioHook() error {
	gpioInit()
	pin, found := gpio.Pins["QSPI_MUX_SEL"]
	if found {
		r, _ := pin.Value()
		if r {
			log.Print("Booted from QSPI1")
		} else {
			log.Print("Booted from QSPI0")
		}

	}

	for name, pin := range gpio.Pins {
		err := pin.SetDirection()
		if err != nil {
			fmt.Printf("%s: %v\n", name, err)
		}
	}
	pin, found = gpio.Pins["LOCAL_I2C_RESET_L"]
	if found {
		pin.SetValue(false)
		time.Sleep(1 * time.Microsecond)
		pin.SetValue(true)
	}

	pin, found = gpio.Pins["FRU_I2C_RESET_L"]
	if found {
		pin.SetValue(false)
		time.Sleep(1 * time.Microsecond)
		pin.SetValue(true)
	}
	err := pubEth0()
	if err != nil {
		return err
	}
	//upgrade.UpdateEnv(false)
	//upgrade.UpdateEnv(true)
	return nil
}

func pubEth0() (err error) {
	var event syscall.EpollEvent
	var events [MaxEpollEvents]syscall.EpollEvent
	var buf [KB]byte

	f, err := os.Open("/dev/kmsg")
	if err != nil {
		return err
	}
	defer f.Close()

	fd := int(f.Fd())
	if err = syscall.SetNonblock(fd, true); err != nil {
		return err
	}
	epfd, err := syscall.EpollCreate1(0)
	if err != nil {
		return err
	}
	defer syscall.Close(epfd)

	event.Events = syscall.EPOLLIN
	event.Fd = int32(fd)
	err = syscall.EpollCtl(epfd, syscall.EPOLL_CTL_ADD, fd, &event)
	if err != nil {
		return err
	}
	nevents, err := syscall.EpollWait(epfd, events[:], -1)
	if err != nil {
		return err
	}
	for ev := 0; ev < nevents; ev++ {
		for {
			nbytes, err := syscall.Read(int(events[ev].Fd), buf[:])
			if nbytes > 0 {
				x := string(buf[:nbytes])
				if strings.Contains(x, "init.redisd") {
					if strings.Contains(x, "eth0") {
						er := pubAddr(x)
						if er != nil {
							return er
						}
					}
				}
			}
			if err != nil {
				break
			}
		}
	}
	return nil
}

func pubAddr(s string) (err error) {
	ip := strings.SplitAfter(s, "[")
	i := ip[2]
	ip = strings.Split(i, "%")
	if strings.Contains(s, "::") {
		err = pubKey("eth0.ipv6", ip[0])
		if err != nil {
			return err
		}
	} else {
		err = pubKey("eth0.ipv4", ip[0])
		if err != nil {
			return err
		}
	}
	return nil
}

func pubKey(k string, v interface{}) (err error) {
	var pub *publisher.Publisher
	if pub, err = publisher.New(); err != nil {
		return err
	}
	pub.Print(k, ": ", v)
	return nil
}
