// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"
)

const (
	punt_2tag_next_error uint = iota
	punt_2tag_next_punt
)
const (
	punt_2tag_error_none uint = iota
	punt_2tag_error_not_double_tagged
	punt_2tag_error_unknown_next
	punt_2tag_error_unknown_index
)

type DoubleTaggedPuntNode struct {
	vnet.InOutNode
	ref_opaque_pool vnet.RefOpaquePool
}

type PuntOpcode uint32

func (o *PuntOpcode) encode(next uint8, oi uint32) {
	*o = PuntOpcode(next) | PuntOpcode(oi)<<8
}
func (o PuntOpcode) decode() (next uint8, oi uint32) {
	next, oi = uint8(o&0xff), uint32(o)>>8
	return
}

func (n *DoubleTaggedPuntNode) AddOpaque(next uint8, o vnet.RefOpaque) PuntOpcode {
	i := n.ref_opaque_pool.GetIndex()
	n.ref_opaque_pool.Entries[i] = o
	return PuntOpcode(next) | PuntOpcode(i)<<8
}

func (n *DoubleTaggedPuntNode) decode_2tag(r0 *vnet.Ref, n_next uint) (next0 uint) {
	type header_no_type struct {
		dst, src Address
	}
	const (
		sizeof_header_no_type = 12
		sizeof_v              = 8
	)
	p0 := *(*vnet.Uint64)(r0.DataOffset(sizeof_header_no_type))

	var t = (vnet.Uint64(TYPE_VLAN)<<48 | vnet.Uint64(TYPE_VLAN)<<16).FromHost()
	var m = (vnet.Uint64(0xffff)<<48 | vnet.Uint64(0xffff)<<16).FromHost()

	error0 := punt_2tag_error_none
	if p0&m != t {
		error0 = punt_2tag_error_not_double_tagged
	}

	o0 := p0 &^ m
	if vnet.HostIsNetworkByteOrder() {
		o0 |= p0 >> 16
	} else {
		o0 |= p0 >> 48
	}

	n0, oi0 := PuntOpcode(o0).decode()

	if uint(oi0) >= n.ref_opaque_pool.Len() {
		error0 = punt_2tag_error_unknown_index
		oi0 = 0
	}

	if uint(n0) >= n_next {
		error0 = punt_2tag_error_unknown_next
	}

	r0.RefOpaque = n.ref_opaque_pool.Entries[oi0]

	n.SetError(r0, error0)
	next0 = uint(n0)
	if error0 != punt_2tag_error_none {
		next0 = punt_2tag_next_error
	}

	*(*header_no_type)(r0.DataOffset(sizeof_v)) = *(*header_no_type)(r0.DataOffset(0))

	r0.Advance(sizeof_v)

	return
}

func (n *DoubleTaggedPuntNode) Init(v *vnet.Vnet, name string) {
	n.Next = []string{
		punt_2tag_next_error: "error",
		punt_2tag_next_punt:  "punt",
	}
	n.Errors = []string{
		punt_2tag_error_none:              "no error",
		punt_2tag_error_not_double_tagged: "not double vlan tagged",
		punt_2tag_error_unknown_next:      "unknown punt next",
		punt_2tag_error_unknown_index:     "unknown packet meta-data index",
	}
	v.RegisterInOutNode(n, name+"-double-tagged-punt")
	n.AddOpaque(0, vnet.RefOpaque{})
}

func (n *DoubleTaggedPuntNode) NodeInput(in *vnet.RefIn, o *vnet.RefOut) {
	for i := uint(0); i < in.InLen(); i++ {
		r := &in.Refs[i]
		x := n.decode_2tag(r, uint(len(o.Outs)))

		o.Outs[x].BufferPool = in.BufferPool
		no := o.Outs[x].AddLen(n.Vnet)
		o.Outs[x].Refs[no] = *r
	}
}
