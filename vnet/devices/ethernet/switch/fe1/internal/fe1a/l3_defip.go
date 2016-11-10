// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

type l3_defip_key_type uint8

const (
	l3_defip_ip4 l3_defip_key_type = iota
	l3_defip_ip6_64bit
	_
	l3_defip_ip6_128bit
	fcoe
)

func (a l3_defip_key_type) tcamEncode(b l3_defip_key_type, isSet bool) (c, d l3_defip_key_type) {
	x, y := m.TcamUint8(a).TcamEncode(m.TcamUint8(b), isSet)
	c, d = l3_defip_key_type(x), l3_defip_key_type(y)
	return
}

type l3_defip_tcam_key struct {
	key_type l3_defip_key_type
	m.Vrf
	m.Ip4Address
}

func (key *l3_defip_tcam_key) tcamEncode(mask *l3_defip_tcam_key, isSet bool) (x, y l3_defip_tcam_key) {
	x.key_type, y.key_type = key.key_type.tcamEncode(mask.key_type, isSet)
	x.Vrf, y.Vrf = key.Vrf.TcamEncode(mask.Vrf, isSet)
	x.Ip4Address, y.Ip4Address = key.Ip4Address.TcamEncode(mask.Ip4Address, isSet)
	return
}

func (x *l3_defip_tcam_key) getSet(b []uint32, lo int, isSet bool) int {
	i := m.MemGetSetUint8((*uint8)(&x.key_type), b, lo+2, lo, isSet)
	i = x.Ip4Address.MemGetSet(b, i, isSet)
	i = x.Vrf.MemGetSet(b, i, isSet)
	return lo + 48
}

type l3_defip_tcam_data struct {
	bucket_has_pipe_counter bool

	drop_on_hit bool

	// Global route for parallel search: global routes are used when VRF specific table misses.
	is_global bool

	// High priority global routes override matches in VRF specific table.
	is_global_high_priority bool

	parallel_search_use_global_search_on_miss bool

	ifp_class_id m.FpClassId

	// Counter for tcam entry.
	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Hit bit index to use when ALPM bucket search misses.
	hit_bit_index uint32

	// ALPM bucket index to use for when this key/mask matches packet.
	bucket_index uint16

	// 2 bit ALPM sub-bucket index.  Used to further refine bucket prefix matching.
	sub_bucket_index uint8

	priority_change m.PriorityChange

	// Next hop to use for this tcam entry when ALPM bucket search misses.
	next_hop m.NextHop
}

func (e *l3_defip_tcam_data) getSetData(b []uint32, lo int, isSet bool) int {
	const hasReservedBit = true
	i := e.next_hop.MemGetSet(b, lo, isSet, hasReservedBit)
	i = e.priority_change.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.drop_on_hit, b, i, isSet)
	i = e.ifp_class_id.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.is_global, b, i, isSet)
	i = e.pipe_counter_ref.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.parallel_search_use_global_search_on_miss, b, i, isSet)
	i = m.MemGetSetUint32(&e.hit_bit_index, b, i+17, i, isSet)
	i = m.MemGetSet1(&e.is_global_high_priority, b, i, isSet)
	i = m.MemGetSet1(&e.bucket_has_pipe_counter, b, i, isSet)
	i = m.MemGetSetUint16(&e.bucket_index, b, i+12, i, isSet)
	i = m.MemGetSetUint8(&e.sub_bucket_index, b, i+1, i, isSet)
	if i != lo+84 {
		panic("84")
	}
	return lo + 84
}

type l3_defip_tcam_search struct {
	// Key and mask to match for this prefix or in ALPM mode to trigger bucket search.
	key, mask l3_defip_tcam_key

	is_valid bool
}

type l3_defip_half_entry struct {
	l3_defip_tcam_search

	// Data to use to forward packet in case ALPM buckets miss or not in ALPM mode.
	l3_defip_tcam_data

	was_hit bool
}

// TCAM size in double entries.
const n_l3_defip_entries = 8 << 10

type l3_defip_entry [2]l3_defip_half_entry

func (e *l3_defip_entry) MemBits() int { return 365 }

func (e *l3_defip_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}
	var keys, masks [2]l3_defip_tcam_key
	if isSet {
		for j := range e {
			keys[j], masks[j] = e[j].key.tcamEncode(&e[j].mask, isSet)
		}
	}
	for j := range e {
		i = keys[j].getSet(b, i, isSet)
	}
	for j := range e {
		i = masks[j].getSet(b, i, isSet)
	}
	if !isSet {
		for j := range e {
			e[j].key, e[j].mask = keys[j].tcamEncode(&masks[j], isSet)
		}
	}

	if i != 194 {
		panic("l3_defip 194")
	}

	for j := range e {
		i = e[j].getSetData(b, i, isSet)
	}

	if i != 362 {
		panic("l3_defip 362")
	}

	i = 363 // skip parity bit
	for j := range e {
		m.MemGetSet1(&e[j].was_hit, b, i, isSet)
	}
}

