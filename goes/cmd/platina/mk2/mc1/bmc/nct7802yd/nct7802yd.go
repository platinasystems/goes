// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package nct7802yd provides access to the H/W Monitor chip on the MC

package nct7802yd

import (
	"errors"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/atsock"
	"github.com/platinasystems/go/goes/cmd"
	"github.com/platinasystems/go/goes/cmd/platina/mk2/mc1/bmc/w83795d"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/eeprom"
	"github.com/platinasystems/gpio"
	"github.com/platinasystems/log"
	"github.com/platinasystems/redis"
	"github.com/platinasystems/redis/publisher"
	"github.com/platinasystems/redis/rpc/args"
	"github.com/platinasystems/redis/rpc/reply"
)

var (
	SlotId         int
	MaxFanTrays    int
	MaxFansPerTray int

	first          int
	hostTemp       float64
	sHostTemp      float64
	hostTempTarget float64
	hostHyst       float64
	qsfpTemp       float64
	sQsfpTemp      float64
	qsfpTempTarget float64
	qsfpHyst       float64
	thTemp         float64
	sThTemp        float64

	lastSpeed string

	hostCtrl bool
	thCtrl   bool

	Vdev I2cDev

	VpageByKey map[string]uint8

	WrRegDv  = make(map[string]string)
	WrRegFn  = make(map[string]string)
	WrRegVal = make(map[string]string)
	WrRegRng = make(map[string][]string)

	command *Command
)

type Command struct {
	Info
	Init func()
	init sync.Once
	Gpio func()
	gpio sync.Once
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
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

func (*Command) String() string { return "nct7802yd" }

func (*Command) Usage() string { return "nct7802yd" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "nct7802y hardware monitoring daemon",
	}
}

func (*Command) Kind() cmd.Kind { return cmd.Daemon }

