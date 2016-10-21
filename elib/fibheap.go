// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"fmt"
)

type fibNode struct {
	// Index of parent or MaxIndex if no parent.
	// Having no parent means this node is a root node.
	sup Index

	// Links to doubly linked list of siblings
	next, prev Index

	// Index of first child node or MaxIndex if no children.
	sub Index

	// Number of children.
	nSub uint16

	// Set when at least one child has been cut since this node was made child of another node.
	// Roots are never marked.
	isMarked bool
}

//go:generate gentemplate -d Package=elib -id fibNode -d VecType=fibNodeVec -d Type=fibNode vec.tmpl

const (
	fibRootIndex = MaxIndex - 1
	MaxNSub      = 32
)

type Ordered interface {
	// Returns negative, zero, positive if element i is less, equal, greater than element j
	Compare(i, j int) int
}

type FibHeap struct {
	root  fibNode
	nodes fibNodeVec

	// Cached min index in heap.
	minIndex Index
	minValid bool
}

func (f *FibHeap) node(ni Index) *fibNode {
	if ni == fibRootIndex {
		return &f.root
	} else {
		return &f.nodes[ni]
	}
}

func (f *FibHeap) linkAfter(pi, xi Index) {
	p := f.node(pi)
	ni := p.next
	n := f.node(ni)
	x := f.node(xi)
	p.next = xi
	x.prev, x.next = pi, ni
	n.prev = xi
}

func (f *FibHeap) addRoot(xi Index) {
	f.linkAfter(fibRootIndex, xi)
}

func (f *FibHeap) unlink(xi Index) {
	x := &f.nodes[xi]
	p := f.node(x.prev)
	n := f.node(x.next)
	p.next = x.next
	n.prev = x.prev
}

// Add a new index to heap.
func (f *FibHeap) Add(xi uint) {
	if len(f.nodes) == 0 {
		f.root.next, f.root.prev = fibRootIndex, fibRootIndex
		f.root.sup = MaxIndex
		f.root.sub = MaxIndex
	}
	f.minValid = false
	f.nodes.Validate(uint(xi))
	x := &f.nodes[xi]
	x.sup = MaxIndex
	x.sub = MaxIndex
	x.nSub = 0
	x.isMarked = false
	f.addRoot(Index(xi))
}

func (f *FibHeap) cutChildren(xi Index) {
	x := &f.nodes[xi]
	bi := x.sub
	if bi == MaxIndex {
		return
	}
	ci := bi
	for {
		c := &f.nodes[ci]
		ni := c.next
		c.sup = MaxIndex
		f.addRoot(ci)
		if ni == bi {
			break
		}
		ci = ni
	}
}

// Del deletes given index from heap.
func (f *FibHeap) Del(i uint) {
	xi := Index(i)
	f.unlink(xi)
	f.cutChildren(xi)

	x := &f.nodes[xi]
	supi := x.sup
	f.minValid = f.minValid && xi != f.minIndex
	if supi == MaxIndex {
		return
	}

	// Adjust parent for deletion of child.
	ni := x.next
	for {
		sup := &f.nodes[supi]
		sup.nSub -= 1
		wasMarked := sup.isMarked
		sup.isMarked = true
		sup.sub = ni
		if sup.nSub == 0 {
			sup.sub = MaxIndex
		}
		sup2i := sup.sup
		if !wasMarked || sup2i == MaxIndex {
			break
		}
		ni = sup.next
		f.unlink(supi)
		sup.sup = MaxIndex
		f.addRoot(supi)
		supi = sup2i
	}
}

// Update node when data ordering changes (lower key)
func (f *FibHeap) Update(xi uint) {
	f.Del(xi)
	f.Add(xi)
}

func (f *FibHeap) Min(data Ordered) (minu uint, valid bool) {
	minu = uint(f.minIndex)
	valid = f.minValid
	if valid || len(f.nodes) == 0 {
		return
	}

	// Degrees seen so far
	var deg [MaxNSub]Index
	// Bitmap of valid degrees, initially zero
	var degValid Word

	ri := f.root.next
	r := f.node(ri)
	ni := r.next

	for ri != fibRootIndex {
		r = &f.nodes[ri]
		n := f.node(ni)
		ns := r.nSub

		m := Word(1) << ns
		nsDegrees := 0 != degValid&m
		degValid ^= m
		if !nsDegrees {
			deg[ns] = ri
			ri = ni
			ni = n.next
		} else {
			ri0 := deg[ns]
			if data.Compare(int(ri0), int(ri)) <= 0 {
				ri, ri0 = ri0, ri
			}
			f.unlink(ri0)
			r0 := &f.nodes[ri0]
			r0.isMarked = false
			r0.sup = ri

			r = &f.nodes[ri]
			if r.sub != MaxIndex {
				r.nSub += 1
				f.linkAfter(r.sub, ri0)
			} else {
				r.sub = ri0
				r.nSub = 1
				r.isMarked = false
				r0.next = ri0
				r0.prev = ri0
			}
		}
	}

	min := MaxIndex
	for degValid != 0 {
		var ns int
		degValid, ns = NextSet(degValid)
		ri := deg[ns]
		if min == MaxIndex || data.Compare(int(ri), int(min)) < 0 {
			min = ri
		}
	}

	valid = min != MaxIndex
	f.minValid = valid
	f.minIndex = min
	minu = uint(min)

	return
}

func (n *fibNode) reloc(i int) {
	d := Index(i)
	if n.sup != MaxIndex {
		n.sup += d
	}
	if n.sub != MaxIndex {
		n.sub += d
	}
	if n.next != fibRootIndex {
		n.next += d
	}
	if n.prev != fibRootIndex {
		n.prev += d
	}
}

func (f *FibHeap) Merge(g *FibHeap) {
	l := len(f.nodes)
	f.nodes.Resize(uint(len(g.nodes)))
	copy(f.nodes[l:], g.nodes)
	for i := l; i < len(f.nodes); i++ {
		f.nodes[i].reloc(l)
	}
	r := g.root
	r.reloc(l)
	for ri := r.next; ri != fibRootIndex; {
		r := &f.nodes[ri]
		f.Add(uint(ri))
		ri = r.next
	}
}

func (f *FibHeap) String() string {
	return fmt.Sprintf("%d elts", len(f.nodes))
}
