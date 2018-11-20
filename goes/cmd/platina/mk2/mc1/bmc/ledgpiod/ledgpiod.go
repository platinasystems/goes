// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package ledgpiod

import (
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/sockfile"
	"github.com/platinasystems/gpio"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
	"github.com/platinasystems/redis/rpc/args"
	"github.com/platinasystems/redis/rpc/reply"
)

const (
	Name    = "ledgpiod"
	Apropos = "FIXME"
	Usage   = "ledgpiod"
)

var apropos = lang.Alt{
	lang.EnUS: Apropos,
}

func New() *Command { return new(Command) }

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

const (
	maxFanTrays = 4
	maxPsu      = 2
)

var (
	lastFanStatus [maxFanTrays]string
	lastPsuStatus [maxPsu]string

	psuLed       = []uint8{0x8, 0x10}
	psuLedYellow = []uint8{0x8, 0x10}
	psuLedOff    = []uint8{0x04, 0x01}

	sysLed       byte = 0x1
	sysLedGreen  byte = 0x1
	sysLedYellow byte = 0xc
	sysLedOff    byte = 0x80

	fanLed       byte = 0x6
	fanLedGreen  byte = 0x2
	fanLedYellow byte = 0x6
	fanLedOff    byte = 0x0

	deviceVer          byte
	forceFanSpeed      bool
	systemFanDirection string
	once               sync.Once
	Init               = func() {}

	first int

	Vdev I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)
)

type Command struct {
	Info
}

type Info struct {
	mutex sync.Mutex
	rpc   *sockfile.RpcServer
	pub   *publisher.Publisher
	stop  chan struct{}
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Kind() cmd.Kind    { return cmd.Daemon }
func (*Command) String() string    { return Name }
func (*Command) Usage() string     { return Name }

func (c *Command) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t

	err := redis.IsReady()
	if err != nil {
		return err
	}

	first = 1

	c.stop = make(chan struct{})
	c.last = make(map[string]float64)
	c.lasts = make(map[string]string)
	c.lastu = make(map[string]uint16)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if c.rpc, err = sockfile.NewRpcServer(Name); err != nil {
		return err
	}

	rpc.Register(&c.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", Name, "Info")
		if err != nil {
			return err
		}
	}

	t := time.NewTicker(2 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if Vdev.Addr != 0 {
				if err = c.update(); err != nil {
				}
			}
		}
	}
	return nil
}

func (c *Command) Close() error {
	close(c.stop)
	return nil
}

