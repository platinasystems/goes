// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"
)

const (
	n_vlan                           = 4096
	n_vlan_spanning_tree_group_entry = 512
)

type vlan_entry_l2_key_type uint8

const (
	vlan_entry_l2_key_type_outer_vlan_mac vlan_entry_l2_key_type = iota
	vlan_entry_l2_key_type_outer_vlan
	vlan_entry_l2_key_type_outer_and_inner_vlan
)

type l3mc_index uint16

func (r *l3mc_index) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint16((*uint16)(r), b, i+13, i, isSet)
}

type rx_vlan_entry struct {
	valid bool

	trill_unknown_unicast_network_receivers_present   bool
	trill_unknown_multicast_network_receivers_present bool
	trill_broadcast_network_receivers_present         bool
	trill_transit_igmp_mld_payload_to_cpu             bool
	trill_access_receivers_present                    bool
	trill_domain_non_unicast_replication_index        uint16
	trill_rbridge_nickname_index                      uint8

	// Enables virtual port replications to this vlan.
	virtual_port_enable bool

	enable_igmp_mld_snooping bool

	// Multicast flood pointers for unknown unicast, unknown multicast and broadcast packets.
	// For virtual port.
	l3mc_index_unknown_unicast   l3mc_index
	l3mc_index_unknown_multicast l3mc_index
	l3mc_index_broadcast         l3mc_index

	l2_entry_key_type vlan_entry_l2_key_type

	spanning_tree_group uint16

	// Forwarding ID for l2 lookup.
	forwarding_id uint16

	ifp_class_id uint16

	vlan_profile_index uint8

	hi_gig_trunk_override_profile_index uint8

	private_vlan_src_port_type uint8

	flex_counter_ref rx_pipe_4p12i_flex_counter_ref

	// Bitmap of virtual port groups that are members of this vlan.
	virtual_port_group_bitmap uint64

	// Rx/Tx vlan membership bitmaps.
	members [n_rx_tx]port_bitmap
}

func (e *rx_vlan_entry) MemBits() int { return 483 }
func (e *rx_vlan_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.members[rx].MemGetSet(b, i, isSet)
	i = e.members[tx].MemGetSet(b, i, isSet)
	i = m.MemGetSetUint16(&e.spanning_tree_group, b, i+8, i, isSet)
	i = m.MemGetSetUint64(&e.virtual_port_group_bitmap, b, i+63, i, isSet)
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.l2_entry_key_type), b, i+1, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&e.private_vlan_src_port_type), b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.hi_gig_trunk_override_profile_index, b, i+7, i, isSet)
	i = m.MemGetSetUint8(&e.vlan_profile_index, b, i+6, i, isSet)
	i = m.MemGetSetUint16(&e.forwarding_id, b, i+11, i, isSet)
	i = m.MemGetSet1(&e.trill_broadcast_network_receivers_present, b, i, isSet)
	i = m.MemGetSet1(&e.trill_unknown_unicast_network_receivers_present, b, i, isSet)
	i = m.MemGetSet1(&e.trill_unknown_multicast_network_receivers_present, b, i, isSet)
	i = m.MemGetSet1(&e.enable_igmp_mld_snooping, b, i, isSet)
	i = 385
	i = e.l3mc_index_broadcast.MemGetSet(b, i, isSet)
	i = e.l3mc_index_unknown_unicast.MemGetSet(b, i, isSet)
	i = e.l3mc_index_unknown_multicast.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.trill_transit_igmp_mld_payload_to_cpu, b, i, isSet)
	i = m.MemGetSetUint16(&e.trill_domain_non_unicast_replication_index, b, i+12, i, isSet)
	i = m.MemGetSet1(&e.trill_access_receivers_present, b, i, isSet)
	i = m.MemGetSetUint8(&e.trill_rbridge_nickname_index, b, i+1, i, isSet)
	i = m.MemGetSet1(&e.virtual_port_enable, b, i, isSet)
	i = e.flex_counter_ref.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint16(&e.ifp_class_id, b, i+11, i, isSet)
	if i != 477 {
		panic("vlan")
	}
}

type rx_vlan_mem m.MemElt

