// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"

	"sync/atomic"
	"unsafe"
)

const (
	punt_1tag_next_punt uint = iota
	punt_1tag_next_error
)
const (
	punt_1tag_error_none uint = iota
	punt_1tag_error_not_single_tagged
	punt_1tag_error_unknown_disposition
)

type vlan_tagged_punt_node struct {
	vnet.InOutNode
	punt_packet_disposition_pool
}

func (n *vlan_tagged_punt_node) add_disposition(cf PuntConfig, n_tags uint) (i uint32) {
	p := &n.punt_packet_disposition_pool
	i = uint32(p.GetIndex())
	d := &p.dispositions[i]
	d.o = cf.RefOpaque
	d.next = uint32(n.Vnet.AddNamedNext(n, cf.Next))
	if cf.AdvanceL3Header {
		d.header_index = 0
		d.data_advance = int32(HeaderBytes + n_tags*VlanHeaderBytes)
	} else {
		d.data_advance = int32(VlanHeaderBytes * (int(n_tags) - int(cf.NReplaceVlanHeaders)))
		d.header_index = d.data_advance
	}
	switch cf.NReplaceVlanHeaders {
	case 1:
		cf.ReplaceVlanHeaders[0].Write(d.replace_tags[VlanHeaderBytes:])
	case 2:
		cf.ReplaceVlanHeaders[0].Write(d.replace_tags[0:])
		cf.ReplaceVlanHeaders[1].Write(d.replace_tags[VlanHeaderBytes:])
	}
	return
}
func (n *vlan_tagged_punt_node) del_disposition(i uint32) (ok bool) {
	return n.punt_packet_disposition_pool.PutIndex(uint(i))
}

type SingleTaggedPuntNode vlan_tagged_punt_node

func (n *SingleTaggedPuntNode) AddDisposition(cf PuntConfig) uint32 {
	return (*vlan_tagged_punt_node)(n).add_disposition(cf, 1)
}
func (n *SingleTaggedPuntNode) DelDisposition(i uint32) (ok bool) {
	return (*vlan_tagged_punt_node)(n).del_disposition(i)
}

// Ethernet header followed by is 1 vlan tag.
// Packet looks like this DST-ETHERNET SRC-ETHERNET 0x8100 TAG ETHERNET-TYPE
type header_no_type struct {
	dst, src Address
}

type type_and_tag struct {
	t Type
	g VlanTag
}

const (
	sizeof_header_no_type = 12
	sizeof_type_and_tag   = 4
)

func (n *SingleTaggedPuntNode) punt_x1(r0 *vnet.Ref) (next0 uint) {
	p0 := (*type_and_tag)(r0.DataOffset(sizeof_header_no_type))

	error0 := punt_1tag_error_none
	if p0.t != TYPE_VLAN.FromHost() {
		error0 = punt_1tag_error_not_single_tagged
	}

	di0 := uint32(p0.g.ToHost())

	if di0 >= uint32(n.punt_packet_disposition_pool.Len()) {
		error0 = punt_1tag_error_unknown_disposition
		di0 = 0
	}

	d0 := &n.dispositions[di0]

	r0.RefOpaque = d0.o

	n.SetError(r0, error0)

	next0 = uint(d0.next)
	if error0 != punt_1tag_error_none {
		next0 = punt_1tag_next_error
	}

	// Copy old src/dst ethernet address (maybe partially over-written by tag replacement).
	h0 := *(*header_no_type)(r0.DataOffset(0))

	// Possibly replace tag(s).
	*(*[2]type_and_tag)(r0.DataOffset(sizeof_header_no_type - sizeof_type_and_tag)) = *(*[2]type_and_tag)(unsafe.Pointer(&d0.replace_tags[0]))

	// Set src and dst ethernet address.
	*(*header_no_type)(r0.DataOffset(uint(d0.header_index))) = h0

	r0.Advance(int(d0.data_advance))

	return
}

