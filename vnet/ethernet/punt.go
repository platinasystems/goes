// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"
)

type DoubleTaggedPuntNext uint

const (
	DoubleTaggedPuntNextPunt DoubleTaggedPuntNext = iota
	DoubleTaggedPuntNextError
)
const (
	doubleTaggedPuntErrorNone uint = iota
	doubleTaggedPuntErrorNotDoubleTagged
	doubleTaggedPuntErrorUnknownNext
	doubleTaggedPuntErrorUnknownIndex
)

type DoubleTaggedPuntNode struct {
	vnet.InOutNode
	ref_opaque_pool vnet.RefOpaquePool
}

type PuntOpcode uint32

func (o *PuntOpcode) encode(next uint8, oi uint32) {
	*o = PuntOpcode(next) | PuntOpcode(oi)<<8
}
func (o PuntOpcode) decode() (next uint, oi uint) {
	next, oi = uint(o&0xff), uint(o)>>8
	return
}

func (n *DoubleTaggedPuntNode) AddPuntOpcode(next DoubleTaggedPuntNext, o vnet.RefOpaque) PuntOpcode {
	i := n.ref_opaque_pool.GetIndex()
	n.ref_opaque_pool.Entries[i] = o
	return PuntOpcode(next) | PuntOpcode(i)<<8
}

// Ethernet header followed by is 2 vlan tags.
// Packet looks like this DST-ETHERNET SRC-ETHERNET 0x8100 TAG0 0x8100 TAG1 ETHERNET-TYPE
type header_no_type struct {
	dst, src Address
}

const (
	sizeof_header_no_type = 12
	sizeof_double_tag     = 8
)

func (n *DoubleTaggedPuntNode) punt_x1(r0 *vnet.Ref) (next0 uint) {
	p0 := *(*vnet.Uint64)(r0.DataOffset(sizeof_header_no_type))

	var t = (vnet.Uint64(TYPE_VLAN)<<48 | vnet.Uint64(TYPE_VLAN)<<16).FromHost()
	var m = (vnet.Uint64(0xffff)<<48 | vnet.Uint64(0xffff)<<16).FromHost()

	error0 := doubleTaggedPuntErrorNone
	if p0&m != t {
		error0 = doubleTaggedPuntErrorNotDoubleTagged
	}

	o0 := p0 &^ m
	if vnet.HostIsNetworkByteOrder() {
		o0 |= p0 >> 16
	} else {
		o0 |= p0 >> 48
	}

	n0, oi0 := PuntOpcode(o0).decode()

	if oi0 >= n.ref_opaque_pool.Len() {
		error0 = doubleTaggedPuntErrorUnknownIndex
		oi0 = 0
	}

	if n0 >= n.MaxNext() {
		error0 = doubleTaggedPuntErrorUnknownNext
	}

	r0.RefOpaque = n.ref_opaque_pool.Entries[oi0]

	next0 = n0
	if error0 != doubleTaggedPuntErrorNone {
		next0 = uint(DoubleTaggedPuntNextError)
		n.SetError(r0, error0)
	}

	// Remove double tag.  After this call packet is untagged.
	*(*header_no_type)(r0.DataOffset(sizeof_double_tag)) = *(*header_no_type)(r0.DataOffset(0))

	r0.Advance(sizeof_double_tag)

	return
}

func (n *DoubleTaggedPuntNode) punt_x2(r0, r1 *vnet.Ref) (next0, next1 uint) {
	p0 := *(*vnet.Uint64)(r0.DataOffset(sizeof_header_no_type))
	p1 := *(*vnet.Uint64)(r1.DataOffset(sizeof_header_no_type))

	var t = (vnet.Uint64(TYPE_VLAN)<<48 | vnet.Uint64(TYPE_VLAN)<<16).FromHost()
	var m = (vnet.Uint64(0xffff)<<48 | vnet.Uint64(0xffff)<<16).FromHost()

	error0, error1 := doubleTaggedPuntErrorNone, doubleTaggedPuntErrorNone
	if p0&m != t {
		error0 = doubleTaggedPuntErrorNotDoubleTagged
	}
	if p1&m != t {
		error1 = doubleTaggedPuntErrorNotDoubleTagged
	}

	o0, o1 := p0&^m, p1&^m
	if vnet.HostIsNetworkByteOrder() {
		o0 |= p0 >> 16
		o1 |= p1 >> 16
	} else {
		o0 |= p0 >> 48
		o1 |= p1 >> 48
	}

	n0, oi0 := PuntOpcode(o0).decode()
	n1, oi1 := PuntOpcode(o1).decode()

	if oi0 >= n.ref_opaque_pool.Len() {
		error0 = doubleTaggedPuntErrorUnknownIndex
		oi0 = 0
	}
	if oi1 >= n.ref_opaque_pool.Len() {
		error1 = doubleTaggedPuntErrorUnknownIndex
		oi1 = 0
	}

	if n0 >= n.MaxNext() {
		error0 = doubleTaggedPuntErrorUnknownNext
	}
	if n1 >= n.MaxNext() {
		error1 = doubleTaggedPuntErrorUnknownNext
	}

	r0.RefOpaque = n.ref_opaque_pool.Entries[oi0]
	r1.RefOpaque = n.ref_opaque_pool.Entries[oi1]

	next0, next1 = n0, n1
	if error0 != doubleTaggedPuntErrorNone {
		next0 = uint(DoubleTaggedPuntNextError)
		n.SetError(r0, error0)
	}
	if error1 != doubleTaggedPuntErrorNone {
		next1 = uint(DoubleTaggedPuntNextError)
		n.SetError(r1, error1)
	}

	// Remove double tag.  After this call packet is untagged.
	*(*header_no_type)(r0.DataOffset(sizeof_double_tag)) = *(*header_no_type)(r0.DataOffset(0))
	*(*header_no_type)(r1.DataOffset(sizeof_double_tag)) = *(*header_no_type)(r1.DataOffset(0))

	r0.Advance(sizeof_double_tag)
	r1.Advance(sizeof_double_tag)

	return
}

func (n *DoubleTaggedPuntNode) Init(v *vnet.Vnet, name string) {
	n.Next = []string{
		DoubleTaggedPuntNextError: "error",
		DoubleTaggedPuntNextPunt:  "punt",
	}
	n.Errors = []string{
		doubleTaggedPuntErrorNone:            "no error",
		doubleTaggedPuntErrorNotDoubleTagged: "not double vlan tagged",
		doubleTaggedPuntErrorUnknownNext:     "unknown punt next",
		doubleTaggedPuntErrorUnknownIndex:    "unknown packet meta-data index",
	}
	v.RegisterInOutNode(n, name+"-double-tagged-punt")
	n.AddPuntOpcode(0, vnet.RefOpaque{})
}

func (n *DoubleTaggedPuntNode) NodeInput(in *vnet.RefIn, o *vnet.RefOut) {
	i, n_left := in.Range()

	for n_left >= 2 {
		r0, r1 := in.Get2(i)
		x0, x1 := n.punt_x2(r0, r1)
		n.Put2(r0, r1, x0, x1)
		n_left -= 2
		i += 2
	}

	for n_left >= 1 {
		r0 := in.Get1(i)
		x0 := n.punt_x1(r0)
		n.Put1(r0, x0)
		n_left -= 1
		i += 1
	}
}
