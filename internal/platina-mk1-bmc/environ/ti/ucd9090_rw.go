// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package ucd9090 provides access to the UCD9090 Power Sequencer/Monitor chip
package ucd9090

import (
	"bufio"
	"encoding/gob"
	"os"
	"time"
	"unsafe"

	"github.com/platinasystems/go/internal/i2c"
	"github.com/platinasystems/go/internal/log"
)

const (
	NONE  = 0
	DO    = 1
	DOMUX = 2
)
const (
	REG8   = 1
	REG16  = 2
	REG16R = 3
)
const MAXTRANS = 10

type I struct {
	Op        int
	RegType   int
	RW        i2c.RW
	RegOffset uint8
	BusSize   i2c.SMBusSize
	Data      [4]byte
	ErrCode   error
	Bus       int
	Addr      int
	MuxBus    int
	MuxAddr   int
	MuxValue  int
}
type R struct {
	F float64
	I uint8
	J uint16
	S string
	E error
}

var b = [4]byte{0, 0, 0, 0}
var i = I{NONE, 0, 0, 0, 0, b, nil, 0, 0, 0, 0, 0}
var j [MAXTRANS]I
var x int
var r = R{float64(0), uint8(0), uint16(0), "string", nil}
var s [MAXTRANS]R

var dummy byte
var regsPointer = unsafe.Pointer(&dummy)
var regsAddr = uintptr(unsafe.Pointer(&dummy))

func getPwmRegs() *pwmRegs      { return (*pwmRegs)(regsPointer) }
func (r *reg8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (r *reg8) get(h *PMon) {
	var data = [4]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{DOMUX, REG8, i2c.Write, 0, i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	j[x] = I{DO, REG8, i2c.Read, r.offset(), i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	//return data[0] ==> move this to server
}

func (r *reg16) get(h *PMon) {
	var data = [4]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{DOMUX, REG16, i2c.Write, 0, i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	j[x] = I{DO, REG16, i2c.Read, r.offset(), i2c.WordData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	//return uint16(data[0])<<8 | uint16(data[1]) ==> move this to server
}

func (r *reg16r) get(h *PMon) {
	var data = [4]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{DOMUX, REG16R, i2c.Write, 0, i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	j[x] = I{DO, REG16R, i2c.Read, r.offset(), i2c.WordData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	//return uint16(data[1])<<8 | uint16(data[0]) ==> move this to server
}

func (r *reg8) set(h *PMon, v uint8) {
	var data = [4]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{DOMUX, REG8, i2c.Write, 0, i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	data[0] = v
	j[x] = I{DO, REG8, i2c.Write, r.offset(), i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
}

func (r *reg16) set(h *PMon, v uint16) {
	var data = [4]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{DOMUX, REG16, i2c.Write, 0, i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	j[x] = I{DO, REG16, i2c.Write, r.offset(), i2c.WordData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
}

func (r *reg16r) set(h *PMon, v uint16) {
	var data = [4]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{DOMUX, REG16R, i2c.Write, 0, i2c.ByteData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	j[x] = I{DO, REG16R, i2c.Write, r.offset(), i2c.WordData, data, nil, h.Bus, h.Addr, h.MuxBus, h.MuxAddr, h.MuxValue}
	x++
}

func clearJS() {
	x = 0
	for k := 0; k < MAXTRANS; k++ {
		j[k] = i
		s[k] = r
	}
}

//FIXME: REPLACE THIS WITH RPC, CHANGE RETURN to error
func DoI2cRpc() {

	f, err := os.Create("/tmp/i2c_ti.dat") //create file, writer
	if err != nil {
		panic(err)
	}
	defer f.Close()
	w := bufio.NewWriter(f)

	enc := gob.NewEncoder(w) //encode i2c ops, write file
	err = enc.Encode(&j)
	if err != nil {
		log.Print("err", "ti2 encode error: ", err)
	}
	w.Flush()
	f.Close()

	for { //wait for file to appear
		_, err = os.Stat("/tmp/i2c_ti.res")
		if !os.IsNotExist(err) {
			break
		}
		if err != nil {
			//log.Print("err", "ti2 server response error: ", err)
		}
		time.Sleep(time.Millisecond * time.Duration(5))
	}
	for {
		fi, er := os.Stat("/tmp/i2c_ti.res")
		if er != nil {
			log.Print("ti error: ", er)
		}
		size := fi.Size()
		if size == 154 {
			break
		}
		time.Sleep(time.Millisecond * time.Duration(5))
	}

	filename := "/tmp/i2c_ti.res" //open, reader
	file, err := os.Open(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	dec := gob.NewDecoder(reader) //read, decode
	err = dec.Decode(&s)
	if err != nil {
		log.Print("ti2 decode error:", err)
	}

	f.Close() //close, remove
	err = os.Remove("/tmp/i2c_ti.res")
	if err != nil {
		log.Print("could not remove i2c_ti.res file, error:", err)
	}
	return
}
