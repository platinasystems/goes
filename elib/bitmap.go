// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package elib is a collection of data structures: bitmaps, pools, heaps, fifos.
package elib

import (
	"fmt"
)

// Bitmaps of <= 63 are stored as positive int64s.
// Negative int64s are indices into pool of all known bitmaps
type Bitmap Word

//go:generate gentemplate -d Package=elib -id Bitmap -d VecType=BitmapVec -d Type=Bitmap vec.tmpl
//go:generate gentemplate -d Package=elib -id BitmapVec -d VecType=BitmapsVec -d Type=[]BitmapVec vec.tmpl

func (b Bitmap) isPoolIndex() bool { return b>>(WordBits-1) != 0 }
func (b Bitmap) isInline() bool    { return !b.isPoolIndex() }

// bothMemory is true iff both arguments are direct non-memory bitmaps
func bothInline(b, c Bitmap) bool { return (b | c).isInline() }

// index gives word index and mask for given bit index
func bitmapIndex(x uint) (i uint, m Word) {
	i = x / WordBits
	m = 1 << (x % WordBits)
	return
}

func bitmapGet(b []Word, x uint) bool {
	i, m := bitmapIndex(x)
	return b[i]&m != 0
}

func bitmapSet(b []Word, x uint) (old bool) {
	i, m := bitmapIndex(x)
	v := b[i]
	old = v&m != 0
	b[i] = v | m
	return
}

func bitmapUnset(b []Word, x uint) (old bool) {
	i, m := bitmapIndex(x)
	v := b[i]
	old = v&m != 0
	b[i] = v &^ m
	return
}

func bitmapMake(nBits uint) []Word {
	n, _ := bitmapIndex(nBits)
	return make([]Word, n)
}

//go:generate gentemplate -d Package=elib -id Bitmap -d PoolType=BitmapPool -d Type=WordVec -d Data=bitmaps pool.tmpl

var Bitmaps = &BitmapPool{}

func (p *BitmapPool) new() uint {
	bi := p.GetIndex()
	p.Validate(bi)
	return bi
}

func (p *BitmapPool) toMem(b Bitmap) Bitmap {
	bi := p.new()
	if len(p.bitmaps[bi]) == 0 {
		p.bitmaps[bi] = append(p.bitmaps[bi], Word(b))
	} else {
		p.bitmaps[bi] = p.bitmaps[bi][:1]
		p.bitmaps[bi][0] = Word(b)
	}
	return ^Bitmap(bi)
}

func (p *BitmapPool) inlineToMem(b Bitmap) (r Bitmap) {
	r = b
	if b.isInline() {
		r = p.toMem(b)
	}
	return
}

// Set2 sets bit X in given bitmap, possibly resizeing and returning new bitmap.
// Second return value is old value of bit X or false if bit is newly created.
func (p *BitmapPool) Set2(b Bitmap, x uint) (r Bitmap, v bool) {
	var bi Bitmap

	r = b
	if b.isInline() {
		if x < WordBits-1 {
			m := Bitmap(1) << x
			v = b&m != 0
			r |= m
			return
		}

		r = p.toMem(b)
	}

	i, m := bitmapIndex(x)
	bi = ^r
	s := p.bitmaps[bi]
	s.Validate(i)
	p.bitmaps[bi] = s
	si := s[i]
	v = si&m != 0
	s[i] = si | m
	return
}

func (p *BitmapPool) Set(b Bitmap, x uint) (r Bitmap) {
	r, _ = p.Set2(b, x)
	return
}

func (p *BitmapPool) Get(b Bitmap, x uint) (v bool) {
	if !b.isPoolIndex() {
		m := Bitmap(1) << x
		// Out of range bits are always zero.
		if x >= WordBits {
			m = 0
		}
		return b&m != 0
	}
	i, m := bitmapIndex(x)
	bm := p.bitmaps[^b]
	if i < uint(len(bm)) {
		v = bm[i]&m != 0
	}
	return
}

