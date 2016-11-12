// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//+build debug

package fe1a

import (
	. "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/debug"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"

	"fmt"
	"unsafe"
)

var base = m.BasePointer

func check(tag string, p unsafe.Pointer, expect uint) {
	CheckAddr(tag, uint(uintptr(p)-uintptr(base)), expect)
}

// Verify top memory map.
func init() {
	r := (*top_controller)(base)
	check("port_reset", unsafe.Pointer(&r.port_reset), 0x2fc)
	check("tsc_disable", unsafe.Pointer(&r.tsc_disable), 0x600)
	check("core_pll_frequency_select", unsafe.Pointer(&r.core_pll_frequency_select), 0x75c)
	check("l1_received_clock_valid_status34", unsafe.Pointer(&r.l1_received_clock_valid_status34), 0x77c)
	check("port_enable[0]", unsafe.Pointer(&r.port_enable[0]), 0x788)
	check("temperature_sensor_interrupt", unsafe.Pointer(&r.temperature_sensor_interrupt.enable), 0x800)
	check("tsc_resolved_speed_status", unsafe.Pointer(&r.tsc_resolved_speed_status[0]), 0x87c)
}

// Verify rx pipe memory maps.
func init() {
	r := (*rx_pipe_controller)(base)
	m := (*rx_pipe_mems)(base)

	check("over_subscription_buffer[0]", unsafe.Pointer(&m.over_subscription_buffer[0]), 0x08000000)
	check("port_table", unsafe.Pointer(&m.port_table[0]), 0x28000000)
	check("source_trunk_map", unsafe.Pointer(&m.source_trunk_map[0]), 0x2c000000)
	check("vlan_protocol", unsafe.Pointer(&m.vlan_protocol[0]), 0x30000000)
	check("vlan_mpls", unsafe.Pointer(&m.vlan_mpls[0]), 0x34000000)
	check("vp_vlan_membership", unsafe.Pointer(&m.vp_vlan_membership[0]), 0x38000000)
	check("dvp_table", unsafe.Pointer(&m.dvp_table), 0x38500000)
	check("ip46_dst_compression", unsafe.Pointer(&m.ip46_dst_compression[0]), 0x39980000)
	check("exact_match_2", unsafe.Pointer(&m.exact_match_2), 0x39c80000)
	check("fib_tcam_bucket_raw", unsafe.Pointer(&m.fib_tcam_bucket_raw), 0x3a6c0000)
	check("initial_l3_ecmp_group", unsafe.Pointer(&m.initial_l3_ecmp_group), 0x40000000)
	check("initial_l3_next_hop", unsafe.Pointer(&m.initial_l3_next_hop), 0x44000000)
	check("dvp_2_table", unsafe.Pointer(&m.dvp_2_table), 0x482c0000)
	check("l3_ecmp_group", unsafe.Pointer(&m.l3_ecmp_group), 0x4c000000)
	check("trunk_group", unsafe.Pointer(&m.trunk_group), 0x50000000)
	check("cos_map", unsafe.Pointer(&m.cos_map), 0x54000000)
	check("pipe_counter_eviction_fifo", unsafe.Pointer(&m.pipe_counter_eviction_fifo), 0x5c000000)
	check("l2_bulk", unsafe.Pointer(&m.l2_bulk), 0x60040000)

	check("rx_buffer_tdm_scheduler", unsafe.Pointer(&r.rx_buffer_tdm_scheduler), 0x04040000)
	check("over_subscription_buffer[0]", unsafe.Pointer(&r.over_subscription_buffer[0]), 0x08000000)
	check("over_subscription_buffer[1]", unsafe.Pointer(&r.over_subscription_buffer[1]), 0x08000000+0x04000000)
	check("rx_config", unsafe.Pointer(&r.rx_config), 0x28000000)
	check("hi_gig_lookup", unsafe.Pointer(&r.hi_gig_lookup), 0x2c000000)
	check("niv_config", unsafe.Pointer(&r.niv_config), 0x30000000)
	check("mim_default_network_svp", unsafe.Pointer(&r.mim_default_network_svp), 0x34000000)
	check("rtag7_hash_field_selection_bitmaps", unsafe.Pointer(&r.rtag7_hash_field_selection_bitmaps[0]), 0x38000000)
	check("shared_lookup_sram_bank_config", unsafe.Pointer(&r.shared_lookup_sram_bank_config), 0x38004100)
	check("shared_table_hash_control", unsafe.Pointer(&r.shared_table_hash_control), 0x3c000000)
	check("shared_lookup_sram_memory_control_0", unsafe.Pointer(&r.shared_lookup_sram_memory_control_0), 0x3d000000)
	check("rep_id_remap_control", unsafe.Pointer(&r.rep_id_remap_control), 0x40000000)
	check("storm_control_meter_mapping", unsafe.Pointer(&r.storm_control_meter_mapping), 0x44000000)
	check("rxf_slice_meter_map_enable", unsafe.Pointer(&r.rxf_slice_meter_map_enable), 0x48000300)
	check("rxf_ecmp_hash_control", unsafe.Pointer(&r.rxf_ecmp_hash_control), 0x4c000000)
	check("mirror_select", unsafe.Pointer(&r.mirror_select), 0x50000000)
	check("sflow_tx_threshold", unsafe.Pointer(&r.sflow_tx_threshold), 0x54000000)
	check("pipe_counter_eviction_control", unsafe.Pointer(&r.pipe_counter_eviction_control), 0x5c000000)
	check("l2_management_control", unsafe.Pointer(&r.l2_management_control), 0x60000000)
}

