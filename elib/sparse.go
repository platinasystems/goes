package elib

import (
	"fmt"
)

// Sparse arrays map sparse indices into dense indices.
type Sparse struct {
	// Bitmap of valid sparse indices.
	valid BitmapVec

	// Count of number of dense indices with smaller indices.
	// COUNT[I] is number of dense indices with sparse INDEX < bitmapBits*I.
	count Int32Vec
}

func (s *Sparse) Get(sparse Index) (dense Index, valid bool) {
	i, m := bitmapIndex(uint(sparse))
	if i >= uint(len(s.valid)) {
		return
	}
	v := s.valid[i]
	valid = v&m != 0
	dense = MaxIndex
	if valid {
		dense = Index(s.count[i]) + Index(NSetBits(Word(v&(m-1))))
	}
	return
}

// TODO optimize sse/avx instructions on x86
func (s *Sparse) inc(start uint, dx int32) {
	for i := start + 1; i < uint(len(s.count)); i++ {
		s.count[i] += dx
	}
}

func (s *Sparse) eltsBefore(i int) Index {
	return Index(s.count[i]) + Index(NSetBits(Word(s.valid[i])))
}

func (s *Sparse) elts() Index {
	return s.eltsBefore(len(s.count) - 1)
}

func (s *Sparse) Set(sparse Index) (dense Index) {
	i, m := bitmapIndex(uint(sparse))

	l := len(s.valid)
	s.valid.Validate(i)
	s.count.Validate(i)

	if l > 0 && len(s.valid) > l {
		if e := s.eltsBefore(l - 1); e != 0 {
			for j := l; j < len(s.valid); j++ {
				s.count[j] = int32(e)
			}
		}
	}

	v := s.valid[i]
	if v&m != 0 {
		return
	}

	s.valid[i] = v | m
	dense = Index(s.count[i]) + Index(NSetBits(Word(v&(m-1))))
	s.inc(i, +1)
	return
}

func (s *Sparse) Unset(sparse Index) (valid bool) {
	i, m := bitmapIndex(uint(sparse))
	v := s.valid[i]
	valid = v&m != 0
	if valid {
		s.valid[i] = v ^ m
		s.inc(i, -1)
	}
	return
}

func (s *Sparse) String() string {
	return fmt.Sprintf("%d elts", s.elts())
}
