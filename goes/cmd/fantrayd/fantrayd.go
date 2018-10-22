// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fantrayd

import (
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/atsock"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
	"github.com/platinasystems/redis/rpc/args"
	"github.com/platinasystems/redis/rpc/reply"
)

var (
	Vdev I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)

	fanTrayA = []string{"not installed", "not installed", "not installed", "not installed"}
)

const nFanTrays = 4

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
	last  map[string]float64
	lasts map[string]string
	lastu map[string]uint16
}

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

func (*Command) String() string { return "fantrayd" }

func (*Command) Usage() string { return "fantrayd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "fantray monitoring daemon, publishes to redis",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	var si syscall.Sysinfo_t

	if c.Init != nil {
		c.init.Do(c.Init)
	}

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

	if c.rpc, err = atsock.NewRpcServer("fantrayd"); err != nil {
		return err
	}

	rpc.Register(&c.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", "fantrayd",
			"Info")
		if err != nil {
			return err
		}
	}

	holdoff := 3
	t := time.NewTicker(5 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if holdoff > 0 {
				holdoff--
			}
			if holdoff == 0 {
				if err = c.update(); err != nil {
					holdoff = 5
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
	for k, i := range VpageByKey {
		v, err := Vdev.FanTrayStatus(i)
		if err != nil {
			return err
		}
		if v != c.lasts[k] {
			c.pub.Print(k, ": ", v)
			c.lasts[k] = v
		}
	}
	return nil
}

const (
	fanTrayLeds = 0x33
	minRpm      = 2000
)

var fanTrayLedOff = []uint8{0x0, 0x0, 0x0, 0x0}
var fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
var fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
var fanTrayLedBits = []uint8{0x30, 0x03, 0x30, 0x03}
var fanTrayDirBits = []uint8{0x80, 0x08, 0x80, 0x08}
var fanTrayAbsBits = []uint8{0x40, 0x04, 0x40, 0x04}
var deviceVer int
var first int

func (h *I2cDev) FanTrayLedInit() error {
	r := getRegs()

	deviceVer = 0x1
	s, _ := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	_, _ = fmt.Sscan(s, &deviceVer)
	if deviceVer == 0xff || deviceVer == 0x00 {
		fanTrayLedGreen = []uint8{0x10, 0x01, 0x10, 0x01}
		fanTrayLedYellow = []uint8{0x20, 0x02, 0x20, 0x02}
	} else {
		fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
		fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
	}

	r.Output[0].set(h, 0xff&(fanTrayLedOff[2]|fanTrayLedOff[3]))
	r.Output[1].set(h, 0xff&(fanTrayLedOff[0]|fanTrayLedOff[1]))
	r.Config[0].set(h, 0xff^fanTrayLeds)
	r.Config[1].set(h, 0xff^fanTrayLeds)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	log.Print("notice: fan tray led init complete")
	return err
}

func (h *I2cDev) FanTrayLedReinit() error {
	r := getRegs()

	deviceVer := 0
	s, _ := redis.Hget(redis.DefaultHash, "eeprom.DeviceVersion")
	_, _ = fmt.Sscan(s, &deviceVer)
	if deviceVer == 0xff || deviceVer == 0x00 {
		fanTrayLedGreen = []uint8{0x10, 0x01, 0x10, 0x01}
		fanTrayLedYellow = []uint8{0x20, 0x02, 0x20, 0x02}
	} else {
		fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
		fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
	}

	r.Config[0].set(h, 0xff^fanTrayLeds)
	r.Config[1].set(h, 0xff^fanTrayLeds)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	log.Print("notice: fan tray led init complete")
	return err
}

func (h *I2cDev) FanTrayStatus(i uint8) (string, error) {
	var w string
	var f string

	if first == 1 {
		err := Vdev.FanTrayLedInit()
		if err != nil {
			return "error", err
		}

		first = 0
	}

	if deviceVer == 0xff || deviceVer == 0x00 {
		fanTrayLedGreen = []uint8{0x10, 0x01, 0x10, 0x01}
		fanTrayLedYellow = []uint8{0x20, 0x02, 0x20, 0x02}
	} else {
		fanTrayLedGreen = []uint8{0x20, 0x02, 0x20, 0x02}
		fanTrayLedYellow = []uint8{0x10, 0x01, 0x10, 0x01}
	}

	r := getRegs()
	n := 0
	i--

	if i < 2 {
		n = 1
	}

	r.Output[n].get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "error", err
	}
	o := s[1].D[0]
	d := 0xff ^ fanTrayLedBits[i]
	o &= d

	r.Input[n].get(h)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return "error", err
	}
	rInputNGet := s[1].D[0]
	if (rInputNGet & fanTrayAbsBits[i]) != 0 {
		//fan tray is not present, turn LED off
		w = "not installed"
		fanTrayA[i] = "not installed"
		o |= fanTrayLedOff[i]
	} else {
		//get fan tray air direction
		if (rInputNGet & fanTrayDirBits[i]) != 0 {
			f = "front->back"
		} else {
			f = "back->front"
		}
		fanTrayA[i] = f

		mismatch := false
		for n := 0; n < nFanTrays; n++ {
			if fanTrayA[n] == "not installed" {
				continue
			} else if fanTrayA[i] != fanTrayA[n] {
				mismatch = true
			}
		}

		//check fan speed is above minimum
		f1 := "fan_tray." + strconv.Itoa(int(i+1)) + ".1.speed.units.rpm"
		f2 := "fan_tray." + strconv.Itoa(int(i+1)) + ".2.speed.units.rpm"
		s1, _ := redis.Hget(redis.DefaultHash, f1)
		s2, _ := redis.Hget(redis.DefaultHash, f2)
		r1, _ := strconv.ParseInt(s1, 10, 64)
		r2, _ := strconv.ParseInt(s2, 10, 64)

		if s1 == "" && s2 == "" {
			o |= fanTrayLedYellow[i]
		} else if ((r1 > minRpm) && (r2 > minRpm)) && !mismatch {
			w = "ok" + "." + f
			o |= fanTrayLedGreen[i]
		} else if mismatch {
			w = "ok" + "." + f
			o |= fanTrayLedYellow[i]
		} else if (r1 <= minRpm) || (r2 < minRpm) {
			w = "warning low rpm detected"
			o |= fanTrayLedYellow[i]
		}
	}

	r.Output[n].set(h, o)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return "error", err
	}
	return w, nil
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