func (c *Command) Main(...string) error {
	var si syscall.Sysinfo_t

	command = c
	if c.Init != nil {
		c.init.Do(c.Init)
	}

	err := redis.IsReady()
	if err != nil {
		log.Print("redis err: ", err)
		return err
	}

	first = 1
	hostTemp = 50
	sHostTemp = 150
	hostTempTarget = 70
	hostHyst = 0

	qsfpTemp = 50
	sQsfpTemp = 150
	qsfpTempTarget = 60
	qsfpHyst = 0

	hostCtrl = false
	sThTemp = 150
	thCtrl = false

	c.stop = make(chan struct{})
	c.last = make(map[string]uint16)
	c.lasts = make(map[string]string)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if c.rpc, err = atsock.NewRpcServer("nct7802yd"); err != nil {
		return err
	}

	rpc.Register(&c.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", "nct7802yd",
			"Info")
		if err != nil {
			return err
		}
	}

	t := time.NewTicker(10 * time.Second)
	for {
		select {
		case <-c.stop:
			return nil
		case <-t.C:
			if err = c.update(); err != nil {
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
			log.Print(err)
		}

		switch d.Fields.ChassisType {
		case CH1_4S:
			MaxFanTrays = 3
			MaxFansPerTray = 2
		case CH1_8S:
			MaxFanTrays = 3
			MaxFansPerTray = 3
		case CH1_16S:
			MaxFanTrays = 6
			MaxFansPerTray = 3
		}

		Vdev.FanInit()
		log.Print("fan init")
		first = 0
	}

	for k, _ := range VpageByKey {
		if strings.Contains(k, "fan_tray.control") {
			v, err := Vdev.getFanControl()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "fan_tray.speed") {
			v, err := Vdev.GetFanSpeed()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "fan_tray.duty") {
			v, err := Vdev.GetFanDuty()
			sv := fmt.Sprintf("0x%x", v)
			if err != nil {
				return err
			}
			if sv != c.lasts[k] {
				c.pub.Print(k, ": ", sv)
				c.lasts[k] = sv
			}
		}
		if strings.Contains(k, "hwmon.front.temp.units.C") {
			v, err := Vdev.FrontTemp()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "hwmon.rear.temp.units.C") {
			v, err := Vdev.RearTemp()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "hwmon.target.units.C") {
			v, err := Vdev.GetHwmTarget()
			if err != nil {
				return err
			}
			if v != c.last[k] {
				c.pub.Print(k, ": ", v)
				c.last[k] = v
			}
		}
		if strings.Contains(k, "host.temp.units.C") {
			v, err := Vdev.CheckHostTemp()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "host.temp.target.units.C") {
			v, err := Vdev.GetHostTempTarget()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "qsfp.temp.units.C") {
			v, err := Vdev.CheckQsfpTemp()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "qsfp.temp.target.units.C") {
			v, err := Vdev.GetQsfpTempTarget()
			if err != nil {
				return err
			}
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
	}
	return nil
}

const (
	fanPoles  = 4
	tempCtrl2 = 0x5f
	high      = 0xff
	med       = 0x80
	low       = 0x50
	//maxFanTrays = 4
)

func fanSpeed(countHi uint8, countLo uint8) uint16 {
	d := ((uint16(countHi) << 4) + uint16(countLo>>4)) * (uint16(fanPoles / 4))
	speed := 1.35E06 / float64(d)
	return uint16(speed)
}

func (h *I2cDev) FrontTemp() (string, error) {
	r := getRegsBank0()
	r.BankSelect.set(h, 0x80)
	r.FrontTemp.get(h)
	r.FractionLSB.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint8(s[3].D[0])
	u := uint8(s[5].D[0])
	v := float64(t) + ((float64(u >> 7)) * 0.25)
	strconv.FormatFloat(v, 'f', 3, 64)
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

func (h *I2cDev) RearTemp() (string, error) {
	r := getRegsBank0()
	r.BankSelect.set(h, 0x80)
	r.RearTemp.get(h)
	r.FractionLSB.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "", err
	}
	t := uint8(s[3].D[0])
	u := uint8(s[5].D[0])
	v := float64(t) + ((float64(u >> 7)) * 0.25)
	return strconv.FormatFloat(v, 'f', 3, 64), nil
}

/*
func (h *I2cDev) FanCount(i uint8) (uint16, error) {
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
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return 0, err
		}
		t := uint8(s[3].D[0])
		u := uint8(s[5].D[0])
		rpm = fanSpeed(t, u)
	}
	return rpm, nil
} */

func (h *I2cDev) FanInit() error {
	//default auto mode
	lastSpeed = "auto"

	//reset hwm to default values
	r0 := getRegsBank0()
	r0.BankSelect.set(h, 0x80)
	r0.SoftReset.set(h, 0x80)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	r2 := getRegsBank0()
	r2.BankSelect.set(h, 0x80)
	//set fan speed output to PWM mode
	r2.FanOutputModeControl.set(h, 0x0)
	//set up clk frequency and dividers
	r2.FanPwmPrescale1.set(h, 0x84)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	//set default speed to auto
	h.SetFanSpeed("auto", true)

	// MC controls all fan trays
	h.SetFanControl("enabled")

	//enable temperature monitoring
	r0.BankSelect.set(h, 0x80)
	r0.FanEnable.set(h, 0x00)
	r0.VmonEnable.set(h, 0x00)
	r0.ModeSel.set(h, 0x0a)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	//temperature monitoring requires a delay before readings are valid
	time.Sleep(500 * time.Millisecond)
	r0.BankSelect.set(h, 0x80)
	r0.StartCtrl.set(h, 0x81)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	return nil
}

func (h *I2cDev) SetLastSpeed() error {
	current, _ := h.GetFanSpeed()
	if current != lastSpeed {
		h.SetFanSpeed(lastSpeed, true)
	}
	return nil
}
func (h *I2cDev) SetFanDuty(d uint8) error {
	for j := 1; j <= MaxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j))+".status")
		if p != "" && !strings.Contains(p, "ok") {
			return nil
			break
		}
	}

	r2 := getRegsBank0()
	r2.BankSelect.set(h, 0x80)
	r2.TempToFanMap1.set(h, 0x0) // manual mode
	r2.TempToFanMap2.set(h, 0x0)
	r2.FanOutValue1.set(h, d)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}
	return nil
}

