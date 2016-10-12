// Package elib is a collection of data structures: bitmaps, pools, heaps, fifos.
package elib

import (
	"fmt"
)

// Number of bits per bitmap element
const bitmapBits = WordBits

// Bitmaps of <= 63 are stored as positive int64s.
// Negative int64s are indices into pool of all known bitmaps
type Bitmap Word

//go:generate gentemplate -d Package=elib -id Bitmap -d VecType=BitmapVec -d Type=Bitmap vec.tmpl
//go:generate gentemplate -d Package=elib -id Bitmaps -d VecType=BitmapsVec -d Type=[]Bitmap vec.tmpl

func (b Bitmap) isPoolIndex() bool {
	if b>>(WordBits-1) != 0 { // sign bit
		return true
	}
	return false
}

func (b Bitmap) isInline() bool {
	return !b.isPoolIndex()
}

// bothMemory is true iff both arguments are direct non-memory bitmaps
func bothInline(b, c Bitmap) bool {
	return (b | c) >= 0
}

func firstSet(x Bitmap) Bitmap {
	return Bitmap(FirstSet(Word(x)))
}

func minLog2(x Bitmap) uint {
	return uint(MinLog2(Word(x)))
}

// index gives word index and mask for given bit index
func bitmapIndex(x uint) (i uint, m Bitmap) {
	i = x / bitmapBits
	m = Bitmap(1) << (x % bitmapBits)
	return
}

func bitmapGet(b []Bitmap, x uint) bool {
	i, m := bitmapIndex(x)
	return b[i]&m != 0
}

func bitmapSet(b []Bitmap, x uint) (old bool) {
	i, m := bitmapIndex(x)
	v := b[i]
	old = v&m != 0
	b[i] = v | m
	return
}

func bitmapUnset(b []Bitmap, x uint) (old bool) {
	i, m := bitmapIndex(x)
	v := b[i]
	old = v&m != 0
	b[i] = v &^ m
	return
}

func bitmapMake(nBits uint) []Bitmap {
	n, _ := bitmapIndex(nBits)
	return make([]Bitmap, n)
}

//go:generate gentemplate -d Package=elib -id Bitmap -d PoolType=BitmapPool -d Type=BitmapVec -d Data=bitmaps pool.tmpl

var Bitmaps = &BitmapPool{}

func (p *BitmapPool) new() uint {
	bi := p.GetIndex()
	p.Validate(bi)
	return bi
}

func (p *BitmapPool) toMem(b Bitmap) Bitmap {
	bi := p.new()
	if len(p.bitmaps[bi]) == 0 {
		p.bitmaps[bi] = append(p.bitmaps[bi], b)
	} else {
		p.bitmaps[bi] = p.bitmaps[bi][:1]
		p.bitmaps[bi][0] = b
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
		if x < bitmapBits-1 {
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
		if x < bitmapBits-1 {
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
		if x < bitmapBits-1 {
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
	var cs []Bitmap
	if c.isPoolIndex() {
		cs = p.bitmaps[^c]
	} else {
		cs = []Bitmap{c}
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
	if l == 1 && !p.bitmaps[bi][0].isInline() {
		return
	}
	r = 0
	if l == 1 {
		r = p.bitmaps[bi][0]
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
	var cs []Bitmap
	if c.isPoolIndex() {
		cs = p.bitmaps[^c]
	} else {
		cs = []Bitmap{c}
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
		if x < bitmapBits-1 {
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
	var s []Bitmap
	if !b.isPoolIndex() {
		s = []Bitmap{b}
	} else {
		s = p.bitmaps[^b]
	}
	for i := range s {
		x := s[i]
		for x != 0 {
			f := firstSet(x)
			l := minLog2(f)
			fn(uint(i*bitmapBits) + l)
			x ^= f
		}
	}
}

func (b Bitmap) ForeachSetBit(fn func(uint)) { Bitmaps.ForeachSetBit(b, fn) }

func (p *BitmapPool) Next(b Bitmap, px *uint) (ok bool) {
	x := *px
	i := uint(0)
	m := ^Bitmap(0)
	if x != ^uint(0) {
		i, m = bitmapIndex(x)
		m = ^(m | (m - 1))
	}
	var s []Bitmap
	if !b.isPoolIndex() {
		s = []Bitmap{b}
	} else {
		s = p.bitmaps[^b]
	}

	m &= s[i]
	if m != 0 {
		*px = i*bitmapBits + minLog2(firstSet(m))
		ok = true
		return
	} else {
		for i++; i < uint(len(s)); i++ {
			m = s[i]
			if m != 0 {
				*px = i*bitmapBits + minLog2(firstSet(m))
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
	var x []Bitmap
	if !b.isPoolIndex() {
		x = []Bitmap{b}
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

func (bm BitmapVec) Get(x uint) (v bool) {
	i, m := bitmapIndex(x)
	if i < uint(len(bm)) {
		v = bm[i]&m != 0
	}
	return
}

func (bm BitmapVec) Set(x uint, v bool) (oldValue bool) {
	i, m := bitmapIndex(x)
	b := bm[i]
	oldValue = b&m != 0
	b |= m
	bm[i] = b
	return
}

func (bm BitmapVec) Unset(x uint) (v bool) {
	i, m := bitmapIndex(x)
	b := bm[i]
	v = b&m != 0
	b &^= m
	bm[i] = b
	return
}

func (bm *BitmapVec) Alloc(nBits uint) {
	if nBits > 0 {
		i, _ := bitmapIndex(nBits - 1)
		bm.Validate(i)
	}
}
