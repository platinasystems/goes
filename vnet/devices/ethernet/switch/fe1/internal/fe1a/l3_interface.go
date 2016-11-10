// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/pipemem"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"errors"
)

const n_l3_interface = 8 << 10

type rx_l3_interface_entry struct {
	vrf uint16

	rx_l3_interface_profile_index uint8

	ifp_class_id uint16

	ip_multicast_interface_for_lookup_keys uint16

	ip_option_profile_index uint8

	// Used for IP multicast PIM bi-directional forwarding.
	active_rx_l3_interface_profile_index uint16

	src_realm_id uint8

	tunnel_termination_ecn_decap_mapping_pointer uint8

	pipe_counter_ref rx_pipe_4p12i_pipe_counter_ref
}

func (e *rx_l3_interface_entry) MemBits() int { return 82 }

func (e *rx_l3_interface_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint16(&e.vrf, b, i+10, i, isSet)
	i = m.MemGetSetUint8(&e.rx_l3_interface_profile_index, b, i+7, i, isSet)
	i = m.MemGetSetUint16(&e.ip_multicast_interface_for_lookup_keys, b, i+12, i, isSet)
	i = e.pipe_counter_ref.MemGetSet(b, i, isSet)

	if i != 52 {
		panic("rx_l3_interface_entry: ip_opt")
	}

	i = m.MemGetSetUint8(&e.ip_option_profile_index, b, i+1, i, isSet)
	i = m.MemGetSetUint16(&e.ifp_class_id, b, i+11, i, isSet)
	i = m.MemGetSetUint16(&e.active_rx_l3_interface_profile_index, b, i+9, i, isSet)
	i = m.MemGetSetUint8(&e.src_realm_id, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.tunnel_termination_ecn_decap_mapping_pointer, b, i+2, i, isSet)

	if i != 81 {
		panic("rx_l3_interface_entry: even-parity")
	}
}

type rx_l3_interface_mem m.MemElt

func (r *rx_l3_interface_mem) geta(q *DmaRequest, v *rx_l3_interface_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_l3_interface_mem) seta(q *DmaRequest, v *rx_l3_interface_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_l3_interface_mem) get(q *DmaRequest, v *rx_l3_interface_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *rx_l3_interface_mem) set(q *DmaRequest, v *rx_l3_interface_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type rx_l3_interface_profile_entry struct {
	// 127 => don't trust dscp; else index into DSCP_TABLE to get internal priority & color.
	trust_dscp_pointer uint8

	ip_urpf_mode
	ip_urpf_check_default_route bool

	allow_global_route_if_no_vrf_specific_route_found bool
	ip_icmp_redirect_to_cpu                           bool
	ip4_multicast_enable                              bool
	ip6_multicast_enable                              bool
	ip4_enable                                        bool
	ip6_enable                                        bool
	ip4_unknown_multicast_to_cpu                      bool
	ip6_unknown_multicast_to_cpu                      bool
	ip_copy_unresolved_src_to_cpu                     bool
	ip6_routing_header_with_type_0_drop               bool
	ip_multicast_remove_vlan_from_lookup_key_override bool
	ip_multicast_miss_as_l2_multicast                 bool
}

func (e *rx_l3_interface_profile_entry) MemBits() int { return 30 }

func (e *rx_l3_interface_profile_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSet1(&e.allow_global_route_if_no_vrf_specific_route_found, b, i, isSet)
	i = m.MemGetSetUint8(&e.trust_dscp_pointer, b, i+6, i, isSet)
	i = e.ip_urpf_mode.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.ip_urpf_check_default_route, b, i, isSet)
	i = m.MemGetSet1(&e.ip_icmp_redirect_to_cpu, b, i, isSet)
	i = m.MemGetSet1(&e.ip6_multicast_enable, b, i, isSet)
	i = m.MemGetSet1(&e.ip4_multicast_enable, b, i, isSet)
	i = m.MemGetSet1(&e.ip6_enable, b, i, isSet)
	i = m.MemGetSet1(&e.ip4_enable, b, i, isSet)
	i = m.MemGetSet1(&e.ip4_unknown_multicast_to_cpu, b, i, isSet)
	i = m.MemGetSet1(&e.ip6_unknown_multicast_to_cpu, b, i, isSet)
	i = m.MemGetSet1(&e.ip_copy_unresolved_src_to_cpu, b, i, isSet)
	i = m.MemGetSet1(&e.ip6_routing_header_with_type_0_drop, b, i, isSet)
	i = m.MemGetSet1(&e.ip_multicast_remove_vlan_from_lookup_key_override, b, i, isSet)
	i = m.MemGetSet1(&e.ip_multicast_miss_as_l2_multicast, b, i, isSet)
	if i != 22 {
		panic("rx_l3_interface_profile")
	}
}

