// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"unsafe"
)

const (
	n_ifp_slice = 12
	// number of 80 bit entries per slice; total of 12*512 = 6144 entries per slice.
	n_ifp_tcam_elts_per_slice = 512
	// IFP supports up to 32 logical tables.
	n_ifp_logical_tables = 32
)

type ifp_tcam_80bit_key [10]uint8
type ifp_tcam_160bit_key [20]uint8

func (a *ifp_tcam_80bit_key) tcamEncode(b *ifp_tcam_80bit_key, isSet bool) (c, d ifp_tcam_80bit_key) {
	for i := range a {
		q, r := m.TcamUint8(a[i]).TcamEncode(m.TcamUint8(b[i]), isSet)
		c[i], d[i] = uint8(q), uint8(r)
	}
	return
}

func (a *ifp_tcam_160bit_key) tcamEncode(b *ifp_tcam_160bit_key, isSet bool) (c, d ifp_tcam_160bit_key) {
	for i := range a {
		q, r := m.TcamUint8(a[i]).TcamEncode(m.TcamUint8(b[i]), isSet)
		c[i], d[i] = uint8(q), uint8(r)
	}
	return
}

func (x *ifp_tcam_80bit_key) MemGetSet(b []uint32, i int, isSet bool) int {
	for i := range x {
		i = m.MemGetSetUint8(&x[i], b, i+7, i, isSet)
	}
	return i
}

func (x *ifp_tcam_160bit_key) MemGetSet(b []uint32, i int, isSet bool) int {
	for i := range x {
		i = m.MemGetSetUint8(&x[i], b, i+7, i, isSet)
	}
	return i
}

type ifp_tcam_80bit_entry struct {
	valid     bool
	key, mask ifp_tcam_80bit_key
}

type ifp_tcam_160bit_entry struct {
	valid     bool
	key, mask ifp_tcam_160bit_key
}

func (e *ifp_tcam_80bit_entry) MemBits() int { return 161 }
func (e *ifp_tcam_80bit_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	var key, mask ifp_tcam_80bit_key
	if isSet {
		key, mask = e.key.tcamEncode(&e.mask, isSet)
	}
	i = key.MemGetSet(b, i, isSet)
	i = mask.MemGetSet(b, i, isSet)
	if !isSet {
		e.key, e.mask = key.tcamEncode(&mask, isSet)
	}
}

func (e *ifp_tcam_160bit_entry) MemBits() int { return 322 }
func (e *ifp_tcam_160bit_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	// Two identical copies of valid bit.
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	var key, mask ifp_tcam_160bit_key
	if isSet {
		key, mask = e.key.tcamEncode(&e.mask, isSet)
	}
	i = key.MemGetSet(b, i, isSet)
	i = mask.MemGetSet(b, i, isSet)
	if !isSet {
		e.key, e.mask = key.tcamEncode(&mask, isSet)
	}
}

type ifp_tcam_80bit_mem m.MemElt
type ifp_tcam_160bit_mem m.MemElt

func (r *ifp_tcam_80bit_mem) geta(q *DmaRequest, e *ifp_tcam_80bit_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, e, BlockRxPipe, t)
}
func (r *ifp_tcam_80bit_mem) seta(q *DmaRequest, e *ifp_tcam_80bit_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, e, BlockRxPipe, t)
}
func (r *ifp_tcam_80bit_mem) get(q *DmaRequest, e *ifp_tcam_80bit_entry) {
	r.geta(q, e, sbus.Duplicate)
}
func (r *ifp_tcam_80bit_mem) set(q *DmaRequest, e *ifp_tcam_80bit_entry) {
	r.seta(q, e, sbus.Duplicate)
}

func (r *ifp_tcam_160bit_mem) geta(q *DmaRequest, e *ifp_tcam_160bit_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, e, BlockRxPipe, t)
}
func (r *ifp_tcam_160bit_mem) seta(q *DmaRequest, e *ifp_tcam_160bit_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, e, BlockRxPipe, t)
}
func (r *ifp_tcam_160bit_mem) get(q *DmaRequest, e *ifp_tcam_160bit_entry) {
	r.geta(q, e, sbus.Duplicate)
}
func (r *ifp_tcam_160bit_mem) set(q *DmaRequest, e *ifp_tcam_160bit_entry) {
	r.seta(q, e, sbus.Duplicate)
}

type drop_precedence uint8

const (
	drop_precedence_none drop_precedence = iota
	drop_precedence_green
	drop_precedence_yellow
	drop_precedence_red
)

func (x *drop_precedence) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type ifp_policy_profile_set_1 struct {
	change_internal_priority                   bool
	per_color_change_internal_congestion_state [n_packet_color]bool
	internal_priority
	per_color_internal_congestion_state [n_packet_color]internal_congestion_state
	per_color_opcode                    [n_packet_color]uint8
	per_color_cos_internal_priority     [n_packet_color]uint8
	per_color_drop_precedence           [n_packet_color]drop_precedence
}

