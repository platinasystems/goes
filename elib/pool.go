// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"errors"
)

type Pool struct {
	// Vector of free indices
	freeIndices []uint32 // Uint32Vec
	// Bitmap of free indices
	freeBitmap Bitmap
	// Non-zero to limit size of pool.
	maxLen uint
}

// ErrTooLarge is passed to panic if pool overflows maxLen
var ErrPoolTooLarge = errors.New("pool: too large")

// Get first free pool index if available.
func (p *Pool) GetIndex(max uint) (i uint) {
	i = max
	l := uint(len(p.freeIndices))
	if l != 0 {
		i = uint(p.freeIndices[l-1])
		p.freeIndices = p.freeIndices[:l-1]
		p.freeBitmap = p.freeBitmap.AndNotx(i)
	}
	if p.maxLen != 0 && i >= p.maxLen {
		panic(ErrPoolTooLarge)
	}
	return
}

// Put (free) given pool index.
func (p *Pool) PutIndex(i uint) (ok bool) {
	if ok = !p.freeBitmap.Get(i); ok {
		p.freeIndices = append(p.freeIndices, uint32(i))
		p.freeBitmap = p.freeBitmap.Orx(i)
	}
	return
}

func (p *Pool) Reset() {
	if p.freeIndices != nil {
		p.freeIndices = p.freeIndices[:0]
	}
	p.freeBitmap = p.freeBitmap.Free()
}

func (p *Pool) IsFree(i uint) (ok bool) { return p.freeBitmap.Get(i) }
func (p *Pool) FreeLen() uint           { return uint(len(p.freeIndices)) }
func (p *Pool) MaxLen() uint            { return p.maxLen }
func (p *Pool) SetMaxLen(x uint)        { p.maxLen = x }
