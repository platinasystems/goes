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

const MAXOPS = 10

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

var b = [34]byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
var i = I{false, i2c.RW(0), 0, 0, b, 0, 0, 0, 0, 0}
var j [MAXOPS]I
var r = R{b, nil}
var s [MAXOPS]R
var x int

var dummy byte
var regsPointer = unsafe.Pointer(&dummy)
var regsAddr = uintptr(unsafe.Pointer(&dummy))

func getPwmRegs() *pwmRegs      { return (*pwmRegs)(regsPointer) }
func (r *reg8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *reg16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (r *reg8) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0, 0, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0, 0, 0}
	x++
}

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

func clearJS() {
	x = 0
	for k := 0; k < MAXOPS; k++ {
		j[k] = i
		s[k] = r
	}
}

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
		if size == 445 {
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
