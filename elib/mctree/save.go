// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mctree

import (
	"github.com/platinasystems/go/elib"

	"encoding/gob"
	"os"
)

type treeSave struct {
	IsSplit   elib.Uint64Vec
	SplitBits []uint8
	nNodes    uint
	nSplit    uint
}

func (t *tree) save_helper(s *treeSave, ni node_index) {
	n := t.get_node(ni)
	if n.is_split() {
		i0, i1 := s.nNodes/64, s.nNodes%64
		s.IsSplit.Validate(i0)
		s.IsSplit[i0] |= uint64(1) << i1
		s.SplitBits = append(s.SplitBits, uint8(n.split_bit))
		s.nNodes++
		t.save_helper(s, n.sub_nodes[0])
		t.save_helper(s, n.sub_nodes[1])
	} else {
		s.nNodes++
	}
}

func (t *tree) restore_helper(s *treeSave, ni node_index) {
	n := t.get_node(ni)
	i0, i1 := s.nNodes/64, uint64(1)<<(s.nNodes%64)
	if i0 < s.IsSplit.Len() && s.IsSplit[i0]&i1 != 0 {
		n.split_bit = uint(s.SplitBits[s.nSplit])
		s.nSplit++
		s.nNodes++
		n, _, _ = n.alloc_subs(t.Main, n.split_bit)
		t.restore_helper(s, n.sub_nodes[0])
		t.restore_helper(s, n.sub_nodes[1])
	} else {
		s.nNodes++
		n.pair_offsets.get(t.Main, 0)
	}
}

func (t *tree) save(path string) (err error) {
	var f *os.File
	if f, err = os.Create(path); err != nil {
		return
	}
	defer f.Close()
	e := gob.NewEncoder(f)
	var s treeSave
	t.save_helper(&s, t.root_node_index)
	err = e.Encode(&s)
	return
}

func (t *tree) restore(path string) (err error) {
	f, err := os.Open(path)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	d := gob.NewDecoder(f)
	var s treeSave
	if err = d.Decode(&s); err != nil {
		return
	}
	t.restore_helper(&s, t.root_node_index)
	return
}

func (m *Main) Save(path string) (err error) {
	t := m.get_tree()
	err = t.save(path)
	return
}

func (m *Main) Restore(path string) (err error) {
	t := m.get_tree()
	t.Main = m
	err = t.restore(path)
	m.tree_sequence = 1
	return
}
