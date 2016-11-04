// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

type port_type uint8

const (
	port_type_ethernet port_type = iota
	port_type_higig
	port_type_loopback
)

type port_operation uint8

const (
	port_operation_normal port_operation = iota
	_
	port_operation_l3_iif_valid
	port_operation_vrf_valid
	port_operation_per_port_vlan_valid
)

// Unicast Reverse Path First (URPF) mode.
type ip_urpf_mode uint8

const (
	ip_urpf_disabled ip_urpf_mode = iota
	// IP address must hit in host/defip table.
	ip_urpf_loose
	// Lookup must hit and vlan must match that in next hop
	// ECMP entries with more than 8 paths will be handled as loose mode.
	ip_urpf_strict_match_next_hop_vlan
	// As above but ECMP entries with more than 8 paths will be punted to CPU.
	ip_urpf_strict_match_next_hop_vlan_punt_to_cpu
)

func (x *ip_urpf_mode) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type vlan_translation_port_key_type uint8

const (
	vlan_translation_port_key_all_ones vlan_translation_port_key_type = iota
	vlan_translation_port_key_sglp
	vlan_translation_port_key_svp
	vlan_translation_port_key_port_group
)

type vlan_translation_key_type uint8

const (
	// {IVID[11:0], OVID[11:0], SOURCE_FIELD}.
	vlan_translation_key_inner_vlan_outer_vlan_source_field vlan_translation_port_key_type = iota
	// {8b0, OTAG[15:0], SOURCE_FIELD}.
	vlan_translation_key_outer_tag_source_field
	// {8b0, ITAG[15:0], SOURCE_FIELD}.
	vlan_translation_key_inner_tag_source_field
	// mac_sa[47:0].
	vlan_translation_key_src_ethernet_address
	// {12b0, OVID[11:0], SOURCE_FIELD}.
	vlan_translation_key_outer_vlan_source_field
	// {IVID[11:0], 12b0, SOURCE_FIELD}.
	vlan_translation_key_inner_vlan_source_field
	// {8b0, PRI, CFI, 12b0, SOURCE_FIELD}.
	vlan_translation_key_pri_cfi_source_field
	// {IPv4_SIP}.
	vlan_translation_key_ip4_src_address
	// {SOURCE_FIELD, SRC_VIF/ETAG_VID}.
	vlan_translation_key_src_field_src_vif
	// {SOURCE_FIELD, SRC_VIF/ETAG_VID, SVLAN}.
	vlan_translation_key_src_field_src_vif_svlan
	// {SOURCE_FIELD, ETAG_VID, CVLAN}.
	vlan_translation_key_src_field_etag_vid_cvlan
	// {SOURCE_FIELD, ETAG_VID/SRC_VIF, OTAG[15:0]}.
	vlan_translation_key_src_field_etag_vid_outer_vlan_tag
	// {SOURCE_FIELD, ETAG_VID/SRC_VIF, ITAG[15:0]}.
	vlan_translation_key_src_field_etag_vid_inner_vlan_tag
	// {DIP}.
	vlan_translation_key_dst_ip4_address
	// {mac_sa[47:0], SOURCE_FIELD}.
	vlan_translation_key_src_ethernet_address_source_field
	// {DIP}.
	vlan_translation_key_dst_ip4_address_1
)

// Receive and transmit.
type rx_tx int

const (
	rx rx_tx = iota
	tx
	n_rx_tx
)

