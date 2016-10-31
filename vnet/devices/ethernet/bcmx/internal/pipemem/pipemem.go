// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pipemem

import (
	"github.com/platinasystems/go/elib"
)

type Ref uint64

const index_bits = 48

func (r Ref) get() (index elib.Word, mask uint) {
	mask = uint(r >> index_bits)
	index = elib.Word(r & ((1 << index_bits) - 1))
	return
}

func (r Ref) Get() (index, mask uint) {
	var i elib.Word
	i, mask = r.get()
	index = uint(i)
	return
}

func (r Ref) Index() (i uint) {
	i, _ = r.Get()
	return
}

func makeRef(mask, index uint) Ref { return Ref(mask<<index_bits | index) }

type Pool struct {
	free   []elib.WordVec
	nPipe  uint
	len    uint
	maxLen uint
}

func (p *Pool) Init(nPipe, maxLen uint) {
	p.nPipe = nPipe
	p.maxLen = maxLen
	p.free = make([]elib.WordVec, nPipe, nPipe)
}

func (p *Pool) Len() uint { return p.len }

// All pipes must have same free length.
func (p *Pool) freeLen() uint { return p.free[0].Len() }

func (p *Pool) Get(pipeMask uint) (Ref, bool) {
	for i := uint(0); i < p.freeLen(); i++ {
		f := ^elib.Word(0)
		for pipe := uint(0); pipe < p.nPipe; pipe++ {
			if pipeMask&(1<<pipe) != 0 {
				f &= p.free[pipe][i]
				if f == 0 {
					break
				}
			}
		}
		if f != 0 {
			i1 := f.FirstSet().MinLog2()
			m1 := elib.Word(1) << i1
			for pipe := uint(0); pipe < p.nPipe; pipe++ {
				if pipeMask&(1<<pipe) != 0 {
					p.free[pipe][i] &^= m1
				}
			}
			return makeRef(pipeMask, i*elib.WordBits+i1), true
		}
	}

	if p.len >= p.maxLen {
		return 0, false
	}

	i := elib.Word(p.len)
	p.len++
	i0, i1 := i.BitmapIndex()
	for pipe := uint(0); pipe < p.nPipe; pipe++ {
		p.free[pipe].Validate(uint(i0))
		if pipeMask&(1<<pipe) == 0 {
			p.free[pipe][i0] |= i1
		}
	}
	return makeRef(pipeMask, uint(i)), true
}

// Find and return index free in both given pools.
func GetCoupled(p0, p1 *Pool, pipeMask0, pipeMask1 uint) (ref0, ref1 Ref, ok bool) {
	fl0, fl1 := p0.freeLen(), p1.freeLen()
	fl := fl0
	if fl > fl1 {
		fl = fl1
	}
	if p0.nPipe != p1.nPipe {
		panic("npipe")
	}

	for i := uint(0); i < fl; i++ {
		f := ^elib.Word(0)
		for pipe := uint(0); pipe < p0.nPipe; pipe++ {
			pm := uint(1) << pipe
			if pipeMask0&pm != 0 {
				f &= p0.free[pipe][i]
				if f == 0 {
					break
				}
			}
			if pipeMask1&pm != 0 {
				f &= p1.free[pipe][i]
				if f == 0 {
					break
				}
			}
		}
		if f != 0 {
			i1 := f.FirstSet().MinLog2()
			m1 := elib.Word(1) << i1
			for pipe := uint(0); pipe < p0.nPipe; pipe++ {
				pm := uint(1) << pipe
				if pipeMask0&pm != 0 {
					p0.free[pipe][i] &^= m1
				}
				if pipeMask1&pm != 0 {
					p1.free[pipe][i] &^= m1
				}
			}
			ri := i*elib.WordBits + i1
			return makeRef(pipeMask0, ri), makeRef(pipeMask1, ri), true
		}
	}

	if p0.len >= p0.maxLen {
		return 0, 0, false
	}

	for p1.len < p0.len && p1.len < p1.maxLen {
		i := elib.Word(p1.len)
		i0, i1 := i.BitmapIndex()
		for pipe := uint(0); pipe < p1.nPipe; pipe++ {
			p1.free[pipe].Validate(uint(i0))
			p1.free[pipe][i0] |= i1
		}
		p1.len++
	}

	if p1.len >= p1.maxLen {
		return 0, 0, false
	}

	i := elib.Word(p0.len)
	p0.len++
	p1.len++
	i0, i1 := i.BitmapIndex()
	for pipe := uint(0); pipe < p0.nPipe; pipe++ {
		p0.free[pipe].Validate(uint(i0))
		p1.free[pipe].Validate(uint(i0))
		if pipeMask0&(1<<pipe) == 0 {
			p0.free[pipe][i0] |= i1
		}
		if pipeMask1&(1<<pipe) == 0 {
			p1.free[pipe][i0] |= i1
		}
	}
	return makeRef(pipeMask0, uint(i)), makeRef(pipeMask1, uint(i)), true
}

func (p *Pool) Put(r Ref) {
	i, pipeMask := r.get()
	i0, i1 := i.BitmapIndex()
	for pipe := uint(0); pipe < p.nPipe; pipe++ {
		if pipeMask&(1<<pipe) != 0 {
			p.free[pipe][i0] |= i1
		}
	}
}

func (p *Pool) Foreach(f func(pipe, index uint)) {
	for i := uint(0); i < p.len; i++ {
		i0, i1 := elib.Word(i).BitmapIndex()
		for pipe := uint(0); pipe < p.nPipe; pipe++ {
			if p.free[pipe][i0]&i1 == 0 {
				f(pipe, i)
			}
		}
	}
}