type rx_l3_interface_profile_mem m.MemElt

func (r *rx_l3_interface_profile_mem) geta(q *DmaRequest, v *rx_l3_interface_profile_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_l3_interface_profile_mem) seta(q *DmaRequest, v *rx_l3_interface_profile_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_l3_interface_profile_mem) get(q *DmaRequest, v *rx_l3_interface_profile_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *rx_l3_interface_profile_mem) set(q *DmaRequest, v *rx_l3_interface_profile_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type tx_l3_interface_vlan_priority_select uint8

const (
	tx_l3_interface_vlan_priority_select_nop tx_l3_interface_vlan_priority_select = iota
	tx_l3_interface_vlan_priority_select_my_values
	tx_l3_interface_vlan_priority_select_via_mapping
)

type tx_l3_interface_vlan_priority_spec struct {
	// 0 => use .1p priority and cfi from packet
	// 1 => use values in this structure
	// 2 => map {int_pri, cng} using egr_mpls_exp_mapping_[12] tables.
	which          tx_l3_interface_vlan_priority_select
	cfi            bool
	dot1p_priority uint8
	mapping_index  uint8
}

func (x *tx_l3_interface_vlan_priority_spec) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSetUint8((*uint8)(&x.which), b, i+1, i, isSet)
	var v uint8
	if isSet {
		v = x.mapping_index
		if x.which == tx_l3_interface_vlan_priority_select_my_values {
			v = x.dot1p_priority
			if x.cfi {
				v |= 1 << 3
			}
		}
	}
	i = m.MemGetSetUint8(&v, b, i+3, i, isSet)
	if !isSet {
		if x.which == tx_l3_interface_vlan_priority_select_my_values {
			x.mapping_index = 0
			x.dot1p_priority = v & 7
			x.cfi = v&(1<<3) != 0
		} else {
			x.mapping_index = v
		}
	}

	return i
}

// Egress L3 Interface
type tx_l3_interface_entry struct {
	// Bits[12:2] are used to index EGR_IP_TUNNEL table. Bits[1:0] select one of 4 MPLS entries embedded in each table location.
	egr_ip_tunnel_index uint16

	ip_ttl_expired_threshold uint8
	ip_dscp
	ip_dscp_mapping_pointer uint8

	// 0 => nop, 1 => set from ip_dscp field, 2 => use mapping_pointer
	ip_dscp_select uint8

	efp_class_id uint16

	// Indicates if the packet needs to be:
	// only L2 Switched and only L2 modifications needs to be done
	// (corresponds to the l3_l2_only interface flag).
	l2_switch bool

	inner_vlan, outer_vlan struct {
		id            m.Vlan
		priority_spec tx_l3_interface_vlan_priority_spec
	}
	// 0 do not modify, 1 replace, 2 delete
	inner_vlan_present_action uint8
	// add if missing inner vlan
	inner_vlan_absent_add bool

	src_ethernet_address m.EthernetAddress
}