func (n *SingleTaggedPuntNode) punt_x2(r0, r1 *vnet.Ref) (next0, next1 uint) {
	p0 := (*type_and_tag)(r0.DataOffset(sizeof_header_no_type))
	p1 := (*type_and_tag)(r1.DataOffset(sizeof_header_no_type))

	error0, error1 := punt_1tag_error_none, punt_1tag_error_none
	if p0.t != TYPE_VLAN.FromHost() {
		error0 = punt_1tag_error_not_single_tagged
	}
	if p1.t != TYPE_VLAN.FromHost() {
		error1 = punt_1tag_error_not_single_tagged
	}

	di0, di1 := uint32(p0.g.ToHost()), uint32(p1.g.ToHost())

	if di0 >= uint32(n.punt_packet_disposition_pool.Len()) {
		error0 = punt_1tag_error_unknown_disposition
		di0 = 0
	}
	if di1 >= uint32(n.punt_packet_disposition_pool.Len()) {
		error1 = punt_1tag_error_unknown_disposition
		di1 = 0
	}

	d0, d1 := &n.dispositions[di0], &n.dispositions[di1]

	r0.RefOpaque = d0.o
	r1.RefOpaque = d1.o

	n.SetError(r0, error0)
	n.SetError(r1, error1)

	next0, next1 = uint(d0.next), uint(d1.next)
	if error0 != punt_1tag_error_none {
		next0 = punt_1tag_next_error
	}
	if error1 != punt_1tag_error_none {
		next1 = punt_1tag_next_error
	}

	// Copy old src/dst ethernet address (maybe partially over-written by tag replacement).
	h0, h1 := *(*header_no_type)(r0.DataOffset(0)), *(*header_no_type)(r1.DataOffset(0))

	// Possibly replace tag(s).
	*(*[2]type_and_tag)(r0.DataOffset(sizeof_header_no_type - sizeof_type_and_tag)) = *(*[2]type_and_tag)(unsafe.Pointer(&d0.replace_tags[0]))
	*(*[2]type_and_tag)(r1.DataOffset(sizeof_header_no_type - sizeof_type_and_tag)) = *(*[2]type_and_tag)(unsafe.Pointer(&d1.replace_tags[0]))

	// Set src and dst ethernet address.
	*(*header_no_type)(r0.DataOffset(uint(d0.header_index))) = h0
	*(*header_no_type)(r1.DataOffset(uint(d1.header_index))) = h1

	r0.Advance(int(d0.data_advance))
	r1.Advance(int(d1.data_advance))

	return
}

func (n *SingleTaggedPuntNode) Init(v *vnet.Vnet, name string) {
	n.Next = []string{
		punt_1tag_next_error: "error",
		punt_1tag_next_punt:  "punt",
	}
	n.Errors = []string{
		punt_1tag_error_none:                "no error",
		punt_1tag_error_not_single_tagged:   "not single vlan tagged",
		punt_1tag_error_unknown_disposition: "unknown packet disposition",
	}
	v.RegisterInOutNode(n, name+"-single-tagged-punt")
	n.AddDisposition(PuntConfig{Next: "punt"})
}

func (n *SingleTaggedPuntNode) NodeInput(in *vnet.RefIn, o *vnet.RefOut) {
	q := n.GetEnqueue(in)
	i, n_left := in.Range()

	for n_left >= 2 {
		r0, r1 := in.Get2(i)
		x0, x1 := n.punt_x2(r0, r1)
		q.Put2(r0, r1, x0, x1)
		n_left -= 2
		i += 2
	}

	for n_left >= 1 {
		r0 := in.Get1(i)
		x0 := n.punt_x1(r0)
		q.Put1(r0, x0)
		n_left -= 1
		i += 1
	}
}

type SingleTaggedInjectNode struct {
	vnet.InOutNode
	inject_disposition_by_si
	sequence uint32
}

type inject_disposition_by_si struct {
	dispositions inject_packet_disposition_vec
}

func (n *SingleTaggedInjectNode) AddDisposition(next uint, offset_valid bool, si vnet.Si, tag VlanTag) {
	n.sw_if_add_del(n.Vnet, si, true)
	d := &n.dispositions[si]
	d.tags[0] = tag
	d.next.set(next, offset_valid)
	return
}