func (p *BitmapPool) Unset2(b Bitmap, x uint) (r Bitmap, old bool) {
	r = b
	if !b.isPoolIndex() {
		m := Bitmap(1) << x
		old = b&m != 0
		r &= ^m
		return
	}

	i, m := bitmapIndex(x)
	s := p.bitmaps[^r]
	si := s[i]
	old = si&m != 0
	s[i] = si &^ m
	return
}

func (p *BitmapPool) Unset(b Bitmap, x uint) (r Bitmap) {
	r, _ = p.Unset2(b, x)
	return
}

func (p *BitmapPool) Invert2(b Bitmap, x uint) (r Bitmap, v bool) {
	var bi Bitmap

	r = b
	if b.isInline() {
		if x < WordBits-1 {
			m := Bitmap(1) << x
			v = b&m != 0
			r ^= m
			return
		}

		r = p.toMem(b)
	}

	i, m := bitmapIndex(x)
	bi = ^r
	s := p.bitmaps[bi]
	s.Validate(i)
	p.bitmaps[bi] = s
	si := s[i]
	v = si&m != 0
	s[i] = si ^ m
	return
}

func (p *BitmapPool) Invert(b Bitmap, x uint) (r Bitmap) {
	r, _ = p.Invert2(b, x)
	return
}

func (b Bitmap) Get(x uint) bool               { return Bitmaps.Get(b, x) }
func (b Bitmap) Set(x uint) Bitmap             { return Bitmaps.Set(b, x) }
func (b Bitmap) Invert(x uint) Bitmap          { return Bitmaps.Invert(b, x) }
func (b Bitmap) Set2(x uint) (Bitmap, bool)    { return Bitmaps.Set2(b, x) }
func (b Bitmap) Invert2(x uint) (Bitmap, bool) { return Bitmaps.Invert2(b, x) }

func (p *BitmapPool) Orx(b Bitmap, x uint) (r Bitmap) {
	r = b
	if !r.isPoolIndex() {
		if x < WordBits-1 {
			r |= Bitmap(1) << x
			return
		}
		r = p.toMem(r)
	}
	bi := uint(^r)
	i, m := bitmapIndex(x)
	p.bitmaps[bi].Validate(i)
	p.bitmaps[bi][i] |= m
	return
}

func (p *BitmapPool) Or(b Bitmap, c Bitmap) (r Bitmap) {
	r = b
	if bothInline(b, c) {
		r |= c
		return
	}
	r = p.inlineToMem(r)
	bi := uint(^r)
	var cs []Word
	if c.isPoolIndex() {
		cs = p.bitmaps[^c]
	} else {
		cs = []Word{Word(c)}
	}
	p.bitmaps[bi].Validate(uint(len(cs) - 1))
	for i := range cs {
		p.bitmaps[bi][i] |= cs[i]
	}
	return
}

func (b Bitmap) Orx(x uint) Bitmap {
	return Bitmaps.Orx(b, x)
}

func (b Bitmap) Or(c Bitmap) Bitmap {
	return Bitmaps.Or(b, c)
}

// Possibly reduce memory bitmap to inline.
func (p *BitmapPool) checkInline(b Bitmap) (r Bitmap) {
	r = b
	bi := uint(^r)
	l := len(p.bitmaps[bi])
	if l > 1 {
		return
	}
	if l == 1 && !Bitmap(p.bitmaps[bi][0]).isInline() {
		return
	}
	r = 0
	if l == 1 {
		r = Bitmap(p.bitmaps[bi][0])
	}
	p.PutIndex(uint(bi))
	return
}

