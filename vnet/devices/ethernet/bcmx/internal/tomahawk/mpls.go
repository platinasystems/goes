package tomahawk

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

	"fmt"
)

// MPLS header:
//   [31:12] 20 bit label
//   [11:9] traffic class (EXP)
//   [8] bottom of stack (BOS) bit
//   [7:0] time to live (TTL)
type mpls_label uint32
type mpls_exp uint8
type mpls_ttl uint8

func (x *mpls_label) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint32((*uint32)(x), b, i+19, i, isSet)
}
func (x *mpls_ttl) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+7, i, isSet)
}
func (x *mpls_exp) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type mpls_tx_next_hop struct {
	l3_intf_index    uint16
	dst_virtual_port uint32

	mac_da_profile_index uint16

	// VC_AND_SWAP_INDEX or PW_INIT_NUM table index
	index uint16

	efp_class_id uint16

	dst_virtual_port_broadcast_drop_enable         bool
	dst_virtual_port_unknown_unicast_drop_enable   bool
	dst_virtual_port_unknown_multicast_drop_enable bool
	dst_virtual_port_is_network_port               bool
	delete_vntag_if_present                        bool

	flex_counter_ref tx_pipe_flex_counter_ref

	hi_gig_2_mode                                   bool
	hi_gig_packet_modify_enable                     bool // set for non-vpls and vplx proxy entries
	hi_gig_add_sys_reserved_vid                     bool
	hi_gig_l3_override                              bool
	hi_gig_change_destination                       bool
	hi_gig_single_copy_vpls_multicast_dst_port      uint8
	hi_gig_single_copy_vpls_multicast_dst_module_id uint8
}

func (e *mpls_tx_next_hop) Type() tx_next_hop_type { return tx_next_hop_type_mpls_mac_da_profile }
func (e *mpls_tx_next_hop) MemGetSet(b []uint32, i int, isSet bool) {
	i = m.MemGetSetUint16(&e.l3_intf_index, b, i+12, i, isSet)
	i = m.MemGetSetUint32(&e.dst_virtual_port, b, i+23, i, isSet)
	i = 32
	i = m.MemGetSet1(&e.hi_gig_packet_modify_enable, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_2_mode, b, i, isSet)
	i = m.MemGetSetUint8(&e.hi_gig_single_copy_vpls_multicast_dst_module_id, b, i+7, i, isSet)
	i = m.MemGetSetUint8(&e.hi_gig_single_copy_vpls_multicast_dst_port, b, i+7, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_add_sys_reserved_vid, b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_l3_override, b, i, isSet)
	i = m.MemGetSetUint16(&e.index, b, i+13, i, isSet)
	i = m.MemGetSetUint16(&e.mac_da_profile_index, b, i+8, i, isSet)
	i = m.MemGetSet1(&e.dst_virtual_port_broadcast_drop_enable, b, i, isSet)
	i = m.MemGetSetUint16(&e.efp_class_id, b, i+11, i, isSet)
	i = 98
	i = m.MemGetSet1(&e.dst_virtual_port_unknown_multicast_drop_enable, b, i, isSet)
	i = m.MemGetSet1(&e.dst_virtual_port_unknown_unicast_drop_enable, b, i, isSet)
	i = m.MemGetSet1(&e.delete_vntag_if_present, b, i, isSet)
	i = e.flex_counter_ref.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.hi_gig_change_destination, b, i, isSet)
	i = m.MemGetSet1(&e.dst_virtual_port_is_network_port, b, i, isSet)
	if i != 122 {
		panic("encoding")
	}
}

type mpls_entry_type uint8

const (
	mpls_entry_type_mpls mpls_entry_type = iota
	mpls_entry_type_mac_in_mac_bvid_bsa
	mpls_entry_type_mac_in_mac_isid
	mpls_entry_type_mac_in_mac_isid_svp
	mpls_entry_type_gre_vpnid_src_ip
	mpls_entry_type_trill_rbridge_nickname
	mpls_entry_type_src_ip
	mpls_entry_type_gre_vpnid
	mpls_entry_type_src_ip_1 // ???
	mpls_entry_type_vxlan_vn_id
	mpls_entry_type_vxlan_vn_id_src_ip
)