func (x *ifp_policy_profile_set_1) MemGetSet(b []uint32, i int, isSet bool) int {
	var c packet_color
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = x.per_color_drop_precedence[c].MemGetSet(b, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = m.MemGetSetUint8(&x.per_color_cos_internal_priority[c], b, i+7, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = m.MemGetSetUint8(&x.per_color_opcode[c], b, i+3, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = x.per_color_internal_congestion_state[c].MemGetSet(b, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = m.MemGetSet1(&x.per_color_change_internal_congestion_state[c], b, i, isSet)
	}
	i = m.MemGetSet1(&x.change_internal_priority, b, i, isSet)
	i = x.internal_priority.MemGetSet(b, i, isSet)
	return i
}

type ifp_policy_profile_set_2 struct {
	per_color_ip_dscp_opcode [n_packet_color]uint8
	per_color_ip_dscp        [n_packet_color]ip_dscp

	per_color_change_ip_ecn_bits [n_packet_color]bool
	per_color_ip_ecn_bits        [n_packet_color]ip_ecn_bits

	per_color_dot1q_priority_opcode [n_packet_color]uint8
	per_color_dot1q_priority        [n_packet_color]dot1q_priority
}

func (x *ifp_policy_profile_set_2) MemGetSet(b []uint32, i int, isSet bool) int {
	var c packet_color
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = x.per_color_ip_ecn_bits[c].MemGetSet(b, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = m.MemGetSet1(&x.per_color_change_ip_ecn_bits[c], b, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = x.per_color_dot1q_priority[c].MemGetSet(b, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = m.MemGetSetUint8(&x.per_color_dot1q_priority_opcode[c], b, i+2, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		n := 1
		if c == packet_color_green {
			n = 2
		}
		i = m.MemGetSetUint8(&x.per_color_ip_dscp_opcode[c], b, i+n, i, isSet)
	}
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = x.per_color_ip_dscp[c].MemGetSet(b, i, isSet)
	}
	return i
}

const n_mirror = 4

type ifp_policy_mirror_set struct {
	mirror_enable [n_mirror]bool
	mtp_index     [n_mirror]uint8
}

func (x *ifp_policy_mirror_set) MemGetSet(b []uint32, i int, isSet bool) int {
	for j := 0; j < n_mirror; j++ {
		i = m.MemGetSetUint8(&x.mtp_index[j], b, i+1, i, isSet)
	}
	for j := 0; j < n_mirror; j++ {
		i = m.MemGetSet1(&x.mirror_enable[j], b, i, isSet)
	}
	return i
}

type ifp_policy_copy_to_cpu_set struct {
	copy_to_cpu_matched_rule uint8
	copy_to_cpu_opcode       [n_packet_color]uint8
}

func (x *ifp_policy_copy_to_cpu_set) MemGetSet(b []uint32, i int, isSet bool) int {
	var c packet_color
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = m.MemGetSetUint8(&x.copy_to_cpu_opcode[c], b, i+2, i, isSet)
	}
	i = m.MemGetSetUint8(&x.copy_to_cpu_matched_rule, b, i+7, i, isSet)
	return i
}

type ifp_policy_l3_switch_change_opcode uint8

const (
	ifp_policy_l3_switch_change_none ifp_policy_l3_switch_change_opcode = iota
	ifp_policy_l3_switch_l2_switch                                      // using next hop or ecmp
	ifp_policy_l3_switch_no_l2_switch
	ifp_policy_l3_switch_add_eh_tag
	ifp_policy_l3_switch_add_classification_tag
	_
	ifp_policy_l3_switch_l3_switch // using next hop or ecmp
	ifp_policy_l3_switch_no_l3_switch
	ifp_policy_l3_switch_class_id
	ifp_policy_l3_switch_bfd_session_index
)

func (x *ifp_policy_l3_switch_change_opcode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+3, i, isSet)
}

type ifp_policy_l3_switch_change_set struct {
	opcode ifp_policy_l3_switch_change_opcode

	ecmp_enable      bool
	ecmp_hash_select uint8

	// Either next hop index or ecmp group depending on ecmp_enable
	index uint32
}

func (x *ifp_policy_l3_switch_change_set) MemGetSet(b []uint32, i int, isSet bool) int {
	var v uint32

	if isSet {
		v = x.index
		if x.ecmp_enable {
			v |= uint32(x.ecmp_hash_select) << 11
			v |= 1 << 17
		}
	}
	i = m.MemGetSetUint32(&v, b, i+17, i, isSet)
	i = 19
	i = x.opcode.MemGetSet(b, i, isSet)
	if !isSet {
		switch x.opcode {
		case ifp_policy_l3_switch_l2_switch, ifp_policy_l3_switch_l3_switch:
			x.ecmp_enable = v&(1<<17) != 0
			if x.ecmp_enable {
				x.ecmp_hash_select = uint8((v >> 11) & 0x7)
				x.index = v & 0x3ff
			} else {
				x.index = v & 0x1ffff
			}
		}
	}
	return i
}

type ifp_policy_drop_opcode uint8

const (
	ifp_policy_drop_noop ifp_policy_drop_opcode = iota
	ifp_policy_drop_yes
	ifp_policy_drop_no
)

func (x *ifp_policy_drop_opcode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type ifp_policy_drop_set struct {
	drop_opcode [n_packet_color]ifp_policy_drop_opcode
}

func (x *ifp_policy_drop_set) MemGetSet(b []uint32, i int, isSet bool) int {
	var c packet_color
	for c = packet_color_red; c >= packet_color_green; c-- {
		i = x.drop_opcode[c].MemGetSet(b, i, isSet)
	}
	return i
}

type ifp_policy_redirect_opcode uint8

const (
	ifp_policy_redirect_noop ifp_policy_redirect_opcode = iota
	ifp_policy_redirect_unicast
	ifp_policy_redirect_cancel_from_lower_priority_tables
	ifp_policy_redirect_multicast
	ifp_policy_redirect_set_egress_port_bitmap
	ifp_policy_redirect_or_egress_port_bitmap
	ifp_policy_redirect_hi_gig_eh_modify
	_
)

func (x *ifp_policy_redirect_opcode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type ifp_policy_unicast_redirect_opcode uint8

const (
	// Index is 17 bit logical port
	ifp_policy_unicast_redirect_to_dst_logical_port ifp_policy_unicast_redirect_opcode = iota
	ifp_policy_unicast_redirect_to_dst_logical_port_with_original_packet
	// Index is next hop
	ifp_policy_unicast_redirect_via_next_hop
	// Index is ecmp group
	ifp_policy_unicast_redirect_via_ecmp
	_
	ifp_policy_unicast_redirect_via_ecmp_with_offset // ?
	// Index is dst virtual port
	ifp_policy_unicast_redirect_to_dst_virtual_port
)

func (x *ifp_policy_unicast_redirect_opcode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type ifp_policy_multicast_redirect_opcode uint8

const (
	// Index is into ifp_redirection_profile table.
	ifp_policy_multicast_redirect_replace_port_bitmap ifp_policy_multicast_redirect_opcode = iota
	ifp_policy_multicast_redirect_broadcast_to_vlan
	// Index is l2mc_index or l3mc_index
	ifp_policy_multicast_redirect_l2_multicast_index
	ifp_policy_multicast_redirect_l3_multicast_index
)

func (x *ifp_policy_multicast_redirect_opcode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type ifp_policy_redirect_set struct {
	opcode           ifp_policy_redirect_opcode
	unicast_opcode   ifp_policy_unicast_redirect_opcode
	multicast_opcode ifp_policy_multicast_redirect_opcode
	index            uint32
}

func (x *ifp_policy_redirect_set) MemGetSet(b []uint32, i int, isSet bool) int {
	i = x.opcode.MemGetSet(b, i, isSet)
	i = x.unicast_opcode.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint32(&x.index, b, i+16, i, isSet)
	i = x.multicast_opcode.MemGetSet(b, i, isSet)
	i += 13 // unused bits
	return i
}

type ifp_policy_counter_set struct {
	pipe_counter_ref          rx_pipe_4p11i_pipe_counter_ref
	per_color_offset_less_one [n_packet_color]uint8
}

func (x *ifp_policy_counter_set) MemGetSet(b []uint32, i int, isSet bool) int {
	var c packet_color
	for c = packet_color_green; c <= packet_color_red; c++ {
		i = m.MemGetSetUint8(&x.per_color_offset_less_one[c], b, i+1, i, isSet)
	}
	i = x.pipe_counter_ref.MemGetSet(b, i, isSet)
	return i
}

type ifp_policy_load_balancing_controls_set struct {
	disable_spray_ecmp         bool
	disable_spray_lag          bool
	disable_spray_hi_gig_trunk bool
}

func (x *ifp_policy_load_balancing_controls_set) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSet1(&x.disable_spray_hi_gig_trunk, b, i, isSet)
	i = m.MemGetSet1(&x.disable_spray_lag, b, i, isSet)
	i = m.MemGetSet1(&x.disable_spray_ecmp, b, i, isSet)
	return i
}

type ifp_meter_pair_mode uint8

const (
	ifp_meter_pair_mode_default ifp_meter_pair_mode = iota
	ifp_meter_pair_mode_flow
	ifp_meter_pair_mode_trtcm_color_blind
	ifp_meter_pair_mode_trtcm_color_aware
	_
	_
	ifp_meter_pair_mode_srtcm_color_blind
	ifp_meter_pair_mode_srtcm_color_aware
)

func (x *ifp_meter_pair_mode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type ifp_policy_meter_set struct {
	test, update [2]bool
	ifp_meter_pair_mode
	pair_mode_modifier bool
	pair_index         uint16
}

func (x *ifp_policy_meter_set) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSet1(&x.pair_mode_modifier, b, i, isSet)
	x.ifp_meter_pair_mode.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint16(&x.pair_index, b, i+9, i, isSet)
	for j := range x.test {
		i = m.MemGetSet1(&x.update[j], b, i, isSet)
		i = m.MemGetSet1(&x.test[j], b, i, isSet)
	}
	return i
}

type ifp_policy_nat_set struct {
	enable bool
	m.NatEditIndex
}

func (x *ifp_policy_nat_set) MemGetSet(b []uint32, i int, isSet bool) int {
	i = x.NatEditIndex.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.enable, b, i, isSet)
	return i
}

type ifp_policy_change_cpu_cos_set struct {
	opcode uint8
	cos    uint8
}

func (x *ifp_policy_change_cpu_cos_set) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSetUint8(&x.opcode, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&x.cos, b, i+5, i, isSet)
	return i
}

// Hits in IFP TCAM direct to corresponding policy table entry.
type ifp_policy_entry struct {
	nat_disable                     bool
	cut_through_disable             bool
	ip_urpf_check_disable           bool
	ip_ttl_decrement_disable        bool
	mirror_override                 bool
	green_to_pid                    bool
	sflow_sample_enable             bool
	instrumentation_triggers_enable bool

	profile_1               ifp_policy_profile_set_1
	profile_2               ifp_policy_profile_set_2
	mirror                  ifp_policy_mirror_set
	copy_to_cpu             ifp_policy_copy_to_cpu_set
	drop                    ifp_policy_drop_set
	l3_switch               ifp_policy_l3_switch_change_set
	redirect                ifp_policy_redirect_set
	counter                 ifp_policy_counter_set
	meter                   ifp_policy_meter_set
	nat                     ifp_policy_nat_set
	load_balancing_controls ifp_policy_load_balancing_controls_set
	change_cpu_cos          ifp_policy_change_cpu_cos_set
}

func (x *ifp_policy_entry) MemBits() int { return 303 }
func (x *ifp_policy_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = x.profile_1.MemGetSet(b, i, isSet)
	i = x.mirror.MemGetSet(b, i, isSet)
	i = x.load_balancing_controls.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.nat_disable, b, i, isSet)
	i = x.copy_to_cpu.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.cut_through_disable, b, i, isSet)
	i = m.MemGetSet1(&x.ip_urpf_check_disable, b, i, isSet)
	i = m.MemGetSet1(&x.ip_ttl_decrement_disable, b, i, isSet)
	i = 101 // skip ecc/parity bits

	i = x.profile_2.MemGetSet(b, i, isSet)
	i = x.l3_switch.MemGetSet(b, i, isSet)
	i = x.change_cpu_cos.MemGetSet(b, i, isSet)
	i = x.drop.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.mirror_override, b, i, isSet)
	i = m.MemGetSet1(&x.green_to_pid, b, i, isSet)
	i = m.MemGetSet1(&x.sflow_sample_enable, b, i, isSet)
	i = m.MemGetSet1(&x.instrumentation_triggers_enable, b, i, isSet)

	i = 202 // skip ecc/parity bits
	i = x.redirect.MemGetSet(b, i, isSet)
	i = x.counter.MemGetSet(b, i, isSet)
	i = x.meter.MemGetSet(b, i, isSet)
	i = x.nat.MemGetSet(b, i, isSet)
}

type ifp_policy_mem m.MemElt

func (r *ifp_policy_mem) geta(q *DmaRequest, v *ifp_policy_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_policy_mem) seta(q *DmaRequest, v *ifp_policy_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_policy_mem) get(q *DmaRequest, v *ifp_policy_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *ifp_policy_mem) set(q *DmaRequest, v *ifp_policy_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type ifp_packet_resolution uint8

const (
	ifp_packet_unknown           ifp_packet_resolution = 0
	ifp_packet_ethernet_control  ifp_packet_resolution = 1
	ifp_packet_oam               ifp_packet_resolution = 2
	ifp_packet_bfd               ifp_packet_resolution = 3
	ifp_packet_bpdu              ifp_packet_resolution = 4
	ifp_packet_icnm              ifp_packet_resolution = 5
	ifp_packet_ieee_1588_message ifp_packet_resolution = 6

	ifp_packet_known_l2_unicast     ifp_packet_resolution = 8
	ifp_packet_unknown_l2_unicast   ifp_packet_resolution = 9
	ifp_packet_known_l2_multicast   ifp_packet_resolution = 10
	ifp_packet_unknown_l2_multicast ifp_packet_resolution = 11
	ifp_packet_l2_broadcast         ifp_packet_resolution = 12

	ifp_packet_known_l3_unicast     ifp_packet_resolution = 16
	ifp_packet_unknown_l3_unicast   ifp_packet_resolution = 17
	ifp_packet_known_l3_multicast   ifp_packet_resolution = 18
	ifp_packet_unknown_l3_multicast ifp_packet_resolution = 19

	ifp_packet_known_l2_mpls  ifp_packet_resolution = 24
	ifp_packet_unknown_mpls   ifp_packet_resolution = 25
	ifp_packet_known_l3_mpls  ifp_packet_resolution = 26
	ifp_packet_known_mpls     ifp_packet_resolution = 28
	ifp_packet_mpls_multicast ifp_packet_resolution = 29

	ifp_packet_known_mac_in_mac   ifp_packet_resolution = 32
	ifp_packet_unknown_mac_in_mac ifp_packet_resolution = 33

	ifp_packet_known_trill   ifp_packet_resolution = 40
	ifp_packet_unknown_trill ifp_packet_resolution = 41

	ifp_packet_known_niv    ifp_packet_resolution = 48
	ifp_packet_unknown_niv  ifp_packet_resolution = 49
	ifp_packet_known_l2_gre ifp_packet_resolution = 50
	ifp_packet_known_vxlan  ifp_packet_resolution = 51
	ifp_packet_known_fcoe   ifp_packet_resolution = 52
	ifp_packet_unknown_fcoe ifp_packet_resolution = 53
)

func (x *ifp_packet_resolution) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+5, i, isSet)
}

