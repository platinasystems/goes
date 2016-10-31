// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

type tunnel_type uint8

const (
	tunnel_type_none tunnel_type = iota // for entry with index 0
	tunnel_type_ip4
	tunnel_type_ip6
	tunnel_type_mpls
)

func (t *tunnel_type) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(t), b, i+1, i, isSet)
}

type ip_tunnel_type uint8

const (
	ip_tunnel_type_ip_in_ip4 ip_tunnel_type = iota
	ip_tunnel_type_6to4
	ip_tunnel_type_isatap
	ip_tunnel_type_6to4_secure
	ip_tunnel_type_gre
	ip_tunnel_type_pim_sim_dr1
	ip_tunnel_type_pim_sim_dr2
	ip_tunnel_type_l2_gre
	_
	_
	ip_tunnel_type_amt
	ip_tunnel_type_vxlan
)

func (t *ip_tunnel_type) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(t), b, i+3, i, isSet)
}

type ip_tunnel_dont_fragment_bit_setting uint8

const (
	ip_tunnel_dont_fragment_bit_set_to_0 ip_tunnel_dont_fragment_bit_setting = iota
	ip_tunnel_dont_fragment_bit_set_to_1
	ip_tunnel_dont_fragment_bit_copy_from_inner_ip_header
	ip_tunnel_dont_fragment_bit_copy_from_inner_ip_header_3 // how different?
)

func (t *ip_tunnel_dont_fragment_bit_setting) ip4MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(t), b, i+1, i, isSet)
}
func (t *ip_tunnel_dont_fragment_bit_setting) ip6MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(t), b, i+0, i, isSet)
}

// 2 bit TPID to select one of 4 16 bit ethernet header types for vlan encapsulations.
type vlan_tpid uint8

func (t *vlan_tpid) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(t), b, i+1, i, isSet)
}

type vlan_priority_spec struct {
	// If set mapping pointer is valid; otherwise cfi/priority fields are used.
	dot1p_mapping_pointer_valid bool
	dot1p_mapping_pointer       uint8
	cfi                         bool
	priority                    uint8
}

func (t *vlan_priority_spec) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSet1(&t.dot1p_mapping_pointer_valid, b, i, isSet)
	var v uint8
	if isSet {
		if t.dot1p_mapping_pointer_valid {
			v = t.dot1p_mapping_pointer
		} else {
			v = t.priority << 1
			if t.cfi {
				v |= 1
			}
		}
	}
	i = m.MemGetSetUint8(&v, b, i+3, i, isSet)
	if !isSet {
		if t.dot1p_mapping_pointer_valid {
			t.dot1p_mapping_pointer = v
		} else {
			t.priority = (v >> 1) & 0x7
			t.cfi = v&1 != 0
		}
	}
	return i
}

type ip_dscp_select uint8

const (
	ip_dscp_from_this_spec ip_dscp_select = iota
	ip_dscp_from_packet
	ip_dscp_from_tx_dscp_table
)

type ip_dscp_spec struct {
	ip_dscp_select
	dscp                uint8
	tx_dscp_table_index uint8
}

func (t *ip_dscp_spec) MemGetSet(b []uint32, i int, isSet bool) int {
	var v uint8
	if isSet {
		switch t.ip_dscp_select {
		case ip_dscp_from_tx_dscp_table:
			v = t.tx_dscp_table_index
		case ip_dscp_from_this_spec:
			v = t.dscp
		}
	}
	i = m.MemGetSetUint8(&v, b, i+7, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&t.ip_dscp_select), b, i+1, i, isSet)
	if !isSet {
		switch t.ip_dscp_select {
		case ip_dscp_from_this_spec:
			t.dscp = v
		case ip_dscp_from_tx_dscp_table:
			t.tx_dscp_table_index = v
		}
	}
	return i
}

// Shared between ip4 & ip6.
type tx_ip_tunnel_entry_common struct {
	tunnel_type

	src_ethernet_address, dst_ethernet_address m.EthernetAddress
	vlan_tpid
	vlan_priority_spec
	vlan_add_header bool

	ip_tunnel_type
	l4_src_port, l4_dst_port m.IpPort
	ip_ttl                   uint8

	ip_dscp_spec
	ip_ecn_encap_mapping_pointer uint8
}

type tx_ip4_tunnel_entry struct {
	tx_ip_tunnel_entry_common

	ip_src_address, ip_dst_address       m.Ip4Address
	ip4_tunnel_dont_fragment_bit_setting ip_tunnel_dont_fragment_bit_setting
	ip6_tunnel_dont_fragment_bit_setting ip_tunnel_dont_fragment_bit_setting
}

func (e *tx_ip4_tunnel_entry) MemBits() int { return 244 }
func (e *tx_ip4_tunnel_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.tunnel_type.MemGetSet(b, i, isSet)
	i = e.ip_tunnel_type.MemGetSet(b, i, isSet)
	i = e.dst_ethernet_address.MemGetSet(b, i, isSet)
	i = e.src_ethernet_address.MemGetSet(b, i, isSet)
	i = e.vlan_tpid.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.vlan_add_header, b, i, isSet)
	i = e.vlan_priority_spec.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_ttl, b, i+7, i, isSet)
	i = e.ip4_tunnel_dont_fragment_bit_setting.ip4MemGetSet(b, i, isSet)
	i = e.ip6_tunnel_dont_fragment_bit_setting.ip6MemGetSet(b, i, isSet)
	i = e.ip_dscp_spec.MemGetSet(b, i, isSet)
	i = e.ip_src_address.MemGetSet(b, i, isSet)
	i = e.ip_dst_address.MemGetSet(b, i, isSet)
	i = e.l4_src_port.MemGetSet(b, i, isSet)
	i = e.l4_dst_port.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_ecn_encap_mapping_pointer, b, i+2, i, isSet)
}