func (h *I2cDev) SetFanControl(w string) error {
	var control string
	var vdevIo [6]w83795d.I2cDev

	vdevIo[0] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x10}
	vdevIo[1] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x20}
	vdevIo[2] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x40}
	vdevIo[3] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x80}
	vdevIo[4] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x10, 1, 0x73, 0x04}
	vdevIo[5] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x10, 1, 0x73, 0x08}

	switch w {
	case "enabled":
		control = "remote.mc" + strconv.Itoa(SlotId)
	case "disabled":
		control = "local"
	}
	for j := 0; j < MaxFanTrays; j++ {
		err := vdevIo[j].SetFanControl(control)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *I2cDev) SetFanSpeed(w string, l bool) error {
	r2 := getRegsBank0()

	//if not all fan trays are ok, only allow high setting
	for j := 1; j <= MaxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j))+".status")
		if p != "" && !strings.Contains(p, "ok") {
			log.Print("warning: fan failure mode, speed fixed at high")
			w = "max"
			break
		}
	}

	if w != "max" {
		lastSpeed = w
	} else {
		w = "high"
	}

	switch w {
	case "auto":
		if !hostCtrl {
			r2.BankSelect.set(h, 0x80)
			//set step up and down time to 1s
			r2.FanStepUpTime.set(h, 0x0a)
			r2.FanStepDownTime.set(h, 0x0a)
			closeMux(h)
			err := DoI2cRpc()
			if err != nil {
				return err
			}

			r2.BankSelect.set(h, 0x80)
			//set fan start speed
			r2.FanStartValue1.set(h, 0x30)
			//set fan stop time to never stop
			r2.FanStopTime1.set(h, 0x0)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}

			r2.BankSelect.set(h, 0x80)
			//set target temps to 50°C
			r2.TargetTemp1.set(h, 0x32)
			r2.TargetTemp2.set(h, 0x32)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}

			r2.BankSelect.set(h, 0x80)
			//set critical temp to set 100% fan speed to 65°C
			r2.FanCritTemp1.set(h, 0x41)
			r2.FanCritTemp2.set(h, 0x41)
			//set target temp hysteresis to +/- 5°C
			r2.TempHyster1.set(h, 0x55)
			r2.TempHyster2.set(h, 0x55)
			//enable temp control of fans
			r2.TempToFanMap1.set(h, 0x11) //Smart-Fan mode
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}
		}
		if l {
			log.Print("notice: fan speed set to ", w)
		}
	//static speed settings below, set hwm to manual mode, then set static speed
	case "high":
		r2.BankSelect.set(h, 0x80)
		r2.TempToFanMap1.set(h, 0x0) //manual
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, high)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to high")
	case "med":
		r2.BankSelect.set(h, 0x80)
		r2.TempToFanMap1.set(h, 0x0) //manual
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, med)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to ", w)
	case "low":
		r2.BankSelect.set(h, 0x80)
		r2.TempToFanMap1.set(h, 0x0) //manual
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, low)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return err
		}
		log.Print("notice: fan speed set to ", w)
	default:
	}

	return nil
}

func (h *I2cDev) GetFanDuty() (uint8, error) {
	r2 := getRegsBank0()
	r2.BankSelect.set(h, 0x80)
	r2.FanOutValue1.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	m := uint8(s[3].D[0])
	return m, nil

}

func (h *I2cDev) getFanControl() (string, error) {
	return "disabled", nil

	var vdevIo [6]w83795d.I2cDev

	vdevIo[0] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x10}
	vdevIo[1] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x20}
	vdevIo[2] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x40}
	vdevIo[3] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x08, 1, 0x73, 0x80}
	vdevIo[4] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x10, 1, 0x73, 0x04}
	vdevIo[5] = w83795d.I2cDev{1, 0x41, 1, 0x70, 0x10, 1, 0x73, 0x08}

	m := 0
	control := "remote.mc" + strconv.Itoa(SlotId)
	for j := 0; j < MaxFanTrays; j++ {
		v, err := vdevIo[j].GetFanControl()
		if err != nil {
			return "", err
		} else if v == control {
			m++
		}
	}
	if m == MaxFanTrays {
		return "enabled", nil
	} else if m == 0 {
		return "disabled", nil
	}
	err := errors.New("invalid config:" + strconv.Itoa(m))
	return "invalid ", err
}

