// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

// 32k next hops in ing_l3_next_hop egr_l3_next_hop tables.
const n_next_hop = 1 << 15

type next_hop_index uint32

// Entry 0 is used to mark invalid next hops and should not be used.
const invalid_next_hop_index next_hop_index = 0

type rx_next_hop_type uint8

const (
	rx_next_hop_type_unicast  rx_next_hop_type = iota // index is vlan
	rx_next_hop_type_tunnel                           // index is l3_oif
	rx_next_hop_type_l2_dvp                           // index is mtu_size
	rx_next_hop_type_gpon_dvp                         // index is dst virtual port resolution info
)

type rx_next_hop_entry struct {
	rx_next_hop_type

	// vlan, l3_oif, mtu or dvp resolution info
	index uint16

	m.LogicalPort

	drop         bool
	copy_to_cpu  bool
	dst_realm_id uint8

	eh_queue_tag_type uint8
	eh_queue_tag      uint32
}

func (e *rx_next_hop_entry) MemBits() int { return 61 }
func (e *rx_next_hop_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint8((*uint8)(&e.rx_next_hop_type), b, i+1, i, isSet)
	i = m.MemGetSetUint16(&e.index, b, i+13, i, isSet)
	i += 1 // skip parity bit 0
	i = e.LogicalPort.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.drop, b, i, isSet)
	i = m.MemGetSet1(&e.copy_to_cpu, b, i, isSet)
	i = m.MemGetSetUint8(&e.dst_realm_id, b, i+1, i, isSet)
	i = m.MemGetSetUint32(&e.eh_queue_tag, b, i+16, i, isSet)
	i = m.MemGetSetUint8(&e.eh_queue_tag_type, b, i+1, i, isSet)
}

type rx_next_hop_mem m.MemElt

func (r *rx_next_hop_mem) geta(q *DmaRequest, v *rx_next_hop_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_next_hop_mem) seta(q *DmaRequest, v *rx_next_hop_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_next_hop_mem) get(q *DmaRequest, v *rx_next_hop_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *rx_next_hop_mem) set(q *DmaRequest, v *rx_next_hop_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type tx_next_hop_type uint8

const (
	tx_next_hop_type_l3_unicast tx_next_hop_type = iota // index is vlan
	tx_next_hop_type_mpls_mac_da_profile
	tx_next_hop_type_mpls_sd_tag_actions
	tx_next_hop_type_mac_in_mac
	tx_next_hop_type_l2_actions
	tx_next_hop_type_normal_proxy
	tx_next_hop_type_ifp_actions
	tx_next_hop_type_l3_multicast
)

type tx_next_hop_entry interface {
	Type() tx_next_hop_type
	MemGetSet(b []uint32, i int, isSet bool)
}

type tx_next_hop_entry_wrapper struct {
	entry tx_next_hop_entry
}

func (e *tx_next_hop_entry_wrapper) MemBits() int { return 144 }
func (e *tx_next_hop_entry_wrapper) MemGetSet(b []uint32, isSet bool) {
	var t tx_next_hop_type
	if isSet {
		t = e.entry.Type()
	}
	i := 0
	i = m.MemGetSetUint8((*uint8)(&t), b, i+3, i, isSet)
	if !isSet {
		switch t {
		case tx_next_hop_type_l3_unicast:
			e.entry = &l3_unicast_tx_next_hop{}
		case tx_next_hop_type_mpls_mac_da_profile:
			e.entry = &mpls_tx_next_hop{}
		default:
			panic(e)
		}
	}
	e.entry.MemGetSet(b, i, isSet)
}

type tx_next_hop_mem m.MemElt

func (r *tx_next_hop_mem) geta(q *DmaRequest, t sbus.AccessType) (v tx_next_hop_entry) {
	g := tx_next_hop_entry_wrapper{}
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, &g, BlockTxPipe, t)
	return g.entry
}
func (r *tx_next_hop_mem) seta(q *DmaRequest, v tx_next_hop_entry, t sbus.AccessType) {
	g := tx_next_hop_entry_wrapper{entry: v}
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, &g, BlockTxPipe, t)
}
func (r *tx_next_hop_mem) get(q *DmaRequest) (v tx_next_hop_entry) { return r.geta(q, sbus.Duplicate) }
func (r *tx_next_hop_mem) set(q *DmaRequest, v tx_next_hop_entry)  { r.seta(q, v, sbus.Duplicate) }

type l3_unicast_tx_next_hop struct {
	disable_dst_ethernet_address_rewrite bool
	disable_src_ethernet_address_rewrite bool

	disable_l3_unicast_vlan_rewrite bool
	disable_ip_ttl_decrement        bool

	etag_de  bool
	etag_pcp uint8

	efp_class_id uint8

	dst_ethernet_address m.EthernetAddress

	l3_intf_index uint16

	pipe_counter_ref tx_pipe_pipe_counter_ref

	vntag_p_bit       bool
	vntag_force_l_bit bool
	vntag_action
	vntag_dst_virtual_interface uint16

	dst_virtual_port_valid bool
	dst_virtual_port       uint16

	hi_gig_2_mode              bool
	hi_gig_vntag_modify_enable bool
}

func (e *l3_unicast_tx_next_hop) Type() tx_next_hop_type { return tx_next_hop_type_l3_unicast }
func (e *l3_unicast_tx_next_hop) MemGetSet(b []uint32, i int, isSet bool) {
	i = m.MemGetSetUint16(&e.l3_intf_index, b, i+12, i, isSet)
	i = e.dst_ethernet_address.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.disable_dst_ethernet_address_rewrite, b, i, isSet)
	i = m.MemGetSet1(&e.disable_ip_ttl_decrement, b, i, isSet)
	i = m.MemGetSet1(&e.disable_l3_unicast_vlan_rewrite, b, i, isSet)
	i = m.MemGetSetUint8(&e.etag_pcp, b, i+2, i, isSet)
	i = m.MemGetSet1(&e.etag_de, b, i, isSet)
	i = m.MemGetSetUint16(&e.vntag_dst_virtual_interface, b, i+13, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_2_mode, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_vntag_modify_enable, b, i, isSet)
	i = m.MemGetSetUint16(&e.dst_virtual_port, b, i+14, i, isSet)
	i = m.MemGetSet1(&e.dst_virtual_port_valid, b, i, isSet)

	if i != 104 {
		panic("104")
	}

	i = e.pipe_counter_ref.MemGetSet(b, i, isSet)
	i = e.vntag_action.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.vntag_p_bit, b, i, isSet)
	i = m.MemGetSet1(&e.vntag_force_l_bit, b, i, isSet)
	i = m.MemGetSet1(&e.disable_src_ethernet_address_rewrite, b, i, isSet)
	i = m.MemGetSetUint8(&e.efp_class_id, b, i+6, i, isSet)

	if i != 135 {
		panic("135")
	}
}
