// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fantray

import (
	"fmt"
	"net/rpc"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
	"github.com/platinasystems/go/internal/redis/rpc/args"
	"github.com/platinasystems/go/internal/redis/rpc/reply"
	"github.com/platinasystems/go/internal/sockfile"
)

const Name = "fantray"

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

var (
	Init = func() {}
	once sync.Once

	Vdev I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
)

type cmd struct {
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

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	once.Do(Init)

	var si syscall.Sysinfo_t
	var err error

	first = 1
	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]float64)
	cmd.lasts = make(map[string]string)
	cmd.lastu = make(map[string]uint16)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if cmd.rpc, err = sockfile.NewRpcServer(Name); err != nil {
		return err
	}

	rpc.Register(&cmd.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", Name, "Info")
		if err != nil {
			return err
		}
	}

	holdoff := 3
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if holdoff > 0 {
				holdoff--
			}
			if holdoff == 0 {
				if err = cmd.update(); err != nil {
					close(cmd.stop)
					return err
				}
			}
		}
	}
	return nil
}

func (cmd *cmd) Close() error {
	close(cmd.stop)
	return nil
}

func (cmd *cmd) update() error {
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
		if v != cmd.lasts[k] {
			cmd.pub.Print(k, ": ", v)
			cmd.lasts[k] = v
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
		o |= fanTrayLedOff[i]
	} else {
		//get fan tray air direction
		if (rInputNGet & fanTrayDirBits[i]) != 0 {
			f = "front->back"
		} else {
			f = "back->front"
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
		} else if (r1 > minRpm) && (r2 > minRpm) {
			w = "ok" + "." + f
			o |= fanTrayLedGreen[i]
		} else {
			w = "warning check fan tray"
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
	err := i.set(args.Field, string(args.Value), false)
	if err == nil {
		*reply = 1
		WrRegVal[args.Field] = string(args.Value)
	}
	return err
}

func (i *Info) set(key, value string, isReadyEvent bool) error {
	i.pub.Print(key, ": ", value)
	return nil
}

func (i *Info) publish(key string, value interface{}) {
	i.pub.Print(key, ": ", value)
}