type tx_ip4_tunnel_mem m.MemElt

func (r *tx_ip4_tunnel_mem) geta(q *DmaRequest, v *tx_ip4_tunnel_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_ip4_tunnel_mem) seta(q *DmaRequest, v *tx_ip4_tunnel_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_ip4_tunnel_mem) get(q *DmaRequest, v *tx_ip4_tunnel_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_ip4_tunnel_mem) set(q *DmaRequest, v *tx_ip4_tunnel_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type tx_ip6_tunnel_entry struct {
	tx_ip_tunnel_entry_common

	ip_src_address, ip_dst_address m.Ip6Address

	// Flow label either comes from ip_flow_label field in this
	// structure or from hash of payload if this bit is set.
	ip_flow_label_from_payload_hash bool
	ip_flow_label                   uint32
}

func (e *tx_ip6_tunnel_entry) MemBits() int { return 2 * 244 }
func (e *tx_ip6_tunnel_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.tunnel_type.MemGetSet(b, i, isSet)
	i = e.ip_tunnel_type.MemGetSet(b, i, isSet)
	i = e.dst_ethernet_address.MemGetSet(b, i, isSet)
	i = e.src_ethernet_address.MemGetSet(b, i, isSet)
	i = e.ip_src_address.MemGetSet(b, i, isSet)
	i = 244

	// Copy of tunnel type for second entry in pair.
	i = e.tunnel_type.MemGetSet(b, i, isSet)
	i = e.vlan_tpid.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.vlan_add_header, b, i, isSet)
	i = e.vlan_priority_spec.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint32(&e.ip_flow_label, b, i+19, i, isSet)
	i = e.ip_dscp_spec.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_ttl, b, i+7, i, isSet)
	i = e.ip_dst_address.MemGetSet(b, i, isSet)
	i = e.l4_src_port.MemGetSet(b, i, isSet)
	i = e.l4_dst_port.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ip_ecn_encap_mapping_pointer, b, i+2, i, isSet)
	i = m.MemGetSet1(&e.ip_flow_label_from_payload_hash, b, i, isSet)
}

type tx_ip6_tunnel_mem m.MemElt

func (r *tx_ip6_tunnel_mem) geta(q *DmaRequest, v *tx_ip6_tunnel_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_ip6_tunnel_mem) seta(q *DmaRequest, v *tx_ip6_tunnel_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_ip6_tunnel_mem) get(q *DmaRequest, v *tx_ip6_tunnel_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_ip6_tunnel_mem) set(q *DmaRequest, v *tx_ip6_tunnel_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type mpls_tunnel_exp_select uint8

const (
	mpls_tunnel_exp_from_this_spec mpls_tunnel_exp_select = iota
	mpls_tunnel_exp_from_mapping
	mpls_tunnel_exp_from_packet
)

type mpls_tunnel_exp_spec struct {
	mpls_tunnel_exp_select
	exp_mapping_pointer uint8
	exp                 uint8
	cfi                 bool
	priority            uint8
}

func (t *mpls_tunnel_exp_spec) MemGetSet(b []uint32, i int, isSet bool) int {
	var v uint8
	if isSet {
		switch t.mpls_tunnel_exp_select {
		case mpls_tunnel_exp_from_mapping:
			v = t.exp_mapping_pointer
		case mpls_tunnel_exp_from_this_spec:
			v = t.priority << 1
			if t.cfi {
				v |= 1
			}
		}
	}
	i = m.MemGetSetUint8((*uint8)(&t.mpls_tunnel_exp_select), b, i+1, i, isSet)
	i = m.MemGetSetUint8(&v, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&t.exp, b, i+2, i, isSet)
	if !isSet {
		switch t.mpls_tunnel_exp_select {
		case mpls_tunnel_exp_from_mapping:
			t.exp_mapping_pointer = v
		case mpls_tunnel_exp_from_this_spec:
			t.priority = (v >> 1) & 0xf
			t.cfi = v&1 != 0
		}
	}
	return i
}

type tx_mpls_tunnel_entry struct {
	// 20 bit mpls label
	label m.MplsLabel

	ttl uint8

	// Push 0 1 or 2 following labels.
	n_labels_to_push uint8

	mpls_tunnel_exp_spec
}

func (e *tx_mpls_tunnel_entry) MemGetSet(b []uint32, i int, isSet bool) int {
	i = e.label.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.n_labels_to_push, b, i+1, i, isSet)
	i = e.mpls_tunnel_exp_spec.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.ttl, b, i+7, i, isSet)
	return i
}

type tx_mpls_tunnel_4entries struct {
	tunnel_type

	entries [4]tx_mpls_tunnel_entry
}

func (e *tx_mpls_tunnel_4entries) MemBits() int { return 244 }
func (e *tx_mpls_tunnel_4entries) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.tunnel_type.MemGetSet(b, i, isSet)
	// skip unused bits
	i = 4
	for j := range e.entries {
		i = e.entries[j].MemGetSet(b, i, isSet)
	}
}

type tx_mpls_tunnel_mem m.MemElt

func (r *tx_mpls_tunnel_mem) geta(q *DmaRequest, v *tx_mpls_tunnel_4entries, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_mpls_tunnel_mem) seta(q *DmaRequest, v *tx_mpls_tunnel_4entries, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_mpls_tunnel_mem) get(q *DmaRequest, v *tx_mpls_tunnel_4entries) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_mpls_tunnel_mem) set(q *DmaRequest, v *tx_mpls_tunnel_4entries) {
	r.seta(q, v, sbus.Duplicate)
}