func (c *Command) update() error {
	stopped := readStopped()
	if stopped == 1 {
		return nil
	}
	if err := writeRegs(); err != nil {
		return err
	}

	if first == 1 {
		err := Vdev.LedFpInit()
		if err != nil {
			return err
		}
		first = 0
	}
	err := Vdev.LedStatus()
	if err != nil {
		return err
	}

	for k, _ := range VpageByKey {
		if strings.Contains(k, "fan_direction") {
			v := Vdev.CheckSystemFans()
			if (v != "") && (v != c.lasts[k]) {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
	}
	return nil

}

func (h *I2cDev) LedFpInit() error {
	var d byte

	pin, found := gpio.Pins["SYSTEM_LED_RST_L"]
	if found {
		pin.SetValue(true)
	}

	ss, _ := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	_, _ = fmt.Sscan(ss, &deviceVer)

	forceFanSpeed = false

	r := getRegs()
	r.Output[0].get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	o := s[1].D[0]

	//on bmc boot up set front panel SYS led to green, FAN led to yellow, let PSU drive PSU LEDs
	d = 0xff ^ (sysLed | fanLed)
	o &= d
	o |= sysLedGreen | fanLedYellow

	r.Output[0].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	r.Config[0].get(h)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	o = s[1].D[0]
	o |= psuLed[0] | psuLed[1]
	o &= (sysLed | fanLed) ^ 0xff

	r.Config[0].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) LedFpReinit() error {

	ss, _ := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	_, _ = fmt.Sscan(ss, &deviceVer)
	r := getRegs()

	r.Config[0].get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	o := s[1].D[0]
	o |= psuLed[0] | psuLed[1]
	o &= (sysLed | fanLed) ^ 0xff

	r.Config[0].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) LedStatus() error {
	r := getRegs()
	var o, c uint8
	var d byte

	if deviceVer == 0xff || deviceVer == 0x00 {
		psuLed = []uint8{0x0c, 0x03}
		psuLedYellow = []uint8{0x00, 0x00}
		psuLedOff = []uint8{0x04, 0x01}
		sysLed = 0xc0
		sysLedGreen = 0x0
		sysLedYellow = 0xc
		sysLedOff = 0x80
		fanLed = 0x30
		fanLedGreen = 0x10
		fanLedYellow = 0x20
		fanLedOff = 0x30
	}

	allFanGood := true
	fanStatChange := false
	for j := 0; j < maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j+1))+".status")
		if !strings.Contains(p, "ok") {
			allFanGood = false
		}
		if lastFanStatus[j] != p {
			fanStatChange = true
			//if any fan tray is failed or not installed, set front panel FAN led to yellow
			if strings.Contains(p, "warning") && !strings.Contains(lastFanStatus[j], "not installed") {
				r.Output[0].get(h)
				closeMux(h)
				err := DoI2cRpc()
				if err != nil {
					return err
				}
				o = s[1].D[0]
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedYellow
				r.Output[0].set(h, o)
				closeMux(h)
				err = DoI2cRpc()
				if err != nil {
					return err
				}
				log.Print("warning: fan tray ", j+1, " failure")
				if !forceFanSpeed {
					redis.Hset(redis.DefaultHash, "fan_tray.speed", "max")
					forceFanSpeed = true
				}
			} else if strings.Contains(p, "not installed") {
				r.Output[0].get(h)
				closeMux(h)
				err := DoI2cRpc()
				if err != nil {
					return err
				}
				o = s[1].D[0]
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedYellow
				r.Output[0].set(h, o)
				closeMux(h)
				err = DoI2cRpc()
				if err != nil {
					return err
				}
				log.Print("warning: fan tray ", j+1, " not installed")
				if !forceFanSpeed {
					redis.Hset(redis.DefaultHash, "fan_tray.speed", "max")
					forceFanSpeed = true
				}
			} else if strings.Contains(lastFanStatus[j], "not installed") && (strings.Contains(p, "warning") || strings.Contains(p, "ok")) {
				log.Print("notice: fan tray ", j+1, " installed")
			}
		}
		lastFanStatus[j] = p
	}

	//if any fan tray is failed or not installed, set front panel FAN led to yellow
	if fanStatChange {
		if allFanGood {
			// if all fan trays have "ok" status, set front panel FAN led to green
			allStat := true
			for i := range lastFanStatus {
				if lastFanStatus[i] == "" {
					allStat = false
				}
			}
			if allStat {
				r.Output[0].get(h)
				closeMux(h)
				err := DoI2cRpc()
				if err != nil {
					return err
				}
				o = s[1].D[0]
				d = 0xff ^ fanLed
				o &= d
				o |= fanLedGreen
				r.Output[0].set(h, o)
				closeMux(h)
				err = DoI2cRpc()
				if err != nil {
					return err
				}
				log.Print("notice: all fan trays up")
				redis.Hset(redis.DefaultHash, "fan_tray.speed.return", "")
				forceFanSpeed = false
			}
		}

	}

	for j := 0; j < maxPsu; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "psu"+strconv.Itoa(j+1)+".status")
		if lastPsuStatus[j] != p {
			r.Output[0].get(h)
			r.Config[0].get(h)
			closeMux(h)
			err := DoI2cRpc()
			if err != nil {
				return err
			}
			o = s[1].D[0]
			c = s[3].D[0]
			//if PSU is not installed or installed and powered on, set front panel PSU led to off or green (PSU drives)
			if strings.Contains(p, "not_installed") || strings.Contains(p, "powered_on") {
				c |= psuLed[j]
			} else if strings.Contains(p, "powered_off") {
				//if PSU is installed but powered off, set front panel PSU led to yellow
				d = 0xff ^ psuLed[j]
				o &= d
				o |= psuLedYellow[j]
				c &= (psuLed[j]) ^ 0xff
			}
			r.Output[0].set(h, o)
			r.Config[0].set(h, c)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}

			lastPsuStatus[j] = p
			if p != "" {
				log.Print("notice: psu", j+1, " ", p)
			}
		}
	}
	return nil
}

