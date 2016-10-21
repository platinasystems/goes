// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/vnet"
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

var debug = true

// Update checksum changing field at even byte offset from 0 to x.
func (c Checksum) AddEven(x Checksum) (d Checksum) {
	d = c - x
	// Fold in carry from high bit.
	if d > c {
		d--
	}
	if debug && d.AddWithCarry(x) != c {
		panic("add even")
	}
	return
}

func (c Checksum) SubEven(x Checksum) Checksum { return c.AddWithCarry(x) }

// Reduce to 16 bits.
func (c Checksum) Fold() vnet.Uint16 {
	m1, m2 := Checksum(0xffffffff), Checksum(0xffff)
	c = (c & m1) + c>>32
	c = (c & m2) + c>>16
	c = (c & m2) + c>>16
	c = (c & m2) + c>>16
	return vnet.Uint16(c)
}