func (p *BitmapPool) AndNot(b Bitmap, c Bitmap) (r Bitmap) {
	r = b
	if bothInline(b, c) {
		r &= ^c
		return
	}
	r = p.inlineToMem(r)
	bi := uint(^r)
	var cs []Word
	if c.isPoolIndex() {
		cs = p.bitmaps[^c]
	} else {
		cs = []Word{Word(c)}
	}
	p.bitmaps[bi].Validate(uint(len(cs) - 1))
	l := 0
	for i := range cs {
		p.bitmaps[bi][i] &= ^cs[i]
		if p.bitmaps[bi][i] != 0 {
			l = i + 1
		}
	}
	// Strip trailing 0s
	p.bitmaps[bi] = p.bitmaps[bi][:l]
	r = p.checkInline(r)
	return
}

func (p *BitmapPool) AndNotx(b Bitmap, x uint) (r Bitmap) {
	r = b
	if !r.isPoolIndex() {
		if x < WordBits-1 {
			r &= ^(Bitmap(1) << x)
			return
		}
		r = p.toMem(r)
	}
	bi := uint(^r)
	i, m := bitmapIndex(x)
	p.bitmaps[bi].Validate(i)
	v := p.bitmaps[bi][i]
	v &= ^m
	p.bitmaps[bi][i] = v

	if l := len(p.bitmaps[bi]); i == uint(l-1) && v == 0 {
		for l > 0 && p.bitmaps[bi][l-1] == 0 {
			l--
		}
		p.bitmaps[bi] = p.bitmaps[bi][:l]
		r = p.checkInline(r)
	}
	return
}

func (b Bitmap) AndNotx(x uint) Bitmap {
	return Bitmaps.AndNotx(b, x)
}

func (b Bitmap) AndNot(c Bitmap) Bitmap {
	return Bitmaps.AndNot(b, c)
}

// Dup copies given bitmap
func (p *BitmapPool) Dup(b Bitmap) Bitmap {
	if b.isPoolIndex() {
		bi := p.new()
		l := uint(len(p.bitmaps[^b]))
		p.bitmaps[bi].Validate(l - 1)
		p.bitmaps[bi] = p.bitmaps[bi][:l]
		copy(p.bitmaps[bi], p.bitmaps[^b])
		b = ^Bitmap(bi)
	}
	return b
}

func (b Bitmap) Dup() Bitmap {
	return Bitmaps.Dup(b)
}

// Free bitmap so it can be reused later.  Returns zero bitmap.
func (p *BitmapPool) Free(b Bitmap) Bitmap {
	if b.isPoolIndex() {
		p.PutIndex(uint(^b))
	}
	return 0
}

func (b Bitmap) Free() Bitmap {
	return Bitmaps.Free(b)
}

func (p *BitmapPool) ForeachSetBit(b Bitmap, fn func(uint)) {
	var s []Word
	if !b.isPoolIndex() {
		s = []Word{Word(b)}
	} else {
		s = p.bitmaps[^b]
	}
	for i := range s {
		x := s[i]
		for x != 0 {
			f := x.FirstSet()
			l := f.MinLog2()
			fn(uint(i*WordBits) + l)
			x ^= f
		}
	}
}

func (b Bitmap) ForeachSetBit(fn func(uint)) { Bitmaps.ForeachSetBit(b, fn) }

func (p *BitmapPool) Next(b Bitmap, px *uint) (ok bool) {
	x := *px
	i := uint(0)
	m := ^Word(0)
	if x != ^uint(0) {
		i, m = bitmapIndex(x)
		m = ^(m | (m - 1))
	}
	var s []Word
	if !b.isPoolIndex() {
		s = []Word{Word(b)}
	} else {
		s = p.bitmaps[^b]
	}

	m &= s[i]
	if m != 0 {
		*px = i*WordBits + m.FirstSet().MinLog2()
		ok = true
		return
	} else {
		for i++; i < uint(len(s)); i++ {
			m = s[i]
			if m != 0 {
				*px = i*WordBits + m.FirstSet().MinLog2()
				ok = true
				return
			}
		}
	}

	*px = ^uint(0)
	ok = false
	return
}

func (b Bitmap) Next(px *uint) bool {
	return Bitmaps.Next(b, px)
}

