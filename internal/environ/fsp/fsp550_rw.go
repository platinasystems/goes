// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package fsp

import (
	"net/rpc"
	"unsafe"

	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
)

const MAXOPS = 30

type I struct {
	InUse     bool
	RW        i2c.RW
	RegOffset uint8
	BusSize   i2c.SMBusSize
	Data      [34]byte
	Bus       int
	Addr      int
	Count     int
	Delay     int
	Eeprom    int
}
type R struct {
	D [34]byte
	E error
}

type I2cReq int

var b = [34]byte{0}
var i = I{false, i2c.RW(0), 0, 0, b, 0, 0, 0, 0, 0}
var j [MAXOPS]I
var r = R{b, nil}
var s [MAXOPS]R
var x int

var dummy byte
var regsPointer = unsafe.Pointer(&dummy)
var regsAddr = uintptr(unsafe.Pointer(&dummy))

var clientA *rpc.Client
var dialed int = 0

// offset function has divide by two for 16-bit offset struct
func getRegs() *regs            { return (*regs)(regsPointer) }
func (r *reg8) offset() uint8   { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg8b) offset() uint8  { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg16) offset() uint8  { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }
func (r *reg16r) offset() uint8 { return uint8((uintptr(unsafe.Pointer(r)) - regsAddr) >> 1) }

func (r *reg8) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

/*
func (r *reg8b) get(h *I2cDev) string {
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
func (r *regi16) get(h *I2cDev) (v int16) { v = int16((*reg16)(r).get(h)); return }
func (r *regi16) set(h *I2cDev, v int16)  { (*reg16)(r).set(h, uint16(v)) }
*/

func (r *reg16) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

func (r *reg16r) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

func (r *reg8) set(h *I2cDev, v uint8) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	data[0] = v
	j[x] = I{true, i2c.Write, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

func (r *reg16) set(h *I2cDev, v uint16) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	j[x] = I{true, i2c.Write, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

func (r *reg16r) set(h *I2cDev, v uint16) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	j[x] = I{true, i2c.Write, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

func clearJ() {
	x = 0
	for k := 0; k < MAXOPS; k++ {
		j[k] = i
	}
}

func DoI2cRpc() {
	if dialed == 0 {
		client, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1233")
		if err != nil {
			log.Print("dialing:", err)
		}
		clientA = client
		dialed = 1
	}
	err := clientA.Call("I2cReq.ReadWrite", &j, &s)
	if err != nil {
		log.Print("i2cReq error:", err)
	}
	clearJ()
	return
}
