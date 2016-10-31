// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"
)

const (
	n_l2_multicast_entry = 16 << 10
)

type l2_multicast_entry struct {
	valid bool

	hi_gig_trunk_override_profile_index uint8

	// Bitmap of l2 ports to flood packets matching this index.
	ports port_bitmap
}

func (e *l2_multicast_entry) MemBits() int { return 162 }
func (e *l2_multicast_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.ports.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.hi_gig_trunk_override_profile_index, b, i+7, i, isSet)
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	if i != 145 {
		panic("l2_multicast_entry")
	}
}

type l2_multicast_mem m.MemElt

func (r *l2_multicast_mem) geta(q *DmaRequest, v *l2_multicast_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l2_multicast_mem) seta(q *DmaRequest, v *l2_multicast_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l2_multicast_mem) get(q *DmaRequest, v *l2_multicast_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l2_multicast_mem) set(q *DmaRequest, v *l2_multicast_entry) {
	r.seta(q, v, sbus.Duplicate)
}

const (
	n_l3_multicast_entry = 8 << 10
)

type l3_multicast_entry struct {
	valid                              bool
	remove_incoming_port_from_l3_ports bool
	disable_clear_l3_bitmap            bool
	disable_mmu_cut_through            bool

	hi_gig_trunk_override_profile_index uint8

	repl_head_base_index uint16

	// Bitmap of l2 and l3 ports to flood packets matching this index.
	l2_ports, l3_ports port_bitmap
}

func (e *l3_multicast_entry) MemBits() int { return 333 }
func (e *l3_multicast_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.l2_ports.MemGetSet(b, i, isSet)
	i = e.l3_ports.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.hi_gig_trunk_override_profile_index, b, i+7, i, isSet)
	i = m.MemGetSet1(&e.remove_incoming_port_from_l3_ports, b, i, isSet)
	i = m.MemGetSet1(&e.disable_clear_l3_bitmap, b, i, isSet)
	i = m.MemGetSet1(&e.disable_mmu_cut_through, b, i, isSet)
	i = m.MemGetSetUint16(&e.repl_head_base_index, b, i+15, i, isSet)
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	if i != 300 {
		panic("l3_multicast_entry")
	}
}

type l3_multicast_mem m.MemElt

func (r *l3_multicast_mem) geta(q *DmaRequest, v *l3_multicast_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_multicast_mem) seta(q *DmaRequest, v *l3_multicast_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_multicast_mem) get(q *DmaRequest, v *l3_multicast_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_multicast_mem) set(q *DmaRequest, v *l3_multicast_entry) {
	r.seta(q, v, sbus.Duplicate)
}
