// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

const n_l3_terminate_tcam_entry = 1 << 10

// Key and mask for station TCAM.  Match when search & mask == key */
type l3_terminate_tcam_key struct {
	m.LogicalPort
	m.Vlan
	m.EthernetAddress
}

func (key *l3_terminate_tcam_key) tcamEncode(mask *l3_terminate_tcam_key, isSet bool) (x, y l3_terminate_tcam_key) {
	x.LogicalPort, y.LogicalPort = key.LogicalPort.TcamEncode(&mask.LogicalPort, isSet)
	x.Vlan, y.Vlan = key.Vlan.TcamEncode(mask.Vlan, isSet)
	x.EthernetAddress, y.EthernetAddress = key.EthernetAddress.TcamEncode(&mask.EthernetAddress, isSet)
	return
}

// What to do when packet dst ethernet address/port/vlan entry matches.
type l3_terminate_tcam_data struct {
	copy_to_cpu bool
	drop        bool
	// Enable termination of various types of packets.
	ip4_unicast_enable   bool
	ip6_unicast_enable   bool
	ip4_multicast_enable bool
	ip6_multicast_enable bool
	mpls_enable          bool
	arp_rarp_enable      bool
	fcoe_enable          bool
	trill_enable         bool
	mac_in_mac_enable    bool
}

type l3_terminate_tcam_entry struct {
	data      l3_terminate_tcam_data
	valid     bool
	key, mask l3_terminate_tcam_key
}

func (r *l3_terminate_tcam_key) getSet(b []uint32, lo int, isSet bool) int {
	i := r.EthernetAddress.MemGetSet(b, lo, isSet)
	i = r.Vlan.MemGetSet(b, i, isSet)
	i = r.LogicalPort.MemGetSet(b, i, isSet)
	return lo + 80
}

func (r *l3_terminate_tcam_data) getSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSet1(&r.mac_in_mac_enable, b, i, isSet)
	i = m.MemGetSet1(&r.mpls_enable, b, i, isSet)
	i = m.MemGetSet1(&r.trill_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip4_unicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip6_unicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.arp_rarp_enable, b, i, isSet)
	i = m.MemGetSet1(&r.fcoe_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip4_multicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.ip6_multicast_enable, b, i, isSet)
	i = m.MemGetSet1(&r.drop, b, i, isSet)
	i = m.MemGetSet1(&r.copy_to_cpu, b, i, isSet)
	// bit 11 is reserved
	i += 1
	return i
}

func (r *l3_terminate_tcam_entry) MemBits() int { return 174 }
func (r *l3_terminate_tcam_entry) MemGetSet(b []uint32, isSet bool) {
	i := m.MemGetSet1(&r.valid, b, 0, isSet)
	var key, mask l3_terminate_tcam_key
	if isSet {
		key, mask = r.key.tcamEncode(&r.mask, isSet)
	}
	i = key.getSet(b, i, isSet)
	if i != 81 {
		panic("81")
	}
	i = mask.getSet(b, i, isSet)
	if i != 161 {
		panic("161")
	}
	if !isSet {
		key, mask = r.key.tcamEncode(&r.mask, isSet)
	}
	i = r.data.getSet(b, i, isSet)
}

type l3_terminate_tcam_mem m.MemElt