type ifp_forwarding_type uint8

const (
	ifp_forwarding_type_l2_vlan_id ifp_forwarding_type = iota
	ifp_forwarding_type_l2_fid
	ifp_forwarding_type_l2_vfi
	ifp_forwarding_type_l2_point_to_point
	ifp_forwarding_type_l3_mpls // PHP on bottom-of-stack label
	ifp_forwarding_type_l3_vrf
	ifp_forwarding_type_vntag_etag
	ifp_forwarding_type_mpls // SWAP or PHP of non bottom-of-stack label
	ifp_forwarding_type_trill
	ifp_forwarding_type_none ifp_forwarding_type = 0xf
)

func (x *ifp_forwarding_type) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+3, i, isSet)
}

type ifp_tunnel_type uint8

const (
	ifp_tunnel_type_none                        ifp_tunnel_type = 0
	ifp_tunnel_type_ip                          ifp_tunnel_type = 1
	ifp_tunnel_type_mpls                        ifp_tunnel_type = 2
	ifp_tunnel_type_mac_in_mac                  ifp_tunnel_type = 3
	ifp_tunnel_type_amt                         ifp_tunnel_type = 6
	ifp_tunnel_type_trill                       ifp_tunnel_type = 7
	ifp_tunnel_type_l2_gre                      ifp_tunnel_type = 8
	ifp_tunnel_type_vxlan                       ifp_tunnel_type = 9
	ifp_tunnel_type_mac_in_mac_loopback         ifp_tunnel_type = 16
	ifp_tunnel_type_trill_network_port_loopback ifp_tunnel_type = 17
	ifp_tunnel_type_trill_access_port_loopback  ifp_tunnel_type = 18
	ifp_tunnel_type_qcn_loopback                ifp_tunnel_type = 23
	ifp_tunnel_type_vxlan_loopback              ifp_tunnel_type = 27
	ifp_tunnel_type_l2_gre_loopback             ifp_tunnel_type = 30
)