type mpls_entry_interface interface {
	Type() mpls_entry_type
	MemGetSet(b []uint32, i int, isSet bool)
}

type mpls_entry_wrapper struct {
	entry mpls_entry_interface
}

func (e *mpls_entry_wrapper) MemBits() int { return 122 }
func (e *mpls_entry_wrapper) MemGetSet(b []uint32, isSet bool) {
	var t mpls_entry_type
	if isSet {
		t = e.entry.Type()
	}
	i := 1
	i = m.MemGetSetUint8((*uint8)(&t), b, i+3, i, isSet)
	if !isSet {
		switch t {
		case mpls_entry_type_mpls:
			e.entry = &mpls_entry_mpls{}
		default:
			panic(e)
		}
	}
	e.entry.MemGetSet(b, i, isSet)
}

// Number of elements in MPLS_ENTRY table.
const n_mpls_entry = 16 << 10

type mpls_entry_mem m.MemElt

func (r *mpls_entry_mem) geta(q *DmaRequest, t sbus.AccessType) (v mpls_entry_interface) {
	g := mpls_entry_wrapper{}
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, &g, BlockRxPipe, t)
	return g.entry
}
func (r *mpls_entry_mem) seta(q *DmaRequest, v mpls_entry_interface, t sbus.AccessType) {
	g := mpls_entry_wrapper{entry: v}
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, &g, BlockRxPipe, t)
}
func (r *mpls_entry_mem) get(q *DmaRequest) (v mpls_entry_interface) { return r.geta(q, sbus.Duplicate) }
func (r *mpls_entry_mem) set(q *DmaRequest, v mpls_entry_interface)  { r.seta(q, v, sbus.Duplicate) }

// Action when BOS bit is set/clear in packet mpls header.  BOS == bottom of label stack.
type mpls_action_bos uint8
type mpls_action_not_bos uint8

const (
	mpls_action_bos_invalid mpls_action_bos = iota
	mpls_action_bos_vpls
	mpls_action_bos_l3_vpn_lookup
	mpls_action_bos_swap_labels
	mpls_action_bos_l3_vpn_next_hop
	mpls_action_bos_l3_vpn_ecmp
	mpls_action_bos_swap_labels_ecmp
)

func (x *mpls_action_bos) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

const (
	mpls_action_not_bos_invalid mpls_action_not_bos = iota
	mpls_action_not_bos_pop
	// PHP = Penultimate Hop Pop
	mpls_action_not_bos_php_next_hop // index is next hop
	mpls_action_not_bos_swap_next_hop
	mpls_action_not_bos_swap_ecmp // index is ecmp group
	mpls_action_not_bos_php_ecmp
)

func (x *mpls_action_not_bos) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type mpls_control_word_check uint8

const (
	mpls_control_word_check_none mpls_control_word_check = iota
	mpls_control_word_check_present_disable_sequence
	mpls_control_word_check_present_check_sequence_ne_0 // strict
	mpls_control_word_check_present_check_sequence      // loose
)

func (x *mpls_control_word_check) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type mpls_entry_mpls struct {
	valid bool

	ip4_payload_enable_for_l2_vpn bool
	ip6_payload_enable_for_l2_vpn bool
	bfd_enable                    bool

	// Key is logical port plus label from mpls header.
	// If all of T bit module & port are zero then logical port is not searched in table.
	m.LogicalPort
	mpls_label

	flex_counter_ref rx_pipe_3p11i_flex_counter_ref

	mpls_action_bos
	mpls_action_not_bos

	mpls_control_word_check

	php_push_exp_into_next_inner_label  bool
	decap_copy_outer_ttl_to_inner_label bool

	// ip dscp from packet or from ing_mpls_exp_mapping table
	disable_ip_dscp_from_ing_mpls_exp_mapping bool

	pop_or_swap_priority_handling uint8
	ing_exp_mapping_pointer       uint8 // upper 4 bits of address in ing_mpls_exp_mapping table
	new_priority                  uint8

	// l3_iif l3 input interface for l3_vpn_lookup
	// next hop index for l3_vpn_next_hop
	// ecmp group number for l3_vpn_ecmp
	// src_virtual_port for vpls
	index uint32

	pseudo_wire_termination_number_valid bool
	pseudo_wire_cc_type                  uint8

	pseudo_wire_termination_number uint16
}

