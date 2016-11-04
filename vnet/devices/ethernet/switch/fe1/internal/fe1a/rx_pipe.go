// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package fe1a

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
)

type rx_pipe_reg32 m.Reg32

func (r *rx_pipe_reg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Reg32)(r).Get(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_reg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Reg32)(r).Set(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_reg32) get(q *DmaRequest, v *uint32) { r.geta(q, sbus.Duplicate, v) }
func (r *rx_pipe_reg32) set(q *DmaRequest, v uint32)  { r.seta(q, sbus.Duplicate, v) }
func (r *rx_pipe_reg32) getDo(q *DmaRequest, c sbus.AccessType) (v uint32) {
	r.geta(q, c, &v)
	q.Do()
	return
}

type rx_pipe_reg64 m.Reg64

func (r *rx_pipe_reg64) geta(q *DmaRequest, c sbus.AccessType, v *uint64) {
	(*m.Reg64)(r).Get(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_reg64) seta(q *DmaRequest, c sbus.AccessType, v uint64) {
	(*m.Reg64)(r).Set(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_reg64) get(q *DmaRequest, v *uint64) { r.geta(q, sbus.Duplicate, v) }
func (r *rx_pipe_reg64) set(q *DmaRequest, v uint64)  { r.seta(q, sbus.Duplicate, v) }
func (r *rx_pipe_reg64) getDo(q *DmaRequest, c sbus.AccessType) (v uint64) {
	r.geta(q, c, &v)
	q.Do()
	return
}

type rx_pipe_preg32 m.Preg32
type rx_pipe_portreg32 [1 << m.Log2NRegPorts]rx_pipe_preg32

func (r *rx_pipe_preg32) address() sbus.Address { return (*m.Preg32)(r).Address() }

func (r *rx_pipe_preg32) geta(q *DmaRequest, c sbus.AccessType, v *uint32) {
	(*m.Preg32)(r).Get(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_preg32) seta(q *DmaRequest, c sbus.AccessType, v uint32) {
	(*m.Preg32)(r).Set(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_preg32) set(q *DmaRequest, v uint32) { r.seta(q, sbus.Duplicate, v) }

type rx_pipe_preg64 m.Preg64
type rx_pipe_portreg64 [1 << m.Log2NRegPorts]rx_pipe_preg64

func (r *rx_pipe_preg64) address() sbus.Address { return (*m.Preg64)(r).Address() }

func (r *rx_pipe_preg64) geta(q *DmaRequest, c sbus.AccessType, v *uint64) {
	(*m.Preg64)(r).Get(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_preg64) seta(q *DmaRequest, c sbus.AccessType, v uint64) {
	(*m.Preg64)(r).Set(&q.DmaRequest, 0, BlockRxPipe, c, v)
}
func (r *rx_pipe_preg64) set(q *DmaRequest, v uint64) { r.seta(q, sbus.Duplicate, v) }

const (
	n_pipe    = 4
	n_rx_pipe = n_pipe
	n_tx_pipe = n_pipe
)

type rx_pipe_regs struct {
	// Buffers data from PHY selected by TDM scheduler to cover latency of rx pipe packet processing.
	rx_buffer struct {
		monitor_config                            rx_pipe_reg64
		force_store_and_forward_config            rx_pipe_reg32
		cell_assembly_cpu, cell_assembly_loopback struct {
			control          rx_pipe_reg32
			cut_thru_control rx_pipe_reg32
		}
		_                              [0x2000 - 0x6]rx_pipe_reg32
		tdm_calendar_init              rx_pipe_reg32
		_                              [0x2100 - 0x2001]rx_pipe_reg32
		aux_arb_control                rx_pipe_reg32
		_                              [0x2200 - 0x2101]rx_pipe_reg32
		hw_reset_control_0             rx_pipe_reg32
		_                              [0x2300 - 0x2201]rx_pipe_reg32
		hw_reset_control_1             rx_pipe_reg32
		_                              [0x2600 - 0x2301]rx_pipe_reg32
		sbus_timer                     rx_pipe_reg32
		_                              [0x2900 - 0x2601]rx_pipe_reg32
		ts_to_core_sync_enable         rx_pipe_reg32
		_                              [0x2b00 - 0x2901]rx_pipe_reg32
		cell_assembly_cpu_control      rx_pipe_reg32
		_                              [0x2c00 - 0x2b01]rx_pipe_reg32
		cell_assembly_cpu_status       rx_pipe_reg32
		_                              [0x2e00 - 0x2c01]rx_pipe_reg32
		cell_assembly_loopback_control rx_pipe_reg32
		_                              [0x2f00 - 0x2e01]rx_pipe_reg32
		cell_assembly_loopback_status  rx_pipe_reg32
		_                              [0x04040000 - 0x002f0100]byte
	}

	rx_buffer_tdm_scheduler tdm_regs

	over_subscription_buffer [8]struct {
		control                           rx_pipe_reg32
		port_config                       [4]rx_pipe_reg32
		niv_ethernet_type                 rx_pipe_reg32
		etag_ethernet_type                rx_pipe_reg32
		outer_tpid                        [4]rx_pipe_reg32
		inner_tpid                        [1]rx_pipe_reg32
		protocol_config                   [3][4]rx_pipe_reg64
		threshold                         [4]rx_pipe_reg64
		cut_through_threshold             [4]rx_pipe_reg32
		flow_control_config               [4]rx_pipe_reg64
		flow_control_threshold            [4]rx_pipe_reg64
		usage                             [4]rx_pipe_reg64
		shared_config                     rx_pipe_reg32
		_                                 [0x2e - 0x2d]rx_pipe_reg32
		shared_usage                      rx_pipe_reg64
		packets_dropped                   [obm_n_priority][4]rx_pipe_reg32
		bytes_dropped                     [obm_n_priority][4]rx_pipe_reg64
		flow_control_event_count          [4]rx_pipe_reg64
		max_usage_select                  rx_pipe_reg32
		max_usage                         rx_pipe_reg64
		monitor_stats_config              [4]rx_pipe_reg32
		force_store_and_forward_config    [4]rx_pipe_reg32
		cell_assembly_control             rx_pipe_reg32
		cell_assembly_cut_through_control rx_pipe_reg32
		_                                 [0x60 - 0x5f]rx_pipe_reg32
		ram_control                       rx_pipe_reg32
		_                                 [0x65 - 0x61]rx_pipe_reg32
		cell_assembly_hw_control          rx_pipe_reg32
		cell_assembly_hw_status           rx_pipe_reg64
		cell_assembly_pointer_status      rx_pipe_reg64
		_                                 [0x4000000 - 0x6800]byte
	}

	rx_config                      rx_pipe_reg64
	dos_control_3                  rx_pipe_reg64
	vlan_control                   rx_pipe_reg32
	flexible_ipv6_extension_header rx_pipe_reg32
	multicast_control_1            rx_pipe_reg32
	ecmp_config                    rx_pipe_reg32
	latency_control                rx_pipe_reg32
	module_remapping_control       rx_pipe_portreg32

	_ [0x2c000000 - 0x28000800]byte

	hi_gig_lookup                   rx_pipe_portreg32
	hi_gig_lookup_destination       rx_pipe_portreg32
	sys_reserved_vlan_id            rx_pipe_reg32
	rtag7_hash_control              rx_pipe_reg64
	global_mpls_range               [2]struct{ lower, upper rx_pipe_reg32 }
	remote_cpu_dst_ethernet_address [2]rx_pipe_reg32
	remote_cpu_ethernet_type        rx_pipe_reg32
	mim_ethertype                   rx_pipe_reg32
	outer_tpid                      [4]rx_pipe_reg32
	mmrp, srp                       struct {
		control [2]rx_pipe_reg32
	}
	niv_ethertype                   rx_pipe_reg32
	fcoe_ethertype                  rx_pipe_reg32
	wesp_proto_control              rx_pipe_reg32
	qcn_cntag_ethertype             rx_pipe_reg32
	hbfc_cntag_ethertype            rx_pipe_reg32
	multicast_control_3             rx_pipe_reg32
	sctp_control                    rx_pipe_reg32
	l2gre_control                   rx_pipe_reg32
	pe_ethertype                    rx_pipe_reg32
	ieee_1588_parsing_control       rx_pipe_reg32
	hi_gig_extension_header_control rx_pipe_reg32
	bfd_rx_udp_control              rx_pipe_reg32
	vxlan_control                   rx_pipe_reg32
	from_remote_cpu_dst_address     [2]rx_pipe_reg64
	from_remote_cpu_ethertype       rx_pipe_reg32
	from_remote_cpu_signature       rx_pipe_reg32
	_                               [0x81 - 0x25]rx_pipe_reg32
	multicast_control_2             rx_pipe_reg32
	_                               [0x30000000 - 0x2c008200]byte

	niv_config                      rx_pipe_reg32
	vlan_translate_hash_control     rx_pipe_reg32
	vfp_slice_control               rx_pipe_reg32
	vfp_key_control_1               rx_pipe_reg32
	vfp_key_control_2               rx_pipe_reg32
	vfp_slice_map                   rx_pipe_reg32
	mpls_entry_hash_control         rx_pipe_reg32
	etag_multicast_range            rx_pipe_reg32
	hash_config_0                   rx_pipe_reg32
	cpu_visibility_packet_profile_1 rx_pipe_reg32
	_                               [0x34000000 - 0x30000a00]byte

	mim_default_network_svp   rx_pipe_reg32
	l2gre_default_network_svp rx_pipe_reg32
	trill_adjacency           rx_pipe_portreg64
	vxlan_default_network_svp rx_pipe_reg32
	mpls_tpid                 [4]rx_pipe_reg32
	mpls_inner_tpid           rx_pipe_reg32
	vrf_mask                  rx_pipe_reg32
	l2_tunnel_parse_control   rx_pipe_reg32
	bfd_rx_ach_type_control0  rx_pipe_reg32
	bfd_rx_ach_type_control1  rx_pipe_reg32
	bfd_rx_ach_type_mplstp    rx_pipe_reg32
	bfd_rx_ach_type_mplstp1   rx_pipe_reg32
	_                         [0x38000000 - 0x34000f00]byte

	rtag7_hash_field_selection_bitmaps      [10]rx_pipe_reg32
	rtag7_hash_seed                         [2]rx_pipe_reg32
	drop_control_0                          rx_pipe_reg32
	rtag7_hash_control_2                    rx_pipe_reg32
	rtag7_hash_control_3                    rx_pipe_reg32
	rtag7_hash_field_selection_bitmaps_1    [0x17 - 0x0f]rx_pipe_reg32
	hash_control                            rx_pipe_reg32
	trill_rbridge_nickname_select           rx_pipe_reg32
	rtag7_hash_control_l2gre_mask           [2]rx_pipe_reg32
	rtag7_hash_field_selection_bitmaps_2    [0x1e - 0x1b]rx_pipe_reg32
	bfd_rx_udp_control_1                    rx_pipe_reg32
	vlan_membership_hash_control            rx_pipe_reg32
	dnat_address_type_hash_control          rx_pipe_reg32
	rtag7_hash_control_4                    rx_pipe_reg32
	rtag7_vxlan_payload_l2_hash_field_bmap  rx_pipe_reg32
	rtag7_vxlan_payload_l3_hash_field_bmap  rx_pipe_reg32
	icmp_error_type                         rx_pipe_reg32
	ipv6_min_fragment_size                  rx_pipe_reg32
	dos_control                             rx_pipe_reg32
	dos_control_2                           rx_pipe_reg32
	l2_table_hash_control                   rx_pipe_reg32
	l3_table_hash_control                   rx_pipe_reg32
	rtag7_hash_select                       rx_pipe_reg32
	ecn_control                             rx_pipe_reg32
	_                                       [0x41 - 0x2c]rx_pipe_reg32
	iss_bank_config                         rx_pipe_reg32
	iss_logical_to_physical_bank_map        rx_pipe_reg32
	_                                       [0x4f - 0x43]rx_pipe_reg32
	gtp_control                             rx_pipe_reg32
	_                                       [0x100 - 0x50]rx_pipe_reg32
	ifp_ethernet_type_map                   [16]rx_pipe_reg32
	_                                       [0x120 - 0x110]rx_pipe_reg32
	ifp_l4_src_port_map                     [16]rx_pipe_reg32
	_                                       [0x140 - 0x130]rx_pipe_reg32
	ifp_l4_dst_port_map                     [16]rx_pipe_reg32
	_                                       [0x160 - 0x150]rx_pipe_reg32
	exact_match_logical_table_select_config rx_pipe_reg32
	_                                       [0x398b0000 - 0x38016100]byte
	ilpm_ser_control                        rx_pipe_reg32
	_                                       [0x398d0000 - 0x398b0100]byte
	l3_defip_control                        rx_pipe_reg32
	_                                       [0x398e0000 - 0x398d0100]byte
	l3_defip_key_select                     rx_pipe_reg32
	_                                       [0x398f0000 - 0x398e0100]byte
	l3_defip_aux_control_0                  rx_pipe_reg32
	_                                       [0x39900000 - 0x398f0100]byte
	l3_defip_aux_control_1                  rx_pipe_reg32
	_                                       [0x39910000 - 0x39900100]byte
	l3_defip_alpm_config                    rx_pipe_reg32
	_                                       [0x3c000000 - 0x39910100]byte
	shared_table_hash_control               rx_pipe_reg32
	_                                       [0x3d000000 - 0x3c000100]byte
	iss_memory_control_0                    rx_pipe_reg32
	_                                       [0x100 - 0x001]rx_pipe_reg32
	iss_memory_control_1                    rx_pipe_reg32
	_                                       [0x200 - 0x101]rx_pipe_reg32
	iss_memory_control_2                    rx_pipe_reg32
	_                                       [0x300 - 0x201]rx_pipe_reg32
	iss_memory_control_3                    rx_pipe_reg32
	_                                       [0x400 - 0x301]rx_pipe_reg32
	iss_memory_control_4                    rx_pipe_reg32
	_                                       [0x500 - 0x401]rx_pipe_reg32
	iss_memory_control_5                    rx_pipe_reg32
	_                                       [0x600 - 0x501]rx_pipe_reg32
	iss_memory_control_84                   rx_pipe_reg32
	_                                       [0x40000000 - 0x3d060100]byte

	rep_id_remap_control                   rx_pipe_reg32
	iss_fpem_logical_to_phyysical_bank_map rx_pipe_reg32
	cpu_control_1                          rx_pipe_reg32
	cpu_control_m                          rx_pipe_reg32
	ing_misc_config2                       rx_pipe_reg32
	mc_control_4                           rx_pipe_reg32
	mc_control_5                           rx_pipe_reg32
	_                                      [0x8 - 0x7]rx_pipe_reg32
	cpu_control_0                          rx_pipe_reg32
	priority_control                       rx_pipe_reg32
	trill_drop_control                     rx_pipe_reg32
	_                                      [0x64 - 0xb]rx_pipe_reg32
	cbl_attribute                          [4]rx_pipe_reg32
	_                                      [0x80 - 0x68]rx_pipe_reg32
	protocol_pkt_control                   [64]rx_pipe_reg32
	igmp_mld_pkt_control                   [64]rx_pipe_reg32
	ifp_logical_table_select_config        rx_pipe_reg32
	_                                      [0x44000000 - 0x40010100]byte

	storm_control_meter_mapping           rx_pipe_reg32
	storm_control_meter_config            rx_pipe_portreg32
	_                                     [0x7 - 0x2]rx_pipe_reg32
	ifp_meter_control                     rx_pipe_portreg32
	_                                     [0xb - 0x8]rx_pipe_reg32
	iss_alpm_logical_to_physical_bank_map rx_pipe_reg32
	_                                     [0x48000000 - 0x44000c00]byte

	_                                 [0x3 - 0x0]rx_pipe_reg32
	ifp_slice_meter_map_enable        rx_pipe_reg32
	_                                 [0x100 - 0x4]rx_pipe_reg32
	ifp_config                        [12]rx_pipe_reg32
	_                                 [0x120 - 0x10c]rx_pipe_reg32
	ifp_logical_table_config          [n_ifp_logical_tables]rx_pipe_reg64
	_                                 [0x4c000000 - 0x48014000]byte
	ifp_ecmp_hash_control             rx_pipe_reg32
	ecmp_random_load_balancing_config rx_pipe_reg32
	_                                 [0x50000000 - 0x4c000200]byte

	mirror_select                     rx_pipe_reg32
	module_port_map_select            rx_pipe_portreg32
	local_sw_disable_control          rx_pipe_portreg32
	src_module_id_egress_select       rx_pipe_portreg32
	sflow_rx_rand_seed                rx_pipe_reg32
	sflow_flex_rand_seed              rx_pipe_reg32
	sflow_mirror_config               rx_pipe_reg32
	misc_config                       rx_pipe_reg32
	sw2_ifp_dst_action_control        rx_pipe_reg32
	trunk_rand_load_balancing_seed    rx_pipe_reg32
	hg_trunk_rand_load_balancing_seed rx_pipe_reg32
	cpu_visibility_packet_profile_2   rx_pipe_reg32
	_                                 [0x510b0000 - 0x50000c00]byte
	sw2_hw_control                    rx_pipe_reg32
	_                                 [0x54000000 - 0x510b0100]byte

	sflow_tx_threshold        rx_pipe_portreg32
	sflow_tx_rand_seed        rx_pipe_reg32
	event_debug               [3]rx_pipe_reg32
	counters                  [0x26 - 0x05]rx_pipe_portreg64
	debug_counter_select      [2][9]rx_pipe_reg32
	cos_mode                  rx_pipe_portreg64
	mirror_cos_control        rx_pipe_reg32
	mirror_cpu_cos_config     rx_pipe_reg32
	tx_sflow_cpu_cos_config   rx_pipe_reg32
	instrument_cpu_cos_config rx_pipe_reg32
	ptr_copytocpu_mask        [2]rx_pipe_reg64
	cpu_rqe_queue_num         [2]rx_pipe_reg32
	mirror_rqe_queue_num      rx_pipe_reg32
	_                         [0xb8 - 0x42]rx_pipe_reg32
	nat_counters              [5][16]rx_pipe_reg32
	_                         [0x54040000 - 0x54010800]byte

	flex_counter                       [5]flex_counter_4pool_control
	_                                  [0x5c000000 - 0x54090000]byte
	flex_counter_eviction_control      rx_pipe_reg32
	flex_counter_eviction_counter_flag rx_pipe_reg32
	_                                  [0x60000000 - 0x5c000200]byte

	l2_management_control          rx_pipe_reg32
	l2_learn_control               rx_pipe_reg32
	l2_bulk_control                rx_pipe_reg32
	l2_bulk_ecc_status             rx_pipe_reg32
	aux_l2_bulk_control            rx_pipe_reg32
	l2_mod_fifo_enable             rx_pipe_reg32
	l2_mod_fifo_read_pointer       rx_pipe_reg32
	l2_mod_fifo_write_pointer      rx_pipe_reg32
	_                              [0x9 - 0x8]rx_pipe_reg32
	l2_mod_fifo_claim_avail        rx_pipe_reg32
	l2_management_hw_reset_control [2]rx_pipe_reg32
	l2_management_ser_fifo_control rx_pipe_reg32
	l2_management_interrupt        rx_pipe_reg32
	l2_management_interrupt_enable rx_pipe_reg32
	_                              [0x20 - 0xf]rx_pipe_reg32
	l2_mod_fifo_memory_control_0   rx_pipe_reg32
	l2_mod_fifo_parity_control     rx_pipe_reg32
	_                              [0x64000000 - 0x60002200]byte
}

type rx_pipe_mem32 m.Mem32

func (x *rx_pipe_mem32) seta(q *DmaRequest, a sbus.AccessType, v uint32) {
	(*m.Mem32)(x).Set(&q.DmaRequest, BlockRxPipe, a, v)
}
func (x *rx_pipe_mem32) geta(q *DmaRequest, a sbus.AccessType, v *uint32) {
	(*m.Mem32)(x).Get(&q.DmaRequest, BlockRxPipe, a, v)
}
func (x *rx_pipe_mem32) set(q *DmaRequest, v uint32) { x.seta(q, sbus.Duplicate, v) }

type rx_pipe_mems struct {
	_                                     [0x00080000 - 0x0]byte
	idb_to_pipe_port_number_mapping_table [n_idb_port]rx_pipe_mem32
	_                                     [m.MemMax - n_idb_port]m.MemElt
	_                                     [0x04000000 - 0xc0000]byte

	tdm_calendar [2]struct {
		entries [128]tdm_calendar_mem
		_       [m.MemMax - 128]m.MemElt
	}
	_ [0x08000000 - 0x04080000]byte

	over_subscription_buffer [8]struct {
		dscp_map                 [4][m.MemMax]m.Mem32
		priority_map             [4][m.MemMax]m.Mem32
		etag_map                 [4][m.MemMax]m.Mem32
		iom_stats_window_results m.Mem
		_                        [0x04000000 - 0x00340000]byte
	}

	port_table [n_pipe_ports + 1]rx_port_table_mem
	_          [m.MemMax - (n_pipe_ports + 1)]m.MemElt

	system_config_table_modbase m.Mem
	system_config_table         m.Mem

	source_trunk_map_modbase [n_global_physical_port]m.Mem32
	_                        [m.MemMax - n_global_physical_port]m.MemElt
	_                        [0x2c000000 - 0x28100000]byte
	source_trunk_map         [8 << 10]source_trunk_map_mem
	_                        [m.MemMax - 8<<10]m.MemElt

	l3_tunnel m.Mem

	udf_tcam   m.Mem
	udf_offset m.Mem

	mod_map          m.Mem
	source_mod_proxy m.Mem

	lport_profile_table [n_pipe_ports + 1]rx_port_table_mem
	_                   [m.MemMax - (n_pipe_ports + 1)]m.MemElt

	ipv4_in_ipv6_prefix_match m.Mem

	vlan_range [256]vlan_range_mem
	_          [m.MemMax - 256]m.MemElt

	cpu_traffic_class_map [256]m.Mem32
	_                     [m.MemMax - 256]m.MemElt

	trill_parse_control m.Mem
	fc_header_type      m.Mem
	source_vp_2         m.Mem

	l3_tunnel_data_only m.Mem
	l3_tunnel_only      m.Mem
	_                   [0x30000000 - 0x2c3c0000]byte

	vlan_protocol [1 << 4]m.Mem32
	_             [m.MemMax - 16]m.MemElt

	vlan_protocol_data [1 << 7][1 << 4]vlan_protocol_data_mem
	_                  [m.MemMax - (1 << 11)]m.MemElt

	vlan_subnet           m.Mem
	vlan_subnet_only      m.Mem
	vlan_subnet_data_only m.Mem

	vlan_mac m.Mem

	vlan_translate m.Mem

	vfp_tcam         m.Mem
	vfp_policy_table m.Mem

	vlan_tag_action_profile [n_vlan_tag_action_profile_entries]rx_vlan_tag_action_mem
	_                       [m.MemMax - n_vlan_tag_action_profile_entries]m.MemElt

	mpls_entry [n_mpls_entry]mpls_entry_mem
	_          [m.MemMax - n_mpls_entry]m.MemElt

	udf_conditional_check_table_cam m.Mem
	udf_conditional_check_table_ram m.Mem

	etag_pcp_mapping m.Mem

	vxlt_remap_table  [2]m.Mem
	vxlt_action_table [2]m.Mem

	mpls_entry_remap_table  [2]m.Mem
	mpls_entry_action_table [2]m.Mem

	vlan_translate_ecc m.Mem
	mpls_entry_ecc     m.Mem

	_ [0x34000000 - 0x30600000]byte

	vlan_mpls m.Mem

	my_station_tcam [n_my_station_tcam_entry]my_station_tcam_mem
	_               [m.MemMax - n_my_station_tcam_entry]m.MemElt

	my_station_tcam_entry_only [n_my_station_tcam_entry]my_station_tcam_entry_only_mem
	_                          [m.MemMax - n_my_station_tcam_entry]m.MemElt

	my_station_tcam_data_only [n_my_station_tcam_entry]my_station_tcam_data_only_mem
	_                         [m.MemMax - n_my_station_tcam_entry]m.MemElt

	source_vp m.Mem

	vfi   m.Mem
	vfi_1 m.Mem

	l3_interface [n_l3_interface]rx_l3_interface_mem
	_            [m.MemMax - n_l3_interface]m.MemElt

	ing_trill_payload_parse_control m.Mem

	visibility_packet_capture_buffer_ivp [2]m.Mem64
	_                                    [m.MemMax - 2]m.MemElt

	_ [0x38000000 - 0x34280000]byte

	ing_vp_vlan_membership m.Mem

	_ m.Mem

	vrf [n_vrf]vrf_mem
	_   [m.MemMax - n_vrf]m.MemElt

	vlan [n_vlan]rx_vlan_mem
	_    [m.MemMax - n_vlan]m.MemElt

	vlan_spanning_tree_group [n_vlan_spanning_tree_group_entry]rx_vlan_spanning_tree_group_mem
	_                        [m.MemMax - n_vlan_spanning_tree_group_entry]m.MemElt

	vlan_profile [128]m.Mem64
	_            [m.MemMax - 128]m.MemElt

	ing_outer_dot1p_mapping_table m.Mem

	vfp_hash_field_bmap_table [2]m.Mem

	gtp_port_table m.Mem

	ip_option_control_profile_table m.Mem

	l3_interface_profile [256]rx_l3_interface_profile_mem
	_                    [m.MemMax - 256]m.MemElt

	ip_multicast_tcam m.Mem

	dnat_address_type m.Mem

	ipv6_multicast_reserved_address m.Mem

	l2_hit_da_only m.Mem
	l2_hit_sa_only m.Mem

	l2_user_entry           [n_l2_user_entry]l2_user_entry_mem
	_                       [m.MemMax - n_l2_user_entry]m.MemElt
	l2_user_entry_only      m.Mem
	l2_user_entry_data_only m.Mem

	dvp_table m.Mem

	l2_entry_tile      m.Mem
	l2_entry_only_tile m.Mem

	l3_entry_only         m.Mem
	l3_entry_ipv4_unicast struct {
		dedicated [n_iss_banks][512][n_iss_bits_per_bucket / 105]l3_ipv4_entry_mem
		shared    [n_iss_banks][n_iss_buckets_per_bank][n_iss_bits_per_bucket / 105]l3_ipv4_entry_mem
		_         [m.MemMax - n_iss_banks*(512+n_iss_buckets_per_bank)*(n_iss_bits_per_bucket/105)]m.MemElt
	}
	l3_entry_ipv4_multicast m.Mem
	l3_entry_ipv6_unicast   m.Mem
	l3_entry_ipv6_multicast m.Mem

	active_l3_interface_profile m.Mem

	fpem_ecc m.Mem

	rtag7_flow_based_hash m.Mem
	rtag7_port_based_hash m.Mem

	l2_entry_only_ecc m.Mem
	l3_entry_only_ecc m.Mem

	ieee_1588_ingress_control m.Mem

	responsive_protocol_match m.Mem

	ip_to_int_cn_mapping m.Mem

	tunnel_ecn_decap   m.Mem
	tunnel_ecn_decap_2 m.Mem

	_ [0x39900000 - 0x389c0000]byte

	ip46_src_compression           m.Mem
	ip46_src_compression_data_only m.Mem
	ip46_dst_compression           m.Mem
	ip46_dst_compression_data_only m.Mem

	exact_match_logical_table_select           m.Mem
	_                                          [0x39b00000 - 0x39a40000]byte
	exact_match_logical_table_select_data_only m.Mem
	exact_match_key_gen_program_profile        m.Mem

	ip_proto_map m.Mem
	_            [0x39c00000 - 0x39bc0000]byte

	exact_match_key_gen_mask m.Mem
	_                        [0x39c80000 - 0x39c40000]byte
	exact_match_2            [4][8 << 10][2]exact_match_mem
	_                        [m.MemMax - 4*(8<<10)*2]m.MemElt
	exact_match_2_entry_only m.Mem
	exact_match_4            m.Mem
	exact_match_4_entry_only m.Mem

	tcp_fn m.Mem
	ttl_fn m.Mem
	tos_fn m.Mem

	ifp_range_check [32]m.Mem64
	_               [m.MemMax - 32]m.MemElt

	src_compression_tcam_only m.Mem
	dst_compression_tcam_only m.Mem

	exact_match_logical_table_select_tcam_only m.Mem

	_ [0x3a080000 - 0x39f40000]byte

	l3_entry_lp     m.Mem
	l3_entry_iss_lp m.Mem
	fpem_lp         m.Mem
	_               [0x3a400000 - 0x3a140000]byte

	l3_defip [n_l3_defip_entries]l3_defip_mem
	_        [m.MemMax - n_l3_defip_entries]m.MemElt

	l3_defip_only [n_l3_defip_entries]l3_defip_tcam_only_mem
	_             [m.MemMax - n_l3_defip_entries]m.MemElt

	l3_defip_data_only [n_l3_defip_entries]l3_defip_tcam_data_only_mem
	_                  [m.MemMax - n_l3_defip_entries]m.MemElt

	l3_defip_pair_128 [n_l3_defip_entries / 2]l3_defip_pair_mem
	_                 [m.MemMax - n_l3_defip_entries/2]m.MemElt

	l3_defip_pair_128_only [n_l3_defip_entries / 2]l3_defip_pair_tcam_only_mem
	_                      [m.MemMax - n_l3_defip_entries/2]m.MemElt

	l3_defip_pair_128_data_only [n_l3_defip_entries / 2]l3_defip_pair_tcam_data_only_mem
	_                           [m.MemMax - n_l3_defip_entries/2]m.MemElt

	l3_defip_alpm_ipv4 [n_iss_bits_per_bucket / 70][n_iss_buckets_per_bank][n_iss_banks]l3_defip_alpm_ip4_mem
	_                  [m.MemMax - (n_iss_bits_per_bucket/70)*n_iss_buckets_per_bank*n_iss_banks]m.MemElt

	l3_defip_alpm_ipv4_with_flex_counters [n_iss_bits_per_bucket / 105][n_iss_buckets_per_bank][n_iss_banks]l3_defip_alpm_ip4_with_flex_counter_mem
	_                                     [m.MemMax - (n_iss_bits_per_bucket/105)*n_iss_buckets_per_bank*n_iss_banks]m.MemElt

	l3_defip_alpm_ipv6_64 [n_iss_bits_per_bucket / 105][n_iss_buckets_per_bank][n_iss_banks]l3_defip_alpm_ip6_64_mem
	_                     [m.MemMax - (n_iss_bits_per_bucket/105)*n_iss_buckets_per_bank*n_iss_banks]m.MemElt

	l3_defip_alpm_ipv6_64_with_flex_counters [n_iss_bits_per_bucket / 140][n_iss_buckets_per_bank][n_iss_banks]l3_defip_alpm_ip6_64_with_flex_counter_mem
	_                                        [m.MemMax - (n_iss_bits_per_bucket/140)*n_iss_buckets_per_bank*n_iss_banks]m.MemElt

	l3_defip_alpm_ipv6_128 [n_iss_bits_per_bucket / 210][n_iss_buckets_per_bank][n_iss_banks]l3_defip_alpm_ip6_128_mem
	_                      [m.MemMax - (n_iss_bits_per_bucket/210)*n_iss_buckets_per_bank*n_iss_banks]m.MemElt

	l3_defip_alpm_raw          m.Mem
	l3_defip_aux_table         m.Mem
	l3_defip_aux_scratch       m.Mem
	l3_defip_aux_hitbit_update m.Mem
	l3_defip_alpm_ecc          m.Mem
	_                          [0x40000000 - 0x3a800000]byte

	initial_l3_ecmp_group [2][1 << 10]ecmp_group_mem
	_                     [m.MemMax - 2<<10]m.MemElt

	initial_l3_ecmp [2][8 << 10]ecmp_mem
	_               [m.MemMax - 16<<10]m.MemElt

	initial_prot_nhi_table   m.Mem
	initial_prot_group_table m.Mem

	trunk_cbl_table m.Mem

	port_cbl_table_modbase m.Mem
	port_cbl_table         m.Mem

	dscp_table m.Mem

	pri_cng_map m.Mem

	untagged_phb m.Mem

	_ [0x41d00000 - 0x40280000]byte

	ifp_logical_table_select           [n_ifp_slice][n_ifp_logical_tables]ifp_logical_table_select_mem
	_                                  [m.MemMax - n_ifp_slice*n_ifp_logical_tables]m.MemElt
	ifp_logical_table_select_data_only [n_ifp_slice][n_ifp_logical_tables]ifp_logical_table_select_data_only_mem
	_                                  [m.MemMax - n_ifp_slice*n_ifp_logical_tables]m.MemElt
	_                                  [0x42040000 - 0x41d80000]byte
	ifp_key_gen_program_profile        [n_ifp_key_generation_profiles]ifp_key_generation_profile_mem
	_                                  [m.MemMax - n_ifp_key_generation_profiles]m.MemElt
	_                                  [0x42140000 - 0x42080000]byte
	ifp_key_gen_program_profile2       [n_ifp_key_generation_profiles]m.Mem32
	_                                  [m.MemMax - n_ifp_key_generation_profiles]m.MemElt
	ifp_logical_table_select_tcam_only [n_ifp_slice][n_ifp_logical_tables]ifp_logical_table_select_tcam_only_mem
	_                                  [m.MemMax - n_ifp_slice*n_ifp_logical_tables]m.MemElt

	exact_match_default_policy m.Mem

	_ [0x44000000 - 0x42200000]byte

	initial_l3_next_hop m.Mem

	l3_entry_hit_only          m.Mem
	l3_defip_hit_only          m.Mem
	l3_defip_pair_128_hit_only m.Mem
	l3_defip_alpm_hit_only     m.Mem

	visibility_packet_capture_buffer_isw1 [4]m.Mem64
	_                                     [m.MemMax - 4]m.MemElt

	_ [0x47200000 - 0x44180000]byte

	exact_match_hit_only m.Mem
	_                    [0x48000000 - 0x47240000]byte

	_ [0x482c0000 - 0x48000000]byte

	dvp_2_table m.Mem

	ifp_i2e_classid_select m.Mem
	ifp_hg_classid_select  m.Mem

	ifp_tcam [n_ifp_slice][n_ifp_tcam_elts_per_slice]ifp_tcam_80bit_mem
	_        [m.MemMax - n_ifp_slice*n_ifp_tcam_elts_per_slice]m.MemElt

	ifp_tcam_wide [n_ifp_slice][n_ifp_tcam_elts_per_slice / 2]ifp_tcam_160bit_mem
	_             [m.MemMax - n_ifp_slice*n_ifp_tcam_elts_per_slice/2]m.MemElt

	_ [1]m.Mem

	ifp_meter_table          m.Mem
	ifp_storm_control_meters m.Mem
	ifp_port_meter_map       m.Mem

	exact_match_qos_actions_profile m.Mem
	_                               [1]m.Mem
	exact_match_action_profile      m.Mem
	_                               [1]m.Mem

	ifp_logical_table_action_priority m.Mem
	_                                 [0x4b200000 - 0x48640000]byte
	ifp_policy_table                  [n_ifp_slice][n_ifp_tcam_elts_per_slice]ifp_policy_mem
	_                                 [m.MemMax - n_ifp_slice*n_ifp_tcam_elts_per_slice]m.MemElt

	eh_mask_profile m.Mem

	_ [0x4c000000 - 0x4b280000]byte

	l3_ecmp_group [2][1 << 10]ecmp_group_mem
	_             [m.MemMax - 2<<10]m.MemElt
	l3_ecmp       [2][8 << 10]ecmp_mem
	_             [m.MemMax - 16<<10]m.MemElt

	_ [1]m.Mem

	l3_next_hop [n_next_hop]rx_next_hop_mem
	_           [m.MemMax - n_next_hop]m.MemElt

	ifp_redirection_profile [m.MemMax]port_bitmap_mem

	l2_multicast [n_l2_multicast_entry]l2_multicast_mem
	_            [m.MemMax - n_l2_multicast_entry]m.MemElt

	l3_multicast_remap m.Mem
	l3_multicast       [n_l3_multicast_entry]l3_multicast_mem
	_                  [m.MemMax - n_l3_multicast_entry]m.MemElt

	icontrol_opcode_bitmap m.Mem

	cpu_port_bitmap [m.MemMax]port_bitmap_mem

	egr_mask_modbase m.Mem

	snat           m.Mem
	snat_only      m.Mem
	snat_data_only m.Mem
	snat_hit_only  m.Mem
	_              [0x50000000 - 0x4c3c0000]byte

	trunk_group   m.Mem
	_             [2]m.Mem
	hg_trunk_mode m.Mem

	dest_trunk_bitmap [m.MemMax]port_bitmap_mem
	egr_mask          [m.MemMax]port_bitmap_mem

	l3_interface_mtu [m.N_cast][8 << 10]m.Mem32
	_                [m.MemMax - 2*(8<<10)]m.MemElt

	modport_map_sw     m.Mem
	modport_map_mirror m.Mem
	modport_map        [4]m.Mem

	pw_term_seq_num m.Mem

	src_modid_egress [m.MemMax]port_bitmap_mem

	hg_trunk_group  m.Mem
	hg_trunk_member m.Mem

	_ [1]m.Mem

	trunk_bitmap_table m.Mem

	mac_block_table [m.MemMax]port_bitmap_mem

	nonucast_trunk_block_mask [m.MemMax]port_bitmap_mem

	im_mtp_index m.Mem

	em_mtp_index m.Mem

	src_modid_ingress_block [m.MemMax]port_bitmap_mem

	alternate_emirror_bitmap m.Mem

	port_lag_failover_set m.Mem

	hg_trunk_failover_set m.Mem

	vlan_profile_2 m.Mem

	unknown_ucast_block_mask [m.MemMax]port_bitmap_mem

	unknown_mcast_block_mask [m.MemMax]port_bitmap_mem

	bcast_block_mask [m.MemMax]port_bitmap_mem

	egrmskbmap [m.MemMax]port_bitmap_mem

	local_sw_disable_default_pbm [m.MemMax]port_bitmap_mem

	known_mcast_block_mask [m.MemMax]port_bitmap_mem

	local_sw_disable_default_pbm_mirr m.Mem

	imirror_bitmap m.Mem

	vlan_membership_check_enable_port_bitmap [m.MemMax]port_bitmap_mem

	higig_trunk_control m.Mem

	_ [1]m.Mem

	hg_trunk_bitmap m.Mem

	hg_trunk_failover_enable m.Mem

	link_status [m.MemMax]port_bitmap_mem

	port_bridge_bmap [m.MemMax]port_bitmap_mem

	multipass_loopback_bitmap [m.MemMax]port_bitmap_mem

	mirror_control m.Mem

	ing_routed_int_pri_mapping m.Mem

	ing_higig_trunk_override_profile m.Mem

	trunk_member m.Mem

	fast_trunk_group m.Mem

	visibility_packet_capture_buffer_isw2 [8]m.Mem64
	_                                     [m.MemMax - 8]m.MemElt

	device_port_by_global_physical [n_global_physical_port]rx_pipe_mem32
	_                              [m.MemMax - n_global_physical_port]m.MemElt

	sflow_ing_data_source m.Mem

	sflow_ing_flex_data_source m.Mem

	_ [0x54000000 - 0x50d40000]byte

	cos_map m.Mem

	cpu_cos_map [n_cpu_cos_map]cpu_cos_map_mem
	_           [m.MemMax - n_cpu_cos_map]m.MemElt

	cpu_cos_map_only m.Mem

	cpu_cos_map_data_only m.Mem

	emirror_control [4]m.Mem

	unknown_hgi_bitmap m.Mem

	epc_link_port_bitmap [m.MemMax]port_bitmap_mem

	port_bridge_mirror_bmap [m.MemMax]port_bitmap_mem

	cos_map_sel m.Mem

	trill_drop_stats m.Mem

	phb2_cos_map m.Mem

	cpu_port_bitmap_1 [m.MemMax]port_bitmap_mem

	ing_flex_counter_pkt_res_map m.Mem

	ing_flex_counter_tos_map m.Mem

	ing_flex_counter_port_map m.Mem

	ing_flex_counter_pkt_pri_map m.Mem

	ing_flex_counter_pri_cng_map m.Mem

	ifp_cos_map m.Mem

	int_cn_to_mmuif_mapping m.Mem

	agm_monitor_table m.Mem

	loopback_port_bitmap [m.MemMax]port_bitmap_mem

	instrumentation_triggers_enable [m.MemMax]port_bitmap_mem

	_ [0x56800000 - 0x54640000]byte

	// pools 0-7:  4k counters per pipe => 8*4k = 32k counters per pipe
	// pools 8-19: 512 counters per pipe => 12*512 = 6k counters per pipe
	flex_counter0 [3]flex_counter_4pool_mems // first 12 pools
	_             [1]flex_counter_4pool_mems
	flex_counter1 [2]flex_counter_4pool_mems // last 8 pools

	_ [0x5c000000 - 0x568c0000]byte

	flex_counter_eviction_fifo m.Mem

	_ [0x60000000 - 0x5c040000]byte

	_ [1]m.Mem

	l2_bulk m.Mem

	_ [0x60100000 - 0x60080000]byte

	l2_mod_fifo m.Mem

	l2_learn_insert_failure m.Mem

	l2_management_ser_fifo m.Mem

	_ [0x60540000 - 0x601c0000]byte

	l2_entry struct {
		dedicated [n_l2_dedicated_banks][n_l2_dedicated_buckets_per_bank][n_l2_entry_per_bucket]l2_entry_mem
		shared    [n_iss_banks][n_iss_buckets_per_bank][n_l2_entry_per_bucket]l2_entry_mem
		_         [m.MemMax - n_l2_entry_per_bucket*(n_l2_dedicated_banks*n_l2_dedicated_buckets_per_bank+n_iss_banks*n_iss_buckets_per_bank)]m.MemElt
	}
}