func (r *rx_vlan_mem) geta(q *DmaRequest, v *rx_vlan_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_vlan_mem) seta(q *DmaRequest, v *rx_vlan_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_vlan_mem) get(q *DmaRequest, v *rx_vlan_entry) {
	r.geta(q, v, sbus.DataSplit0)
}
func (r *rx_vlan_mem) set(q *DmaRequest, v *rx_vlan_entry) {
	r.seta(q, v, sbus.DataSplit0)
}

type tx_vlan_entry struct {
	valid                         bool // entry is valid
	modifiy_cfi_bit               bool
	modifiy_dot1p                 bool
	outer_tpid_index              uint8
	dot1p_mapping_pointer         uint8
	spanning_tree_group           uint16
	virtual_port_group_membership uint64
	flex_counter_ref              tx_pipe_flex_counter_ref
	members                       port_bitmap
	untagged_members              port_bitmap
}

func (e *tx_vlan_entry) MemBits() int { return 381 }
func (e *tx_vlan_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSet1(&e.valid, b, i, isSet)
	i = m.MemGetSetUint16(&e.spanning_tree_group, b, i+8, i, isSet)
	i = m.MemGetSetUint8(&e.outer_tpid_index, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.dot1p_mapping_pointer, b, i+3, i, isSet)
	i = m.MemGetSet1(&e.modifiy_cfi_bit, b, i, isSet)
	i = m.MemGetSetUint64(&e.virtual_port_group_membership, b, i+63, i, isSet)
	i = m.MemGetSet1(&e.modifiy_dot1p, b, i, isSet)
	i = e.flex_counter_ref.MemGetSet(b, i, isSet)

	i = 103
	i = e.untagged_members.MemGetSet(b, i, isSet)
	i = e.members.MemGetSet(b, i, isSet)
}

type tx_vlan_mem m.MemElt

func (r *tx_vlan_mem) geta(q *DmaRequest, v *tx_vlan_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_vlan_mem) seta(q *DmaRequest, v *tx_vlan_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_vlan_mem) get(q *DmaRequest, v *tx_vlan_entry) {
	r.geta(q, v, sbus.DataSplit0)
}
func (r *tx_vlan_mem) set(q *DmaRequest, v *tx_vlan_entry) {
	r.seta(q, v, sbus.DataSplit0)
}

type vlan_spanning_tree_group_entry [n_phys_ports]m.SpanningTreeState

func (e *vlan_spanning_tree_group_entry) MemBits() int { return 273 }
func (e *vlan_spanning_tree_group_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for ei := range e {
		i = e[ei].MemGetSet(b, i, isSet)
	}
}

type rx_vlan_spanning_tree_group_mem m.MemElt

