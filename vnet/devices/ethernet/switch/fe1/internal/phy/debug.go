// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build debug

package phy

import (
	. "github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/debug"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
)

func (r *pcs_reg) offset() uint      { return uint((*u16)(r).offset()) }
func (r *pcs_lane_reg) offset() uint { return uint((*u16)(r).offset()) }
func (r *pmd_lane_reg) offset() uint { return uint((*u16)(r).offset()) }
func (r *pcs_reg_32) offset() uint   { return uint(r[0].offset()) }
func (r *pmd_reg_32) offset() uint   { return uint(r[0].offset()) }

// Check TSCF memory map.
func init() {
	r := (*tscf_regs)(m.BasePointer)
	CheckAddr("main", r.main.setup.offset(), 0x9000)
	CheckAddr("pmd_x1", r.pmd_x1.reset.offset(), 0x9010)
	CheckAddr("packet_generator", r.packet_generator.control[0].offset(), 0x9030)
	CheckAddr("rx_cl82_alignment_marker_timer", r.rx_cl82_alignment_marker_timer.offset(), 0x9123)
	CheckAddr("tx_x1", r.tx_x1.pma_fifo_watermark.offset(), 0x9200)
	CheckAddr("rx_x1", r.rx_x1.decode_control.offset(), 0x9221)
	CheckAddr("an_x1", r.an_x1.oui.offset(), 0x9240)
	CheckAddr("speed_change", r.speed_change.pll_lock_timer_period.offset(), 0x9260)
	CheckAddr("rx_x1a", r.rx_x1a.forward_error_correction_alignment_status.offset(), 0x92b0)
	CheckAddr("tx_x2", r.tx_x2.mld_swap_count.offset(), 0xa000)
	CheckAddr("tx_x2", r.tx_x2.cl82_control.offset(), 0xa002)
	CheckAddr("rx_x2", r.rx_x2.misc_control[0].offset(), 0xa023)
	CheckAddr("rx_cl82", r.rx_cl82.live_deskew_decoder_status.offset(), 0xa080)
	CheckAddr("pmd_x4", r.pmd_x4.lane_reset_control.offset(), 0xc010)
	CheckAddr("speed_change_x4", r.speed_change_x4.control.offset(), 0xc050)
	CheckAddr("tx_x4", r.tx_x4.mac_credit_clock_count[0].offset(), 0xc100)
	CheckAddr("rx_x4", r.rx_x4.pcs_control.offset(), 0xc130)
	CheckAddr("test1", r.test1.tx_packet_count[0].offset(), 0xc1b0)
	CheckAddr("an_x4", r.an_x4.cl73_auto_negotiation_control.offset(), 0xc1c0)
	CheckAddr("dsc", r.dsc_afe3.rx_peak_filter_control.offset(), 0xd000)
	CheckAddr("uc_cmd", r.uc_cmd.control.offset(), 0xd03d)
	CheckAddr("dsc_b", r.dsc_b.training_sum_interleave_abcd[0][0].offset(), 0xd040)
	CheckAddr("dsc_c", r.dsc_c.cdr_control[0].offset(), 0xd050)
	CheckAddr("dsc_d", r.dsc_d.state_machine.control[0].offset(), 0xd060)
	CheckAddr("dsc_e", r.dsc_e.rx_phase_slicer_counter.offset(), 0xd070)
	CheckAddr("cl93_rx", r.cl93n72_rx.control[0].offset(), 0xd080)
	CheckAddr("cl93_tx", r.cl93n72_tx.local_update_to_link_partner.offset(), 0xd090)
	CheckAddr("tx_phase_interpolator", r.tx_phase_interpolator.control.offset(), 0xd0a0)
	CheckAddr("clock_and_reset", r.clock_and_reset.over_sampling_mode_control.offset(), 0xd0b0)
	CheckAddr("ams_rx", r.ams_rx.control[0].offset(), 0xd0c0)
	CheckAddr("ams_tx", r.ams_tx.control[0].offset(), 0xd0d0)
	CheckAddr("sigdet", r.sigdet.control[0].offset(), 0xd0e0)
	CheckAddr("dig", r.dig.revision_id0.offset(), 0xd100)
	CheckAddr("ams_pll", r.ams_pll.control[0].offset(), 0xd110)
	CheckAddr("tx_pattern", r.tx_pattern.data[0].offset(), 0xd120)
	CheckAddr("tx_equalizer", r.tx_equalizer.control[0].offset(), 0xd130)
	CheckAddr("pll", r.pll.calibration_control[0].offset(), 0xd140)
	CheckAddr("tx_common_control", r.tx_common.control[0].offset(), 0xd150)
	CheckAddr("tlb_rx", r.tlb_rx.pseudo_random_bitstream_checker_count_control.offset(), 0xd160)
	CheckAddr("tlb_tx", r.tlb_tx.pattern_gen_config.offset(), 0xd170)
	CheckAddr("uc", r.uc.clock_control.offset(), 0xd200)
	CheckAddr("mdio", r.mdio.mask_data.offset(), 0xffdb)
}

