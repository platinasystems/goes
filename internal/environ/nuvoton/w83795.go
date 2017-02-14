// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package w83795 provides access to the H/W Monitor chip

package w83795

import (
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/log"
	"github.com/platinasystems/go/internal/redis"
	"github.com/platinasystems/go/internal/redis/publisher"
)

const Name = "w83795"

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
	stop  chan struct{}
	pub   *publisher.Publisher
	last  map[string]uint16
	lasts map[string]string
}

func New() *cmd { return new(cmd) }

func (*cmd) Kind() goes.Kind { return goes.Daemon }
func (*cmd) String() string  { return Name }
func (*cmd) Usage() string   { return Name }

func (cmd *cmd) Main(...string) error {
	var si syscall.Sysinfo_t
	var err error

	cmd.stop = make(chan struct{})
	cmd.last = make(map[string]uint16)
	cmd.lasts = make(map[string]string)

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
	t := time.NewTicker(10 * time.Second)
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
		if strings.Contains(k, "rpm") {
			v := Vdev.FanCount(i)
			if v != cmd.last[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.last[k] = v
			}
		}
		if strings.Contains(k, "speed") {
			v := Vdev.GetFanSpeed()
			if v != cmd.lasts[k] {
				cmd.pub.Print(k, ": ", v)
				cmd.lasts[k] = v
			}
		}
	}
	return nil
}

const (
	fanPoles    = 4
	tempCtrl2   = 0x5f
	high        = 0xff
	med         = 0x80
	low         = 0x50
	maxFanTrays = 4
)

func fanSpeed(countHi uint8, countLo uint8) uint16 {
	d := ((uint16(countHi) << 4) + (uint16(countLo & 0xf))) * (uint16(fanPoles / 4))
	speed := 1.35E06 / float64(d)
	return uint16(speed)
}

func (h *I2cDev) FrontTemp() float64 {
	r := getRegsBank0()
	r.BankSelect.set(h, 0x80)
	r.FrontTemp.get(h)
	r.FractionLSB.get(h)
	DoI2cRpc()
	t := uint8(s[3].D[0])
	u := uint8(s[5].D[0])
	return (float64(t) + ((float64(u >> 7)) * 0.25))
}

func (h *I2cDev) RearTemp() float64 {
	r := getRegsBank0()
	r.BankSelect.set(h, 0x80)
	r.RearTemp.get(h)
	r.FractionLSB.get(h)
	DoI2cRpc()
	t := uint8(s[3].D[0])
	u := uint8(s[5].D[0])
	return (float64(t) + ((float64(u >> 7)) * 0.25))
}

func (h *I2cDev) FanCount(i uint8) uint16 {
	var rpm uint16

	if i > 14 {
		panic("FanCount subscript out of range\n")
	}
	i--

	n := i/2 + 1
	w := "fan_tray." + strconv.Itoa(int(n)) + ".status"
	p, _ := redis.Hget(redis.DefaultHash, w)

	//set fan speed to max and return 0 rpm if fan tray is not present or failed
	if strings.Contains(p, "not installed") {
		rpm = uint16(0)
	} else {
		//remap physical to logical, 0:7 -> 7:0
		i = i + 7 - (2 * i)
		r := getRegsBank0()
		r.BankSelect.set(h, 0x80)
		r.FanCount[i].get(h)
		r.FractionLSB.get(h)
		DoI2cRpc()
		t := uint8(s[3].D[0])
		u := uint8(s[5].D[0])
		rpm = fanSpeed(t, u)
	}
	return rpm
}

func (h *I2cDev) FanInit() {
	r0 := getRegsBank0()
	r0.BankSelect.set(h, 0x80)

	//reset hwm to default values
	r0.Configuration.set(h, 0x9c)
	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)

	//set fan speed output to PWM mode
	r2.FanOutputModeControl.set(h, 0x0)

	//set up clk frequency and dividers

	r2.FanPwmPrescale1.set(h, 0x84)
	r2.FanPwmPrescale2.set(h, 0x84)

	//set default speed to auto
	h.SetFanSpeed("auto")

	//enable temperature monitoring
	r2.BankSelect.set(h, 0x80)
	r0.TempCntl2.set(h, tempCtrl2)
	DoI2cRpc()

	//temperature monitoring requires a delay before readings are valid
	time.Sleep(500 * time.Millisecond)
	r0.Configuration.set(h, 0x1d)
	DoI2cRpc()

	//	time.Sleep(1 * time.Second)
}

func (h *I2cDev) SetFanSpeed(w string) {
	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)

	//if not all fan trays are installed or fan is failed, fan speed is fixed at high

	p, _ := redis.Hget(redis.DefaultHash, w)

	//set fan speed to max and return 0 rpm if fan tray is not present or failed
	if strings.Contains(p, "not installed") || strings.Contains(p, "not installed") {
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, high)
		r2.FanOutValue2.set(h, high)
		DoI2cRpc()
		log.Print("notice: fan speed set to high due to a fan missing or a fan failure")
		return
	}

	switch w {
	case "auto":
		r2 := getRegsBank2()
		r2.BankSelect.set(h, 0x82)
		//set thermal cruise
		r2.FanControlModeSelect1.set(h, 0x00)
		r2.FanControlModeSelect2.set(h, 0x00)

		//set step up and down time to 1s
		r2.FanStepUpTime.set(h, 0x0a)
		r2.FanStepDownTime.set(h, 0x0a)

		//set fan start speed
		r2.FanStartValue1.set(h, 0x30)
		r2.FanStartValue2.set(h, 0x30)

		//set fan stop speed
		r2.FanStopValue1.set(h, 0x30)
		r2.FanStopValue2.set(h, 0x30)

		//set fan stop time to never stop
		r2.FanStopTime1.set(h, 0x0)
		r2.FanStopTime2.set(h, 0x0)

		//set target temps to 50°C
		r2.TargetTemp1.set(h, 0x32)
		r2.TargetTemp2.set(h, 0x32)

		//set critical temp to set 100% fan speed to 65°C
		r2.FanCritTemp1.set(h, 0x41)
		r2.FanCritTemp2.set(h, 0x41)

		//set target temp hysteresis to +/- 5°C
		r2.TempHyster1.set(h, 0x55)
		r2.TempHyster2.set(h, 0x55)

		//enable temp control of fans
		r2.TempToFanMap1.set(h, 0xff)
		r2.TempToFanMap2.set(h, 0xff)
		log.Print("notice: fan speed set to ", w)
	//static speed settings below, set hwm to manual mode, then set static speed
	case "high":
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, high)
		r2.FanOutValue2.set(h, high)
		DoI2cRpc()
		log.Print("notice: fan speed set to ", w)
	case "med":
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, med)
		r2.FanOutValue2.set(h, med)
		DoI2cRpc()
		log.Print("notice: fan speed set to ", w)
	case "low":
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, low)
		r2.FanOutValue2.set(h, low)
		DoI2cRpc()
		log.Print("notice: fan speed set to ", w)
	default:
		DoI2cRpc()
	}

	return
}

func (h *I2cDev) GetFanSpeed() string {
	var speed string

	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
	r2.TempToFanMap1.get(h)
	r2.FanOutValue1.get(h)
	DoI2cRpc()
	t := uint8(s[3].D[0])
	m := uint8(s[5].D[0])

	if t == 0xff {
		speed = "auto"
	} else if m == high {
		speed = "high"
	} else if m == med {
		speed = "med"
	} else if m == low {
		speed = "low"
	} else {
		speed = "invalid " + strconv.Itoa(int(m))
	}
	return speed
}
