// Copyright © 2015-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package w83795d provides access to the H/W Monitor chip
package w83795d

import (
	"errors"
	"fmt"
	"net/rpc"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/external/atsock"
	"github.com/platinasystems/goes/external/log"
	"github.com/platinasystems/goes/external/redis"
	"github.com/platinasystems/goes/external/redis/publisher"
	"github.com/platinasystems/goes/external/redis/rpc/args"
	"github.com/platinasystems/goes/external/redis/rpc/reply"
	"github.com/platinasystems/goes/lang"
	"github.com/platinasystems/gpio"
)

var (
	pollInterval time.Duration = 30

	hostTemp       uint8 = 50
	hostTempTarget uint8 = 70
	qsfpTemp       uint8 = 50
	qsfpTempTarget uint8 = 60
	hwmTarget      uint8

	configuredSpeed string

	hostCtrl           bool
	dutyAtThermalEvent int
	dutyIncrement      int

	fanDutyIncrement int   = 0x20
	hysteresis       uint8 = 5

	thTempTarget uint8 = 55

	setSpeed     bool
	setHwmTarget bool
	hostReset    bool

	Vdev I2cDev

	VpageByKey map[string]uint8

	WrRegDv = make(map[string]string)
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
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

func (*Command) String() string { return "w83795d" }

func (*Command) Usage() string { return "w83795d" }

func (*Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "w83795 hardware monitoring daemon",
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

	c.stop = make(chan struct{})
	c.last = make(map[string]uint16)
	c.lasts = make(map[string]string)

	if c.pub, err = publisher.New(); err != nil {
		return err
	}

	if err = syscall.Sysinfo(&si); err != nil {
		return err
	}

	if c.rpc, err = atsock.NewRpcServer("w83795d"); err != nil {
		return err
	}

	rpc.Register(&c.Info)
	for _, v := range WrRegDv {
		err = redis.Assign(redis.DefaultHash+":"+v+".", "w83795d", "Info")
		if err != nil {
			return err
		}
	}

	Vdev.FanInit()

	t := time.NewTicker(pollInterval * time.Second)
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
	c.Info.mutex.Lock()
	defer c.Info.mutex.Unlock()

	stopped := readStopped()
	if stopped == 1 {
		return nil
	}

	if setSpeed {
		Vdev.SetConfiguredSpeed()
		setSpeed = false
	}

	if setHwmTarget {
		Vdev.SetHwmTarget()
		setHwmTarget = false
	}

	if hostReset {
		doHostReset()
		hostReset = false
	}

	if err := Vdev.PollThermal(); err != nil {
		log.Print("PollThermal: Err: ", err)
	}
	for k, i := range VpageByKey {
		if strings.Contains(k, "rpm") {
			v, err := Vdev.FanCount(i)
			if err != nil {
				continue
			}
			if v != c.last[k] {
				c.pub.Print(k, ": ", v)
				c.last[k] = v
			}
		}
		if strings.Contains(k, "fan_tray.speed") {
			v := configuredSpeed
			if hostCtrl {
				v = "thermal_override"
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
		if strings.Contains(k, "host.temp.units.C") {
			v := Vdev.CheckHostTemp()
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "host.temp.target.units.C") {
			v := Vdev.GetHostTempTarget()
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "qsfp.temp.units.C") {
			v := Vdev.CheckQsfpTemp()
			if v != c.lasts[k] {
				c.pub.Print(k, ": ", v)
				c.lasts[k] = v
			}
		}
		if strings.Contains(k, "qsfp.temp.target.units.C") {
			v := Vdev.GetQsfpTempTarget()
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
	d := ((uint16(countHi) << 4) + uint16(countLo>>4)) * (uint16(fanPoles / 4))
	speed := 1.35e06 / float64(d)
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

func (h *I2cDev) FanCount(i uint8) (uint16, error) {
	var rpm uint16
	var t, u byte

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
		r.FanCount[i].get(h)
		r.FractionLSB.get(h)
		r.FanCount[i].get(h)
		r.FractionLSB.get(h)
		r.FanCount[i].get(h)
		r.FractionLSB.get(h)
		closeMux(h)
		err := DoI2cRpc()
		if err != nil {
			return 0, err
		}
		c := [4]byte{s[3].D[0], s[7].D[0], s[11].D[0], s[15].D[0]}
		l := [4]byte{s[5].D[0], s[9].D[0], s[13].D[0], s[17].D[0]}

		if c[0] == c[1] && l[0] == l[1] {
			t = c[0]
			u = l[0]
		} else if c[1] == c[2] && l[1] == l[2] {
			t = c[1]
			u = l[1]
		} else if c[2] == c[3] && l[2] == l[3] {
			t = c[2]
			u = l[2]
		} else {
			return 0, fmt.Errorf("rpm read mismatch")
		}
		rpm = fanSpeed(t, u)
	}
	return rpm, nil
}

func (h *I2cDev) FanInit() error {
	//default auto mode
	configuredSpeed = "auto"

	//reset hwm to default values
	r0 := getRegsBank0()
	r0.BankSelect.set(h, 0x80)
	r0.Configuration.set(h, 0x9c)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
	//set fan speed output to PWM mode
	r2.FanOutputModeControl.set(h, 0x0)
	//set up clk frequency and dividers
	r2.FanPwmPrescale1.set(h, 0x84)
	r2.FanPwmPrescale2.set(h, 0x84)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	//set default speed to auto
	h.SetConfiguredSpeed()

	//enable temperature monitoring
	r0.BankSelect.set(h, 0x80)
	r0.TempCntl2.set(h, tempCtrl2)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	//temperature monitoring requires a delay before readings are valid
	time.Sleep(500 * time.Millisecond)
	r0.BankSelect.set(h, 0x80)
	r0.Configuration.set(h, 0x1d)
	closeMux(h)
	err = DoI2cRpc()
	if err != nil {
		return err
	}

	return nil
}

func (h *I2cDev) SetConfiguredSpeed() error {
	current, _ := h.GetFanSpeed()
	if current != configuredSpeed {
		h.SetFanSpeed(configuredSpeed)
	}
	return nil
}

func (h *I2cDev) SetFanDuty(d uint8) error {
	for j := 1; j <= maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j))+".status")
		if p != "" && !strings.Contains(p, "ok") {
			return nil
			break
		}
	}

	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
	r2.TempToFanMap1.set(h, 0x0)
	r2.TempToFanMap2.set(h, 0x0)
	r2.FanOutValue1.set(h, d)
	r2.FanOutValue2.set(h, d)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	return nil
}

func (h *I2cDev) SetFanSpeed(w string) error {
	r2 := getRegsBank2()

	//if not all fan trays are ok, only allow high setting
	for j := 1; j <= maxFanTrays; j++ {
		p, _ := redis.Hget(redis.DefaultHash, "fan_tray."+strconv.Itoa(int(j))+".status")
		if p != "" && !strings.Contains(p, "ok") {
			log.Print("warning: fan failure mode, speed fixed at high")
			w = "high"
			break
		}
	}

	switch w {
	case "auto":
		if !hostCtrl {
			r2.BankSelect.set(h, 0x82)
			//set thermal cruise
			r2.FanControlModeSelect1.set(h, 0x00)
			r2.FanControlModeSelect2.set(h, 0x00)
			//set step up and down time to 1s
			r2.FanStepUpTime.set(h, 0x0a)
			r2.FanStepDownTime.set(h, 0x0a)
			closeMux(h)
			err := DoI2cRpc()
			if err != nil {
				return err
			}

			r2.BankSelect.set(h, 0x82)
			//set fan start speed
			r2.FanStartValue1.set(h, 0x30)
			r2.FanStartValue2.set(h, 0x30)
			//set fan stop speed
			r2.FanStopValue1.set(h, 0x30)
			r2.FanStopValue2.set(h, 0x30)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}

			r2.BankSelect.set(h, 0x82)
			//set fan stop time to never stop
			r2.FanStopTime1.set(h, 0x0)
			r2.FanStopTime2.set(h, 0x0)
			//set target temps to 50°C
			r2.TargetTemp1.set(h, 0x32)
			r2.TargetTemp2.set(h, 0x32)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}

			r2.BankSelect.set(h, 0x82)
			//set critical temp to set 100% fan speed to 65°C
			r2.FanCritTemp1.set(h, 0x41)
			r2.FanCritTemp2.set(h, 0x41)
			//set target temp hysteresis to +/- 5°C
			r2.TempHyster1.set(h, 0x55)
			r2.TempHyster2.set(h, 0x55)
			//enable temp control of fans
			r2.TempToFanMap1.set(h, 0xff)
			r2.TempToFanMap2.set(h, 0xff)
			closeMux(h)
			err = DoI2cRpc()
			if err != nil {
				return err
			}
		}

	//static speed settings below, set hwm to manual mode, then set static speed
	case "high":
		h.SetFanDuty(high)

	case "med":
		h.SetFanDuty(med)

	case "low":
		h.SetFanDuty(low)
	default:
	}

	log.Print("notice: fan speed set to ", w)
	return nil
}