func (e *mpls_entry_mpls) Type() mpls_entry_type { return mpls_entry_type_mpls }
func (e *mpls_entry_mpls) MemGetSet(b []uint32, i int, isSet bool) {
	m.MemGetSet1(&e.valid, b, 0, isSet)
	// type is set by wrapper and i points after type.
	if expect := 5; i != expect {
		panic(fmt.Errorf("got %d != expected %d", i, expect))
	}
	i = e.LogicalPort.MemGetSet(b, i, isSet)
	i = e.mpls_label.MemGetSet(b, i, isSet)
	i = e.mpls_action_bos.MemGetSet(b, i, isSet)
	i = e.mpls_control_word_check.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.pop_or_swap_priority_handling, b, i+1, i, isSet)
	i = m.MemGetSetUint8(&e.ing_exp_mapping_pointer, b, i+4, i, isSet)
	i = m.MemGetSetUint8(&e.new_priority, b, i+3, i, isSet)
	i = m.MemGetSetUint32(&e.index, b, i+16, i, isSet)
	i = m.MemGetSet1(&e.ip4_payload_enable_for_l2_vpn, b, i, isSet)
	i = m.MemGetSet1(&e.ip6_payload_enable_for_l2_vpn, b, i, isSet)
	i = m.MemGetSetUint16(&e.pseudo_wire_termination_number, b, i+10, i, isSet)
	i = m.MemGetSet1(&e.pseudo_wire_termination_number_valid, b, i, isSet)
	i = 91
	i = e.mpls_action_not_bos.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.php_push_exp_into_next_inner_label, b, i, isSet)
	i = m.MemGetSet1(&e.bfd_enable, b, i, isSet)
	i = 97
	i = e.flex_counter_ref.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.pseudo_wire_cc_type, b, i+1, i, isSet)
	i = m.MemGetSet1(&e.decap_copy_outer_ttl_to_inner_label, b, i, isSet)
	i = m.MemGetSet1(&e.disable_ip_dscp_from_ing_mpls_exp_mapping, b, i, isSet)
}

type tx_mpls_vc_and_swap_label_action uint8

const (
	tx_mpls_vc_and_swap_label_action_none tx_mpls_vc_and_swap_label_action = iota
	tx_mpls_vc_and_swap_label_action_push
	tx_mpls_vc_and_swap_label_action_swap
)

func (x *tx_mpls_vc_and_swap_label_action) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type mpls_control_word_append uint8

const (
	mpls_control_word_append_none mpls_control_word_append = iota
	mpls_control_word_append_with_sequence_0
	mpls_control_word_append_with_sequence_from_pseudo_wire_counter
)

func (x *mpls_control_word_append) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type mpls_exp_select uint8

const (
	mpls_exp_select_this_entry mpls_exp_select = iota
	mpls_exp_select_via_exp_mapping_pointer

	// If packet has no inner label, then use EXP from this entry.
	// If outermost label, use the MPLS_EXP_MAPPING_PTR and new EXP (from inner or table) to get a new outer PRI/CFI.
	mpls_exp_select_next_inner_label

	// Keep the old EXP value from the last label decapsulated (swap label).
	// If outermost label, use old EXP to also map the new outer PRI/CFI from the TX_MPLS_PRI_MAPPING table.
	mpls_exp_select_keep
)

func (x *mpls_exp_select) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type mpls_exp_spec struct {
	mpls_exp_select
	exp_mapping_pointer uint8
	cfi                 bool
	priority            uint8
}

func (x *mpls_exp_spec) MemGetSet(b []uint32, i int, isSet bool) int {
	var v uint8
	if isSet {
		if x.mpls_exp_select == mpls_exp_select_this_entry {
			v = x.priority << 1
			if x.cfi {
				v |= 1
			}
		} else {
			v = x.exp_mapping_pointer
		}
	}
	i = x.mpls_exp_select.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&v, b, i+3, i, isSet)
	if !isSet {
		if x.mpls_exp_select == mpls_exp_select_this_entry {
			x.priority = (v >> 1) & 0x7
			x.cfi = v&1 != 0
		} else {
			x.exp_mapping_pointer = v
		}
	}
	return i
}