// Verify tx pipe memory maps.
func init() {
	r := (*tx_pipe_controller)(base)
	m := (*tx_pipe_mems)(base)

	check("latency_mode", unsafe.Pointer(&r.latency_mode), 0x00000000)
	check("config", unsafe.Pointer(&r.config), 0x04000000)
	check("outer_tpid", unsafe.Pointer(&r.outer_tpid), 0x08000000)
	check("port_debug", unsafe.Pointer(&r.port_debug), 0x0c000000)
	check("trill_header_attributes", unsafe.Pointer(&r.trill_header_attributes), 0x10000000)
	check("tunnel_pimdr", unsafe.Pointer(&r.tunnel_pimdr), 0x14000000)
	check("multicast_control_1", unsafe.Pointer(&r.multicast_control_1), 0x18000000)
	check("wesp_proto_control", unsafe.Pointer(&r.wesp_proto_control), 0x1c000000)
	check("txf_slice_control", unsafe.Pointer(&r.txf_slice_control), 0x20000000)
	check("event_debug", unsafe.Pointer(&r.event_debug), 0x24000000)
	check("counters", unsafe.Pointer(&r.counters), 0x28000000)
	check("pipe_counters", unsafe.Pointer(&r.pipe_counter), 0x28040000)

	check("l3_next_hop", unsafe.Pointer(&m.l3_next_hop), 0x04000000)
	check("int_cn_update", unsafe.Pointer(&m.int_cn_update), 0x05900000)
	check("vlan_translate", unsafe.Pointer(&m.vlan_translate), 0x08000000)
	check("vlan_translate_action_table", unsafe.Pointer(&m.vlan_translate_action_table), 0x08780000)
	check("mpls_exp_pri_mapping", unsafe.Pointer(&m.mpls_exp_pri_mapping), 0x10000000)
	check("trill_parse_control_2", unsafe.Pointer(&m.trill_parse_control_2), 0x1c000000)
	check("pipe_counter_maps.packet_resolution", unsafe.Pointer(&m.pipe_counter_maps.packet_resolution), 0x24080000)
	check("txf_counter_table", unsafe.Pointer(&m.txf_counter_table), 0x28000000)
	check("port_enable", unsafe.Pointer(&m.port_enable), 0x28200000)
	check("pipe_counter", unsafe.Pointer(&m.pipe_counter), 0x2a800000)
}

// Verify mmu global memory maps.
func init() {
	r := (*mmu_global_controller)(base)

	check("misc_config", unsafe.Pointer(&r.misc_config), 0x08000000)
	check("global_physical_port_by_mmu_port", unsafe.Pointer(&r.global_physical_port_by_mmu_port), 0x08120000)
}

