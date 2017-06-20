// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"

	"unsafe"
)

// Incremental checksum update.
type Checksum uint64

func (sum Checksum) AddWithCarry(x Checksum) (t Checksum) {
	t = sum + x
	if t < x {
		t++
	}
	return
}

func (c Checksum) AddBytes(b []byte) (d Checksum) {
	d = c
	i, n_left := 0, len(b)

	var sum0, sum1, sum2, sum3 Checksum

	for n_left >= 8*4 {
		sum0 = sum0.AddWithCarry(*(*Checksum)(unsafe.Pointer(&b[i+8*0])))
		sum1 = sum1.AddWithCarry(*(*Checksum)(unsafe.Pointer(&b[i+8*1])))
		sum2 = sum2.AddWithCarry(*(*Checksum)(unsafe.Pointer(&b[i+8*2])))
		sum3 = sum3.AddWithCarry(*(*Checksum)(unsafe.Pointer(&b[i+8*3])))
		n_left -= 8 * 4
		i += 8 * 4
	}

	for n_left >= 8 {
		sum0 = sum0.AddWithCarry(*(*Checksum)(unsafe.Pointer(&b[i+8*0])))
		n_left -= 8 * 1
		i += 8 * 1
	}

	for n_left >= 2 {
		sum0 = sum0.AddWithCarry(Checksum(*(*uint16)(unsafe.Pointer(&b[i]))))
		n_left -= 2
		i += 2
	}

	if n_left > 0 {
		v := Checksum(b[i])
		if vnet.HostIsNetworkByteOrder() {
			v <<= 8
		}
		sum0 = sum0.AddWithCarry(v)
	}

	d = d.AddWithCarry(sum0)
	d = d.AddWithCarry(sum1)
	d = d.AddWithCarry(sum2)
	d = d.AddWithCarry(sum3)
	return
}

func (c Checksum) AddRef(first *vnet.Ref, o_first uint) (d Checksum) {
	d = c
	first.Foreach(func(r *vnet.Ref, i uint) {
		o := uint(0)
		if i == 0 {
			o = o_first
		}
		d = d.AddBytes(r.DataOffsetSlice(o))
	})
	return
}

// Reduce to 16 bits.
func (c Checksum) Fold() vnet.Uint16 {
	m1, m2 := Checksum(0xffffffff), Checksum(0xffff)
	c = (c & m1) + c>>32
	c = (c & m2) + c>>16
	c = (c & m2) + c>>16
	c = (c & m2) + c>>16
	return vnet.Uint16(c)
}

// Update checksum changing field at even byte offset from 0 to x.
func (c Checksum) AddEven(x Checksum) (d Checksum) {
	d = c - x
	// Fold in carry from high bit.
	if d > c {
		d--
	}
	if elib.Debug && d.AddWithCarry(x) != c {
		panic("add even")
	}
	return
}

func (c Checksum) SubEven(x Checksum) Checksum { return c.AddWithCarry(x) }