func (e *tx_l3_interface_entry) MemBits() int { return 144 }
func (e *tx_l3_interface_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint16(&e.egr_ip_tunnel_index, b, i+12, i, isSet)
	i = e.outer_vlan.id.MemGetSet(b, i, isSet)
	i = e.src_ethernet_address.MemGetSet(b, i, isSet)
	i = e.inner_vlan.id.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_ttl_expired_threshold, b, i+7, i, isSet)
	i = e.ip_dscp.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_dscp_mapping_pointer, b, i+7, i, isSet)
	i = e.outer_vlan.priority_spec.MemGetSet(b, i, isSet)
	i = e.inner_vlan.priority_spec.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.l2_switch, b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_dscp_select, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.inner_vlan_present_action, b, i+1, i, isSet)
	i = m.MemGetSet1(&e.inner_vlan_absent_add, b, i, isSet)
	i = m.MemGetSetUint16(&e.efp_class_id, b, i+11, i, isSet)
}

type tx_l3_interface_mem m.MemElt

func (r *tx_l3_interface_mem) geta(q *DmaRequest, v *tx_l3_interface_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_l3_interface_mem) seta(q *DmaRequest, v *tx_l3_interface_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_l3_interface_mem) get(q *DmaRequest, v *tx_l3_interface_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_l3_interface_mem) set(q *DmaRequest, v *tx_l3_interface_entry) {
	r.seta(q, v, sbus.Duplicate)
}

const n_vrf = 2 << 10

type vrf_entry struct {
	pipe_counter_ref rx_pipe_4p12i_pipe_counter_ref
}

func (e *vrf_entry) MemBits() int { return 25 }

func (e *vrf_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.pipe_counter_ref.MemGetSet(b, i, isSet)
}

type vrf_mem m.MemElt