func (r *l3_terminate_tcam_mem) geta(q *DmaRequest, v *l3_terminate_tcam_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_terminate_tcam_mem) seta(q *DmaRequest, v *l3_terminate_tcam_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_terminate_tcam_mem) get(q *DmaRequest, v *l3_terminate_tcam_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *l3_terminate_tcam_mem) set(q *DmaRequest, v *l3_terminate_tcam_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_terminate_tcam_entry_only_mem m.MemElt
type l3_terminate_tcam_data_only_mem m.MemElt

//go:generate gentemplate -d Package=fe1a -id l3_terminate_tcam -d PoolType=l3_terminate_tcam_pool -d Type=l3_terminate_tcam_entry -d Data=entries github.com/platinasystems/go/elib/pool.tmpl

type l3_terminate_tcam_main struct {
	pool             l3_terminate_tcam_pool
	poolIndexByEntry map[l3_terminate_tcam_entry]uint
}

func (tm *l3_terminate_tcam_main) addDel(t *fe1a, e *l3_terminate_tcam_entry, isDel bool) (i uint, ok bool) {
	if tm.poolIndexByEntry == nil {
		tm.poolIndexByEntry = make(map[l3_terminate_tcam_entry]uint)
		tm.pool.SetMaxLen(n_l3_terminate_tcam_entry)
	}

	q := t.getDmaReq()
	f := l3_terminate_tcam_entry{}
	f.key = e.key
	f.mask = e.mask
	if i, ok = tm.poolIndexByEntry[f]; !ok && isDel {
		return
	}
	if isDel {
		pe := &tm.pool.entries[i]
		pe.valid = false
		t.rx_pipe_mems.l3_terminate_tcam[i].set(q, pe)
		tm.pool.PutIndex(i)
		delete(tm.poolIndexByEntry, f)
	} else {
		if !ok {
			i = tm.pool.GetIndex()
			tm.poolIndexByEntry[f] = i
		}
		pe := &tm.pool.entries[i]
		*pe = *e
		pe.valid = true
		t.rx_pipe_mems.l3_terminate_tcam[i].set(q, pe)
	}
	q.Do()
	return
}

type fib_tcam_key_type uint8

const (
	fib_tcam_ip4 fib_tcam_key_type = iota
	fib_tcam_ip6_64bit
	_
	fib_tcam_ip6_128bit
	fcoe
)

func (a fib_tcam_key_type) tcamEncode(b fib_tcam_key_type, isSet bool) (c, d fib_tcam_key_type) {
	x, y := m.TcamUint8(a).TcamEncode(m.TcamUint8(b), isSet)
	c, d = fib_tcam_key_type(x), fib_tcam_key_type(y)
	return
}

type fib_tcam_tcam_key struct {
	key_type fib_tcam_key_type
	m.Vrf
	m.Ip4Address
}

func (key *fib_tcam_tcam_key) tcamEncode(mask *fib_tcam_tcam_key, isSet bool) (x, y fib_tcam_tcam_key) {
	x.key_type, y.key_type = key.key_type.tcamEncode(mask.key_type, isSet)
	x.Vrf, y.Vrf = key.Vrf.TcamEncode(mask.Vrf, isSet)
	x.Ip4Address, y.Ip4Address = key.Ip4Address.TcamEncode(mask.Ip4Address, isSet)
	return
}

func (x *fib_tcam_tcam_key) getSet(b []uint32, lo int, isSet bool) int {
	i := m.MemGetSetUint8((*uint8)(&x.key_type), b, lo+2, lo, isSet)
	i = x.Ip4Address.MemGetSet(b, i, isSet)
	i = x.Vrf.MemGetSet(b, i, isSet)
	return lo + 48
}

type fib_tcam_tcam_data struct {
	bucket_has_pipe_counter bool

	drop_on_hit bool

	// Global route for parallel search: global routes are used when VRF specific table misses.
	is_global bool

	// High priority global routes override matches in VRF specific table.
	is_global_high_priority bool

	parallel_search_use_global_search_on_miss bool

	rxf_class_id m.FpClassId

	// Counter for tcam entry.
	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Hit bit index to use when bucket search misses.
	hit_bit_index uint32

	// BUCKET bucket index to use for when this key/mask matches packet.
	bucket_index uint16

	// 2 bit BUCKET sub-bucket index.  Used to further refine bucket prefix matching.
	sub_bucket_index uint8

	priority_change m.PriorityChange

	// Next hop to use for this tcam entry when BUCKET bucket search misses.
	next_hop m.NextHop
}

func (e *fib_tcam_tcam_data) getSetData(b []uint32, lo int, isSet bool) int {
	const hasReservedBit = true
	i := e.next_hop.MemGetSet(b, lo, isSet, hasReservedBit)
	i = e.priority_change.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.drop_on_hit, b, i, isSet)
	i = e.rxf_class_id.MemGetSet(b, i, isSet)
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

type fib_tcam_tcam_search struct {
	// Key and mask to match for this prefix or in BUCKET mode to trigger bucket search.
	key, mask fib_tcam_tcam_key

	is_valid bool
}

type fib_tcam_half_entry struct {
	fib_tcam_tcam_search

	// Data to use to forward packet in case BUCKET buckets miss or not in BUCKET mode.
	fib_tcam_tcam_data

	was_hit bool
}

// TCAM size in double entries.
const n_fib_tcam_entries = 8 << 10

type fib_tcam_entry [2]fib_tcam_half_entry

func (e *fib_tcam_entry) MemBits() int { return 365 }

func (e *fib_tcam_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}
	var keys, masks [2]fib_tcam_tcam_key
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
		panic("fib_tcam 194")
	}

	for j := range e {
		i = e[j].getSetData(b, i, isSet)
	}

	if i != 362 {
		panic("fib_tcam 362")
	}

	i = 363 // skip parity bit
	for j := range e {
		m.MemGetSet1(&e[j].was_hit, b, i, isSet)
	}
}