type rx_port_table_entry struct {
	port_type
	port_operation

	// Port associated with this Device Port Number. It is combined with MY_MODID to form SGPP for ethernet ports.
	global_physical_port uint8

	// Rx packet/byte counters for this port.
	flex_counter rx_pipe_4p12i_flex_counter_ref

	// Index into protocol_pkt_control register array.
	protocol_packet_index uint8

	// When enabled all rx traffic is sent to dst logical port.
	enable_tx_dst_logical_port bool
	tx_dst_logical_port        m.LogicalPort

	// Remote CPU
	enable_remote_cpu         bool
	enable_remote_cpu_parsing bool

	// 4 bit mirror enable
	rx_mirror_enable uint8

	// L3 section
	enable_ip4_multicast bool
	enable_ip6_multicast bool
	enable_l3_ip4        bool
	enable_l3_ip6        bool
	enable_mpls          bool

	// Search {S,VLAN,G} else {S,0,G} for multicast lookups.
	include_vlan_in_ip_multicast_lookup bool

	trust_packet_ip4_dscp bool
	trust_packet_ip6_dscp bool

	// If trust packet dscp is set, 7 high bits of dscp table index.  Lo 6 bits come from dscp: tos byte [7:2]
	dscp_table_pointer uint8

	ip_urpf_mode

	// If set routes marked with default bit in defip table will not pass URPF check.
	ip_urpf_check_default_route bool

	// L2 section
	port_bridge bool

	// Default (outer) VLAN and priority for this port.
	default_vlan          m.Vlan
	default_vlan_priority uint8
	default_vlan_cfi      bool

	outer_vlan_pri_mapping [8]uint8
	outer_vlan_cfi_mapping [2]uint8

	default_inner_vlan          m.Vlan
	default_inner_vlan_priority uint8
	default_inner_vlan_cfi      bool

	inner_vlan_pri_mapping [8]uint8
	inner_vlan_cfi_mapping [2]uint8

	trust_packet_outer_and_inner_vlan_id bool

	// Pointer to select one of 64 profiles in the ING_OUTER_DOT1P_MAPPING_TABLE.
	dot1p_remap_pointer              uint8
	dot1p_prohibited_priority_bitmap uint8

	// Index into vlan_protocol_data table
	vlan_protocol_data_index uint8

	enable_vlan_membership_check                     bool
	disable_vlan_membership_and_spanning_tree_checks bool
	drop_all_vlan_tagged_packets                     bool
	drop_all_vlan_untagged_packets                   bool
	pass_pause_frames_without_dropping               bool
	disable_static_move_drop                         bool

	subnet_based_vlan_enable                           bool
	mac_based_vlan_enable                              bool
	subnet_based_vlan_has_priority_over_mac_based_vlan bool

	private_vlan_enable bool

	use_inner_vlan_for_learning_and_forwarding     bool
	use_inner_vlan_prioiry                         bool
	verify_outer_vlan_tpid_matches_vlan_table_tpid bool

	vlan_translation_enable           bool
	vlan_translation_miss_drop_enable bool

	vlan_translation_drop_miss_key          [2]bool
	vlan_translation_drop_miss_keys_0_and_1 bool
	vlan_translation_key_types              [2]vlan_translation_key_type
	vlan_translation_port_key_types         [2]vlan_translation_port_key_type

	rx_drop_bpdu bool

	// Bitmap of ing_outer_tpid register to enable matching outer vlan header.
	outer_tpid_enable         uint8
	outer_tpid_cfi_bit_as_cng uint8

	enable_l2_gre_termination          bool
	enable_l2_gre_src_ip_in_lookup_key bool
	enable_l2_gre_svp                  bool

	enable_phb_from_etag                             bool
	etag_internal_priority_and_cng_derived_from_etag bool
	etag_de_for_adds                                 bool
	etag_pcp_for_adds                                uint8
	etag_pcp_de_source                               uint8
	etag_ing_dot1p_mapping_pointer                   uint8

	ieee_1588_profile_index uint8

	ieee_802_1as_trap_to_cpu bool

	rtag7_port_load_balancing_number uint8
	ecmp_random_bit_offset           [2]uint8
	trunk_random_bit_offset          uint8

	// IFP section
	// IPBM valid values: 0-33
	port_ipbm_index uint8
	ifp_class_id    uint8

	vfp_port_group_id uint8

	enable_ifp bool
	enable_vfp bool
	enable_efp bool

	//   0 == Use port_group_id from SOURCE_TRUNK_MAP table for VFP key.
	//   1 == Use port_group_id from LPORT table for VFP key.
	vfp_use_lport_table_group_id bool

	// misc
	my_module_id                                           uint8
	enable_class_based_station_movement                    bool
	enable_dual_module_id_mode                             bool
	module_header_ingress_tagged_bit_legacy_double_tagging bool
	mac_ip_bind_lookup_miss_drop                           bool

	ing_vlan_tag_action_profile_pointer uint8

	hi_gig_disable_drop_of_packets_with_my_module_id bool
	hi_gig_remove_header_src_port                    bool
	hi_gig_port_is_trunk_member                      bool
	hi_gig_trunk_id                                  uint8
	enable_hi_gig_2_mode                             bool

	cpu_managed_learning_control_for_new_entries        uint8
	cpu_managed_learning_control_for_station_moves      uint8
	cpu_managed_learning_control_for_bmac_new_entries   uint8
	cpu_managed_learning_control_for_bmac_station_moves uint8

	rx_pri_cng_map_pointer uint8

	niv_is_uplink_port          bool
	niv_rpf_check_enable        bool
	niv_interface_lookup_enable bool
	niv_interface_id            uint16
	niv_namespace               uint16

	vntag_drop_packet_if_tag_present     bool
	vntag_drop_packet_if_tag_not_present bool
	vntag_actions_if_not_present         uint8
	vntag_actions_if_present             uint8

	trill_port_is_rbridge        bool
	trill_allow_trill_frames     bool
	trill_allow_non_trill_frames bool
	trill_copy_core_is_is_to_cpu bool

	// mac in mac encapsulation
	mac_in_mac_enable_default_network_svp   bool
	mac_in_mac_enable_multicast_termination bool
	mac_in_mac_enable_termination           bool

	vxlan_enable_termination       bool
	vxlan_include_src_ip_in_lookup bool
	vxlan_enable_default_svp       bool
}