func (x *ifp_tunnel_type) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+4, i, isSet)
}

type ifp_l3_type uint8

const (
	ifp_l3_type_ip4_no_options            ifp_l3_type = 0
	ifp_l3_type_ip4_with_options          ifp_l3_type = 1
	ifp_l3_type_ip6_no_extension_headers  ifp_l3_type = 4
	ifp_l3_type_ip6_1_extension_header    ifp_l3_type = 5
	ifp_l3_type_ip6_gt1_extension_headers ifp_l3_type = 6
	ifp_l3_type_arp_request               ifp_l3_type = 8
	ifp_l3_type_arp_reply                 ifp_l3_type = 9
	ifp_l3_type_trill                     ifp_l3_type = 10
	ifp_l3_type_fcoe                      ifp_l3_type = 11
	ifp_l3_type_mpls_unicast              ifp_l3_type = 12
	ifp_l3_type_mpls_multicast            ifp_l3_type = 13
	ifp_l3_type_mac_in_mac                ifp_l3_type = 14
	ifp_l3_type_none                      ifp_l3_type = 31
)

func (x *ifp_l3_type) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+4, i, isSet)
}

type ifp_lookup_status_vector struct {
	mpls_bottom_of_stack_terminated bool
	mpls_entry_lookup_hit           [2]bool
	l3_tunnel_hit                   bool
	dos_attack                      bool
	unresolved_sa                   bool
	l3_defip_hit                    bool
	l3_entry_lookup_dst_hit         bool
	l3_entry_lookup_src_hit         bool
	l2_user_entry_lookup_hit        bool
	l2_lookup_dst_hit               bool
	l2_lookup_src_static_hit        bool
	l2_lookup_src_hit               bool
	l2_vlan_id_valid                bool
	l2_spanning_tree_state          spanning_tree_state

	// [0,1] => lookup 1,2 hit, [2] => lookup 1 or lookup 2 hit
	l2_vlan_translation_lookup_hit [3]bool
}

func (x *ifp_lookup_status_vector) MemGetSet(b []uint32, i int, isSet bool) int {
	for j := range x.l2_vlan_translation_lookup_hit {
		i = m.MemGetSet1(&x.l2_vlan_translation_lookup_hit[j], b, i, isSet)
	}
	i = m.MemGetSet1(&x.l2_vlan_id_valid, b, i, isSet)
	x.l2_spanning_tree_state.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.l2_lookup_src_hit, b, i, isSet)
	i = m.MemGetSet1(&x.l2_lookup_src_static_hit, b, i, isSet)
	i = m.MemGetSet1(&x.l2_lookup_dst_hit, b, i, isSet)
	i = m.MemGetSet1(&x.l2_user_entry_lookup_hit, b, i, isSet)
	i = m.MemGetSet1(&x.l3_entry_lookup_src_hit, b, i, isSet)
	i = m.MemGetSet1(&x.l3_entry_lookup_dst_hit, b, i, isSet)
	i = m.MemGetSet1(&x.l3_defip_hit, b, i, isSet)
	i = m.MemGetSet1(&x.unresolved_sa, b, i, isSet)
	i = m.MemGetSet1(&x.dos_attack, b, i, isSet)
	i = m.MemGetSet1(&x.l3_tunnel_hit, b, i, isSet)
	for j := range x.mpls_entry_lookup_hit {
		i = m.MemGetSet1(&x.mpls_entry_lookup_hit[j], b, i, isSet)
	}
	i = m.MemGetSet1(&x.mpls_bottom_of_stack_terminated, b, i, isSet)
	i += 1 // skip reserved bit
	return i
}