type mpls_sd_tag_action_if_present uint8

const (
	mpls_sd_tag_action_if_present_none mpls_sd_tag_action_if_present = iota
	mpls_sd_tag_action_if_present_replace_vid_tpid
	mpls_sd_tag_action_if_present_replace_vid_only
	mpls_sd_tag_action_if_present_delete
	mpls_sd_tag_action_if_present_replace_vid_pri_tpid
	mpls_sd_tag_action_if_present_replace_vid_pri_only
	mpls_sd_tag_action_if_present_replace_pri_only
	mpls_sd_tag_action_if_present_replace_tpid_only
)

func (x *mpls_sd_tag_action_if_present) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+2, i, isSet)
}

type mpls_sd_tag_action_if_not_present uint8

const (
	mpls_sd_tag_action_if_not_present_none mpls_sd_tag_action_if_not_present = iota
	mpls_sd_tag_action_if_not_present_add_vid_tpid
)

func (x *mpls_sd_tag_action_if_not_present) MemGetSet(b []uint32, i int, isSet bool) int {
	return m.MemGetSetUint8((*uint8)(x), b, i+1, i, isSet)
}

type tx_mpls_vc_and_swap_label_entry struct {
	mpls_ttl
	mpls_exp
	mpls_label
	action              tx_mpls_vc_and_swap_label_action
	control_word_append mpls_control_word_append
	mpls_exp_spec

	// must be set for mpls_control_word_append_with_sequence
	pseudo_wire_update_init_counters bool

	mpls_sd_tag_action_if_present
	mpls_sd_tag_action_if_not_present

	// Selects one of 4 TPIDs (e.g. ethernet types) for TPID replacement.
	sd_tag_add_or_replace_tpid_index uint8
	sd_tag_add_or_replace_vlan       m.Vlan

	sd_tag_vlan_priority vlan_priority_spec
	sd_tag_remark_cfi    bool
}

func (e *tx_mpls_vc_and_swap_label_entry) MemBits() int { return 82 }
func (e *tx_mpls_vc_and_swap_label_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = e.mpls_label.MemGetSet(b, i, isSet)
	i = e.mpls_exp.MemGetSet(b, i, isSet)
	i = e.mpls_ttl.MemGetSet(b, i, isSet)
	i = e.action.MemGetSet(b, i, isSet)
	i = e.control_word_append.MemGetSet(b, i, isSet)
	i = e.mpls_exp_spec.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.pseudo_wire_update_init_counters, b, i, isSet)
	i = e.mpls_sd_tag_action_if_present.MemGetSet(b, i, isSet)
	i = e.mpls_sd_tag_action_if_not_present.MemGetSet(b, i, isSet)
	i = m.MemGetSetUint8(&e.sd_tag_add_or_replace_tpid_index, b, i+1, i, isSet)
	i = e.sd_tag_add_or_replace_vlan.MemGetSet(b, i, isSet)
	i = e.sd_tag_vlan_priority.MemGetSet(b, i, isSet)
	i = m.MemGetSet1(&e.sd_tag_remark_cfi, b, i, isSet)
}

const n_tx_mpls_vc_and_swap_label = 16 << 10

type tx_mpls_vc_and_swap_label_mem m.MemElt

func (r *tx_mpls_vc_and_swap_label_mem) geta(q *DmaRequest, e *tx_mpls_vc_and_swap_label_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, e, BlockTxPipe, t)
}
func (r *tx_mpls_vc_and_swap_label_mem) seta(q *DmaRequest, e *tx_mpls_vc_and_swap_label_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, e, BlockTxPipe, t)
}
func (r *tx_mpls_vc_and_swap_label_mem) get(q *DmaRequest, e *tx_mpls_vc_and_swap_label_entry) {
	r.geta(q, e, sbus.Duplicate)
}
func (r *tx_mpls_vc_and_swap_label_mem) set(q *DmaRequest, e *tx_mpls_vc_and_swap_label_entry) {
	r.seta(q, e, sbus.Duplicate)
}
