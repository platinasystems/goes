// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package fsp provides access to the power supply unit
package fsp

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"unsafe"

	"github.com/platinasystems/go/goes/internal/gpio"
	"github.com/platinasystems/go/goes/internal/log"
	"github.com/platinasystems/go/i2c"
)

var (
	dummy       byte
	regsPointer = unsafe.Pointer(&dummy)
	regsAddr    = uintptr(unsafe.Pointer(&dummy))
)

type Psu struct {
	Slot       int
	Installed  int
	Id         string
	Bus        int
	Addr       int
	MuxBus     int
	MuxAddr    int
	MuxValue   int
	GpioPwrok  string
	GpioPrsntL string
	GpioPwronL string
	GpioIntL   string
}

const ()

func getPsuRegs() *psuRegs { return (*psuRegs)(regsPointer) }

// offset function has divide by two for 16-bit offset struct
func (r *reg8) offset() uint8   { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg8b) offset() uint8  { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg16) offset() uint8  { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg16r) offset() uint8 { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }

func (h *Psu) i2cDo(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
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

func (h *Psu) i2cDoMux(rw i2c.RW, regOffset uint8, size i2c.SMBusSize, data *i2c.SMBusData) (err error) {
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

func (r *reg8) get(h *Psu) byte {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return 0
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: get8: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: get8 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("fsp550: get8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	return data[0]
}

func (r *reg8b) get(h *Psu) string {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return "not installed"
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: get8b: ", rc, " addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: get8b MuxWr #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	data[0] = 2
	for i := 0; i < 5000; i++ { // read count into data[0]
		err := h.i2cDo(i2c.Read, r.offset(), i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: get8b Count #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	count := data[0] + 1
	for i := 0; i < 5000; i++ { // recover bus
		err := h.i2cDo(i2c.Read, 0, i2c.ByteData, &data)
		if err == nil {
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	if (count == 0) || (count == 1) {
		s := "Not Supported"
		return s
	}
	data[0] = count
	for i := 0; i < 5000; i++ { // read block
		err := h.i2cDo(i2c.Read, r.offset(), i2c.BlockData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: get8b #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	s := string(data[1:(data[0])])
	for i := 0; i < 5000; i++ { // recover bus
		err := h.i2cDo(i2c.Read, 0, i2c.ByteData, &data)
		if err == nil {
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	return s
}

func (r *reg16) get(h *Psu) (v uint16) {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return 0
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: get16: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: get16 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("fsp550: get16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	return uint16(data[0])<<8 | uint16(data[1])
}

func (r *reg16r) get(h *Psu) (v uint16) {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return 0
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: get16r: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: get16r MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("fsp550: get16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
	return uint16(data[1])<<8 | uint16(data[0])
}

func (r *reg8) set(h *Psu, v uint8) {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: set8: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: set8 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("fsp550: set8 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16) set(h *Psu, v uint16) {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: set16: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: set16 MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("fsp550: set16 #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *reg16r) set(h *Psu, v uint16) {
	var data i2c.SMBusData
	if h.Installed == 0 {
		return
	}
	i2c.Lock.Lock()
	defer func() {
		if rc := recover(); rc != nil {
			log.Print("Recovered in fsp550: set16r: ", rc, ", addr: ", r.offset())
		}
		i2c.Lock.Unlock()
	}()
	data[0] = byte(h.MuxValue)
	for i := 0; i < 5000; i++ {
		err := h.i2cDoMux(i2c.Write, 0, i2c.ByteData, &data)
		if err == nil {
			if i > 0 {
				log.Print("fsp550: set16r MuxWr #retries: ", i, ", addr: ", r.offset())
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
				log.Print("fsp550: set16r #retries: ", i, ", addr: ", r.offset())
			}
			break
		}
		if i == 4999 {
			panic(err)
		}
	}
}

func (r *regi16) get(h *Psu) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *Psu, v int16)  { (*reg16)(r).set(h, uint16(v)) }

func (h *Psu) convert(v uint16) float64 {
	if h.Id == "Great Wall" {
		v = v & 0xfff
		nn := float64(6) * (-1)
		vv := float64(v) * (math.Exp2(nn))
		vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
		return vv
	} else if h.Id == "FSP" {
		r := getPsuRegs()
		n := r.VoutMode.get(h) & 0x1f
		n = (n ^ 0x1f) & 0x1f
		nn := float64(n) * (-1)
		vv := (float64(v) * (math.Exp2(nn))) / 2
		vv, _ = strconv.ParseFloat(fmt.Sprintf("%.3f", vv), 64)
		return vv
	} else {
		return 0
	}
}

func (h *Psu) Page() uint16 {
	r := getPsuRegs()
	t := r.Page.get(h)
	return uint16(t)
}

func (h *Psu) PageWr(i uint16) {
	r := getPsuRegs()
	r.Page.set(h, uint8(i))
	return
}

func (h *Psu) StatusWord() uint16 {
	r := getPsuRegs()
	t := r.StatusWord.get(h)
	return uint16(t)
}

func (h *Psu) StatusVout() uint16 {
	r := getPsuRegs()
	t := r.StatusVout.get(h)
	return uint16(t)
}

func (h *Psu) StatusIout() uint16 {
	r := getPsuRegs()
	t := r.StatusIout.get(h)
	return uint16(t)
}

func (h *Psu) StatusInput() uint16 {
	r := getPsuRegs()
	t := r.StatusInput.get(h)
	return uint16(t)
}

func (h *Psu) StatusTemp() uint16 {
	r := getPsuRegs()
	t := r.StatusTemp.get(h)
	return uint16(t)
}

func (h *Psu) StatusFans() uint16 {
	r := getPsuRegs()
	t := r.StatusFans.get(h)
	return uint16(t)
}

func (h *Psu) Vin() float64 {
	r := getPsuRegs()
	t := r.Vin.get(h)
	v := h.convert(t)
	return v
}

func (h *Psu) Iin() uint16 {
	r := getPsuRegs()
	t := r.Iin.get(h)
	return uint16(t)
}

func (h *Psu) Vout() float64 {
	r := getPsuRegs()
	t := r.Vout.get(h)
	v := h.convert(t)
	return v
}

func (h *Psu) Iout() uint16 {
	r := getPsuRegs()
	t := r.Iout.get(h)
	return uint16(t)
}

func (h *Psu) Temp1() uint16 {
	r := getPsuRegs()
	t := r.Temp1.get(h)
	return uint16(t)
}

func (h *Psu) Temp2() uint16 {
	r := getPsuRegs()
	t := r.Temp2.get(h)
	return uint16(t)
}

func (h *Psu) FanSpeed() uint16 {
	r := getPsuRegs()
	t := r.FanSpeed.get(h)
	return uint16(t)
}

func (h *Psu) Pout() float64 {
	r := getPsuRegs()
	t := r.Pout.get(h)
	v := h.convert(t)
	return v
}

func (h *Psu) Pin() float64 {
	r := getPsuRegs()
	t := r.Pin.get(h)
	v := h.convert(t)
	return v
}

func (h *Psu) PoutRaw() uint16 {
	r := getPsuRegs()
	t := r.Pout.get(h)
	return t
}

func (h *Psu) PinRaw() uint16 {
	r := getPsuRegs()
	t := r.Pin.get(h)
	return t
}

func (h *Psu) ModeRaw() uint16 {
	if h.Id == "Great Wall" {
		r := getPsuRegs()
		t := r.Pin.get(h)
		return t
	} else {
		return 0
	}
}

func (h *Psu) PMBusRev() uint16 {
	r := getPsuRegs()
	t := r.PMBusRev.get(h)
	return uint16(t)
}

func (h *Psu) MfgIdent() string {
	r := getPsuRegs()
	t := r.MfgId.get(h)
	if t == "Not Supported" {
		t = "FSP"
	}
	t = strings.Trim(t, "#")
	h.Id = t
	return t
}

func (h *Psu) MfgModel() string {
	r := getPsuRegs()
	t := r.MfgMod.get(h)
	if t == "Not Supported" {
		t = "FSP"
	}
	t = strings.Trim(t, "#")
	return t
}

func (h *Psu) PsuStatus() string {
	pin, found := gpio.Pins[h.GpioPrsntL]
	if !found {
		h.Installed = 0
		return "not_found"
	} else {
		t, err := pin.Value()
		if err != nil {
			h.Installed = 0
			return err.Error()
		} else if t {
			h.Installed = 0
			return "not_installed"
		}
	}

	h.Installed = 1
	pin, found = gpio.Pins[h.GpioPwrok]
	if !found {
		return "undetermined"
	}
	t, err := pin.Value()
	if err != nil {
		return err.Error()
	}
	if !t {
		return "powered_off"
	}
	return "powered_on"
}

func (h *Psu) SetAdminState(s string) {
	pin, found := gpio.Pins[h.GpioPwronL]
	if found {
		switch s {
		case "disable":
			pin.SetValue(true)
			log.Print("notice: psu", h.Slot, " ", s)
		case "enable":
			pin.SetValue(false)
			log.Print("notice: psu", h.Slot, " ", s)
		}
	}
}

func (h *Psu) GetAdminState() string {
	pin, found := gpio.Pins[h.GpioPwronL]
	if !found {
		return "not found"
	}
	t, err := pin.Value()
	if err != nil {
		return err.Error()
	}
	if t {
		return "disabled"
	}
	return "enabled"
}
