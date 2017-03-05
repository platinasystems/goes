// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package fantray

import (
	"strconv"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "fantray"

type I2cDev struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

var Vdev I2cDev

var VpageByKey map[string]uint8

type cmd struct {
	stop chan struct{}
	pub  *publisher.Publisher
	last map[string]string
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	var err error

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]string)

	if cmd.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	//if err = cmd.update(); err != nil {
	//	close(cmd.stop)
	//	return err
	//}
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-cmd.stop:
			return nil
		case <-t.C:
			if err = cmd.update(); err != nil {
				close(cmd.stop)
				return err
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
	for k, i := range VpageByKey {
		v := Vdev.FanTrayStatus(i)
		if v != cmd.last[k] {
			cmd.pub.Print(k, ": ", v)
			cmd.last[k] = v
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
var deviceVer byte

func (h *I2cDev) FanTrayLedInit() {
	r := getRegs()

	//e := eeprom.Device{
	//	BusIndex:   0,
	//	BusAddress: 0x55,
	//}
	//e.GetInfo()
	//deviceVer = e.Fields.DeviceVersion
	deviceVer := 0xff //
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
	log.Print("notice: fan tray led init complete")
}

func (h *I2cDev) FanTrayStatus(i uint8) string {
	var w string
	var f string

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
	DoI2cRpc()
	o := s[1].D[0]
	d := 0xff ^ fanTrayLedBits[i]
	o &= d

	r.Input[n].get(h)
	closeMux(h)
	DoI2cRpc()
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
		f1 := "fan_tray." + strconv.Itoa(int(i+1)) + ".1.rpm"
		f2 := "fan_tray." + strconv.Itoa(int(i+1)) + ".2.rpm"
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
	DoI2cRpc()
	return w
}