func (n *SingleTaggedInjectNode) sw_if_add_del(v *vnet.Vnet, si vnet.Si, isUp bool) (err error) {
	var zero inject_packet_disposition
	zero.next.set(uint(inject_1tag_next_error), false)
	n.dispositions.ValidateInit(uint(si), zero)
	return
}

func (n *SingleTaggedInjectNode) inject_x1(r0 *vnet.Ref, next_offset uint) (next0 uint) {
	d0 := &n.dispositions[r0.Si]

	// Copy src, dst addresses.
	h0 := *(*header_no_type)(r0.DataOffset(0))

	// Make space for 4 byte vlan header.
	r0.Advance(-sizeof_type_and_tag)

	// Insert tag.
	t0 := (*type_and_tag)(r0.DataOffset(sizeof_header_no_type))

	t0.t = TYPE_VLAN.FromHost()

	t0.g = d0.tags[0]

	// Copy back src, dst addresses.
	*(*header_no_type)(r0.DataOffset(0)) = h0

	n.SetError(r0, inject_1tag_error_unknown_interface)

	next0 = d0.next.get(next_offset)

	return
}
func (n *SingleTaggedInjectNode) inject_x2(r0, r1 *vnet.Ref, next_offset uint) (next0, next1 uint) {
	d0, d1 := &n.dispositions[r0.Si], &n.dispositions[r1.Si]

	h0, h1 := *(*header_no_type)(r0.DataOffset(0)), *(*header_no_type)(r1.DataOffset(0))

	r0.Advance(-sizeof_type_and_tag)
	r1.Advance(-sizeof_type_and_tag)

	t0 := (*type_and_tag)(r0.DataOffset(sizeof_header_no_type))
	t1 := (*type_and_tag)(r1.DataOffset(sizeof_header_no_type))

	t0.t, t1.t = TYPE_VLAN.FromHost(), TYPE_VLAN.FromHost()
	t0.g, t1.g = d0.tags[0], d1.tags[0]

	*(*header_no_type)(r0.DataOffset(0)) = h0
	*(*header_no_type)(r1.DataOffset(0)) = h1

	n.SetError(r0, inject_1tag_error_unknown_interface)
	n.SetError(r1, inject_1tag_error_unknown_interface)

	next0 = d0.next.get(next_offset)
	next1 = d1.next.get(next_offset)

	return
}

const (
	inject_1tag_next_error uint = iota
)
const (
	inject_1tag_error_none uint = iota
	inject_1tag_error_unknown_interface
)

func (n *SingleTaggedInjectNode) Init(v *vnet.Vnet, name string) {
	n.Next = []string{
		inject_1tag_next_error: "error",
	}
	n.Errors = []string{
		inject_1tag_error_none:              "no error",
		inject_1tag_error_unknown_interface: "unknown interface",
	}
	v.RegisterInOutNode(n, name+"-single-tagged-inject")
	v.RegisterSwIfAddDelHook(n.sw_if_add_del)
}

func (n *SingleTaggedInjectNode) NodeInput(in *vnet.RefIn, o *vnet.RefOut) {
	q := n.GetEnqueue(in)
	i, n_left := in.Range()
	next_offset := uint(atomic.AddUint32(&n.sequence, 1)) & 1

	for n_left >= 2 {
		r0, r1 := in.Get2(i)
		x0, x1 := n.inject_x2(r0, r1, next_offset)
		q.Put2(r0, r1, x0, x1)
		n_left -= 2
		i += 2
	}

	for n_left >= 1 {
		r0 := in.Get1(i)
		x0 := n.inject_x1(r0, next_offset)
		q.Put1(r0, x0)
		n_left -= 1
		i += 1
	}
}

type SingleTaggedPuntInjectNodes struct {
	Punt   SingleTaggedPuntNode
	Inject SingleTaggedInjectNode
}

func (n *SingleTaggedPuntInjectNodes) Init(v *vnet.Vnet, name string) {
	n.Punt.Init(v, name)
	n.Inject.Init(v, name)
}
