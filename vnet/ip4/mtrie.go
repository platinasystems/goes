// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"github.com/platinasystems/go/vnet/ip"
)

type leaf uint32

const (
	emptyLeaf    leaf = leaf(1 + 2*ip.AdjMiss)
	rootPlyIndex uint = 0
)

func (l leaf) isTerminal() bool    { return l&1 != 0 }
func (l leaf) ResultIndex() ip.Adj { return ip.Adj(l >> 1) }
func setResult(i ip.Adj) leaf      { return leaf(1 + 2*i) }
func (l *leaf) setResult(i ip.Adj) { *l = setResult(i) }
func (l leaf) isPly() bool         { return !l.isTerminal() }
func (l leaf) plyIndex() uint      { return uint(l >> 1) }
func setPlyIndex(i uint) leaf      { return leaf(0 + 2*i) }
func (l *leaf) setPlyIndex(i uint) { *l = setPlyIndex(i) }

const plyLeaves = 1 << 8

type ply struct {
	leaves [plyLeaves]leaf

	// Prefix length of leaves.
	lens [plyLeaves]uint8

	// Number of non-empty leaves.
	nNonEmpty int

	poolIndex uint
}

//go:generate gentemplate -d Package=ip4 -id ply -d PoolType=plyPool -d Type=ply -d Data=plys github.com/platinasystems/go/elib/pool.tmpl

type mtrie struct {
	// Pool of plies.  Index zero is root ply.
	plyPool

	// Special case leaf for default route 0.0.0.0/0.
	// This is to avoid having to paint default leaf in all plys of trie.
	defaultLeaf leaf
}

func (m *mtrie) LookupStep(l leaf, dst byte) (lʹ leaf) {
	pi := uint(0)
	it := l.isTerminal()
	if !it {
		pi = l.plyIndex()
	}
	lʹ = m.plys[pi].leaves[dst]
	if it {
		lʹ = l
	}
	return
}

func (p *ply) init(l leaf, n uint8) {
	p.nNonEmpty = 0
	if l != emptyLeaf {
		p.nNonEmpty = len(p.leaves)
	}
	for i := 0; i < plyLeaves; i += 4 {
		p.lens[i+0] = n
		p.lens[i+1] = n
		p.lens[i+2] = n
		p.lens[i+3] = n
		p.leaves[i+0] = l
		p.leaves[i+1] = l
		p.leaves[i+2] = l
		p.leaves[i+3] = l
	}
}

func (m *mtrie) newPly(l leaf, n uint8) (lʹ leaf, ply *ply) {
	pi := m.plyPool.GetIndex()
	ply = &m.plys[pi]
	ply.poolIndex = pi
	ply.init(l, n)
	lʹ = setPlyIndex(pi)
	return
}

func (m *mtrie) plyForLeaf(l leaf) *ply { return &m.plys[l.plyIndex()] }

func (m *mtrie) freePly(p *ply) {
	isRoot := p.poolIndex == 0
	for _, l := range p.leaves {
		if !l.isTerminal() {
			m.freePly(m.plyForLeaf(l))
		}
	}
	if isRoot {
		p.init(emptyLeaf, 0)
	} else {
		m.plyPool.PutIndex(p.poolIndex)
	}
}

func (m *mtrie) Free() { m.freePly(&m.plys[0]) }

func (m *mtrie) lookup(dst *Address) (a ip.Adj) {
	a = ip.AdjMiss
	if len(m.plys) == 0 {
		return
	}
	p := &m.plys[0]
	for i := range dst {
		l := p.leaves[dst[i]]
		if l.isTerminal() {
			a = l.ResultIndex()
			return
		}
		p = m.plyForLeaf(l)
	}
	panic("no terminal leaf found")
}

func (m *mtrie) setPlyWithMoreSpecificLeaf(p *ply, l leaf, n uint8) {
	for i, pl := range p.leaves {
		if !pl.isTerminal() {
			m.setPlyWithMoreSpecificLeaf(m.plyForLeaf(pl), l, n)
		} else if n >= p.lens[i] {
			p.leaves[i] = l
			p.lens[i] = n
			if pl != emptyLeaf {
				p.nNonEmpty++
			}
		}
	}
}

func (p *ply) replaceLeaf(new, old leaf, i uint8) {
	p.leaves[i] = new
	if old != emptyLeaf {
		p.nNonEmpty++
	}
}

type addDelLeaf struct {
	key    Address
	keyLen uint8
	result ip.Adj
}

