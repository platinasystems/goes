// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

type tx_pipe_reg32 m.Reg32

func (r *tx_pipe_reg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Reg32)(r).Get(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_reg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Reg32)(r).Set(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_reg32) set(q *DmaRequest, v uint32) { r.seta(q, sbus.Duplicate, v) }
func (r *tx_pipe_reg32) getDo(q *DmaRequest, c sbus.AccessType) (v uint32) {
	r.geta(q, c, &v)
	q.Do()
	return
}

type tx_pipe_reg64 m.Reg64

func (r *tx_pipe_reg64) geta(q *DmaRequest, c sbus.AccessType, v *uint64) {
	(*m.Reg64)(r).Get(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_reg64) seta(q *DmaRequest, c sbus.AccessType, v uint64) {
	(*m.Reg64)(r).Set(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_reg64) set(q *DmaRequest, v uint64) { r.seta(q, sbus.Duplicate, v) }
func (r *tx_pipe_reg64) getDo(q *DmaRequest, c sbus.AccessType) (v uint64) {
	r.geta(q, c, &v)
	q.Do()
	return
}

type tx_pipe_preg32 m.Preg32
type tx_pipe_portreg32 [1 << m.Log2NRegPorts]tx_pipe_preg32

func (r *tx_pipe_preg32) address() sbus.Address { return (*m.Preg32)(r).Address() }

func (r *tx_pipe_preg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_preg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_preg32) set(q *DmaRequest, v uint32) { r.seta(q, sbus.Duplicate, v) }

type tx_pipe_preg64 m.Preg64
type tx_pipe_portreg64 [1 << m.Log2NRegPorts]tx_pipe_preg64

func (r *tx_pipe_preg64) address() sbus.Address { return (*m.Preg64)(r).Address() }

func (r *tx_pipe_preg64) geta(q *DmaRequest, c sbus.AccessType, v *uint64) {
	(*m.Preg64)(r).Get(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_preg64) seta(q *DmaRequest, c sbus.AccessType, v uint64) {
	(*m.Preg64)(r).Set(&q.DmaRequest, 0, BlockTxPipe, c, v)
}
func (r *tx_pipe_preg64) set(q *DmaRequest, v uint64) { r.seta(q, sbus.Duplicate, v) }

type tx_pipe_regs struct {
	latency_mode tx_pipe_reg32
	_            [0x01000000 - 0x00000100]byte

	hw_reset_control_0      tx_pipe_reg32
	_                       [0x01010000 - 0x01000100]byte
	hw_reset_control_1      tx_pipe_reg32
	_                       [0x01030000 - 0x01010100]byte
	arbiter_timeout_control tx_pipe_reg32
	_                       [0x01040000 - 0x01030100]byte
	arbiter_misc_control    tx_pipe_reg32

	_ [0x04000000 - 0x01040100]byte

	config                     tx_pipe_reg32
	config_1                   tx_pipe_reg32
	_                          [1]tx_pipe_reg32
	vlan_control_1             tx_pipe_portreg32
	mirror_select              tx_pipe_reg32
	l3_tunnel_pfm_vid          tx_pipe_reg32
	niv_ethertype              tx_pipe_reg32
	_                          [0x10 - 0x07]tx_pipe_reg32
	ip_multicast_config_2      tx_pipe_portreg32
	qcn_cntag_ethertype        tx_pipe_reg32
	_                          [0x18 - 0x12]tx_pipe_reg32
	port_to_next_hop_mapping   tx_pipe_portreg32
	port_extender_ethertype    tx_pipe_reg32
	_                          [0x20 - 0x1a]tx_pipe_reg32
	ieee_1588_ingress_control  tx_pipe_portreg32
	hg_eh_control              tx_pipe_reg64
	_                          [0x23 - 0x22]tx_pipe_reg32
	vlan_tag_action_for_bypass tx_pipe_portreg32
	_                          [0x08000000 - 0x04002400]byte

	outer_tpid                      [4]tx_pipe_reg32
	_                               [0x10 - 0x04]tx_pipe_reg32
	vlan_translate_hash_control     tx_pipe_reg32
	vlan_control_2                  tx_pipe_portreg32
	vlan_control_3                  tx_pipe_portreg32
	pvlan_eport_control             tx_pipe_portreg32
	ingress_port_tpid_select        tx_pipe_portreg32
	tunnel_id_mask                  tx_pipe_reg32
	vp_vlan_membership_hash_control tx_pipe_reg32
	_                               [0x0c000000 - 0x08001700]byte

	port_debug           tx_pipe_portreg32
	niv_config           tx_pipe_reg32
	sys_reserved_vid     tx_pipe_reg32
	etag_multicast_range tx_pipe_reg32
	_                    [0x10000000 - 0x0c000400]byte

	trill_header_attributes tx_pipe_reg32
	l2gre_control           tx_pipe_reg32
	vxlan_control           tx_pipe_reg32
	ecn_control             tx_pipe_reg32
	_                       [0x14000000 - 0x10000400]byte

	tunnel_pimdr       [2][2]tx_pipe_reg32
	modmap_control     tx_pipe_portreg32
	sf_src_modid_check tx_pipe_portreg32
	mim_ethertype      tx_pipe_reg32
	_                  [0x18000000 - 0x14000700]byte

	multicast_control_1           tx_pipe_reg32
	multicast_control_2           tx_pipe_reg32
	shaping_control               tx_pipe_portreg32
	counter_control               tx_pipe_portreg32
	packet_modification_control   tx_pipe_reg32
	port_mtu                      tx_pipe_portreg32
	_                             [0x7 - 0x6]tx_pipe_reg32
	niv_ethertype_2               tx_pipe_reg32
	pe_ethertype_2                tx_pipe_reg32
	hg_hdr_prot_status_tx_control tx_pipe_reg32
	_                             [0x10 - 0x0a]tx_pipe_reg32
	qcn_cntag_ethertype_2         tx_pipe_reg32
	hbfc_cntag_ethertype_2        tx_pipe_reg32
	port_outer_tpid_enable        tx_pipe_portreg32
	_                             [0x1c000000 - 0x18001300]byte

	wesp_proto_control            tx_pipe_reg32
	_                             [0x3 - 0x1]tx_pipe_reg32
	flexible_ip6_extension_header tx_pipe_reg32
	_                             [0x5 - 0x4]tx_pipe_reg32
	ieee_1588_parsing_control     tx_pipe_reg32
	_                             [0x20000000 - 0x1c000600]byte

	txf_slice_control        tx_pipe_reg32
	txf_meter_control        tx_pipe_reg32
	txf_slice_map            tx_pipe_reg32
	txf_class_id_selector    tx_pipe_reg32
	txf_key4_dvp_selector    tx_pipe_reg32
	txf_key4_mdl_selector    tx_pipe_reg32
	ieee_1588_egress_control tx_pipe_portreg32
	ieee_1588_link_delay_64  tx_pipe_portreg64
	_                        [0x13 - 0x08]tx_pipe_reg32
	txf_key8_dvp_selector    tx_pipe_reg32
	_                        [0x24000000 - 0x20001400]byte

	event_debug                            tx_pipe_reg32
	_                                      [0x28000000 - 0x24000100]byte
	counters                               [0x11 - 0x0]tx_pipe_portreg64
	_                                      [0x20 - 0x11]tx_pipe_reg32
	debug_counter_select                   [12]tx_pipe_reg32
	_                                      [0x30 - 0x2c]tx_pipe_reg32
	debug_counter_select_hi                [12]tx_pipe_reg32
	_                                      [0x41 - 0x3c]tx_pipe_reg32
	device_to_physical_port_number_mapping tx_pipe_portreg32
	_                                      [0x50 - 0x42]tx_pipe_reg32
	mmu_max_cell_credit                    tx_pipe_portreg32
	perq_counter, txf_counter              pipe_counter_1pool_control
	_                                      [0x113 - 0x59]tx_pipe_reg32
	data_buffer_misc_control               tx_pipe_reg32
	_                                      [0x400 - 0x114]tx_pipe_reg32

	pipe_counter pipe_counter_4pool_control
	_            [0x29040000 - 0x28050000]byte

	data_buffer_control_parity_enable tx_pipe_reg32
	_                                 [0x29500000 - 0x29040100]byte
	data_buffer_1dbg_a                tx_pipe_reg32
}

type tx_pipe_mems struct {
	_ [0x04000000 - 0]byte

	l3_next_hop [n_next_hop]tx_next_hop_mem
	_           [m.MemMax - n_next_hop]m.MemElt

	l3_interface [n_l3_interface]tx_l3_interface_mem
	_            [m.MemMax - n_l3_interface]m.MemElt

	mpls_vc_and_swap_label_table [n_tx_mpls_vc_and_swap_label]tx_mpls_vc_and_swap_label_mem
	_                            [m.MemMax - n_tx_mpls_vc_and_swap_label]m.MemElt

	dvp_attribute m.Mem

	vfi m.Mem

	port_table [n_pipe_ports]tx_port_mem
	_          [m.MemMax - n_pipe_ports]m.MemElt

	trill_parse_control m.Mem

	mpls_dst_ethernet_address [512]m.Mem64
	_                         [m.MemMax - 512]m.MemElt

	ip_multicast m.Mem

	rx_port [n_pipe_ports + 1]m.Mem32
	_       [m.MemMax - (n_pipe_ports + 1)]m.MemElt

	_ [1]m.Mem

	trill_tree_profile m.Mem

	map_mh m.Mem

	_ [2]m.Mem

	nat_packet_edit_info m.Mem

	_ [1]m.Mem

	outer_pri_cfi_mapping_for_bypass m.Mem

	_ [1]m.Mem

	dvp_attribute_1 m.Mem

	_ [0x05900000 - 0x04500000]byte

	int_cn_update m.Mem

	macda_oui_profile m.Mem

	vntag_etag_profile m.Mem

	_ [0x08000000 - 0x059c0000]byte

	vlan_translate m.Mem

	vlan [n_vlan]tx_vlan_mem
	_    [m.MemMax - n_vlan]m.MemElt

	vlan_spanning_tree_group [n_vlan_spanning_tree_group_entry]tx_vlan_spanning_tree_group_mem
	_                        [m.MemMax - n_vlan_spanning_tree_group_entry]m.MemElt

	pri_cng_map m.Mem

	ip_tunnel [2 << 10]tx_ip4_tunnel_mem
	_         [m.MemMax - 2<<10]m.MemElt

	ip_tunnel_ip6 [1 << 10]tx_ip6_tunnel_mem
	_             [m.MemMax - 1<<10]m.MemElt

	ip_tunnel_mpls [2 << 10]tx_mpls_tunnel_mem
	_              [m.MemMax - 2<<10]m.MemElt

	mpls_exp_mapping [2]m.Mem

	mpls_pri_mapping m.Mem

	im_mtp_index m.Mem

	em_mtp_index m.Mem

	_ [1]m.Mem

	vlan_tag_action_profile [n_vlan_tag_action_profile_entries]tx_vlan_tag_action_mem
	_                       [m.MemMax - n_vlan_tag_action_profile_entries]m.MemElt

	mirror_encap_control m.Mem

	fragment_id_table m.Mem

	dscp_table m.Mem

	gpp_attributes_modbase m.Mem

	gpp_attributes m.Mem

	vplag_group m.Mem

	vplag_member m.Mem

	vp_vlan_membership m.Mem

	_ [1]m.Mem

	etag_pcp_mapping m.Mem

	vlan_translate_remap_table [2]m.Mem

	vlan_translate_ecc m.Mem

	_ [0x08780000 - 0x086c0000]byte

	vlan_translate_action_table [2]m.Mem

	_ [0x10000000 - 0x08800000]byte

	mpls_exp_pri_mapping m.Mem

	trill_rbridge_nicknames m.Mem

	int_cn_to_ip_mapping m.Mem

	tunnel_ecn_encap [2]m.Mem

	ip_to_int_cn_mapping m.Mem

	_ [0x14000000 - 0x10180000]byte

	module_map m.Mem

	_ [0x18000000 - 0x14040000]byte

	mirror_encap_data [2]m.Mem

	_ [0x1c000000 - 0x18080000]byte

	trill_parse_control_2 m.Mem

	_ [0x20000000 - 0x1c040000]byte

	txf_tcam m.Mem

	txf_policy_table m.Mem

	txf_meter_table m.Mem

	l2_mpls_pseudo_wire_sequence_numbers m.Mem

	ieee_1588_sa m.Mem

	_ [0x24080000 - 0x20140000]byte

	pipe_counter_maps struct {
		packet_resolution m.Mem
		ip_tos            m.Mem
		port              m.Mem
		priority          m.Mem
		priority_cng      m.Mem
	}

	ip_cut_thru_class [n_phys_ports]m.Mem32
	_                 [m.MemMax - n_phys_ports]m.MemElt

	tx_start_count [34][16]m.Mem32
	_              [m.MemMax - 34*16]m.MemElt

	pipe_counter_cos_map m.Mem

	_ [0x28000000 - 0x24280000]byte

	txf_counter_table [1024]tx_pipe_pipe_counter_mem
	_                 [m.MemMax - 1024]m.MemElt

	_ [1]m.Mem

	per_tx_queue_counters struct {
		cpu   [mmu_n_cpu_queues]tx_pipe_pipe_counter_mem
		ports [n_rx_pipe_mmu_port]struct {
			unicast   [mmu_n_tx_queues]tx_pipe_pipe_counter_mem
			multicast [mmu_n_tx_queues]tx_pipe_pipe_counter_mem
		}
	}
	_ [m.MemMax - (48 + n_rx_pipe_mmu_port*2*mmu_n_tx_queues)]m.MemElt

	_ [0x28200000 - 0x280c0000]byte

	port_enable [n_pipe_ports]m.Mem32
	_           [m.MemMax - n_pipe_ports]m.MemElt

	_ [1]m.Mem

	port_mmu_cell_requests_outstanding [n_phys_ports]m.Mem32
	_                                  [m.MemMax - n_phys_ports]m.MemElt

	max_used_entries [n_pipe_ports]m.Mem32
	_                [m.MemMax - n_pipe_ports]m.MemElt

	per_port_buffer_soft_reset [n_pipe_ports]m.Mem32
	_                          [m.MemMax - n_pipe_ports]m.MemElt

	_ [0x287c0000 - 0x28340000]byte

	data_buffer_1dbg_b [n_pipe_ports]m.Mem32
	_                  [m.MemMax - n_pipe_ports]m.MemElt

	_ [0x2a800000 - 0x28800000]byte

	// pools 0-1: 4k counters per pool per pipe
	// pools 2-3: 1k counters per pool per pipe
	pipe_counter pipe_counter_4pool_mems
}