func (e *rx_port_table_entry) MemBits() int { return 424 }

func (e *rx_port_table_entry) MemGetSet(b []uint32, isSet bool) {
	i := m.MemGetSet1(&e.enable_ifp, b, 0, isSet)
	i = m.MemGetSet1(&e.vlan_translation_miss_drop_enable, b, i, isSet)
	i = m.MemGetSet1(&e.vlan_translation_enable, b, i, isSet)
	i = m.MemGetSet1(&e.trust_packet_ip4_dscp, b, i, isSet)
	i = m.MemGetSet1(&e.trust_packet_ip6_dscp, b, i, isSet)
	i = m.MemGetSet1(&e.enable_vlan_membership_check, b, i, isSet)
	i = m.MemGetSetUint8(&e.rx_mirror_enable, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&e.default_vlan_priority, b, i+2, i, isSet)
	i = m.MemGetSet1(&e.include_vlan_in_ip_multicast_lookup, b, i, isSet)

	if i != 14 {
		panic("port_table: ip6_multicast")
	}

	i = m.MemGetSet1(&e.enable_ip6_multicast, b, i, isSet)
	i = m.MemGetSet1(&e.enable_ip4_multicast, b, i, isSet)
	i = m.MemGetSet1(&e.module_header_ingress_tagged_bit_legacy_double_tagging, b, i, isSet)
	i = m.MemGetSet1(&e.enable_class_based_station_movement, b, i, isSet)
	i = m.MemGetSet1(&e.enable_l3_ip6, b, i, isSet)
	i = m.MemGetSet1(&e.enable_l3_ip4, b, i, isSet)
	i = m.MemGetSet1(&e.rx_drop_bpdu, b, i, isSet)
	i = m.MemGetSet1(&e.drop_all_vlan_tagged_packets, b, i, isSet)
	i = m.MemGetSet1(&e.drop_all_vlan_untagged_packets, b, i, isSet)
	i = m.MemGetSet1(&e.pass_pause_frames_without_dropping, b, i, isSet)
	i = m.MemGetSet1(&e.subnet_based_vlan_enable, b, i, isSet)
	i = m.MemGetSet1(&e.mac_based_vlan_enable, b, i, isSet)

	if i != 26 {
		panic("port_table: default_vlan")
	}

	i = e.default_vlan.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.port_type), b, i+1, i, isSet)
	i = m.MemGetSet1(&e.enable_dual_module_id_mode, b, i, isSet)
	i = m.MemGetSetUint8(&e.dot1p_remap_pointer, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.private_vlan_enable, b, i, isSet)
	i = m.MemGetSetUint8(&e.my_module_id, b, i+7, i, isSet)
	i = m.MemGetSet1(&e.subnet_based_vlan_has_priority_over_mac_based_vlan, b, i, isSet)
	i = m.MemGetSet1(&e.port_bridge, b, i, isSet)
	i = m.MemGetSet1(&e.mac_in_mac_enable_default_network_svp, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_port_is_trunk_member, b, i, isSet)
	i = m.MemGetSetUint8(&e.hi_gig_trunk_id, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_disable_drop_of_packets_with_my_module_id, b, i, isSet)
	i = m.MemGetSetUint8(&e.rtag7_port_load_balancing_number, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_remove_header_src_port, b, i, isSet)
	i = m.MemGetSet1(&e.enable_vfp, b, i, isSet)
	i = m.MemGetSetUint8(&e.vfp_port_group_id, b, i+7, i, isSet)

	if i != 83 {
		panic("port_table: ip_urpf_mode")
	}

	i = e.ip_urpf_mode.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.ip_urpf_check_default_route, b, i, isSet)
	i = m.MemGetSetUint8(&e.outer_tpid_cfi_bit_as_cng, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&e.outer_tpid_enable, b, i+3, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.vlan_translation_port_key_types[1]), b, i+1, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.vlan_translation_port_key_types[0]), b, i+1, i, isSet)

	if i != 98 {
		panic("port_table: ecc")
	}

	i = 106 // skip ecc/parity 0
	i = m.MemGetSet1(&e.trust_packet_outer_and_inner_vlan_id, b, i, isSet)
	i = e.default_inner_vlan.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.vlan_translation_key_types[0]), b, i+4, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.vlan_translation_key_types[1]), b, i+4, i, isSet)

	i = m.MemGetSetUint8(&e.cpu_managed_learning_control_for_new_entries, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&e.cpu_managed_learning_control_for_station_moves, b, i+3, i, isSet)
	for j := range e.outer_vlan_pri_mapping {
		i = m.MemGetSetUint8(&e.outer_vlan_pri_mapping[j], b, i+2, i, isSet)
	}
	i = m.MemGetSetUint8(&e.outer_vlan_cfi_mapping[0], b, i+0, i, isSet)
	i = m.MemGetSetUint8(&e.outer_vlan_cfi_mapping[1], b, i+0, i, isSet)
	i = m.MemGetSet1(&e.enable_mpls, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.port_operation), b, i+2, i, isSet)
	i = m.MemGetSet1(&e.disable_vlan_membership_and_spanning_tree_checks, b, i, isSet)
	i = m.MemGetSet1(&e.enable_remote_cpu_parsing, b, i, isSet)
	i = m.MemGetSet1(&e.disable_static_move_drop, b, i, isSet)
	i = m.MemGetSetUint8(&e.cpu_managed_learning_control_for_bmac_new_entries, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.cpu_managed_learning_control_for_bmac_station_moves, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.protocol_packet_index, b, i+5, i, isSet)
	i = m.MemGetSetUint8(&e.rx_pri_cng_map_pointer, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.enable_l2_gre_termination, b, i, isSet)
	i = m.MemGetSet1(&e.enable_l2_gre_src_ip_in_lookup_key, b, i, isSet)
	i = m.MemGetSet1(&e.enable_l2_gre_svp, b, i, isSet)
	i = m.MemGetSet1(&e.etag_internal_priority_and_cng_derived_from_etag, b, i, isSet)
	i = m.MemGetSetUint8(&e.etag_ing_dot1p_mapping_pointer, b, i+5, i, isSet)
	i = m.MemGetSetUint8(&e.ieee_1588_profile_index, b, i+5, i, isSet)

	if i != 202 {
		panic("port_table: rsvd1")
	}

	i = 212 // skip parity/ecc 1

	i = m.MemGetSetUint8(&e.default_inner_vlan_priority, b, i+2, i, isSet)
	i = m.MemGetSet1(&e.ieee_802_1as_trap_to_cpu, b, i, isSet)
	i = m.MemGetSet1(&e.vfp_use_lport_table_group_id, b, i, isSet)

	i = m.MemGetSetUint8(&e.ifp_class_id, b, i+7, i, isSet)
	i = m.MemGetSetUint8(&e.vlan_protocol_data_index, b, i+6, i, isSet)

	i = m.MemGetSet1(&e.default_vlan_cfi, b, i, isSet)
	i = m.MemGetSet1(&e.default_inner_vlan_cfi, b, i, isSet)

	i = m.MemGetSet1(&e.niv_is_uplink_port, b, i, isSet)
	i = m.MemGetSetUint16(&e.niv_interface_id, b, i+11, i, isSet)
	i = m.MemGetSetUint16(&e.niv_namespace, b, i+11, i, isSet)

	i = m.MemGetSetUint8(&e.dot1p_prohibited_priority_bitmap, b, i+7, i, isSet)

	i = m.MemGetSet1(&e.niv_rpf_check_enable, b, i, isSet)
	i = m.MemGetSet1(&e.niv_interface_lookup_enable, b, i, isSet)

	i = e.tx_dst_logical_port.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.enable_tx_dst_logical_port, b, i, isSet)

	if i != 287 {
		panic("port_table: vntag")
	}

	i = m.MemGetSetUint8(&e.vntag_actions_if_not_present, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.vntag_actions_if_present, b, i+1, i, isSet)
	i = m.MemGetSet1(&e.vntag_drop_packet_if_tag_present, b, i, isSet)
	i = m.MemGetSet1(&e.vntag_drop_packet_if_tag_not_present, b, i, isSet)

	i = m.MemGetSet1(&e.mac_in_mac_enable_multicast_termination, b, i, isSet)

	i = m.MemGetSet1(&e.use_inner_vlan_for_learning_and_forwarding, b, i, isSet)
	i = m.MemGetSet1(&e.use_inner_vlan_prioiry, b, i, isSet)
	i = m.MemGetSet1(&e.verify_outer_vlan_tpid_matches_vlan_table_tpid, b, i, isSet)
	i = m.MemGetSetUint8(&e.ing_vlan_tag_action_profile_pointer, b, i+5, i, isSet)

	i = m.MemGetSet1(&e.trill_allow_non_trill_frames, b, i, isSet)
	i = m.MemGetSet1(&e.trill_allow_trill_frames, b, i, isSet)
	i = m.MemGetSet1(&e.trill_port_is_rbridge, b, i, isSet)
	i = m.MemGetSet1(&e.trill_copy_core_is_is_to_cpu, b, i, isSet)

	i = m.MemGetSet1(&e.mac_in_mac_enable_termination, b, i, isSet)

	if i != 308 {
		panic("port_table: data2")
	}

	i = 319 // skip parity/ecc 2 + reserved bit

	i = m.MemGetSet1(&e.mac_ip_bind_lookup_miss_drop, b, i, isSet)

	i = m.MemGetSetUint8(&e.inner_vlan_cfi_mapping[1], b, i+0, i, isSet)
	i = m.MemGetSetUint8(&e.inner_vlan_cfi_mapping[0], b, i+0, i, isSet)
	for j := range e.inner_vlan_pri_mapping {
		i = m.MemGetSetUint8(&e.inner_vlan_pri_mapping[j], b, i+2, i, isSet)
	}
	i = m.MemGetSetUint8(&e.dscp_table_pointer, b, i+6, i, isSet)

	i = e.flex_counter.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.enable_hi_gig_2_mode, b, i, isSet)
	i = m.MemGetSetUint8(&e.etag_pcp_de_source, b, i+1, i, isSet)

	i = m.MemGetSet1(&e.vxlan_enable_termination, b, i, isSet)
	i = m.MemGetSet1(&e.vxlan_include_src_ip_in_lookup, b, i, isSet)
	i = m.MemGetSet1(&e.vxlan_enable_default_svp, b, i, isSet)
	i = m.MemGetSet1(&e.etag_de_for_adds, b, i, isSet)
	i = m.MemGetSetUint8(&e.etag_pcp_for_adds, b, i+2, i, isSet)

	if i != 383 {
		panic("port_table: ecmp random")
	}

	i = m.MemGetSetUint8(&e.ecmp_random_bit_offset[1], b, i+3, i, isSet)
	i = m.MemGetSetUint8(&e.ecmp_random_bit_offset[0], b, i+3, i, isSet)
	i = m.MemGetSetUint8(&e.trunk_random_bit_offset, b, i+3, i, isSet)
	i = m.MemGetSetUint8(&e.global_physical_port, b, i+7, i, isSet)
	i = m.MemGetSetUint8(&e.port_ipbm_index, b, i+5, i, isSet)

	i = m.MemGetSet1(&e.enable_remote_cpu, b, i, isSet)

	i = m.MemGetSet1(&e.vlan_translation_drop_miss_keys_0_and_1, b, i, isSet)
	i = m.MemGetSet1(&e.vlan_translation_drop_miss_key[1], b, i, isSet)
	i = m.MemGetSet1(&e.vlan_translation_drop_miss_key[0], b, i, isSet)

	if i != 413 {
		panic("port_table: rsvd413")
	}
}

