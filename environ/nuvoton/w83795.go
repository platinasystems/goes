// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package w83795 provides access to the H/W Monitor chip
package w83795

import (
	"strconv"
	"strings"
	"time"
	"unsafe"

	"github.com/platinasystems/go/i2c"
	"github.com/platinasystems/go/log"
	"github.com/platinasystems/go/redis"
)

var (
	dummy            byte
	regsPointer      = unsafe.Pointer(&dummy)
	regsAddr         = uintptr(unsafe.Pointer(&dummy))
	lastSpeed        string
	fanFail          = false
	allInstalled     = true
	lastAllInstalled = true
)

type HwMonitor struct {
	Bus      int
	Addr     int
	MuxBus   int
	MuxAddr  int
	MuxValue int
}

const (
	fanPoles    = 4
	tempCtrl2   = 0x5f
	high        = 0xff
	med         = 0x80
	low         = 0x50
	maxFanTrays = 4
)

func getHwmRegsBank0() *hwmRegsBank0 { return (*hwmRegsBank0)(regsPointer) }
func getHwmRegsBank2() *hwmRegsBank2 { return (*hwmRegsBank2)(regsPointer) }

func (r *reg8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (h *HwMonitor) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(h.Bus)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(h.Addr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (h *HwMonitor) i2cDoMux(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
	var bus i2c.Bus

	err = bus.Open(h.MuxBus)
	if err != nil {
		return
	}
	defer bus.Close()

	err = bus.ForceSlaveAddress(h.MuxAddr)
	if err != nil {
		return
	}

	err = bus.Do(rw, regOffset, size, data)
	return
}

func (r *reg8) get(h *HwMonitor) byte {
	var data i2c.SMBusData
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in w83795.get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: get8 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: get8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	return data[0]
}

func (r *reg16) get(h *HwMonitor) (v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in w83795.get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: get16 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: get16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16r) get(h *HwMonitor) (v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in w83795.get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: get16 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: get16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	return uint16(data[1])<<8 | uint16(data[0])
}

func (r *reg8) set(h *HwMonitor, v uint8) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in w83795.get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: set8 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	data[0] = v
	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Write, r.offset(), i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: set8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16) set(h *HwMonitor, v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in w83795.get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: set16 MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: set16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16r) set(h *HwMonitor, v uint16) {
	var data i2c.SMBusData

	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in w83795.get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()

	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: set16r MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}

	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	for i := 0; i < 5000; i++ {
		err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
		if err == nil {
			if i > 0 {
				log.Print("w83795: set16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *regi16) get(h *HwMonitor) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *HwMonitor, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func fanSpeed(countHi uint8, countLo uint8) uint16 {
	d := ((uint16(countHi) << 4) + (uint16(countLo & 0xf))) * (uint16(fanPoles / 4))
	speed := 1.35E06 / float64(d)
	return uint16(speed)
}

func (h *HwMonitor) FrontTemp() float64 {
	r := getHwmRegsBank0()
	r.BankSelect.set(h, 0x80)
	t := r.FrontTemp.get(h)
	u := r.FractionLSB.get(h)
	return (float64(t) + ((float64(u >> 7)) * 0.25))
}

func (h *HwMonitor) RearTemp() float64 {
	r := getHwmRegsBank0()
	r.BankSelect.set(h, 0x80)
	t := r.RearTemp.get(h)
	u := r.FractionLSB.get(h)
	return (float64(t) + ((float64(u >> 7)) * 0.25))
}

func (h *HwMonitor) FanCount(i uint8) uint16 {
	var rpm uint16

	if i > 14 {
		panic("FanCount subscript out of range\n")
	}
	i--

	n := i/2 + 1
	s := "fan_tray." + strconv.Itoa(int(n)) + ".status"
	p, _ := redis.Hget("platina", s)

	//set fan speed to max and return 0 rpm if fan tray is not present
	if strings.Contains(p, "not installed") {
		allInstalled = false
		if lastAllInstalled != allInstalled {
			redis.Hset("platina", "fan_tray.speed", "high")
			lastAllInstalled = false
		}
		rpm = uint16(0)
	} else {

		//if all fan trays are present, return to previous fan speed
		fanFail = false
		for j := 1; j <= maxFanTrays; j++ {
			s = "fan_tray." + strconv.Itoa(int(j)) + ".status"
			p, _ = redis.Hget("platina", s)
			if strings.Contains(p, "not installed") {
				fanFail = true
				break
			}
		}

		if !fanFail {
			allInstalled = true
			if lastAllInstalled != allInstalled {
				lastAllInstalled = true
				redis.Hset("platina", "fan_tray.speed", lastSpeed)
			}
		}

		//remap physical to logical, 0:7 -> 7:0
		i = i + 7 - (2 * i)
		r := getHwmRegsBank0()
		r.BankSelect.set(h, 0x80)
		t := r.FanCount[i].get(h)
		u := r.FractionLSB.get(h)
		rpm = fanSpeed(t, u)
	}

	return rpm
}

func (h *HwMonitor) FanInit() {
	r0 := getHwmRegsBank0()
	r0.BankSelect.set(h, 0x80)

	//reset hwm to default values
	r0.Configuration.set(h, 0x9c)

	r2 := getHwmRegsBank2()
	r2.BankSelect.set(h, 0x82)

	//set fan speed output to PWM mode
	r2.FanOutputModeControl.set(h, 0x0)

	//set up clk frequency and dividers
	r2.FanPwmPrescale1.set(h, 0x84)
	r2.FanPwmPrescale2.set(h, 0x84)

	//set default speed to medium
	r2.FanOutValue1.set(h, med)
	r2.FanOutValue2.set(h, med)
	lastSpeed = "med"

	//enable temperature monitoring
	r2.BankSelect.set(h, 0x80)
	r0.TempCntl2.set(h, tempCtrl2)

	//temperature monitoring requires a delay before readings are valid
	time.Sleep(500 * time.Millisecond)
	r0.Configuration.set(h, 0x1d)

	//	time.Sleep(1 * time.Second)
}

func (h *HwMonitor) SetFanSpeed(s string) {
	r2 := getHwmRegsBank2()
	r2.BankSelect.set(h, 0x82)

	//if not all fan trays are installed, fan speed is fixed at high
	if !allInstalled {
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, high)
		r2.FanOutValue2.set(h, high)
		return
	} else {
		//if all fan trays present, save the set speed
		lastSpeed = s
	}

	switch s {
	case "auto":
		r2 := getHwmRegsBank2()
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
		r2.FanStopValue1.set(h, 0x20)
		r2.FanStopValue2.set(h, 0x20)

		//set fan stop time to max 25.5s
		r2.FanStopTime1.set(h, 0xff)
		r2.FanStopTime2.set(h, 0xff)

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

	//static speed settings below, set hwm to manual mode, then set static speed
	case "high":
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, high)
		r2.FanOutValue2.set(h, high)
	case "med":
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, med)
		r2.FanOutValue2.set(h, med)
	case "low":
		r2.TempToFanMap1.set(h, 0x0)
		r2.TempToFanMap2.set(h, 0x0)
		r2.FanOutValue1.set(h, low)
		r2.FanOutValue2.set(h, low)
	default:
	}

	return
}

func (h *HwMonitor) GetFanSpeed() string {
	var speed string

	r2 := getHwmRegsBank2()
	r2.BankSelect.set(h, 0x82)
	t := r2.TempToFanMap1.get(h)
	m := r2.FanOutValue1.get(h)

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