func (r *vrf_mem) geta(q *DmaRequest, v *vrf_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *vrf_mem) seta(q *DmaRequest, v *vrf_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *vrf_mem) get(q *DmaRequest, v *vrf_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *vrf_mem) set(q *DmaRequest, v *vrf_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type l3_interface struct {
	// Index in rx_l3_interface/tx_l3_interface tables.
	ref pipemem.Ref
	rx  rx_l3_interface_entry
	tx  tx_l3_interface_entry
}

func (i *l3_interface) index() uint { return i.ref.Index() }

//go:generate gentemplate -d Package=fe1a -id l3_interface -d VecType=l3_interface_vec -d Type=l3_interface github.com/platinasystems/go/elib/vec.tmpl

type l3_interface_main struct {
	pool          pipemem.Pool
	if_by_si      l3_interface_vec
	si_by_counter [n_pipe]elib.Uint32Vec
}

func (t *fe1a) l3_interface_init() {
	im := &t.l3_interface_main
	q := t.getDmaReq()

	// Setup a default rx_l3_interface profile entry to be shared amongst interfaces
	index := 0
	profile := rx_l3_interface_profile_entry{
		allow_global_route_if_no_vrf_specific_route_found: true,
		trust_dscp_pointer:                                0xff,
		ip_urpf_mode:                                      0,
		ip_urpf_check_default_route:                       false,
		ip_icmp_redirect_to_cpu:                           true,
		ip_copy_unresolved_src_to_cpu:                     false,
		ip_multicast_remove_vlan_from_lookup_key_override: false,
		ip_multicast_miss_as_l2_multicast:                 false,
		ip4_enable:                                        true,
		ip6_enable:                                        true,
		ip4_multicast_enable:                              true,
		ip6_multicast_enable:                              true,
		ip4_unknown_multicast_to_cpu:                      true,
		ip6_unknown_multicast_to_cpu:                      true,
		ip6_routing_header_with_type_0_drop:               true,
	}
	t.rx_pipe_mems.l3_interface_profile[index].set(q, &profile)
	q.Do()

	im.pool.Init(n_pipe, n_l3_interface)

	// Allocate index 0 in each pipe which hardware uses as invalid.
	if ref, ok := im.pool.Get(1<<n_pipe - 1); !ok || ref.Index() != 0 {
		panic("alloc invalid index")
	}
}

const (
	rx_if_counter = iota
)

func (p *Port) GetSwInterfaceCounterNames() (nm vnet.InterfaceCounterNames) {
	// Unfortunately there is no way to hook up a pipe counter to an egress l3 interface.
	// So, we only count rx packet/bytes for l3 interfaces.
	nm.Combined = []string{
		rx_if_counter: rx_counter_prefix + "l3 interface",
	}
	return
}

func (t *fe1a) swIfCounterSync(v *vnet.Vnet) {
	t.update_pool_counter_values(pipe_counter_pool_rx_l3_interface, BlockRxPipe)
	p := &t.pipe_counter_main.rx_pipe.pools[pipe_counter_pool_rx_l3_interface]
	im := &t.l3_interface_main
	th := v.GetIfThread(0)
	p.mu.Lock()
	p.Pool.Foreach(func(pipe, index uint) {
		c := &p.counters[pipe][index]
		if c.value.Packets != 0 {
			kind := v.HwSwCombinedIfCounter(rx_if_counter)
			kind.Add64(th, vnet.Si(im.si_by_counter[pipe][index]), c.value.Packets, c.value.Bytes)
			c.value.Zero()
		}
	})
	p.mu.Unlock()
}

func (t *fe1a) l3_interface_for_si(si vnet.Si) *l3_interface {
	return &t.l3_interface_main.if_by_si[si]
}

func (t *fe1a) swIfAddDel(v *vnet.Vnet, si vnet.Si, isDel bool) (err error) {
	port := port_for_si(v, si)
	if port == nil {
		return
	}

	im := &t.l3_interface_main

	im.if_by_si.Validate(uint(si))
	i := &im.if_by_si[si]

	if isDel {
		if i.rx.pipe_counter_ref.is_valid() {
			i.rx.pipe_counter_ref.free(t, BlockRxPipe)
		}
		im.pool.Put(i.ref)
		*i = l3_interface{}
		return
	}

	pipe_port := port.physical_port_number.toPipe()
	pipe := uint(port.physical_port_number.pipe())
	pipe_mask := uint(1) << pipe
	var ok bool
	if i.ref, ok = im.pool.Get(pipe_mask); !ok {
		err = errors.New("l3 interface overflow")
		return
	}
	_, pipe_ok := i.rx.pipe_counter_ref.alloc(t, pipe_counter_pool_rx_l3_interface, pipe_mask, BlockRxPipe)
	if pipe_ok {
		index := uint(i.rx.pipe_counter_ref.index)
		im.si_by_counter[pipe].Validate(index)
		im.si_by_counter[pipe][index] = uint32(si)
	}

	i.rx.rx_l3_interface_profile_index = 0
	// i.rx.vrf = 0 // fill in with fib index
	i.tx.outer_vlan.id = m.Vlan(v.SwIf(si).Id(v))
	i.tx.src_ethernet_address = m.EthernetAddress(port.EthernetAddress())

	q := t.getDmaReq()
	ri := i.ref.Index()

	// For sup-interfaces initialize source trunk map.
	if v.SupSi(si) == si {
		e := source_trunk_map_entry{
			index: uint16(ri),
		}
		e.lport_profile_index = uint8(pipe_port)
		t.rx_pipe_mems.source_trunk_map[pipe_port].set(q, &e)
	}

	t.rx_pipe_mems.l3_interface[ri].seta(q, &i.rx, sbus.Unique(pipe))
	// large mtu => no drops
	t.rx_pipe_mems.l3_interface_mtu[m.Unicast][ri].Set(&q.DmaRequest, BlockRxPipe, sbus.Unique(pipe), 0x3fff)
	t.tx_pipe_mems.l3_interface[ri].seta(q, &i.tx, sbus.Unique(pipe))
	q.Do()

	return
}