type rx_port_table_mem m.MemElt

func (r *rx_port_table_mem) geta(q *DmaRequest, v *rx_port_table_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_port_table_mem) seta(q *DmaRequest, v *rx_port_table_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_port_table_mem) get(q *DmaRequest, v *rx_port_table_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *rx_port_table_mem) set(q *DmaRequest, v *rx_port_table_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type tx_port_table_entry struct {
	port_type

	// Tx packet/byte counters for this port.
	flex_counter tx_pipe_flex_counter_ref

	vlan_xlate_port_group_id uint8
	cntag_delete_pri_bitmap  uint8

	efp_port_group_id uint8
	efp_enable        bool

	preserve_cpu_tag bool

	enable_vlan_membership_check bool

	vxlan_vfi_lookup_key bool

	vntag_actions_if_present     uint8
	niv_prune_enable             bool
	niv_vif_id                   uint16
	niv_uplink_port              bool
	trill_rbridge_enable         bool
	trill_allow_trill_frames     bool
	trill_allow_non_trill_frames bool

	l2_gre_vfi_lookup_key_type bool

	mirror_encap_index  uint8
	mirror_encap_enable bool

	hi_gig_src_modid                            uint8 // modid for hi gig header when change_src_modid is set
	hi_gig_dual_modid_enable                    bool
	hi_gig_change_src_modid                     bool
	enable_hi_gig_2_mode                        bool
	hi_gig_2_eh_extension_header_enable         bool
	hi_gig_2_eh_extension_header_learn_override bool
}

func (e *tx_port_table_entry) MemBits() int { return 108 }

func (e *tx_port_table_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint8((*uint8)(&e.port_type), b, i+1, i, isSet)
	i = m.MemGetSet1(&e.enable_hi_gig_2_mode, b, i, isSet)
	i = m.MemGetSet1(&e.enable_vlan_membership_check, b, i, isSet)
	i = m.MemGetSet1(&e.preserve_cpu_tag, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_change_src_modid, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.hi_gig_src_modid), b, i+7, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.efp_port_group_id), b, i+7, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.vlan_xlate_port_group_id), b, i+7, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_dual_modid_enable, b, i, isSet)
	i = m.MemGetSet1(&e.efp_enable, b, i, isSet)
	i = 49 // skip reserved bits
	i = m.MemGetSetUint8((*uint8)(&e.cntag_delete_pri_bitmap), b, i+7, i, isSet)
	i = 59
	i = m.MemGetSet1(&e.vxlan_vfi_lookup_key, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_2_eh_extension_header_enable, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_2_eh_extension_header_learn_override, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.vntag_actions_if_present), b, i+1, i, isSet)
	i = m.MemGetSet1(&e.niv_prune_enable, b, i, isSet)
	i = m.MemGetSetUint16((*uint16)(&e.niv_vif_id), b, i+11, i, isSet)
	i = m.MemGetSet1(&e.niv_uplink_port, b, i, isSet)
	i = m.MemGetSet1(&e.trill_rbridge_enable, b, i, isSet)
	i = m.MemGetSet1(&e.trill_allow_trill_frames, b, i, isSet)
	i = m.MemGetSet1(&e.trill_allow_non_trill_frames, b, i, isSet)
	i = m.MemGetSet1(&e.mirror_encap_enable, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.mirror_encap_index), b, i+3, i, isSet)
	i = m.MemGetSet1(&e.l2_gre_vfi_lookup_key_type, b, i, isSet)
	i = e.flex_counter.MemGetSet(b, i, isSet)

	if i != 106 {
		panic("106")
	}
}