type fib_tcam_mem m.MemElt

func (r *fib_tcam_mem) geta(q *DmaRequest, v *fib_tcam_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_mem) seta(q *DmaRequest, v *fib_tcam_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_mem) get(q *DmaRequest, v *fib_tcam_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *fib_tcam_mem) set(q *DmaRequest, v *fib_tcam_entry) { r.seta(q, v, sbus.Duplicate) }

type fib_tcam_tcam_only_entry [2]fib_tcam_tcam_search

func (e *fib_tcam_tcam_only_entry) MemBits() int { return 194 }

func (e *fib_tcam_tcam_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}
	var keys, masks [2]fib_tcam_tcam_key
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

type fib_tcam_tcam_only_mem m.MemElt

func (r *fib_tcam_tcam_only_mem) geta(q *DmaRequest, v *fib_tcam_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_tcam_only_mem) seta(q *DmaRequest, v *fib_tcam_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_tcam_only_mem) get(q *DmaRequest, v *fib_tcam_tcam_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_tcam_only_mem) set(q *DmaRequest, v *fib_tcam_tcam_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_tcam_data_only_entry [2]fib_tcam_tcam_data

func (e *fib_tcam_tcam_data_only_entry) MemBits() int { return 169 }

func (e *fib_tcam_tcam_data_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = e[j].getSetData(b, i, isSet)
	}
}

type fib_tcam_tcam_data_only_mem m.MemElt