func (s *addDelLeaf) setLeafHelper(m *mtrie, oldPlyIndex, keyByteIndex uint) {
	nBits := int(s.keyLen) - 8*int(keyByteIndex+1)
	k := s.key[keyByteIndex]
	oldPly := &m.plys[oldPlyIndex]

	// Number of bits next plies <= 0 => insert leaves this ply.
	if nBits <= 0 {
		nBits = -nBits
		for i := k; i < k+1<<uint(nBits); i++ {
			oldLeaf := oldPly.leaves[i]
			oldTerm := oldLeaf.isTerminal()

			// Is leaf to be inserted more specific?
			if s.keyLen >= oldPly.lens[i] {
				newLeaf := setResult(s.result)
				if oldTerm {
					oldPly.lens[i] = s.keyLen
					oldPly.replaceLeaf(newLeaf, oldLeaf, i)
				} else {
					// Existing leaf points to another ply.
					// We need to place new_leaf into all more specific slots.
					newPly := m.plyForLeaf(oldLeaf)
					m.setPlyWithMoreSpecificLeaf(newPly, newLeaf, s.keyLen)
				}
			} else if !oldTerm {
				s.setLeafHelper(m, oldLeaf.plyIndex(), keyByteIndex+1)
			}
		}
	} else {
		oldLeaf := oldPly.leaves[k]
		oldTerm := oldLeaf.isTerminal()
		var newPly *ply
		if !oldTerm {
			newPly = m.plyForLeaf(oldLeaf)
		} else {
			var newLeaf leaf
			newLeaf, newPly = m.newPly(oldLeaf, oldPly.lens[k])
			// Refetch since newPly may move pool.
			oldPly = &m.plys[oldPlyIndex]
			oldPly.leaves[k] = newLeaf
			oldPly.lens[k] = 0
			if oldLeaf != emptyLeaf {
				oldPly.nNonEmpty--
			}
			// Account for the ply we just created.
			oldPly.nNonEmpty++
		}
		s.setLeafHelper(m, newPly.poolIndex, keyByteIndex+1)
	}
}

func (s *addDelLeaf) unsetLeafHelper(m *mtrie, oldPlyIndex, keyByteIndex uint) (oldPlyWasDeleted bool) {
	k := s.key[keyByteIndex]
	nBits := int(s.keyLen) - 8*int(keyByteIndex+1)
	if nBits <= 0 {
		nBits = -nBits
		k &^= 1<<uint(nBits) - 1
		if nBits > 8 {
			nBits = 8
		}
	}
	delLeaf := setResult(s.result)
	oldPly := &m.plys[oldPlyIndex]
	for i := k; i < k+1<<uint(nBits); i++ {
		oldLeaf := oldPly.leaves[i]
		oldTerm := oldLeaf.isTerminal()
		if oldLeaf == delLeaf ||
			(!oldTerm && s.unsetLeafHelper(m, oldLeaf.plyIndex(), keyByteIndex+1)) {
			oldPly.leaves[i] = emptyLeaf
			oldPly.lens[i] = 0
			oldPly.nNonEmpty--
			oldPlyWasDeleted = oldPly.nNonEmpty == 0 && keyByteIndex > 0
			if oldPlyWasDeleted {
				m.plyPool.PutIndex(oldPly.poolIndex)
				// Nothing more to do.
				break
			}
		}
	}

	return
}

func (s *addDelLeaf) set(m *mtrie)        { s.setLeafHelper(m, rootPlyIndex, 0) }
func (s *addDelLeaf) unset(m *mtrie) bool { return s.unsetLeafHelper(m, rootPlyIndex, 0) }

func (l *leaf) remap(from, to ip.Adj) (remapEmpty int) {
	if l.isTerminal() {
		if adj := l.ResultIndex(); adj == from {
			if to == ip.AdjNil {
				remapEmpty = 1
				*l = emptyLeaf
			} else {
				l.setResult(to)
			}
		}
	}
	return
}

func (p *ply) remap(from, to ip.Adj) {
	nRemapEmpty := 0
	for i := range p.leaves {
		nRemapEmpty += p.leaves[i].remap(from, to)
	}
	p.nNonEmpty -= nRemapEmpty
}

func (t *mtrie) remapAdjacency(from, to ip.Adj) {
	for i := uint(0); i < t.plyPool.Len(); i++ {
		if !t.plyPool.IsFree(i) {
			t.plyPool.plys[i].remap(from, to)
		}
	}
}

func (m *mtrie) init() {
	m.defaultLeaf = emptyLeaf
	// Make root ply.
	l, _ := m.newPly(emptyLeaf, 0)
	if l.plyIndex() != 0 {
		panic("root ply must be index 0")
	}
}

func (m *mtrie) reset() {
	m.plyPool.Reset()
	m.defaultLeaf = emptyLeaf
}
