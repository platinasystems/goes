// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package w83795 provides access to the H/W Monitor chip
package w83795

import (
	"unsafe"

	"github.com/platinasystems/go/i2c"
)

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

type HwMonitor struct {
	Bus      int
	Addr     int
	MuxAddr  int
	MuxValue int
}

const (
	fanPoles  = 4
	tempCtrl2 = 0x5f
)

func getHwmRegs() *hwmRegs { return (*hwmRegs)(regsPointer) }
func getGenRegs() *genRegs { return (*genRegs)(regsPointer) }

func (r *reg8) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

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

	err = bus.Open(h.Bus)
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
	err := h.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
	if err != nil {
		panic(err)
	}
	return data[0]
}

func (r *reg8) setErr(h *HwMonitor, v uint8) error {
	var data i2c.SMBusData
	data[0] = v
	return h.i2cDo(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg8) set(h *HwMonitor, v uint8) {
	err := r.setErr(h, v)
	if err != nil {
		panic(err)
	}
}

func (r *reg8) setMux(h *HwMonitor) error {
	var data i2c.SMBusData
	data[0] = byte(h.MuxValue)
	return h.i2cDoMux(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg16) get(h *HwMonitor) (v uint16) {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16) set(h *HwMonitor, v uint16) {
	var data i2c.SMBusData
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
}

func (r *regi16) get(h *HwMonitor) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *HwMonitor, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func fanSpeed(countHi uint8, countLo uint8) uint16 {
	if countHi == 0xff {
		return 0 //fan tray missing
	} else {
		d := ((uint16(countHi) << 4) + (uint16(countLo & 0xf))) * (uint16(fanPoles / 4))
		speed := 1.35E06 / float64(d)
		return uint16(speed)
	}
}

func (h *HwMonitor) FrontTemp() float64 {
	q := getGenRegs()
	q.Reg.setMux(h)
	r := getHwmRegs()
	r.TempCntl2.set(h, tempCtrl2)
	t := r.FrontTemp.get(h)
	u := r.FractionLSB.get(h)
	return (float64(t) + ((float64(u >> 7)) * 0.25))
}

func (h *HwMonitor) RearTemp() float64 {
	q := getGenRegs()
	q.Reg.setMux(h)
	r := getHwmRegs()
	r.TempCntl2.set(h, tempCtrl2)
	t := r.RearTemp.get(h)
	u := r.FractionLSB.get(h)
	return (float64(t) + ((float64(u >> 7)) * 0.25))
}

func (h *HwMonitor) FanCount(i uint8) uint16 {
	if i > 14 {
		panic("FanCount subscript out of range\n")
	}
	i--
	q := getGenRegs()
	q.Reg.setMux(h)
	r := getHwmRegs()
	t := r.FanCount[i].get(h)
	u := r.FractionLSB.get(h)
	return fanSpeed(t, u)
}
