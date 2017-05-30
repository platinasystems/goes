// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"fmt"
	"sort"
)

type MaskedStringer interface {
	MaskedString(mask MaskedStringer) string
}

type masked_string_pair struct{ v, m MaskedStringer }
type MaskedStrings struct {
	m map[string]masked_string_pair
}

func (x *MaskedStrings) Add(key string, v, m MaskedStringer) {
	if x.m == nil {
		x.m = make(map[string]masked_string_pair)
	}
	x.m[key] = masked_string_pair{v: v, m: m}
}

func (x *MaskedStrings) String() (s string) {
	type t struct{ k, v string }
	var ts []t
	for k, v := range x.m {
		ts = append(ts, t{k: k, v: v.v.MaskedString(v.m)})
	}
	sort.Slice(ts, func(i, j int) bool { return ts[i].k < ts[j].k })
	for i := range ts {
		t := &ts[i]
		if s != "" {
			s += ", "
		}
		s += t.k + ": " + t.v
	}
	return
}

type Bool bool

func (v Bool) MaskedString(r MaskedStringer) (s string) {
	m := r.(Bool)
	s = fmt.Sprintf("%v", v)
	if m != true {
		s += fmt.Sprintf("/%v", m)
	}
	return
}

func (v Uint8) MaskedString(r MaskedStringer) string {
	m := r.(Uint8)
	return fmt.Sprintf("0x%x/%x", v, m)
}
func (v Uint16) MaskedString(r MaskedStringer) string {
	m := r.(Uint16)
	return fmt.Sprintf("0x%x/%x", v.ToHost(), m.ToHost())
}
func (v Uint32) MaskedString(r MaskedStringer) string {
	m := r.(Uint32)
	return fmt.Sprintf("0x%x/%x", v.ToHost(), m.ToHost())
}
func (v Uint64) MaskedString(r MaskedStringer) string {
	m := r.(Uint64)
	return fmt.Sprintf("0x%x/%x", v.ToHost(), m.ToHost())
}

// As above but host byte order.
type HostUint16 uint16
type HostUint32 uint32
type HostUint64 uint64

func (v HostUint16) MaskedString(r MaskedStringer) string {
	m := r.(HostUint16)
	return fmt.Sprintf("0x%x/%x", v, m)
}
func (v HostUint32) MaskedString(r MaskedStringer) string {
	m := r.(HostUint32)
	return fmt.Sprintf("0x%x/%x", v, m)
}
func (v HostUint64) MaskedString(r MaskedStringer) string {
	m := r.(HostUint64)
	return fmt.Sprintf("0x%x/%x", v, m)
}