func (r *fib_tcam_tcam_data_only_mem) geta(q *DmaRequest, v *fib_tcam_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_tcam_data_only_mem) seta(q *DmaRequest, v *fib_tcam_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_tcam_data_only_mem) get(q *DmaRequest, v *fib_tcam_tcam_data_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_tcam_data_only_mem) set(q *DmaRequest, v *fib_tcam_tcam_data_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_pair_entry [4]fib_tcam_half_entry

func (e *fib_tcam_pair_entry) MemBits() int { return 474 }

func (e *fib_tcam_pair_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}

	var keys, masks [4]fib_tcam_tcam_key
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

type fib_tcam_pair_mem m.MemElt

func (r *fib_tcam_pair_mem) geta(q *DmaRequest, v *fib_tcam_pair_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_pair_mem) seta(q *DmaRequest, v *fib_tcam_pair_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_pair_mem) get(q *DmaRequest, v *fib_tcam_pair_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *fib_tcam_pair_mem) set(q *DmaRequest, v *fib_tcam_pair_entry) { r.seta(q, v, sbus.Duplicate) }

type fib_tcam_pair_tcam_only_entry [4]fib_tcam_tcam_search

func (e *fib_tcam_pair_tcam_only_entry) MemBits() int { return 388 }

func (e *fib_tcam_pair_tcam_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e {
		i = m.MemGetSet1(&e[j].is_valid, b, i, isSet)
	}
	var keys, masks [4]fib_tcam_tcam_key
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

type fib_tcam_pair_tcam_only_mem m.MemElt

func (r *fib_tcam_pair_tcam_only_mem) geta(q *DmaRequest, v *fib_tcam_pair_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_pair_tcam_only_mem) seta(q *DmaRequest, v *fib_tcam_pair_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_pair_tcam_only_mem) get(q *DmaRequest, v *fib_tcam_pair_tcam_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_pair_tcam_only_mem) set(q *DmaRequest, v *fib_tcam_pair_tcam_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_pair_tcam_data_only_entry fib_tcam_tcam_data

func (e *fib_tcam_pair_tcam_data_only_entry) MemBits() int { return 85 }

func (e *fib_tcam_pair_tcam_data_only_entry) MemGetSet(b []uint32, isSet bool) {
	(*fib_tcam_tcam_data)(e).getSetData(b, 0, isSet)
}

type fib_tcam_pair_tcam_data_only_mem m.MemElt

func (r *fib_tcam_pair_tcam_data_only_mem) geta(q *DmaRequest, v *fib_tcam_pair_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_pair_tcam_data_only_mem) seta(q *DmaRequest, v *fib_tcam_pair_tcam_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_pair_tcam_data_only_mem) get(q *DmaRequest, v *fib_tcam_pair_tcam_data_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_pair_tcam_data_only_mem) set(q *DmaRequest, v *fib_tcam_pair_tcam_data_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_bucket_common struct {
	is_valid    bool
	drop_on_hit bool

	rxf_class_id m.FpClassId

	// 2 bit sub-bucket index
	sub_bucket_index uint8

	priority_change m.PriorityChange

	next_hop m.NextHop

	// Number of bits of destination to match (e.g. 24 for /24)
	dst_length uint8
}

func (e *fib_tcam_bucket_common) memGetSet(b []uint32, isSet bool, dst []m.Ip4Address, pipeCounter *rx_pipe_3p11i_pipe_counter_ref) {
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
	i = e.rxf_class_id.MemGetSet(b, i, isSet)
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

type fib_tcam_6_entry_bucket_entry struct {
	fib_tcam_bucket_common

	was_hit bool

	// Ip4 destination
	dst m.Ip4Address
}

func (e *fib_tcam_6_entry_bucket_entry) MemBits() int { return 71 }
func (e *fib_tcam_6_entry_bucket_entry) MemGetSet(b []uint32, isSet bool) {
	var tmp [1]m.Ip4Address
	if isSet {
		tmp[0] = e.dst
	}
	e.fib_tcam_bucket_common.memGetSet(b, isSet, tmp[:] /* no pipe counter */, nil)
	if !isSet {
		e.dst = tmp[0]
	}
	m.MemGetSet1(&e.was_hit, b, 70, isSet)
}

type fib_tcam_6_entry_bucket_mem m.MemElt

func (r *fib_tcam_6_entry_bucket_mem) geta(q *DmaRequest, v *fib_tcam_6_entry_bucket_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_6_entry_bucket_mem) seta(q *DmaRequest, v *fib_tcam_6_entry_bucket_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_6_entry_bucket_mem) get(q *DmaRequest, v *fib_tcam_6_entry_bucket_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_6_entry_bucket_mem) set(q *DmaRequest, v *fib_tcam_6_entry_bucket_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_4_ip4_entry_bucket_entry struct {
	fib_tcam_bucket_common

	was_hit bool

	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Ip4 destination
	dst m.Ip4Address
}

func (e *fib_tcam_4_ip4_entry_bucket_entry) MemBits() int { return 106 }
func (e *fib_tcam_4_ip4_entry_bucket_entry) MemGetSet(b []uint32, isSet bool) {
	var tmp [1]m.Ip4Address
	if isSet {
		tmp[0] = e.dst
	}
	e.fib_tcam_bucket_common.memGetSet(b, isSet, tmp[:], &e.pipe_counter_ref)
	if !isSet {
		e.dst = tmp[0]
	}
	m.MemGetSet1(&e.was_hit, b, 105, isSet)
}

type fib_tcam_4_ip4_entry_bucket_mem m.MemElt

func (r *fib_tcam_4_ip4_entry_bucket_mem) geta(q *DmaRequest, v *fib_tcam_4_ip4_entry_bucket_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_4_ip4_entry_bucket_mem) seta(q *DmaRequest, v *fib_tcam_4_ip4_entry_bucket_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_4_ip4_entry_bucket_mem) get(q *DmaRequest, v *fib_tcam_4_ip4_entry_bucket_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_4_ip4_entry_bucket_mem) set(q *DmaRequest, v *fib_tcam_4_ip4_entry_bucket_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_bucket_ip6_64_entry struct {
	fib_tcam_bucket_common

	was_hit bool

	// Ip6/64 destination
	dst [2]m.Ip4Address
}

func (e *fib_tcam_bucket_ip6_64_entry) MemBits() int { return 106 }
func (e *fib_tcam_bucket_ip6_64_entry) MemGetSet(b []uint32, isSet bool) {
	e.fib_tcam_bucket_common.memGetSet(b, isSet, e.dst[:], nil)
	m.MemGetSet1(&e.was_hit, b, 105, isSet)
}

type fib_tcam_bucket_ip6_64_mem m.MemElt

func (r *fib_tcam_bucket_ip6_64_mem) geta(q *DmaRequest, v *fib_tcam_bucket_ip6_64_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_bucket_ip6_64_mem) seta(q *DmaRequest, v *fib_tcam_bucket_ip6_64_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_bucket_ip6_64_mem) get(q *DmaRequest, v *fib_tcam_bucket_ip6_64_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_bucket_ip6_64_mem) set(q *DmaRequest, v *fib_tcam_bucket_ip6_64_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_bucket_ip6_64_with_pipe_counter_entry struct {
	fib_tcam_bucket_common

	was_hit bool

	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Ip6/64 destination
	dst [2]m.Ip4Address
}

func (e *fib_tcam_bucket_ip6_64_with_pipe_counter_entry) MemBits() int { return 141 }
func (e *fib_tcam_bucket_ip6_64_with_pipe_counter_entry) MemGetSet(b []uint32, isSet bool) {
	e.fib_tcam_bucket_common.memGetSet(b, isSet, e.dst[:], &e.pipe_counter_ref)
	m.MemGetSet1(&e.was_hit, b, 140, isSet)
}

type fib_tcam_bucket_ip6_64_with_pipe_counter_mem m.MemElt

func (r *fib_tcam_bucket_ip6_64_with_pipe_counter_mem) geta(q *DmaRequest, v *fib_tcam_bucket_ip6_64_with_pipe_counter_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_bucket_ip6_64_with_pipe_counter_mem) seta(q *DmaRequest, v *fib_tcam_bucket_ip6_64_with_pipe_counter_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_bucket_ip6_64_with_pipe_counter_mem) get(q *DmaRequest, v *fib_tcam_bucket_ip6_64_with_pipe_counter_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_bucket_ip6_64_with_pipe_counter_mem) set(q *DmaRequest, v *fib_tcam_bucket_ip6_64_with_pipe_counter_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type fib_tcam_bucket_ip6_128_entry struct {
	fib_tcam_bucket_common

	was_hit bool

	pipe_counter_ref rx_pipe_3p11i_pipe_counter_ref

	// Ip6/128 destination
	dst [4]m.Ip4Address
}

func (e *fib_tcam_bucket_ip6_128_entry) MemBits() int { return 211 }
func (e *fib_tcam_bucket_ip6_128_entry) MemGetSet(b []uint32, isSet bool) {
	e.fib_tcam_bucket_common.memGetSet(b, isSet, e.dst[:], &e.pipe_counter_ref)
	m.MemGetSet1(&e.was_hit, b, 210, isSet)
}

type fib_tcam_bucket_ip6_128_mem m.MemElt

func (r *fib_tcam_bucket_ip6_128_mem) geta(q *DmaRequest, v *fib_tcam_bucket_ip6_128_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_bucket_ip6_128_mem) seta(q *DmaRequest, v *fib_tcam_bucket_ip6_128_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *fib_tcam_bucket_ip6_128_mem) get(q *DmaRequest, v *fib_tcam_bucket_ip6_128_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *fib_tcam_bucket_ip6_128_mem) set(q *DmaRequest, v *fib_tcam_bucket_ip6_128_entry) {
	r.seta(q, v, sbus.Duplicate)
}

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
	rxf_class_id    uint8
	priority_change m.PriorityChange
	// Next hop or ecmp index depending on is_ecmp.
	index uint32
}

type l3_ip4_entry struct {
	key_type l3_entry_key_type
	valid    bool
	l3_entry_data
	m.Vrf
	m.Ip4Address
}

func (e *l3_ip4_entry) MemBits() int { return 106 }

func (e *l3_ip4_entry) MemGetSet(b []uint32, isSet bool) {
	i := 1 // skip parity bit
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.key_type), b, i+4, i, isSet)
	i = e.Ip4Address.MemGetSet(b, i, isSet)
	i = e.Vrf.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.rxf_class_id, b, i+5, i, isSet)
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

type l3_ip4_entry_mem m.MemElt

func (r *l3_ip4_entry_mem) geta(q *DmaRequest, v *l3_ip4_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_ip4_entry_mem) seta(q *DmaRequest, v *l3_ip4_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *l3_ip4_entry_mem) get(q *DmaRequest, v *l3_ip4_entry) { r.geta(q, v, sbus.Duplicate) }
func (r *l3_ip4_entry_mem) set(q *DmaRequest, v *l3_ip4_entry) { r.seta(q, v, sbus.Duplicate) }

func (t *fe1a) shared_lookup_sram_init() {
	r := t.rx_pipe_controller
	q := t.getDmaReq()

	// Bypass iss memory lp.
	r.shared_lookup_sram_memory_control_84.set(q, r.shared_lookup_sram_memory_control_84.getDo(q, sbus.Duplicate)|(0xf<<0))

	// All banks BUCKET
	r.shared_lookup_sram_bank_config.set(q, 0xf<<8)

	// BUCKET uses 4 banks (and not 2)
	r.shared_lookup_sram_logical_to_physical_bank_map.set(q, 0<<24)
	r.shared_lookup_sram_bucket_logical_to_physical_bank_map.set(q, 0)

	// Enable algorithmic LPM mode and combined search mode.
	r.fib_tcam.control.set(q, 1<<1|(0<<2))

	q.Do()
}
