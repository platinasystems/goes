package elib

// Underlying machine word, typically 32 or 64 bits.
type Word uintptr

//go:generate gentemplate -d Package=elib -id Word -d VecType=WordVec -d Type=Word vec.tmpl

const (
	// Compute the size _S of a Word in bytes.
	_m           = ^Word(0)
	Log2WordBits = 3 + (_m>>8&1 + _m>>16&1 + _m>>32&1)
	WordBits     = 1 << Log2WordBits
	_n           = ^uint(0)
	Log2IntBits  = 3 + (_n>>8&1 + _n>>16&1 + _n>>32&1)
	IntBits      = 1 << Log2IntBits
)

// Helper to index word sized bitmaps.
func (i Word) BitmapIndex() (Word, Word) {
	return i / WordBits, 1 << (i % WordBits)
}

func NSetBits(y Word) uint {
	x := uint64(y)
	x = (x & 0x5555555555555555) + ((x & 0xAAAAAAAAAAAAAAAA) >> 1)
	x = (x & 0x3333333333333333) + ((x & 0xCCCCCCCCCCCCCCCC) >> 2)
	x = (x & 0x0F0F0F0F0F0F0F0F) + ((x & 0xF0F0F0F0F0F0F0F0) >> 4)
	x *= 0x0101010101010101
	return uint(((x >> 56) & 0xFF))
}
func (x Word) NSetBits() uint { return NSetBits(x) }

func NLeadingZeros(x Word) uint {
	n := uint(WordBits)
	var y Word
	if WordBits > 32 {
		y = x >> 32
		if y != 0 {
			n -= 32
			x = y
		}
	}
	y = x >> 16
	if y != 0 {
		n -= 16
		x = y
	}
	y = x >> 8
	if y != 0 {
		n = n - 8
		x = y
	}

	y = x >> 4
	if y != 0 {
		n = n - 4
		x = y
	}
	y = x >> 2
	if y != 0 {
		n = n - 2
		x = y
	}
	y = x >> 1
	if y != 0 {
		return n - 2
	}
	return n - uint(x)
}
func (x Word) NLeadingZeros() uint { return NLeadingZeros(x) }

// firstSet gives 2^f where f is the lowest 1 bit in x
func FirstSet(x Word) Word    { return x & -x }
func (x Word) FirstSet() Word { return FirstSet(x) }

// isPow2 true for x a power of 2
func IsPow2(x Word) bool    { return 0 == x&(x-1) }
func (x Word) IsPow2() bool { return IsPow2(x) }

// roundPow2 rounds x to next power of two >= x
func RoundPow2(x, p Word) Word       { return (x + p - 1) &^ (p - 1) }
func (x Word) RoundPow2(p Word) Word { return RoundPow2(x, p) }

func MinLog2(x Word) uint    { return WordBits - 1 - NLeadingZeros(x) }
func (x Word) MinLog2() uint { return MinLog2(x) }

func MaxLog2(x Word) uint {
	l := MinLog2(x)
	if x > Word(1)<<l {
		l++
	}
	return l
}
func (x Word) MaxLog2() uint { return MaxLog2(x) }

func MaxPow2(x Word) Word {
	z := MinLog2(x)
	y := Word(1) << z
	if x > y {
		y += y
	}
	return y
}
func (x Word) MaxPow2() Word { return MaxPow2(x) }

func NextSet(x Word) (v Word, i int) {
	f := x & -x
	v = x ^ f
	i = int(WordBits - 1 - NLeadingZeros(f))
	return
}
func (x Word) NextSet() (Word, int) { return NextSet(x) }

func (x Word) ForeachSetBit(fn func(i uint)) {
	for x != 0 {
		f := x.FirstSet()
		i := f.MinLog2()
		x ^= f
		fn(i)
	}
}

/* Bit compress: based on "Hacker's Delight" code from GLS. */
const (
	log2Uint32Bits = 5
	log2Uint64Bits = 6
)

type BitCompressUint64 struct {
	state [1 + log2Uint64Bits]uint64
}

func (s *BitCompressUint64) SetMask(mask uint64) {
	m := ^mask
	zm := mask
	s.state[0] = mask
	for i := uint(1); i < uint(len(s.state)); i++ {
		q := m
		m ^= m << 1
		m ^= m << 2
		m ^= m << 4
		m ^= m << 8
		m ^= m << 16
		m ^= m << 32
		n := (m << 1) & zm
		s.state[i] = n
		m = q &^ m
		q = zm & n
		zm = zm ^ q ^ (q >> (1 << (i - 1)))
	}
}

func (s *BitCompressUint64) Mask() uint64 { return s.state[0] }

func (s *BitCompressUint64) Compress(x uint64) (r uint64) {
	r = x & s.state[0]
	q := r & s.state[1]
	r ^= q ^ (q >> 1)
	q = r & s.state[2]
	r ^= q ^ (q >> 2)
	q = r & s.state[3]
	r ^= q ^ (q >> 4)
	q = r & s.state[4]
	r ^= q ^ (q >> 8)
	q = r & s.state[5]
	r ^= q ^ (q >> 16)
	q = r & s.state[6]
	r ^= q ^ (q >> 32)
	return r
}
