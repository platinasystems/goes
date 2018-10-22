// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package qsfp

import (
	"net/rpc"
	"unsafe"

	"github.com/platinasystems/i2c"
	"github.com/platinasystems/log"
)

var bio = [34]byte{0}
var iio = I{false, i2c.RW(0), 0, 0, bio, 0, 0, 0}
var jio [MAXOPS]I
var rio = R{bio, nil}
var sio [MAXOPS]R
var xio int

var dummyio byte
var regsPointerio = unsafe.Pointer(&dummyio)
var regsAddrio = uintptr(unsafe.Pointer(&dummyio))

var clientAio *rpc.Client
var dialedio int = 0

func getRegs() *regsio            { return (*regsio)(regsPointerio) }
func (r *regio8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddrio) }
func (r *regio16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddrio) }
func (r *regio16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddrio) }

func closeMuxio(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(0)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++

}

func (r *regio8) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++
	jio[xio] = I{true, i2c.Read, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0}
	xio++
}

func (r *regio16) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++
	jio[xio] = I{true, i2c.Read, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	xio++
}

func (r *regio16r) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++
	jio[xio] = I{true, i2c.Read, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	xio++
}

func (r *regio8) set(h *I2cDev, v uint8) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++
	data[0] = v
	jio[xio] = I{true, i2c.Write, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0}
	xio++
}

func (r *regio16) set(h *I2cDev, v uint16) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	jio[xio] = I{true, i2c.Write, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	xio++
}

func (r *regio16r) set(h *I2cDev, v uint16) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	jio[xio] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	xio++
	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	jio[xio] = I{true, i2c.Write, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	xio++
}

func clearJio() {
	xio = 0
	for k := 0; k < MAXOPS; k++ {
		jio[k] = iio
	}
}

func readStoppedio() byte {
	var data = [34]byte{0, 0, 0, 0}
	jio[0] = I{true, i2c.Write, 0, i2c.ByteData, data, int(0x98), int(0), 0}
	DoI2cRpcio()
	return sio[0].D[0]
}

func DoI2cRpcio() (err error) {
	if dialedio == 0 {
		client, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1233")
		if err != nil {
			log.Print("dialing:", err)
		}
		clientAio = client
		dialedio = 1
	}
	err = clientAio.Call("I2cReq.ReadWrite", &jio, &sio)
	if err != nil {
		log.Print("i2cReq error:", err)
	}
	clearJio()
	return err
}
