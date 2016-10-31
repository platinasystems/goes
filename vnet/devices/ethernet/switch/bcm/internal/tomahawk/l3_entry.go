// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"
)

type l3_entry_key_type uint8

const (
	l3_entry_ip4_unicast l3_entry_key_type = iota
	l3_entry_ip4_unicast_extended
	l3_entry_ip6_unicast
	l3_entry_ip6_unicast_extended
	l3_entry_ip4_multicast
	l3_entry_ip6_multicast
	l3_entry_trill
	_
)

type l3_entry_data struct {
	is_hit          bool
	is_local        bool
	bfd_enable      bool
	is_ecmp         bool
	drop            bool
	ifp_class_id    uint8
	priority_change m.PriorityChange
	// Next hop or ecmp index depending on is_ecmp.
	index uint32
}

type l3_ipv4_entry struct {
	key_type l3_entry_key_type
	valid    bool
	l3_entry_data
	m.Vrf
	m.Ip4Address
}

func (e *l3_ipv4_entry) MemBits() int { return 106 }

func (e *l3_ipv4_entry) MemGetSet(b []uint32, isSet bool) {
	i := 1 // skip parity bit
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.key_type), b, i+4, i, isSet)
	i = e.Ip4Address.MemGetSet(b, i, isSet)
	i = e.Vrf.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ifp_class_id, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.is_local, b, i, isSet)
	i = m.MemGetSet1(&e.bfd_enable, b, i, isSet)
	i = e.priority_change.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint32(&e.index, b, i+16, i, isSet)
	i = m.MemGetSet1(&e.drop, b, i, isSet)
	i = m.MemGetSet1(&e.is_ecmp, b, i, isSet)
	if i != 82 {
		panic("82")
	}
}

type l3_ipv4_entry_mem m.MemElt

func (r *l3_ipv4_entry_mem) geta(q *DmaRequest, v *l3_ipv4_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_ipv4_entry_mem) seta(q *DmaRequest, v *l3_ipv4_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_ipv4_entry_mem) get(q *DmaRequest, v *l3_ipv4_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *l3_ipv4_entry_mem) set(q *DmaRequest, v *l3_ipv4_entry) { r.seta(q, v, sbus.Duplicate) }
