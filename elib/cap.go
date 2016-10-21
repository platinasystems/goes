// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

// Vector capacities of the form 2^i + 2^j
type Cap uint32

// True if removing first set bit gives a power of 2.
func (c Cap) IsValid() bool {
	f := c & -c
	c ^= f
	return 0 == c&(c-1)
}

const CapNil = ^Cap(0)

func (n Cap) Round(log2Unit Cap) Cap {
	u := Word(1<<log2Unit - 1)
	w := (Word(n) + u) &^ u
	// Power of 2?
	if w&(w-1) != 0 {
		l0 := MinLog2(w)
		m0 := Word(1) << l0
		l1 := MaxLog2(w ^ m0)
		w = m0 + 1<<l1
	}
	return Cap(w)
}

func (c Cap) Pow2() (i, j Cap) {
	j = c & -c
	i = c ^ j
	if i == 0 {
		i = j
		j = CapNil
	}
	return
}

func (c Cap) Log2() (i, j Cap) {
	i, j = c.Pow2()
	i = Cap(Word(i).MinLog2())
	if j != CapNil {
		j = Cap(Word(j).MinLog2())
	}
	return
}

func (c Cap) NextUnit(log2Min, log2Unit Cap) (n Cap) {
	n = c

	// Double every 2/8/16 expands depending on table size.
	min := Cap(1)<<log2Min - 1
	switch {
	case n < min:
		n = min
	case n < 256:
		n = Cap(float64(n) * 1.41421356237309504878) /* exp (log2 / 2) */
	case n < 1024:
		n = Cap(float64(n) * 1.09050773266525765919) /* exp (log2 / 8) */
	default:
		n = Cap(float64(n) * 1.04427378242741384031) /* exp (log2 / 16) */
	}

	n = n.Round(log2Unit)
	return
}
func (c Cap) Next() Cap { return c.NextUnit(3, 2) }

// NextResizeCap gives next larger resizeable array capacity.
func NextResizeCap(x Index) Index { return Index(Cap(x).Next()) }
