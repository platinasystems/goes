// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package qsfp

import (
	"unsafe"

	"github.com/platinasystems/go/internal/i2c"
)

func getRegs() *regsio            { return (*regsio)(regsPointer) }
func (r *regio8) offset() uint8   { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *regio16) offset() uint8  { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }
func (r *regio16r) offset() uint8 { return uint8(uintptr(unsafe.Pointer(r)) - regsAddr) }

func (r *regio8) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0}
	x++
	data[0] = byte(0)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
}

func (r *regio16) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	x++
}

func (r *regio16r) get(h *I2cDev) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
	j[x] = I{true, i2c.Read, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	x++
}

func (r *regio8) set(h *I2cDev, v uint8) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
	data[0] = v
	j[x] = I{true, i2c.Write, r.offset(), i2c.ByteData, data, h.Bus, h.Addr, 0}
	x++
	data[0] = byte(0)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
}

func (r *regio16) set(h *I2cDev, v uint16) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
	data[0] = uint8(v >> 8)
	data[1] = uint8(v)
	j[x] = I{true, i2c.Write, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	x++
}

func (r *regio16r) set(h *I2cDev, v uint16) {
	var data = [34]byte{0, 0, 0, 0}

	data[0] = byte(h.MuxValue)
	j[x] = I{true, i2c.Write, 0, i2c.ByteData, data, h.MuxBus, h.MuxAddr, 0}
	x++
	data[1] = uint8(v >> 8)
	data[0] = uint8(v)
	j[x] = I{true, i2c.Write, r.offset(), i2c.WordData, data, h.Bus, h.Addr, 0}
	x++
}
