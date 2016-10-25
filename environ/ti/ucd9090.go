// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package ucd9090

import (
	"fmt"
	"math"
	"strconv"
	"unsafe"

	"github.com/platinasystems/go/i2c"
)

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

type PowerMon struct {
	Bus      int
	Addr     int
	MuxAddr  int
	MuxValue int
}

func getPwmRegs() *pwmRegs { return (*pwmRegs)(regsPointer) }
func getGenRegs() *genRegs { return (*genRegs)(regsPointer) }

func (r *reg8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (h *PowerMon) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
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

func (h *PowerMon) i2cDoMux(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
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

func (r *reg8) get(h *PowerMon) byte {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
	if err != nil {
		panic(err)
	}
	return data[0]
}

func (r *reg8) setErr(h *PowerMon, v uint8) error {
	var data i2c.SMBusData
	data[0] = v
	return h.i2cDo(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg8) set(h *PowerMon, v uint8) {
	err := r.setErr(h, v)
	if err != nil {
		panic(err)
	}
}

func (r *reg8) setMux(h *PowerMon) error {
	var data i2c.SMBusData
	data[0] = byte(h.MuxValue)
	return h.i2cDoMux(i2c.Write, r.offset(), i2c.ByteData, &data)
}

func (r *reg16) get(h *PowerMon) (v uint16) {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16r) get(h *PowerMon) (v uint16) {
	var data i2c.SMBusData
	err := h.i2cDo(i2c.Read, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
	return uint16(data[1])<<8 | uint16(data[0])
}

func (r *reg16) set(h *PowerMon, v uint16) {
	var data i2c.SMBusData
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
}

func (r *reg16r) set(h *PowerMon, v uint16) {
	var data i2c.SMBusData
	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	err := h.i2cDo(i2c.Write, r.offset(), i2c.WordData, &data)
	if err != nil {
		panic(err)
	}
}

func (r *regi16) get(h *PowerMon) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *PowerMon, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func (h *PowerMon) Vout(i uint8) float64 {
	i2c.Lock.Lock()
	defer i2c.Lock.Unlock()

	if i > 10 {
		panic("Voltage rail subscript out of range\n")
	}
	i--
	q := getGenRegs()
	q.Reg.setMux(h)
	r := getPwmRegs()
	r.Page.set(h, i)

	n := r.VoutMode.get(h) & 0xf
	n--
	n = (n ^ 0xf) & 0xf
	v := r.ReadVout.get(h)

	nn := float64(n) * (-1)
	vv := float64(v) * (math.Exp2(nn))
	vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
	return float64(vv)
}