func (h *I2cDev) CheckSystemFans() string {

	mismatch := false
	var n string
	for j := 0; j < maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j+1))+".status")
		var d string

		if strings.Contains(p, "back->front") {
			d = "back->front"
		} else if strings.Contains(p, "front->back") {
			d = "front->back"
		}
		if n == "" && d != "" {
			n = d
		} else if d != "" {
			if n != d {
				systemFanDirection = "mixed"
				mismatch = true
				break
			}
		}
	}
	if !mismatch {
		for i := 0; i < maxPsu; i++ {
			var d string
			p, _ := redis.Hget(redis.DefaultHash, "psu"+strconv.Itoa(i+1)+".fan_direction")
			if strings.Contains(p, "back->front") {
				d = "back->front"
			} else if strings.Contains(p, "front->back") {
				d = "front->back"
			}
			if n == "" && d != "" {
				n = d
			} else if d != "" {
				if n != d {
					systemFanDirection = "mixed"
					mismatch = true
					break
				}
			}

		}
	}
	if mismatch {
		systemFanDirection = "mixed"
		p, _ := redis.Hget(redis.DefaultHash, "system.fan_direction")
		if !strings.Contains(p, "mixed") {
			log.Print("warning: mismatching fan direction detected, check fan trays and PSUs")
		}
	} else {
		if n != "" {
			systemFanDirection = n
		}
	}

	return systemFanDirection
}

func writeRegs() error {
	for k, v := range WrRegVal {
		switch WrRegFn[k] {
		case "speed":
			if false {
				log.Print("test", k, v)
			}
		}
		delete(WrRegVal, k)
	}
	return nil
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	_, p := WrRegFn[args.Field]
	if !p {
		return fmt.Errorf("cannot hset: %s", args.Field)
	}
	_, q := WrRegRng[args.Field]
	if !q {
		err := i.set(args.Field, string(args.Value), false)
		if err == nil {
			*reply = 1
			WrRegVal[args.Field] = string(args.Value)
		}
		return err
	}
	var a [2]int
	var e [2]error
	if len(WrRegRng[args.Field]) == 2 {
		for i, v := range WrRegRng[args.Field] {
			a[i], e[i] = strconv.Atoi(v)
		}
		if e[0] == nil && e[1] == nil {
			val, err := strconv.Atoi(string(args.Value))
			if err != nil {
				return err
			}
			if val >= a[0] && val <= a[1] {
				err := i.set(args.Field,
					string(args.Value), false)
				if err == nil {
					*reply = 1
					WrRegVal[args.Field] =
						string(args.Value)
				}
				return err
			}
			return fmt.Errorf("Cannot hset.  Valid range is: %s",
				WrRegRng[args.Field])
		}
	}
	for _, v := range WrRegRng[args.Field] {
		if v == string(args.Value) {
			err := i.set(args.Field, string(args.Value), false)
			if err == nil {
				*reply = 1
				WrRegVal[args.Field] = string(args.Value)
			}
			return err
		}
	}
	return fmt.Errorf("Cannot hset.  Valid values are: %s",
		WrRegRng[args.Field])
}

func (i *Info) set(key, value string, isReadyEvent bool) error {
	i.pub.Print(key, ": ", value)
	return nil
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}