type l3_defip_mem m.MemElt

func (r *l3_defip_mem) geta(q *DmaRequest, v *l3_defip_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_mem) seta(q *DmaRequest, v *l3_defip_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_mem) get(q *DmaRequest, v *l3_defip_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *l3_defip_mem) set(q *DmaRequest, v *l3_defip_entry) { r.seta(q, v, sbus.Duplicate) }

type l3_defip_tcam_only_entry [2]l3_defip_tcam_search

func (e *l3_defip_tcam_only_entry) MemBits() int { return 194 }

func (e *l3_defip_tcam_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}
	var keys, masks [2]l3_defip_tcam_key
	if isSet {
		for j := range e {
			keys[j], masks[j] = e[j].key.tcamEncode(&e[j].mask, isSet)
		}
	}
	for j := range e {
		i = keys[j].getSet(b, i, isSet)
	}
	for j := range e {
		i = masks[j].getSet(b, i, isSet)
	}
	if !isSet {
		for j := range e {
			e[j].key, e[j].mask = keys[j].tcamEncode(&masks[j], isSet)
		}
	}
}

type l3_defip_tcam_only_mem m.MemElt

func (r *l3_defip_tcam_only_mem) geta(q *DmaRequest, v *l3_defip_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_tcam_only_mem) seta(q *DmaRequest, v *l3_defip_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_tcam_only_mem) get(q *DmaRequest, v *l3_defip_tcam_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_tcam_only_mem) set(q *DmaRequest, v *l3_defip_tcam_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_tcam_data_only_entry [2]l3_defip_tcam_data

func (e *l3_defip_tcam_data_only_entry) MemBits() int { return 169 }

func (e *l3_defip_tcam_data_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = e[j].getSetData(b, i, isSet)
	}
}

type l3_defip_tcam_data_only_mem m.MemElt

func (r *l3_defip_tcam_data_only_mem) geta(q *DmaRequest, v *l3_defip_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_tcam_data_only_mem) seta(q *DmaRequest, v *l3_defip_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_tcam_data_only_mem) get(q *DmaRequest, v *l3_defip_tcam_data_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_tcam_data_only_mem) set(q *DmaRequest, v *l3_defip_tcam_data_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_pair_entry [4]l3_defip_half_entry

func (e *l3_defip_pair_entry) MemBits() int { return 474 }

func (e *l3_defip_pair_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}

	var keys, masks [4]l3_defip_tcam_key
	if isSet {
		for j := range e {
			keys[j], masks[j] = e[j].key.tcamEncode(&e[j].mask, isSet)
		}
	}
	for j := range e {
		i = keys[j].getSet(b, i, isSet)
	}
	for j := range e {
		i = masks[j].getSet(b, i, isSet)
	}
	if !isSet {
		for j := range e {
			e[j].key, e[j].mask = keys[j].tcamEncode(&masks[j], isSet)
		}
	}

	i = 473 // skip parity bit
	m.MemGetSet1(&e[0].was_hit, b, i, isSet)
}

type l3_defip_pair_mem m.MemElt

func (r *l3_defip_pair_mem) geta(q *DmaRequest, v *l3_defip_pair_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_pair_mem) seta(q *DmaRequest, v *l3_defip_pair_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_pair_mem) get(q *DmaRequest, v *l3_defip_pair_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *l3_defip_pair_mem) set(q *DmaRequest, v *l3_defip_pair_entry) { r.seta(q, v, sbus.Duplicate) }

type l3_defip_pair_tcam_only_entry [4]l3_defip_tcam_search

func (e *l3_defip_pair_tcam_only_entry) MemBits() int { return 388 }

func (e *l3_defip_pair_tcam_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}
	var keys, masks [4]l3_defip_tcam_key
	if isSet {
		for j := range e {
			keys[j], masks[j] = e[j].key.tcamEncode(&e[j].mask, isSet)
		}
	}
	for j := range e {
		i = keys[j].getSet(b, i, isSet)
	}
	for j := range e {
		i = masks[j].getSet(b, i, isSet)
	}
	if !isSet {
		for j := range e {
			e[j].key, e[j].mask = keys[j].tcamEncode(&masks[j], isSet)
		}
	}
}

type l3_defip_pair_tcam_only_mem m.MemElt