type ifp_logical_table_select_tcam_key struct {
	exact_match_lookup_performed           bool
	hg_lookup                              bool
	is_cpu_masquerade_or_visibility_packet bool
	is_mirror_packet                       bool
	is_hi_gig_packet                       bool
	hi_gig_lookup                          bool
	is_drop                                bool
	src_virtual_port_valid                 bool
	my_station_tcam_hit                    bool
	ip_l4_valid                            bool

	ifp_packet_resolution
	ifp_forwarding_type
	ifp_lookup_status_vector
	ifp_tunnel_type
	ifp_l3_type

	// 32 bit combination of classes configured via ifp_logical_table_select_config register.
	source_class uint32

	exact_match_logical_table_id [2]uint8
}

func (a *ifp_logical_table_select_tcam_key) tcamEncode(b *ifp_logical_table_select_tcam_key, isSet bool) (c, d ifp_logical_table_select_tcam_key) {
	panic("not yet")
	return
}

func (x *ifp_logical_table_select_tcam_key) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSetUint32(&x.source_class, b, i+31, i, isSet)
	i = m.MemGetSet1(&x.is_cpu_masquerade_or_visibility_packet, b, i, isSet)
	i = m.MemGetSet1(&x.is_mirror_packet, b, i, isSet)
	i = m.MemGetSet1(&x.is_hi_gig_packet, b, i, isSet)
	i = m.MemGetSet1(&x.is_drop, b, i, isSet)
	i = m.MemGetSet1(&x.hi_gig_lookup, b, i, isSet)
	i += 43 - 38
	i = x.ifp_forwarding_type.MemGetSet(b, i, isSet)
	i = x.ifp_lookup_status_vector.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.src_virtual_port_valid, b, i, isSet)
	i = m.MemGetSet1(&x.my_station_tcam_hit, b, i, isSet)
	i = x.ifp_tunnel_type.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&x.ip_l4_valid, b, i, isSet)
	i = x.ifp_l3_type.MemGetSet(b, i, isSet)
	i += 80 - 79 //  unused bit

	return i
}

type ifp_ipbm_source uint8

const (
	ifp_ipbm_source_rx_port ifp_ipbm_source = iota
	ifp_ipbm_source_source_trunk_map_table
	ifp_ipbm_source_src_virtual_port_table
	ifp_ipbm_source_port_table
)

func (x *ifp_ipbm_source) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type ifp_aux_ab_select uint8

const (
	ifp_aux_ab_select_assigned_vntag_etag ifp_aux_ab_select = iota
	ifp_aux_ab_select_rx_cntag
	ifp_aux_ab_select_hi_gig_eh_tag
	ifp_aux_ab_select_mpls_label
	ifp_aux_ab_select_mpls_control_word
	ifp_aux_ab_select_rtag7_hash_a
	ifp_aux_ab_select_rtag7_hash_b
	ifp_aux_ab_select_vxlan_flags
	ifp_aux_ab_select_vxlan_reserved_1_and_2
	ifp_aux_ab_select_ip6_flow_label
	ifp_aux_ab_select_bfd_your_discriminator
)

func (x *ifp_aux_ab_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+3, i, isSet)
}

type ifp_aux_cd_select uint8

const (
	ifp_aux_cd_select_compressed_ip_protocol_l4_ports ifp_aux_cd_select = iota
	ifp_aux_cd_select_ip_ttl_and_tos
	ifp_aux_cd_select_total_packet_length
	ifp_aux_cd_select_rtag7_hash_a_31_16
	ifp_aux_cd_select_rtag7_hash_a_15_0
	ifp_aux_cd_select_rtag7_hash_b_31_16
	ifp_aux_cd_select_rtag7_hash_b_15_0
	ifp_aux_cd_select_fcoe_fields
)

func (x *ifp_aux_cd_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+3, i, isSet)
}

type ifp_class_id_ab_select uint8

const (
	ifp_class_id_ab_from_source_trunk_map ifp_class_id_ab_select = iota
	ifp_class_id_ab_from_l3_iif
	ifp_class_id_ab_select_src_virtual_port
	ifp_class_id_ab_select_vlan
	ifp_class_id_ab_select_vfp_lo
	ifp_class_id_ab_select_vfp_hi
	ifp_class_id_ab_select_port_ifp_vfp
	ifp_class_id_ab_select_udf
)