func (h *I2cDev) GetFanDuty() (uint8, error) {
	r2 := getRegsBank2()

	r2.BankSelect.set(h, 0x82)
	r2.FanOutValue1.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return 0, err
	}
	m := uint8(s[3].D[0])
	return m, nil

}

func (h *I2cDev) GetFanSpeed() (string, error) {
	r2 := getRegsBank2()

	r2.BankSelect.set(h, 0x82)
	r2.TempToFanMap1.get(h)
	r2.FanOutValue1.get(h)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return "error", err
	}
	t := uint8(s[3].D[0])
	m := uint8(s[5].D[0])

	if t == 0xff {
		return "auto", nil
	}

	switch m {
	case high:
		return "high", nil
	case med:
		return "med", nil
	case low:
		return "low", nil
	}
	return "invalid " + strconv.Itoa(int(m)), nil
}

func (h *I2cDev) PollThermal() error {
	hostHot := hostTemp >= hostTempTarget
	qsfpHot := qsfpTemp >= qsfpTempTarget

	ft, err := h.FrontTemp()
	if err != nil {
		return err
	}
	f, err := strconv.ParseFloat(ft, 64)
	if err != nil {
		return err
	}
	rt, err := h.RearTemp()
	if err != nil {
		return err
	}
	r, err := strconv.ParseFloat(rt, 64)
	if err != nil {
		return err
	}

	thTemp := uint8(f)
	if r > f {
		thTemp = uint8(r)
	}
	thHot := thTemp > thTemp

	if hostHot || qsfpHot || thHot {
		if !hostCtrl {
			d, err := h.GetFanDuty()
			if err != nil {
				return err
			}
			dutyAtThermalEvent = int(d)
			dutyIncrement = 0
			hostCtrl = true
		}
		dutyIncrement += fanDutyIncrement
		if dutyAtThermalEvent+dutyIncrement > 0xff {
			dutyIncrement = 0xff - dutyAtThermalEvent
		}

		h.SetFanDuty(uint8(dutyAtThermalEvent + dutyIncrement))
		log.Print("thermal event: fan duty set to ",
			dutyAtThermalEvent+dutyIncrement,
			" (duty increment) ", dutyIncrement)
	}

	if hostCtrl && (hostTemp < (hostTempTarget - hysteresis)) &&
		(qsfpTemp < (qsfpTempTarget - hysteresis)) &&
		(thTemp < thTempTarget-hysteresis) {
		dutyIncrement -= fanDutyIncrement
		if dutyIncrement > 0 {
			h.SetFanDuty(uint8(dutyAtThermalEvent + dutyIncrement))
			log.Print("thermal resolving: fan duty set to ",
				dutyAtThermalEvent+dutyIncrement,
				" (duty increment) ", dutyIncrement)
		} else {
			hostCtrl = false
			h.SetConfiguredSpeed()
		}
	}
	return nil
}

