// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mctree

import (
	"github.com/platinasystems/go/elib"

	"unsafe"
)

const word_bits = 32

type word uint32

// Keys are made up of multiple pairs each having a mask and a value.
// A keys k matches k & mask == value for all pairs.
type Pair struct{ Mask, Value word }

func (a *Pair) equal(b *Pair) bool { return a.Value == b.Value && a.Mask == b.Mask }
func (a *Pair) poison() {
	a.Value = ^word(0)
	a.Mask = 0
}
func (a *Pair) Set(value, mask uint) {
	a.Value = word(value)
	a.Mask = word(mask)
}

//go:generate gentemplate -d Package=mctree -id Pair -d VecType=pair_vec -d Type=Pair github.com/platinasystems/go/elib/vec.tmpl

func (p pair_vec) poison() {
	for i := range p {
		p[i].poison()
	}
}

func (p pair_vec) get_pairs_for_index(i, n uint) pair_vec { return pair_vec(p[i*n : (i+1)*n]) }

type pair_hash struct {
	hash elib.Hash

	pair_offset_by_hash_index []pair_offset

	n_pairs_per_key uint

	pairs_pool elib.Pool

	pairs pair_vec
}

func (h *pair_hash) get_pairs_for_offset(o pair_offset) pair_vec {
	return pair_vec(h.pairs[o : o+pair_offset(h.n_pairs_per_key)])
}

func (m *pair_hash) get_pairs_for_hash_index(hash_index uint) pair_vec {
	return m.get_pairs_for_offset(m.pair_offset_by_hash_index[hash_index])
}

func (m *pair_hash) HashIndex(s *elib.HashState, i uint) {
	p := m.get_pairs_for_hash_index(i)
	p.HashKey(s)
}

func (m *pair_hash) HashResize(newCap uint, rs []elib.HashResizeCopy) {
	new := make([]pair_offset, newCap, newCap)
	old := m.pair_offset_by_hash_index
	for i := range new {
		new[i] = pair_offset_invalid
	}
	for i := range rs {
		new[rs[i].Dst] = old[rs[i].Src]
	}
	m.pair_offset_by_hash_index = new
}

func (p pair_vec) HashKey(s *elib.HashState) {
	switch len(p) {
	case 1:
		s.HashUint64(uint64(p[0].Value), uint64(p[0].Mask), 0, 0)
	case 2:
		s.HashUint64(uint64(p[0].Value), uint64(p[0].Mask), uint64(p[1].Value), uint64(p[1].Mask))
	default:
		s.HashPointer(unsafe.Pointer(&p[0]), uintptr(len(p))*unsafe.Sizeof(p[0]))
	}
}

func (a pair_vec) equal(b pair_vec) bool {
	if la, lb := a.Len(), b.Len(); la != lb {
		return false
	} else {
		for i := uint(0); i < la; i++ {
			if !a[i].equal(&b[i]) {
				return false
			}
		}
	}
	return true
}
func (a pair_vec) HashKeyEqual(h elib.Hasher, i uint) bool {
	b := pair_vec(h.(*pair_hash).get_pairs_for_hash_index(i))
	return a.equal(b)
}

type pair_offset uint32

const pair_offset_invalid pair_offset = ^pair_offset(0)

func (m *Main) pair_offset(i uint) pair_offset { return pair_offset(i * m.n_pairs_per_key) }

//go:generate gentemplate -d Package=mctree -id pair_offset -d VecType=pair_offset_vec -d Type=pair_offset github.com/platinasystems/go/elib/vec.tmpl

func (m *pair_hash) is_masked(o pair_offset, bit uint) bool {
	b0, b1 := index(bit)
	return m.pairs[o+pair_offset(b0)].Mask&b1 != 0
}

func (h *pair_hash) init(n, cap uint) {
	if cap < 32 {
		cap = 32
	}
	h.hash.Init(h, cap)
	h.n_pairs_per_key = n
}

func (h *pair_hash) set(p []Pair) (o pair_offset, exists bool) {
	if i, ok := h.hash.Set(pair_vec(p)); ok {
		exists = true
		o = h.pair_offset_by_hash_index[i]
	} else {
		pi := h.pairs_pool.GetIndex(h.pairs.Len())
		o = pair_offset(pi * h.n_pairs_per_key)
		h.pairs.Validate(uint(o) + h.n_pairs_per_key - 1)
		h.pair_offset_by_hash_index[i] = o
	}
	copy(h.get_pairs_for_offset(o), p)
	return
}

func (h *pair_hash) unset(p []Pair) (o pair_offset, ok bool) {
	var i uint
	if i, ok = h.hash.Unset(pair_vec(p)); ok {
		if elib.Debug {
			v := h.get_pairs_for_hash_index(i)
			v.poison()
		}
		o = h.pair_offset_by_hash_index[i]
		h.pairs_pool.PutIndex(uint(o) / h.n_pairs_per_key)
		h.pair_offset_by_hash_index[i] = pair_offset_invalid
	}
	return
}

func (h *pair_hash) get(p []Pair) (o pair_offset, ok bool) {
	var i uint
	if i, ok = h.hash.Get(pair_vec(p)); ok {
		o = h.pair_offset_by_hash_index[i]
	}
	return
}