func (x *ifp_class_id_ab_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type ifp_class_id_cd_select uint8

const (
	ifp_class_id_cd_select_from_l2_src_lookup ifp_class_id_cd_select = iota
	ifp_class_id_cd_select_from_l2_dst_lookup
	ifp_class_id_cd_select_from_l3_src_lookup
	ifp_class_id_cd_select_from_l3_dst_lookup
	ifp_class_id_cd_select_from_vfp
	ifp_class_id_cd_select_from_l2_lookup
	ifp_class_id_cd_select_from_l3_lookup
)

func (x *ifp_class_id_cd_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type ifp_src_select uint8

const (
	ifp_src_from_src_virtual_port_if_defined_else_src_physical_port ifp_src_select = iota
	ifp_src_from_src_logical_port
	ifp_src_from_src_virtual_port
)

func (x *ifp_src_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type ifp_src_dst_select uint8

const (
	ifp_src_dst_from_src_physical_port ifp_src_dst_select = iota
	ifp_src_dst_from_src_logical_port
	ifp_src_dst_from_src_virtual_port
	_
	ifp_src_dst_from_dst_logical_port
	ifp_src_dst_from_dst_virtual_port_lag
	ifp_src_dst_from_dst_virtual_port
	ifp_src_dst_from_ecmp_1
	ifp_src_dst_from_ecmp_2
	ifp_src_dst_from_next_hop
	ifp_src_dst_from_ip_multicast_group
	ifp_src_dst_from_l2_multicast_group
)

func (x *ifp_src_dst_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

// For TCP flags, IP TOS byte and TTL
type ifp_misc_func_select uint8

const (
	ifp_misc_func_select_0 ifp_misc_func_select = iota
	ifp_misc_func_select_1
	ifp_misc_func_select_from_packet
)

func (x *ifp_misc_func_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

// Result data when logical table select tcam matches.
type ifp_logical_table_select_tcam_result struct {
	enable bool

	// Normalize ip address and ports such that a < b
	normalize_l3_addresses_and_ports bool
	// Same for l2 addresses.
	normalize_l2_addresses bool

	key_generation_profile_index uint8
	table_id                     uint8
	class_id                     uint8

	// 0 => 1 slice, 1 => first slice, 2 => 2nd of 2, 3 => 2nd of 3, 4 => 3rd of 3 slices
	multi_entry_mode uint8

	insert_ipbm_in_key bool
	ifp_ipbm_source

	aux_ab_select      [2]ifp_aux_ab_select
	aux_cd_select      [2]ifp_aux_cd_select
	class_id_ab_select [2]ifp_class_id_ab_select
	class_id_cd_select [2]ifp_class_id_cd_select
	src_select         [2]ifp_src_select
	src_dst_select     [2]ifp_src_dst_select

	ip_ttl_select    ifp_misc_func_select
	ip_tos_select    ifp_misc_func_select
	tcp_flags_select ifp_misc_func_select
}

func (x *ifp_logical_table_select_tcam_result) MemGetSet(b []uint32, i int, isSet bool) int {
	i = m.MemGetSet1(&x.enable, b, i, isSet)
	i = m.MemGetSetUint8(&x.key_generation_profile_index, b, i+5, i, isSet)
	i = m.MemGetSetUint8(&x.table_id, b, i+4, i, isSet)
	i = m.MemGetSetUint8(&x.class_id, b, i+1, i, isSet)
	i = m.MemGetSet1(&x.insert_ipbm_in_key, b, i, isSet)
	i = x.ifp_ipbm_source.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&x.multi_entry_mode, b, i+2, i, isSet)
	i = m.MemGetSet1(&x.normalize_l3_addresses_and_ports, b, i, isSet)
	i = m.MemGetSet1(&x.normalize_l2_addresses, b, i, isSet)
	for j := range x.aux_ab_select {
		i = x.aux_ab_select[j].MemGetSet(b, i, isSet)
	}
	i = x.ip_ttl_select.MemGetSet(b, i, isSet)
	i = x.ip_tos_select.MemGetSet(b, i, isSet)
	i = x.tcp_flags_select.MemGetSet(b, i, isSet)
	for j := range x.class_id_ab_select {
		i = x.class_id_ab_select[j].MemGetSet(b, i, isSet)
	}
	for j := range x.class_id_cd_select {
		i = x.class_id_cd_select[j].MemGetSet(b, i, isSet)
	}
	// Note: order is b/a unlike others
	for j := 1; j >= 0; j-- {
		i = x.src_select[j].MemGetSet(b, i, isSet)
	}
	for j := range x.src_dst_select {
		i = x.src_dst_select[j].MemGetSet(b, i, isSet)
	}
	for j := range x.aux_cd_select {
		i = x.aux_cd_select[j].MemGetSet(b, i, isSet)
	}
	i += 1 // skip parity bit
	return i
}

type ifp_logical_table_select_entry struct {
	ifp_logical_table_select_tcam_only_entry
	ifp_logical_table_select_data_only_entry
}

func (e *ifp_logical_table_select_entry) MemBits() int { return 262 }
func (e *ifp_logical_table_select_entry) MemGetSet(b []uint32, isSet bool) {
	i := m.MemGetSet1(&e.valid, b, 0, isSet)
	var key, mask ifp_logical_table_select_tcam_key
	if isSet {
		key, mask = e.key.tcamEncode(&e.mask, isSet)
	}
	i = key.MemGetSet(b, i, isSet)
	i = mask.MemGetSet(b, i, isSet)
	if !isSet {
		e.key, e.mask = key.tcamEncode(&mask, isSet)
	}
	i = e.result.MemGetSet(b, i, isSet)
}

type ifp_logical_table_select_mem m.MemElt

func (r *ifp_logical_table_select_mem) geta(q *DmaRequest, v *ifp_logical_table_select_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_logical_table_select_mem) seta(q *DmaRequest, v *ifp_logical_table_select_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_logical_table_select_mem) get(q *DmaRequest, v *ifp_logical_table_select_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *ifp_logical_table_select_mem) set(q *DmaRequest, v *ifp_logical_table_select_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type ifp_logical_table_select_data_only_entry struct {
	result ifp_logical_table_select_tcam_result
}

func (e *ifp_logical_table_select_data_only_entry) MemBits() int { return 69 }
func (e *ifp_logical_table_select_data_only_entry) MemGetSet(b []uint32, isSet bool) {
	e.result.MemGetSet(b, 0, isSet)
}

type ifp_logical_table_select_data_only_mem m.MemElt

func (r *ifp_logical_table_select_data_only_mem) geta(q *DmaRequest, v *ifp_logical_table_select_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_logical_table_select_data_only_mem) seta(q *DmaRequest, v *ifp_logical_table_select_data_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_logical_table_select_data_only_mem) get(q *DmaRequest, v *ifp_logical_table_select_data_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *ifp_logical_table_select_data_only_mem) set(q *DmaRequest, v *ifp_logical_table_select_data_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type ifp_logical_table_select_tcam_only_entry struct {
	valid     bool
	key, mask ifp_logical_table_select_tcam_key
}

func (e *ifp_logical_table_select_tcam_only_entry) MemBits() int { return 193 }
func (e *ifp_logical_table_select_tcam_only_entry) MemGetSet(b []uint32, isSet bool) {
	i := m.MemGetSet1(&e.valid, b, 0, isSet)
	i = e.key.MemGetSet(b, i, isSet)
	i = e.mask.MemGetSet(b, i, isSet)
}

type ifp_logical_table_select_tcam_only_mem m.MemElt

func (r *ifp_logical_table_select_tcam_only_mem) geta(q *DmaRequest, v *ifp_logical_table_select_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_logical_table_select_tcam_only_mem) seta(q *DmaRequest, v *ifp_logical_table_select_tcam_only_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ifp_logical_table_select_tcam_only_mem) get(q *DmaRequest, v *ifp_logical_table_select_tcam_only_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *ifp_logical_table_select_tcam_only_mem) set(q *DmaRequest, v *ifp_logical_table_select_tcam_only_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type ifp_field byte
type ifp_u5 ifp_field
type ifp_u4 ifp_field
type ifp_u3 ifp_field
type ifp_u2 ifp_field
type ifp_u1 ifp_field
type ifp_u0 ifp_field

func get_ifp_field_extractor_l0_bus() *ifp_field_extractor_l0_bus {
	return (*ifp_field_extractor_l0_bus)(m.RegsBasePointer)
}

func (x *ifp_field) offset() uint { return uint(uintptr(unsafe.Pointer(x)) - m.RegsBaseAddress) }
func (x *ifp_u5) offset() uint    { return (*ifp_field)(x).offset() }
func (x *ifp_u4) offset() uint    { return (*ifp_field)(x).offset() }
func (x *ifp_u3) offset() uint    { return (*ifp_field)(x).offset() }
func (x *ifp_u2) offset() uint    { return (*ifp_field)(x).offset() }
func (x *ifp_u1) offset() uint    { return (*ifp_field)(x).offset() }

type ifp_field_extractor_l0_bus struct {
	// Level 0 32 bit input bus.
	u32 struct {
		aux_ab struct{ b, a ifp_u5 }

		ethernet_address_31_0 struct{ src, dst ifp_u5 }

		itag_isid_vpnid_vnid ifp_u5

		udf_chunks_0_5 [2][3]ifp_u5

		// ip4 address is always index 0.
		ip_address struct{ src, dst [4]ifp_u5 }

		_ [32 - 19]ifp_u5
	}

	// Level 0 16 bit input bus.
	u16 struct {
		aux_cd                    struct{ d, c ifp_u4 }
		aux_ab                    struct{ b, a [2]ifp_u4 }
		udf_6_7                   [2][2]ifp_u4
		src_ab_15_0               struct{ a, b ifp_u4 }
		class_id_cd               struct{ d, c ifp_u4 }
		class_id_ab               struct{ b, a ifp_u4 }
		exact_match_lookup_status ifp_u4
		forwarding_type           ifp_u4
		l3_iif                    ifp_u4

		// 36 bit output from src, dst compression tcam lookups.
		compression_tcam_result_31_0 struct{ dst, src [2]ifp_u4 }

		// 32 bit output from 32 entries of ifp_range_check table.
		range_check_results [2]ifp_u4

		vlan_tag               struct{ outer, inner ifp_u4 }
		l4_ports               struct{ src, dst ifp_u4 }
		ethernet_type          ifp_u4
		ethernet_address_47_32 struct{ src, dst ifp_u4 }
	}

	u8 struct {
		rx_physical_port ifp_u3
		vfi_7_0          ifp_u3
		vrf_7_0          ifp_u3
		vlan_id_7_0      struct{ inner, outer ifp_u3 }
		// [7:6] class id, [4:0] logical table id
		table_id_class_id   ifp_u3
		class_id_7_0        struct{ d, c, b, a ifp_u3 }
		tunnel_type         ifp_u3
		packet_resolution   ifp_u3
		ip_tos_fn           ifp_u3
		ip_ttl_fn           ifp_u3
		ip_next_header_last ifp_u3
		ip_next_header_2    ifp_u3
		ip_next_header_1    ifp_u3
		ip_first_subcode    ifp_u3
		iq_bus_7_0          ifp_u3
		spare               ifp_u3
		_                   [32 - 20]ifp_u3
	}

	u4 struct {
		vfi_11_8             ifp_u2
		l3_type_3_0          ifp_u2
		tcp_flags_fn_7_4     ifp_u2
		tcp_flags_fn_3_0     ifp_u2
		vlan_id_11_8         struct{ inner, outer ifp_u2 }
		class_id_3_0         struct{ d, c, b, a ifp_u2 }
		lookup_status_vector [4]ifp_u2
		forwarding_type      ifp_u2

		// Compressed l4 dst/src port for IP packets; RX_ID/OX_ID for FCoE packets.
		compressed_ip_ports           struct{ dst, src ifp_u2 }
		compressed_ethernet_type      ifp_u2
		compressed_ip_protocol        ifp_u2
		compression_tcam_result_35_32 struct{ dst, src ifp_u2 }

		udf_chunk_valid [2][2]ifp_u2

		// 0x0: Reserved
		// 0x1: Reserved
		// 0x2: PHP
		// 0x3: SWAP
		// 0x4: POP
		// 0x5: POP BoS label; Payload is L2
		// 0x6: POP BoS label: Payload is L3
		// 0x7: Reserved
		mpls_forwarding_label_action ifp_u2

		// *
		internal_priority_field ifp_u2

		// *
		mh_opcode ifp_u2

		vrf_10_7 ifp_u2

		_ [32 - 29]ifp_u2
	}

	u2 struct {
		// Source Realm ID from L3_IIF table
		nat_src_realm_id ifp_u1
		vfi_13_12        ifp_u1
		_                ifp_u1
		src_ab_16        ifp_u1 // [1] a [0] b
		l3_type_5_4      ifp_u1

		// [1] REP_COPY
		// [0] IPV4_ICMP_ERROR
		l3_control_status ifp_u1

		_ ifp_u1

		// [1] hg_lookup
		// [0] cpu_pkt_profile_1.ifp_key_control
		packet_profile ifp_u1

		// [1] l4 valid
		// [0] ip first fragment
		l4_status ifp_u1

		// When packet is received on a front panel port, this field carries assigned value of INT_CN.
		// When packet is received on a HiGig port, this field carries Congestion Class field from HiGig header
		int_cn ifp_u1

		// [1] dvp valid, [0] svp valid
		src_dst_virtual_port_valid ifp_u1

		// IFP_LOGICAL_TABLE_CLASS_ID from Logical Table Select TCAM
		ifp_logical_table_class_id ifp_u1

		_ ifp_u1

		// [1] d, [0] c
		aux_cd_valid ifp_u1

		// [1] b, [0] a
		aux_ab_valid ifp_u1

		// [0] L2 fields are Normalized.
		// [1] L3/L4 fields are Normalized (valid for IP and FCoE packets)
		address_is_normalized ifp_u1

		// Bit 1: dst ip is local *
		l3_address_exception ifp_u1

		// [1] my station tcam hit, [0] l3 routable *
		my_station_hit_l3_routable ifp_u1

		// [1] df [0] mf
		ip_flags ifp_u1

		// [1] whole packet, [0] first fragment
		ip_frag_info ifp_u1

		// [1] higig packet, [0] mirror
		hg_status ifp_u1

		// 0 => ether II 802.3 version 2
		// 1 => snap 802.2 packet with llc header 0xaa 0xaa 0x03 0x0 0x0 0x0
		// 2 => LLC all packets not covered by 0 & 1
		// 3 => reserved
		l2_packet_format ifp_u1

		// VLAN Tag status of the packet received on the wire.
		incoming_tag_status ifp_u1

		// VLAN tag status of the packet after VLAN assignment and Ingress VLAN translation.
		// [1] outer tag present, [0] inner tag present
		switching_tag_status ifp_u1

		// 0 => 0x8100, 1 => 0x9100, 2 => 0x88a8, 3 => other
		outer_tpid_encode ifp_u1
		inner_tpid_encode ifp_u1

		// green yellow red
		packet_color ifp_u1

		// [1] ral is present [0] gal is present
		ral_gal ifp_u1

		mpls_pseudo_wire_control_word_valid ifp_u1

		ipv4_checksum_ok ifp_u1

		_ [32 - 30]ifp_u1
	}
}

type ifp_field_extractor_key struct {
	l0_data [5 * 32]uint32
}

func (x *ifp_u5) set(k *ifp_field_extractor_key, v uint32) { k.l0_data[x.offset()] = v }
func (x *ifp_u5) get(k *ifp_field_extractor_key) uint32    { return k.l0_data[x.offset()] }
func (x *ifp_u4) set(k *ifp_field_extractor_key, v uint16) { k.l0_data[x.offset()] = uint32(v) }
func (x *ifp_u4) get(k *ifp_field_extractor_key) uint16    { return uint16(k.l0_data[x.offset()]) }
func (x *ifp_u3) set(k *ifp_field_extractor_key, v uint8)  { k.l0_data[x.offset()] = uint32(v) }
func (x *ifp_u3) get(k *ifp_field_extractor_key) uint8     { return uint8(k.l0_data[x.offset()]) }
func (x *ifp_u2) set(k *ifp_field_extractor_key, v uint8)  { k.l0_data[x.offset()] = uint32(v) }
func (x *ifp_u2) get(k *ifp_field_extractor_key) uint8     { return uint8(k.l0_data[x.offset()]) }
func (x *ifp_u1) set(k *ifp_field_extractor_key, v uint8)  { k.l0_data[x.offset()] = uint32(v) }
func (x *ifp_u1) get(k *ifp_field_extractor_key) uint8     { return uint8(k.l0_data[x.offset()]) }

type ifp_field_extractor_l1_bus_section_a struct {
	// Chosen by selectors from l0 bus u8 section.
	l1_e8 [7]ifp_u3
	l1_e4 [8]ifp_u2
	l1_e2 [8]ifp_u2
}

type ifp_field_extractor_l1_bus_section_b struct {
	l1_e32 [4]ifp_u5
	l1_e16 [7]ifp_u2
}

type ifp_field_extractor_l2_bus_section_a struct {
	// Low 96 bits selected from ifp_field_extractor_l1_bus_section_b (16 & 32 bit fields).
	l2_e16_4_9 [6]ifp_u2

	// High 104 bits passed through from ifp_field_extractor_l1_bus_section_a
	ifp_field_extractor_l1_bus_section_a
}

type ifp_field_extractor_l2_bus_section_b struct {
	// 64 bits selected from ifp_field_extractor_l1_bus_section_b; passed through to output.
	l2_e16_0_3 [4]ifp_u2
}

type ifp_field_extractor_l3_bus_section_a struct {
	// Low 80 bits of 160 bit key.
	l2_section_b_e16_0  ifp_u2
	l3_e4_0             ifp_u2
	l2_section_b_e16_1  ifp_u2
	l3_e4_3_1           [3]ifp_u2
	l2_section_b_e16_23 [2]ifp_u2
}

type ifp_field_extractor_l3_bus_section_b struct {
	// High 80 bits of 160 bit key or all of 80 bit key.
	// All selected from ifp_field_extractor_l2_bus_section_a
	l3_e4_4_20 [17]ifp_u2
	l3_e2      [5]ifp_u2
	l3_e1      [2]ifp_u1
}

type ifp_field_extractor_80bit_output struct {
	ifp_field_extractor_l3_bus_section_b
}

type ifp_field_extractor_160bit_output struct {
	ifp_field_extractor_l3_bus_section_a
	ifp_field_extractor_l3_bus_section_b
}

const n_ifp_key_generation_profiles = 64

type ifp_key_generation_profile_entry struct {
	// All l1 selects are 5 bit numbers i
	l1_e8_select_a [7]uint8 // l1 section a bits  0 + 8*i passed through to l2 section a at bit 96
	l1_e4_select_a [8]uint8 // l1 section a bits 56 + 4*i passed through to l2 section a at bit 96+56
	l1_e2_select_a [8]uint8 // l1 section a bits 88 + 2*i passed through to l2 section a at bit 96+88

	l1_e32_select_b [4]uint8 // l1 section b bits   0 + 32*i
	l1_e16_select_b [7]uint8 // l1 section b bits 128 + 16*i

	// All l2 selects are 4 bit numbers i which select l1 section b bits starting at 16*i
	l2_e16_select_a [6]uint8 // l2 section a bits 0 + 16*i
	l2_e16_select_b [4]uint8 // l2 section b bits 0 + 16*i passed through to l3 section b

	// Sizes e4: 6 bits, e2: 7 bits, e1: 8 bits
	l3_e4_select_a [17]uint8 // l3 section a bits 0 + 4*i
	l3_e2_select_a [5]uint8  // l3 section a bits 68 + 2*i
	l3_e1_select_a [2]uint8  // l3 section a bits 78 + 1*i
	l3_e4_select_b [4]uint8  // l3 section b
}

func (e *ifp_key_generation_profile_entry) MemBits() int { return 387 }
func (e *ifp_key_generation_profile_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for j := range e.l1_e32_select_b {
		i = m.MemGetSetUint8(&e.l1_e32_select_b[j], b, i+4, i, isSet)
	}
	for j := range e.l1_e16_select_b {
		i = m.MemGetSetUint8(&e.l1_e16_select_b[j], b, i+4, i, isSet)
	}
	for j := range e.l1_e8_select_a {
		i = m.MemGetSetUint8(&e.l1_e8_select_a[j], b, i+4, i, isSet)
	}
	for j := range e.l1_e4_select_a {
		i = m.MemGetSetUint8(&e.l1_e4_select_a[j], b, i+4, i, isSet)
	}
	for j := range e.l1_e2_select_a {
		i = m.MemGetSetUint8(&e.l1_e2_select_a[j], b, i+4, i, isSet)
	}
	for j := range e.l2_e16_select_b {
		i = m.MemGetSetUint8(&e.l2_e16_select_b[j], b, i+3, i, isSet)
	}
	for j := range e.l3_e4_select_b {
		i = m.MemGetSetUint8(&e.l3_e4_select_b[j], b, i+5, i, isSet)
	}
	for j := range e.l3_e4_select_a {
		i = m.MemGetSetUint8(&e.l3_e4_select_a[j], b, i+5, i, isSet)
	}
	for j := range e.l3_e2_select_a {
		i = m.MemGetSetUint8(&e.l3_e2_select_a[j], b, i+6, i, isSet)
	}
	for j := range e.l3_e1_select_a {
		i = m.MemGetSetUint8(&e.l3_e1_select_a[j], b, i+7, i, isSet)
	}
}

type ifp_key_generation_profile_mem m.MemElt

func (r *ifp_key_generation_profile_mem) geta(q *DmaRequest, e *ifp_key_generation_profile_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, e, BlockRxPipe, t)
}
func (r *ifp_key_generation_profile_mem) seta(q *DmaRequest, e *ifp_key_generation_profile_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, e, BlockRxPipe, t)
}
func (r *ifp_key_generation_profile_mem) get(q *DmaRequest, e *ifp_key_generation_profile_entry) {
	r.geta(q, e, sbus.Duplicate)
}
func (r *ifp_key_generation_profile_mem) set(q *DmaRequest, e *ifp_key_generation_profile_entry) {
	r.seta(q, e, sbus.Duplicate)
}