// Verify mmu pipe memory maps.
func init() {
	r := (*mmu_pipe_regs)(base)
	m := (*mmu_pipe_mems)(base)

	check("cut_through_purge_count", unsafe.Pointer(&r.cut_through_purge_count), 0x10002b00)
	check("time_domain", unsafe.Pointer(&r.time_domain), 0x20000000)
	check("pqe[0]", unsafe.Pointer(&r.pqe[0]), 0x24000000)
	check("clear_counters", unsafe.Pointer(&r.clear_counters), 0x28000000)
	check("rqe_priority_scheduling_type", unsafe.Pointer(&r.rqe_priority_scheduling_type), 0x2c003600)
	check("thdu_bypass", unsafe.Pointer(&r.tx_admission_control.bypass), 0x38001000)
	check("mmu_thdm_db_pool_shared_limit", unsafe.Pointer(&r.multicast_admission_control.db.service_pool_shared_limit), 0x3c000400)
	check("mmu_thdm_mcqe_pool_shared_limit", unsafe.Pointer(&r.multicast_admission_control.mcqe.pool_shared_limit), 0x40000400)
	check("db", unsafe.Pointer(&r.db), 0x44000000)
	check("db.calculated_color_resume_limits", unsafe.Pointer(&r.db.calculated_color_resume_limits), 0x44008000)
	check("qe", unsafe.Pointer(&r.qe), 0x48010000)
	check("interrupt_enable", unsafe.Pointer(&r.interrupt_enable), 0x4c000100)

	check("enqx_pipemem", unsafe.Pointer(&m.enqx_pipemem), 0x04040000)
	check("packet_header", unsafe.Pointer(&m.packet_header), 0x10040000)
	check("copy_count", unsafe.Pointer(&m.copy_count), 0x14000000)
	check("mmu_wred_config", unsafe.Pointer(&m.wred.config), 0x20000000)
	check("mmu_pqe_mem", unsafe.Pointer(&m.mmu_pqe_mem), 0x24000000)
	check("unicast_tx_drops", unsafe.Pointer(&m.unicast_tx_drops), 0x28000000)
	check("replication_fifo_bank", unsafe.Pointer(&m.replication_fifo_bank), 0x2c000000)
	check("wred_drop_curve_profile", unsafe.Pointer(&m.wred_drop_curve_profile), 0x30000000)
	check("tx_purge_queue_memory", unsafe.Pointer(&m.tx_purge_queue_memory), 0x34000000)
	check("mmu_thdu_queue_to_queue_group_map", unsafe.Pointer(&m.tx_admission_control.queue_to_queue_group_map), 0x38000000)
	check("mmu_thdm_db_queue_config", unsafe.Pointer(&m.multicast_admission_control.db_queue_config), 0x3c000000)
	check("mmu_thdm_mcqe_queue_config", unsafe.Pointer(&m.multicast_admission_control.mcqe_queue_config), 0x40000000)
}

// Verify mmu sc memory maps.
func init() {
	r := (*mmu_slice_controller)(base)
	m := (*mmu_slice_mems)(base)

	check("toq.fatal_error", unsafe.Pointer(&r.toq.fatal_error), 0x08000200)
	check("l3_multicast_port_aggregate_id", unsafe.Pointer(&r.l3_multicast_port_aggregate_id), 0x0c100000)
	check("cfap.config", unsafe.Pointer(&r.cfap.config), 0x10000000)
	check("mtro_refresh_config", unsafe.Pointer(&r.mtro_refresh_config), 0x14000000)
	check("prio2cos_profile", unsafe.Pointer(&r.prio2cos_profile), 0x28000000)
	check("queue_scheduler.port_flush", unsafe.Pointer(&r.queue_scheduler.port_flush), 0x34000100)
	check("mmu_port_credit", unsafe.Pointer(&r.mmu_port_credit), 0x38000000)
	check("misc_config", unsafe.Pointer(&r.misc_config), 0x3c000000)

	check("cell_link", unsafe.Pointer(&m.cell_link), 0x08040000)
	check("replication.state", unsafe.Pointer(&m.replication.state), 0x0c040000)
	check("cfap_bank", unsafe.Pointer(&m.cfap_bank), 0x10040000)
	check("tx_metering.config", unsafe.Pointer(&m.tx_metering.config), 0x14000000)
	check("queue_scheduler.l2_accumulated_compensation", unsafe.Pointer(&m.queue_scheduler.l2_accumulated_compensation), 0x34000000)
	check("tdm_calendar", unsafe.Pointer(&m.tdm_calendar), 0x38000000)
	check("cbp_data_slices_01", unsafe.Pointer(&m.cbp_data_slices[0]), 0x50000000)
	check("cbp_data_slices_23", unsafe.Pointer(&m.cbp_data_slices[1]), 0x54000000)
}

// Verify rxf bit extractor offsets.
func init() {
	x := get_rxf_field_extractor_l0_bus()
	if got, want := x.u32.aux_ab.b.offset(), uint(0*32); got != want {
		panic(fmt.Errorf("u32 %d %d", got, want))
	}
	if got, want := x.u16.aux_cd.d.offset(), uint(1*32); got != want {
		panic(fmt.Errorf("u16 %d %d", got, want))
	}
	if got, want := x.u8.rx_physical_port.offset(), uint(2*32); got != want {
		panic(fmt.Errorf("u8 %d %d", got, want))
	}
	if got, want := x.u4.vfi_11_8.offset(), uint(3*32); got != want {
		panic(fmt.Errorf("u4 %d %d", got, want))
	}
	if got, want := x.u2.nat_src_realm_id.offset(), uint(4*32); got != want {
		panic(fmt.Errorf("u2 %d %d", got, want))
	}
}