// Check TSCE memory map.
func init() {
	r := (*tsce_regs)(m.BasePointer)
	CheckAddr("main", r.main.setup.offset(), 0x9000)
	CheckAddr("pmd_x1", r.pmd_x1.reset.offset(), 0x9010)
	CheckAddr("packet_generator", r.packet_generator.control[0].offset(), 0x9030)
	CheckAddr("mem_ecc", r.mem_ecc.twobit_ecc_error_interrupt_enable.offset(), 0x9050)
	CheckAddr("mem_debug", r.mem_debug.tm_deskew_memory.offset(), 0x9060)
	CheckAddr("rx_cl82_alignment_marker_timer", r.rx_cl82_alignment_marker_timer.offset(), 0x9123)
	CheckAddr("rx_x1", r.rx_x1.sync_state_machine.offset(), 0x9220)
	CheckAddr("an_x1", r.an_x1.oui.offset(), 0x9240)
	CheckAddr("speed_change", r.speed_change.pll_lock_timer_period.offset(), 0x9260)
	CheckAddr("tx_x2", r.tx_x2.mld_swap_count.offset(), 0xa000)
	CheckAddr("rx_x2", r.rx_x2.qreserved[0].offset(), 0xa020)
	CheckAddr("rx_cl82", r.rx_cl82.rx_decoder_status.offset(), 0xa080)
	CheckAddr("pmd_x4", r.pmd_x4.lane_reset_control.offset(), 0xc010)
	CheckAddr("speed_change_x4", r.speed_change_x4.control.offset(), 0xc050)
	CheckAddr("tx_x4", r.tx_x4.mac_credit_clock_count[0].offset(), 0xc100)
	CheckAddr("rx_x4", r.rx_x4.pcs_control.offset(), 0xc130)
	CheckAddr("an_x4", r.an_x4.enables.offset(), 0xc180)
	CheckAddr("cl72_link", r.cl72_link.control.offset(), 0xc253)
	CheckAddr("digital_control", r.digital_control.ctl_1000x.offset(), 0xc301)
	CheckAddr("interlaken_common", r.interlaken.control.offset(), 0xc330)
	CheckAddr("dsc_a", r.dsc_a.cdr_control[0].offset(), 0xd001)
	CheckAddr("uc_cmd", r.uc_cmd.control.offset(), 0xd00d)
	CheckAddr("dsc_b", r.dsc_b.state_machine.control[0].offset(), 0xd010)
	CheckAddr("dsc_c", r.dsc_c.dfe_common_control.offset(), 0xd020)
	CheckAddr("dsc_d", r.dsc_d.training_sum_control[0].offset(), 0xd030)
	CheckAddr("dsc_e", r.dsc_e.control.offset(), 0xd040)
	CheckAddr("cl72_rx", r.cl72_rx.receive_status.offset(), 0xd050)
	CheckAddr("cl72_tx", r.cl72_tx.coefficient_update.offset(), 0xd060)
	CheckAddr("tx_phase_interpolator", r.tx_phase_interpolator.control[0].offset(), 0xd070)
	CheckAddr("clock_and_reset", r.clock_and_reset.over_sampling_mode_control.offset(), 0xd080)
	CheckAddr("ams_rx", r.ams_rx.control[0].offset(), 0xd090)
	CheckAddr("ams_tx", r.ams_tx.control[0].offset(), 0xd0a0)
	CheckAddr("ams_com", r.ams_com.pll_control[0].offset(), 0xd0b0)
	CheckAddr("sigdet", r.sigdet.control[0].offset(), 0xd0c0)
	CheckAddr("tlb_rx", r.tlb_rx.pseudo_random_bitstream_checker_count_control.offset(), 0xd0d0)
	CheckAddr("tlb_tx", r.tlb_tx.pattern_gen_config.offset(), 0xd0e0)
	CheckAddr("dig", r.dig.revision_id0.offset(), 0xd0f0)
	CheckAddr("tx_pattern", r.tx_pattern.data[0].offset(), 0xd100)
	CheckAddr("tx_equalizer", r.tx_equalizer.control[0].offset(), 0xd110)
	CheckAddr("pll", r.pll.calibration_control[0].offset(), 0xd120)
	CheckAddr("tx_common_control", r.tx_common.cl72_tap_limit_control[0].offset(), 0xd130)
	CheckAddr("uc", r.uc.ram_word.offset(), 0xd200)
	CheckAddr("mdio", r.mdio.mask_data.offset(), 0xffdb)
}