func (r *l3_defip_pair_tcam_only_mem) geta(q *DmaRequest, v *l3_defip_pair_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_pair_tcam_only_mem) seta(q *DmaRequest, v *l3_defip_pair_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_pair_tcam_only_mem) get(q *DmaRequest, v *l3_defip_pair_tcam_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_pair_tcam_only_mem) set(q *DmaRequest, v *l3_defip_pair_tcam_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_pair_tcam_data_only_entry l3_defip_tcam_data

func (e *l3_defip_pair_tcam_data_only_entry) MemBits() int { return 85 }

func (e *l3_defip_pair_tcam_data_only_entry) MemGetSet(b []uint32, isSet bool) {
	(*l3_defip_tcam_data)(e).getSetData(b, 0, isSet)
}

type l3_defip_pair_tcam_data_only_mem m.MemElt

func (r *l3_defip_pair_tcam_data_only_mem) geta(q *DmaRequest, v *l3_defip_pair_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_pair_tcam_data_only_mem) seta(q *DmaRequest, v *l3_defip_pair_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_pair_tcam_data_only_mem) get(q *DmaRequest, v *l3_defip_pair_tcam_data_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_pair_tcam_data_only_mem) set(q *DmaRequest, v *l3_defip_pair_tcam_data_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_alpm_common struct {
	is_valid    bool
	drop_on_hit bool

	ifp_class_id m.FpClassId

	// 2 bit sub-bucket index
	sub_bucket_index uint8

	priority_change m.PriorityChange

	next_hop m.NextHop

	// Number of bits of destination to match (e.g. 24 for /24)
	dst_length uint8
}

func (e *l3_defip_alpm_common) memGetSet(b []uint32, isSet bool, dst []m.Ip4Address, pipeCounter *rx_pipe_3p11i_pipe_counter_ref) {
	i := m.MemGetSet1(&e.is_valid, b, 1, isSet)
	for j := range dst {
		i = dst[j].MemGetSet(b, i, isSet)
	}
	i = m.MemGetSetUint8(&e.sub_bucket_index, b, i+1, i, isSet)
	l := int(elib.MinLog2(elib.Word(32 * len(dst))))
	i = m.MemGetSetUint8(&e.dst_length, b, i+l, i, isSet)
	i = e.next_hop.MemGetSet(b, i, isSet /* hasReservedBit */, false)
	i = e.priority_change.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.drop_on_hit, b, i, isSet)
	i = e.ifp_class_id.MemGetSet(b, i, isSet)
	if pipeCounter != nil {
		i = pipeCounter.MemGetSet(b, i, isSet)
		if i != 88 {
			panic("88")
		}
	} else {
		if i != 70 {
			panic("70")
		}
	}
}

type l3_defip_alpm_ip4_entry struct {
	l3_defip_alpm_common

	was_hit bool

	// Ip4 destination
	dst m.Ip4Address
}

func (e *l3_defip_alpm_ip4_entry) MemBits() int { return 71 }
func (e *l3_defip_alpm_ip4_entry) MemGetSet(b []uint32, isSet bool) {
	var tmp [1]m.Ip4Address
	if isSet {
		tmp[0] = e.dst
	}
	e.l3_defip_alpm_common.memGetSet(b, isSet, tmp[:] /* no pipe counter */, nil)
	if !isSet {
		e.dst = tmp[0]
	}
	m.MemGetSet1(&e.was_hit, b, 70, isSet)
}

type l3_defip_alpm_ip4_mem m.MemElt

func (r *l3_defip_alpm_ip4_mem) geta(q *DmaRequest, v *l3_defip_alpm_ip4_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip4_mem) seta(q *DmaRequest, v *l3_defip_alpm_ip4_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip4_mem) get(q *DmaRequest, v *l3_defip_alpm_ip4_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_alpm_ip4_mem) set(q *DmaRequest, v *l3_defip_alpm_ip4_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_alpm_ip4_with_pipe_counter_entry struct {
	l3_defip_alpm_common

	was_hit bool

	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Ip4 destination
	dst m.Ip4Address
}

func (e *l3_defip_alpm_ip4_with_pipe_counter_entry) MemBits() int { return 106 }
func (e *l3_defip_alpm_ip4_with_pipe_counter_entry) MemGetSet(b []uint32, isSet bool) {
	var tmp [1]m.Ip4Address
	if isSet {
		tmp[0] = e.dst
	}
	e.l3_defip_alpm_common.memGetSet(b, isSet, tmp[:], &e.pipe_counter_ref)
	if !isSet {
		e.dst = tmp[0]
	}
	m.MemGetSet1(&e.was_hit, b, 105, isSet)
}

type l3_defip_alpm_ip4_with_pipe_counter_mem m.MemElt

func (r *l3_defip_alpm_ip4_with_pipe_counter_mem) geta(q *DmaRequest, v *l3_defip_alpm_ip4_with_pipe_counter_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip4_with_pipe_counter_mem) seta(q *DmaRequest, v *l3_defip_alpm_ip4_with_pipe_counter_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip4_with_pipe_counter_mem) get(q *DmaRequest, v *l3_defip_alpm_ip4_with_pipe_counter_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_alpm_ip4_with_pipe_counter_mem) set(q *DmaRequest, v *l3_defip_alpm_ip4_with_pipe_counter_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_alpm_ip6_64_entry struct {
	l3_defip_alpm_common

	was_hit bool

	// Ip6/64 destination
	dst [2]m.Ip4Address
}

func (e *l3_defip_alpm_ip6_64_entry) MemBits() int { return 106 }
func (e *l3_defip_alpm_ip6_64_entry) MemGetSet(b []uint32, isSet bool) {
	e.l3_defip_alpm_common.memGetSet(b, isSet, e.dst[:], nil)
	m.MemGetSet1(&e.was_hit, b, 105, isSet)
}

type l3_defip_alpm_ip6_64_mem m.MemElt

func (r *l3_defip_alpm_ip6_64_mem) geta(q *DmaRequest, v *l3_defip_alpm_ip6_64_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip6_64_mem) seta(q *DmaRequest, v *l3_defip_alpm_ip6_64_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip6_64_mem) get(q *DmaRequest, v *l3_defip_alpm_ip6_64_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_alpm_ip6_64_mem) set(q *DmaRequest, v *l3_defip_alpm_ip6_64_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_alpm_ip6_64_with_pipe_counter_entry struct {
	l3_defip_alpm_common

	was_hit bool

	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Ip6/64 destination
	dst [2]m.Ip4Address
}

func (e *l3_defip_alpm_ip6_64_with_pipe_counter_entry) MemBits() int { return 141 }
func (e *l3_defip_alpm_ip6_64_with_pipe_counter_entry) MemGetSet(b []uint32, isSet bool) {
	e.l3_defip_alpm_common.memGetSet(b, isSet, e.dst[:], &e.pipe_counter_ref)
	m.MemGetSet1(&e.was_hit, b, 140, isSet)
}

type l3_defip_alpm_ip6_64_with_pipe_counter_mem m.MemElt

func (r *l3_defip_alpm_ip6_64_with_pipe_counter_mem) geta(q *DmaRequest, v *l3_defip_alpm_ip6_64_with_pipe_counter_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip6_64_with_pipe_counter_mem) seta(q *DmaRequest, v *l3_defip_alpm_ip6_64_with_pipe_counter_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip6_64_with_pipe_counter_mem) get(q *DmaRequest, v *l3_defip_alpm_ip6_64_with_pipe_counter_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_alpm_ip6_64_with_pipe_counter_mem) set(q *DmaRequest, v *l3_defip_alpm_ip6_64_with_pipe_counter_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_defip_alpm_ip6_128_entry struct {
	l3_defip_alpm_common

	was_hit bool

	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Ip6/128 destination
	dst [4]m.Ip4Address
}

func (e *l3_defip_alpm_ip6_128_entry) MemBits() int { return 211 }
func (e *l3_defip_alpm_ip6_128_entry) MemGetSet(b []uint32, isSet bool) {
	e.l3_defip_alpm_common.memGetSet(b, isSet, e.dst[:], &e.pipe_counter_ref)
	m.MemGetSet1(&e.was_hit, b, 210, isSet)
}

type l3_defip_alpm_ip6_128_mem m.MemElt

func (r *l3_defip_alpm_ip6_128_mem) geta(q *DmaRequest, v *l3_defip_alpm_ip6_128_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip6_128_mem) seta(q *DmaRequest, v *l3_defip_alpm_ip6_128_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_defip_alpm_ip6_128_mem) get(q *DmaRequest, v *l3_defip_alpm_ip6_128_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_defip_alpm_ip6_128_mem) set(q *DmaRequest, v *l3_defip_alpm_ip6_128_entry) {
	r.seta(q, v, sbus.Duplicate)
}

func (t *fe1a) iss_init() {
	r := t.rx_pipe_regs
	q := t.getDmaReq()

	// Bypass iss memory lp.
	r.iss_memory_control_84.set(q, r.iss_memory_control_84.getDo(q, sbus.Duplicate)|(0xf<<0))

	// All banks ALPM
	r.iss_bank_config.set(q, 0xf<<8)

	// ALPM uses 4 banks (and not 2)
	r.iss_logical_to_physical_bank_map.set(q, 0<<24)
	r.iss_alpm_logical_to_physical_bank_map.set(q, 0)

	// Enable algorithmic LPM mode and combined search mode.
	r.l3_defip_control.set(q, 1<<1|(0<<2))

	q.Do()
}