func (p *BitmapPool) String(b Bitmap) string {
	s := "{"
	p.ForeachSetBit(b, func(x uint) {
		if len(s) > 1 {
			s += ", "
		}
		s += fmt.Sprintf("%d", x)
	})
	s += "}"
	return s
}

func (b Bitmap) String() string {
	return Bitmaps.String(b)
}

func (p *BitmapPool) HexString(b Bitmap) string {
	var x []Word
	if !b.isPoolIndex() {
		x = []Word{Word(b)}
	} else {
		x = p.bitmaps[^b]
	}
	i := len(x) - 1
	s := fmt.Sprintf("0x%x", uint64(x[i]))
	for i--; i >= 0; i-- {
		s += fmt.Sprintf("%016x", uint64(x[i]))
	}
	return s
}

func (b Bitmap) HexString() string {
	return Bitmaps.HexString(b)
}

func (bm WordVec) GetBit(x uint) (v bool) {
	i, m := bitmapIndex(x)
	if i < uint(len(bm)) {
		v = bm[i]&m != 0
	}
	return
}

func (bm WordVec) SetBit(x uint, v bool) (oldValue bool) {
	i, m := bitmapIndex(x)
	b := bm[i]
	oldValue = b&m != 0
	b |= m
	bm[i] = b
	return
}

func (bm WordVec) UnsetBit(x uint) (v bool) {
	i, m := bitmapIndex(x)
	b := bm[i]
	v = b&m != 0
	b &^= m
	bm[i] = b
	return
}

// Fetch bits I through I + N_BITS.
func (bm *WordVec) GetMultiple(x, n_bits uint) (v Word) {
	i0, i1 := x/WordBits, x%WordBits
	l := bm.Len()
	m := ^Word(0) >> (WordBits - n_bits)
	b := *bm

	// Check first word.
	if i0 < l {
		v = (b[i0] >> i1) & m
	}

	// Check for overlap into next word.
	i0++
	if i1+n_bits > WordBits && i0 < l {
		r := WordBits - i1
		n_bits -= r
		v |= Word((b[i0] & (1<<n_bits - 1)) << r)
	}
	return
}

func (b Bitmap) GetMultiple(x, n_bits uint) (v Word) {
	if b.isInline() {
		v = (Word(b) >> x) & (1<<n_bits - 1)
	} else {
		v = Bitmaps.bitmaps[^b].GetMultiple(x, n_bits)
	}
	return
}

// Give bits I through I + N_BITS to new value; return old value.
func (bm *WordVec) SetMultiple(x, n_bits uint, new_value Word) (old_value Word) {
	i0, i1 := x/WordBits, x%WordBits
	m := ^Word(0) >> (WordBits - n_bits)
	v := new_value & m

	bm.Validate((x + n_bits - 1) / WordBits)
	l := bm.Len()
	b := *bm

	// Insert into first word.
	t := b[i0]
	old_value |= (t >> i1) & m
	t &^= m << i1
	t |= v << i1
	b[i0] = t

	// Insert into second word.
	i0++
	if i1+n_bits > WordBits && i0 < l {
		r := WordBits - i1
		v >>= r
		m >>= r
		n_bits -= r

		u := b[i0]
		old_value |= (u & m) << r
		u &^= m
		u |= v
		b[i0] = u
	}
	return
}

func (b Bitmap) SetMultiple(x, n_bits uint, new_value Word) (r Bitmap, old_value Word) {
	if b.isInline() {
		if x+n_bits <= WordBits-1 {
			m := Word(1)<<n_bits - 1
			old_value = (Word(b) >> x) & m
			r = Bitmap((Word(b) &^ (m << x)) | (new_value << x))
			return
		}
		b = Bitmaps.toMem(b)
	}
	r = b
	old_value = Bitmaps.bitmaps[^r].SetMultiple(x, n_bits, new_value)
	return
}

func (bm *WordVec) Alloc(nBits uint) {
	if nBits > 0 {
		i, _ := bitmapIndex(nBits - 1)
		bm.Validate(i)
	}
}