func (h *I2cDev) SetHwmTarget() error {
	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
	r2.TargetTemp1.set(h, hwmTarget)
	r2.TargetTemp2.set(h, hwmTarget)
	closeMux(h)
	err := DoI2cRpc()
	if err != nil {
		return err
	}

	return nil
}

func (h *I2cDev) GetHwmTarget() (uint16, error) {
	r2 := getRegsBank2()
	r2.BankSelect.set(h, 0x82)
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

func (h *I2cDev) CheckHostTemp() string {
	return fmt.Sprintf("%d", hostTemp)
}

func (h *I2cDev) CheckQsfpTemp() string {
	return fmt.Sprintf("%d", qsfpTemp)
}

func (h *I2cDev) GetHostTempTarget() string {
	return fmt.Sprintf("%d", hostTempTarget)
}

func (h *I2cDev) GetQsfpTempTarget() string {
	return fmt.Sprintf("%d", qsfpTempTarget)
}

func doHostReset() error {
	// FIXME cmd.Init("gpio")
	log.Print("notice: issue hard reset to host")
	pin, found := gpio.FindPin("BMC_TO_HOST_RST_L")
	if found {
		pin.SetValue(false)
	}
	time.Sleep(50 * time.Millisecond)
	if found {
		pin.SetValue(true)
	}
	return nil
}

func parseTemp(t string, low uint8, high uint8) (uint8, error) {
	f, err := strconv.ParseFloat(t, 64)
	if err != nil {
		return high, err
	}
	if f < float64(low) || f > float64(high) {
		return high, fmt.Errorf("Temperature must between %d and %d",
			low, high)
	}
	return uint8(f), nil
}

func (i *Info) Hset(args args.Hset, reply *reply.Hset) error {
	i.mutex.Lock()
	defer i.mutex.Unlock()
	v := string(args.Value)
	v = strings.TrimRight(v, "\n") // Be conservative in what we accept

	switch args.Field {
	case "fan_tray.speed":
		if v == "auto" || v == "high" || v == "med" || v == "low" || v == "max" {
			configuredSpeed = v
			setSpeed = true
		} else {
			return errors.New("Invalid speed")
		}
	case "host.temp.units.C":
		f, err := parseTemp(v, 0, 255)
		if err != nil {
			return err
		}
		hostTemp = f

	case "host.temp.target.units.C":
		f, err := parseTemp(v, 25, 85)
		if err != nil {
			return err
		}
		hostTempTarget = f

	case "qsfp.temp.units.C":
		f, err := parseTemp(v, 0, 255)
		if err != nil {
			return err
		}
		qsfpTemp = f

	case "qsfp.temp.target.units.C":
		f, err := parseTemp(v, 25, 85)
		if err != nil {
			return err
		}
		qsfpTempTarget = f

	case "fan_tray.speed.return":
		if v == "" {
			setSpeed = true
		}

	case "hwmon.target.units.C":
		t, err := parseTemp(v, 25, 85)
		if err != nil {
			return err
		}

		if t > 60 {
			return errors.New("Only integers up to 60 accepted")
		}
		hwmTarget = t
		setHwmTarget = true

	case "host.reset":
		if v == "true" {
			hostReset = true
		}
	default:
		return fmt.Errorf("Don't know how to set %s", args.Field)
	}

	err := i.set(args.Field, v, false)
	if err == nil {
		*reply = 1
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