func (h *I2cDev) GetFanSpeed() (string, error) {
	var speed string
	var duty uint8
	r2 := getRegsBank0()

	if !hostCtrl {
		r2.BankSelect.set(h, 0x80)
		r2.TempToFanMap1.get(h)
		r2.FanOutValue1.get(h)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return "error", err
		}
		t := uint8(s[3].D[0])
		m := uint8(s[5].D[0])

		if t == 0x11 {
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
	}
	if hostCtrl || (!hostCtrl && speed == "auto") {
		hostHot := false
		hostHotter := false
		hostHot = hostTemp > hostTempTarget
		hostHotter = hostTemp > sHostTemp
		qsfpHot := false
		qsfpHotter := false
		qsfpHot = qsfpTemp > qsfpTempTarget
		qsfpHotter = qsfpTemp > sQsfpTemp
		if (!hostCtrl && (hostHot || qsfpHot)) || (hostCtrl && (hostHotter || qsfpHotter)) {
			var err error
			duty, err = h.GetFanDuty()
			if err != nil {
				return "auto", err
			}
			if hostHot || hostHotter {
				hostHyst = 5
				sHostTemp = hostTemp

			}
			if qsfpHot || qsfpHotter {
				qsfpHyst = 5
				sQsfpTemp = qsfpTemp
			}
			if duty < 0xff {
				if duty <= 0xdf {
					h.SetFanDuty(duty + 0x20)
				} else {
					h.SetFanDuty(0xff)
				}
			} else {
			}

			if !hostCtrl {
				hostCtrl = true
			}
		} else if hostCtrl && (hostTemp <= (hostTempTarget - hostHyst)) && (qsfpTemp <= (qsfpTempTarget - qsfpHyst)) {
			hostCtrl = false
			thCtrl = false
			sHostTemp = 150
			hostHyst = 0
			sQsfpTemp = 150
			qsfpHyst = 0
			sThTemp = 150
			//set fan speed to thermal cruise (auto)
			h.SetFanSpeed("auto", false)
		}
		if hostCtrl {
			ft, err := h.FrontTemp()
			if err != nil {
				return "auto", err
			}
			f, err := strconv.ParseFloat(ft, 64)
			if err != nil {
				return "auto", err
			}
			rt, err := h.RearTemp()
			if err != nil {
				return "auto", err
			}
			r, err := strconv.ParseFloat(rt, 64)
			if err != nil {
				return "auto", err
			}
			if r > f {
				thTemp = r
			} else {
				thTemp = f
			}
			if (!thCtrl && (thTemp > 55)) || (thCtrl && (thTemp > sThTemp)) {
				//increase fan speed
				sThTemp = thTemp
				duty, err = h.GetFanDuty()
				if err != nil {
					return "auto", err
				}
				if duty < 0xff {
					if duty <= 0xdf {
						h.SetFanDuty(duty + 0x20)
					} else {
						h.SetFanDuty(0xff)
					}
				} else {
				}
				if !thCtrl {
					thCtrl = true
				}
			}
		}
		return "auto", nil
	}

	return speed, nil
}

func (h *I2cDev) SetHwmTarget(t string) error {
	v, err := strconv.ParseUint(t, 10, 8)

	if err != nil {
		log.Print("Unable to set target temperature, only integers up to 60 accepted")
		return err
	} else {
		if v > 60 {
			log.Print("Unable to set target temperature, only integers up to 60 accepted")
			return nil
		} else {
			r2 := getRegsBank0()
			r2.BankSelect.set(h, 0x80)
			r2.TargetTemp1.set(h, uint8(v))
			r2.TargetTemp2.set(h, uint8(v))
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (h *I2cDev) GetHwmTarget() (uint16, error) {
	r2 := getRegsBank0()
	r2.BankSelect.set(h, 0x80)
	r2.TargetTemp1.get(h)
	r2.TargetTemp2.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}

	m := uint16(0)
	if s[3].D[0] == s[5].D[0] {
		m = uint16(s[3].D[0])
	}
	return m, nil
}

func (h *I2cDev) CheckHostTemp() (string, error) {
	v := hostTemp
	return strconv.FormatFloat(v, 'f', 2, 64), nil
}

func (h *I2cDev) CheckQsfpTemp() (string, error) {
	v := qsfpTemp
	return strconv.FormatFloat(v, 'f', 2, 64), nil
}

func (h *I2cDev) GetHostTempTarget() (string, error) {
	v := hostTempTarget
	return strconv.FormatFloat(v, 'f', 2, 64), nil
}

func (h *I2cDev) GetQsfpTempTarget() (string, error) {
	v := qsfpTempTarget
	return strconv.FormatFloat(v, 'f', 2, 64), nil
}

func hostReset() error {
	command.gpio.Do(command.Gpio)
	log.Print("issue hard reset to host")
	pin, found := gpio.Pins["BMC_TO_HOST_RST_L"]
	if found {
		pin.SetValue(false)
	}
	time.Sleep(50 * time.Millisecond)
	if found {
		pin.SetValue(true)
	}
	return nil
}

func writeRegs() error {
	for k, v := range WrRegVal {
		switch WrRegFn[k] {
		case "control":
			if v == "enabled" || v == "disabled" {
				log.Print("control", v)
			}
		case "speed":
			if v == "auto" || v == "high" || v == "med" || v == "low" || v == "max" {
				Vdev.SetFanSpeed(v, true)
			}
		case "host.temp.units.C":
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				hostTemp = f
			}
		case "host.temp.target.units.C":
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				hostTempTarget = f
			}
		case "qsfp.temp.units.C":
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				qsfpTemp = f
			}
		case "qsfp.temp.target.units.C":
			log.Print("write qsfp target: ", v)
			f, err := strconv.ParseFloat(v, 64)
			if err == nil {
				qsfpTempTarget = f
			}
		case "speed.return":
			if v == "" {
				Vdev.SetLastSpeed()
			}
		case "target.units.C":
			Vdev.SetHwmTarget(v)
		case "reset":
			if v == "true" {
				hostReset()
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