type tx_port_mem m.MemElt

func (r *tx_port_mem) geta(q *DmaRequest, v *tx_port_table_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_port_mem) seta(q *DmaRequest, v *tx_port_table_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_port_mem) get(q *DmaRequest, v *tx_port_table_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_port_mem) set(q *DmaRequest, v *tx_port_table_entry) {
	r.seta(q, v, sbus.Duplicate)
}

func (t *fe1a) init_port_table() {
	q := t.getDmaReq()

	rx_pipe_defaults := rx_port_table_entry{
		port_type:      port_type_ethernet,
		port_operation: port_operation_l3_iif_valid,

		enable_ifp:           true,
		enable_l3_ip4:        true,
		enable_l3_ip6:        true,
		enable_ip4_multicast: true,
		enable_ip6_multicast: true,
		enable_mpls:          true,

		// VLAN stuff.
		enable_vfp:                           true,
		trust_packet_outer_and_inner_vlan_id: true,
		default_vlan:                         1,
		outer_tpid_enable:                    0x3,

		// Identity maps for pri and cfi.
		outer_vlan_pri_mapping: [...]uint8{0, 1, 2, 3, 4, 5, 6, 7},
		outer_vlan_cfi_mapping: [...]uint8{0, 1},
		inner_vlan_pri_mapping: [...]uint8{0, 1, 2, 3, 4, 5, 6, 7},
		inner_vlan_cfi_mapping: [...]uint8{0, 1},

		// PVP_CML_SWITCH by default (hardware learn and forward).
		cpu_managed_learning_control_for_new_entries:   0x8,
		cpu_managed_learning_control_for_station_moves: 0x8,
	}

	tx_pipe_defaults := tx_port_table_entry{
		port_type: rx_pipe_defaults.port_type,
	}

	// Initialize all data ports and management ports.
	for i := range t.Ports {
		p := t.Ports[i].(*Port)
		pipe_port := p.physical_port_number.toPipe()
		pipe_mask := uint(1) << p.physical_port_number.pipe()

		// Rx_pipe
		{
			e := rx_pipe_defaults
			e.global_physical_port = uint8(pipe_port.toGpp())
			e.flex_counter.alloc(t, flex_counter_pool_rx_port_table, pipe_mask, BlockRxPipe)
			t.rx_pipe_port_table[pipe_port] = e
			t.rx_pipe_mems.port_table[pipe_port].set(q, &e)
			t.rx_pipe_mems.lport_profile_table[pipe_port].set(q, &e)
		}

		// Tx_pipe
		{
			e := tx_pipe_defaults
			e.flex_counter.alloc(t, flex_counter_pool_tx_port_table, pipe_mask, BlockTxPipe)
			t.tx_pipe_port_table[pipe_port] = e
			t.tx_pipe_mems.port_table[pipe_port].set(q, &e)
		}
	}
	q.Do()

	// Per-pipe loopback ports.
	for pipe := uint(0); pipe < n_pipe; pipe++ {
		phys_port := phys_port_loopback_for_pipe(pipe)
		pipe_port := phys_port.toPipe()
		pipe_mask := uint(1) << phys_port.pipe()

		// Rx_pipe
		{
			e := rx_pipe_defaults
			e.port_type = port_type_loopback
			e.global_physical_port = uint8(pipe_port.toGpp())
			e.flex_counter.alloc(t, flex_counter_pool_rx_port_table, pipe_mask, BlockRxPipe)
			t.rx_pipe_port_table[pipe_port] = e
			t.rx_pipe_mems.port_table[pipe_port].set(q, &e)
			t.rx_pipe_mems.lport_profile_table[pipe_port].set(q, &e)
		}

		// Tx_pipe
		{
			e := tx_pipe_defaults
			e.port_type = port_type_loopback
			e.flex_counter.alloc(t, flex_counter_pool_tx_port_table, pipe_mask, BlockTxPipe)
			t.tx_pipe_port_table[pipe_port] = e
			t.tx_pipe_mems.port_table[pipe_port].set(q, &e)
		}
	}
	q.Do()

	// Cpu port.
	{
		pipe_port := phys_port_cpu.toPipe()
		pipe_mask := uint(1) << phys_port_cpu.pipe()

		// Rx pipe
		{
			e := rx_pipe_defaults
			e.global_physical_port = uint8(pipe_port.toGpp())
			e.flex_counter.alloc(t, flex_counter_pool_rx_port_table, pipe_mask, BlockRxPipe)
			f := e

			t.rx_pipe_port_table[pipe_port] = e
			t.rx_pipe_mems.port_table[pipe_port].set(q, &e)
			t.rx_pipe_mems.lport_profile_table[pipe_port].set(q, &e)

			f.port_type = port_type_higig
			f.flex_counter.alloc(t, flex_counter_pool_rx_port_table, pipe_mask, BlockRxPipe)
			t.rx_pipe_mems.port_table[pipe_port_cpu_hi_gig_index].set(q, &f)
			t.rx_pipe_mems.lport_profile_table[pipe_port_cpu_hi_gig_index].set(q, &f)
		}

		// Tx pipe
		{
			e := tx_pipe_defaults
			e.flex_counter.alloc(t, flex_counter_pool_tx_port_table, pipe_mask, BlockTxPipe)
			t.tx_pipe_port_table[pipe_port] = e
			t.tx_pipe_mems.port_table[pipe_port].set(q, &e)
		}

		q.Do()
	}

	// Port counters depend on port table due to rx/tx flex counters.
	t.port_counter_init()
}