func (r *rx_vlan_spanning_tree_group_mem) geta(q *DmaRequest, v *vlan_spanning_tree_group_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_vlan_spanning_tree_group_mem) seta(q *DmaRequest, v *vlan_spanning_tree_group_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_vlan_spanning_tree_group_mem) get(q *DmaRequest, v *vlan_spanning_tree_group_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *rx_vlan_spanning_tree_group_mem) set(q *DmaRequest, v *vlan_spanning_tree_group_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type tx_vlan_spanning_tree_group_mem m.MemElt

func (r *tx_vlan_spanning_tree_group_mem) geta(q *DmaRequest, v *vlan_spanning_tree_group_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_vlan_spanning_tree_group_mem) seta(q *DmaRequest, v *vlan_spanning_tree_group_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_vlan_spanning_tree_group_mem) get(q *DmaRequest, v *vlan_spanning_tree_group_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_vlan_spanning_tree_group_mem) set(q *DmaRequest, v *vlan_spanning_tree_group_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type vlan_protocol_data_entry struct {
	tag_action_profile_pointer               uint8
	outer_vlan_cfi, inner_vlan_cfi           bool
	outer_vlan_priority, inner_vlan_priority uint8
	outer_vlan, inner_vlan                   m.Vlan
}

func (e *vlan_protocol_data_entry) MemBits() int { return 39 }
func (e *vlan_protocol_data_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint8(&e.outer_vlan_priority, b, i+2, i, isSet)
	i = e.outer_vlan.MemGetSet(b, i, isSet)
	i = e.inner_vlan.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.tag_action_profile_pointer, b, i+5, i, isSet)
	i = m.MemGetSet1(&e.outer_vlan_cfi, b, i, isSet)
	i = m.MemGetSet1(&e.inner_vlan_cfi, b, i, isSet)
	i = m.MemGetSetUint8(&e.inner_vlan_priority, b, i+2, i, isSet)
}

type vlan_protocol_data_mem m.MemElt

func (r *vlan_protocol_data_mem) geta(q *DmaRequest, v *vlan_protocol_data_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *vlan_protocol_data_mem) seta(q *DmaRequest, v *vlan_protocol_data_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *vlan_protocol_data_mem) get(q *DmaRequest, v *vlan_protocol_data_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *vlan_protocol_data_mem) set(q *DmaRequest, v *vlan_protocol_data_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type vlan_range_entry [8]struct{ min, max m.Vlan }

func (e *vlan_range_entry) MemBits() int { return 193 }
func (e *vlan_range_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	for ei := range e {
		i = e[ei].min.MemGetSet(b, i, isSet)
		i = e[ei].max.MemGetSet(b, i, isSet)
	}
}

type vlan_range_mem m.MemElt

func (r *vlan_range_mem) geta(q *DmaRequest, v *vlan_range_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *vlan_range_mem) seta(q *DmaRequest, v *vlan_range_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *vlan_range_mem) get(q *DmaRequest, v *vlan_range_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *vlan_range_mem) set(q *DmaRequest, v *vlan_range_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type vlan_tag_action_type uint8

const (
	// 67:66 DT_IPRI_ACTION Specifies the inner 802.1p/PCP action if incoming packet is double tagged.
	dt_ipri_action vlan_tag_action_type = iota
	// 65:64 DT_OPRI_ACTION Specifies the outer 802.1p/PCP action if incoming packet is double tagged.
	dt_opri_action
	// 63:62 SOT_IPRI_ACTION Specifies the inner 802.1p/PCP action if incoming packet is single outer-tagged.
	sot_ipri_action
	// 61:60 SOT_OPRI_ACTION Specifies the outer 802.1p/PCP action if incoming packet is single outer-tagged.
	sot_opri_action
	// 59:58 SIT_IPRI_ACTION Specifies the inner 802.1p/PCP action if incoming packet is single inner-tagged.
	sit_ipri_action
	// 57:56 SIT_OPRI_ACTION Specifies the outer 802.1p/PCP action if incoming packet is single inner-tagged.
	sit_opri_action
	// 51:50 UT_IPRI_ACTION Specifies the inner 802.1p/PCP action if incoming packet is untagged.
	ut_ipri_action
	// 49:48 UT_OPRI_ACTION Specifies the outer 802.1p/PCP action if incoming packet is untagged.
	ut_opri_action

	// 47:46 DT_ICFI_ACTION Specifies the inner CFI/DE action if incoming packet is double tagged.
	dt_icfi_action
	// 45:44 DT_OCFI_ACTION Specifies the outer CFI/DE action if incoming packet is double tagged.
	dt_ocfi_action
	// 43:42 SOT_ICFI_ACTION Specifies the inner CFI/DE action if incoming packet is single outer-tagged.
	sot_icfi_action
	// 41:40 SOT_OCFI_ACTION Specifies the outer CFI/DE action if incoming packet is single outer-tagged.
	sot_ocfi_action
	// 39:38 SIT_ICFI_ACTION Specifies the inner CFI/DE action if incoming packet is single inner-tagged.
	sit_icfi_action
	// 37:36 SIT_OCFI_ACTION Specifies the outer CFI/DE action if incoming packet is single inner-tagged.
	sit_ocfi_action
	// 35:34 UT_ICFI_ACTION Specifies the inner CFI/DE action if incoming packet is untagged.
	ut_icfi_action
	// 33:32 UT_OCFI_ACTION Specifies the outer CFI/DE action if incoming packet is untagged.
	ut_ocfi_action

	// 29:27 DT_OTAG_ACTION Specifies the outer VLAN tag action if incoming packet is double tagged.
	dt_otag_action
	// 26:24 DT_POTAG_ACTION Specifies the outer VLAN tag action if incoming packet is double tagged and priority-tagged.
	dt_potag_action
	// 23:21 DT_ITAG_ACTION Specifies the inner VLAN tag action if incoming packet is double tagged.
	dt_itag_action
	// 20:18 DT_PITAG_ACTION Specifies the inner VLAN tag action if incoming packet is double tagged and inner tag is a priority tag.
	dt_pitag_action
	// 17:16 SOT_OTAG_ACTION Specifies the outer VLAN tag action if incoming packet is single outer-tagged.
	sot_otag_action
	// 15:14 SOT_POTAG_ACTION Specifies the outer VLAN tag action if incoming packet is single outer-tagged and priority-tagged.
	sot_potag_action
	// 13:11 SOT_ITAG_ACTION Specifies the inner VLAN tag action if incoming packet is single outer-tagged.
	sot_itag_action
	// 10:8 SIT_OTAG_ACTION Specifies the inner VLAN tag action if incoming packet is single inner-tagged and not priority-tagged.
	sit_otag_action
	// 7:6 SIT_ITAG_ACTION Specifies the inner VLAN tag action if incoming packet is single inner-tagged and not priority-tagged.
	sit_itag_action
	// 5:4 SIT_PITAG_ACTION Specifies the inner VLAN tag action if incoming packet is single inner-tagged and priority-tagged.
	sit_pitag_action
	// 3:2 UT_OTAG_ACTION Specifies the outer VLAN tag action if incoming packet untagged.
	ut_otag_action
	// 1:0 UT_ITAG_ACTION Specifies the inner VLAN tag action if incoming packet is untagged.
	ut_itag_action

	n_action_type
)

const (
	vlan_tag_action_none    = 0
	vlan_tag_action_add     = 1
	vlan_tag_action_replace = 2
	vlan_tag_action_delete  = 3
)

type vlan_tag_action_entry [n_action_type]uint8

type rx_vlan_tag_action_entry vlan_tag_action_entry
type tx_vlan_tag_action_entry vlan_tag_action_entry

const n_vlan_tag_action_profile_entries = 64

func (e *rx_vlan_tag_action_entry) MemBits() int { return 90 }
func (e *rx_vlan_tag_action_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint8(&e[ut_itag_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[ut_otag_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_pitag_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_itag_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_otag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_itag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_potag_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_otag_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_pitag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_itag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_potag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_otag_action], b, i+2, i, isSet)
	if i != 30 {
		panic("30")
	}

	i = 32
	i = m.MemGetSetUint8(&e[ut_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[ut_icfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_icfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_icfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_icfi_action], b, i+1, i, isSet)

	i = m.MemGetSetUint8(&e[ut_opri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[ut_ipri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_opri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_ipri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_opri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_ipri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_opri_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_ipri_action], b, i+1, i, isSet)
}

func (e *tx_vlan_tag_action_entry) MemBits() int { return 90 }
func (e *tx_vlan_tag_action_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint8(&e[ut_itag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[ut_otag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sit_pitag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sit_itag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sit_otag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_itag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_potag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_otag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_pitag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_itag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_potag_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_otag_action], b, i+2, i, isSet)
	if i != 36 {
		panic("36")
	}

	i = m.MemGetSetUint8(&e[ut_opri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[ut_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[ut_ipri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[ut_icfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_opri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sit_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sit_ipri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sit_icfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_opri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[sot_ipri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[sot_icfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_opri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_ocfi_action], b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e[dt_ipri_action], b, i+2, i, isSet)
	i = m.MemGetSetUint8(&e[dt_icfi_action], b, i+1, i, isSet)

	if i != 76 {
		panic("76")
	}
}

type rx_vlan_tag_action_mem m.MemElt
type tx_vlan_tag_action_mem m.MemElt

func (r *rx_vlan_tag_action_mem) geta(q *DmaRequest, v *rx_vlan_tag_action_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_vlan_tag_action_mem) seta(q *DmaRequest, v *rx_vlan_tag_action_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *rx_vlan_tag_action_mem) get(q *DmaRequest, v *rx_vlan_tag_action_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *rx_vlan_tag_action_mem) set(q *DmaRequest, v *rx_vlan_tag_action_entry) {
	r.seta(q, v, sbus.Duplicate)
}

func (r *tx_vlan_tag_action_mem) geta(q *DmaRequest, v *tx_vlan_tag_action_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_vlan_tag_action_mem) seta(q *DmaRequest, v *tx_vlan_tag_action_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockTxPipe, t)
}
func (r *tx_vlan_tag_action_mem) get(q *DmaRequest, v *tx_vlan_tag_action_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *tx_vlan_tag_action_mem) set(q *DmaRequest, v *tx_vlan_tag_action_entry) {
	r.seta(q, v, sbus.Duplicate)
}
