// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
)

func get_hundred_gig_controller() *hundred_gig_controller {
	return (*hundred_gig_controller)(m.BasePointer)
}

// Address map for 100g core
type hundred_gig_controller struct {
	_ [0x0002 - 0x0000]pad_u16

	phyid_common

	_ [0x000d - 0x0004]pad_u16

	acc_mdio_common

	_ [0x0096 - 0x000f]pad_u16

	cl93n72_common

	_ [0x9000 - 0x009c]pad_u16

	main struct {
		setup              pcs_u16
		synce_control      [2]pcs_u16
		rx_lane_swap       pcs_u16
		devices_in_package pcs_u16

		_ [0x9007 - 0x9005]pad_u16

		tick_generation_control [2]pcs_u16

		loopback_control pcs_u16

		mdio_broadcast      pcs_u16
		mdio_timeout        pcs_u16
		mdio_timeout_status pcs_u16

		_ [0x900e - 0x900d]pad_u16

		serdes_id pcs_u16
	}

	_ [0x9010 - 0x900f]pad_u16

	pmd_x1 struct {
		reset    pcs_u16
		mode     pcs_u16
		status   pcs_u16
		override pcs_u16
	}

	_ [0x9030 - 0x9014]pad_u16

	packet_generator struct {
		control [3]pcs_u16

		pseudo_random_test_pattern_control pcs_u16

		rx_crc_errors pcs_u16

		_ [0x9037 - 0x9035]pad_u16

		testpatt_seed [2][4]pcs_u16

		_ [0x9040 - 0x903f]pad_u16

		repeated_payload_bytes pcs_u16

		error_mask   [5]pcs_u16
		error_inject [2]pcs_u16
	}

	_ [0x9123 - 0x9048]pad_u16

	rx_cl82_am_common

	_ [0x9200 - 0x9133]pad_u16

	tx_x1 struct {
		pma_fifo_watermark        pcs_u16
		pma_delay_after_watermark pcs_u16
		cl91_fec_enable           pcs_u16
	}

	_ [0x9221 - 0x9203]pad_u16

	rx_x1 struct {
		decode_control pcs_u16
		deskew_windows pcs_u16
		cl91_config    pcs_u16

		cl91_symbol_error_threshold pcs_u16
		cl91_symbol_error_timer     pcs_u16

		_ [0x9230 - 0x9226]pad_u16

		forward_error_correction_mem_ecc_status [4]pcs_u16

		deskew_mem_ecc_status [4]pcs_u16

		interrupt_status [2]pcs_u16
		interrupt_enable [2]pcs_u16

		ecc_disable pcs_u16

		ecc_error_inject pcs_u16
	}

	_ [0x9240 - 0x923e]pad_u16

	an_x1 struct {
		oui pcs_u32

		priority_remap [5]pcs_u16

		_ [0x9250 - 0x9247]pad_u16

		cl73 struct {
			break_link_timer                    pcs_u16
			auto_negotiation_error_timer        pcs_u16
			parallel_detect_dme_lock_timer      pcs_u16
			parallel_detect_signal_detect_timer pcs_u16

			ignore_link_timer pcs_u16

			qualify_link_timer_yes_cl72_training pcs_u16
			qualify_link_timer_no_cl72_training  pcs_u16
			page_timers                          pcs_u16
		}
	}

	_ [0x9260 - 0x9258]pad_u16

	speed_change struct {
		pll_lock_timer_period       pcs_u16
		pmd_rx_lock_timer_period    pcs_u16
		pipeline_reset_timer_period pcs_u16
		tx_pipeline_reset_count     pcs_u16
		sc_status                   pcs_u16

		_ [0x9270 - 0x9265]pad_u16

		lanes [4]struct {
			speed [4]pcs_u16

			credit_clock_count          [2]pcs_u16
			credit_loop_count_01        pcs_u16
			credit_mac_generation_count pcs_u16

			_ [0x10 - 0x08]pad_u16
		}
	}

	rx_x1a struct {
		forward_error_correction_alignment_status pcs_u16

		cl91_status pcs_u16

		n_corrected_symbols   pcs_u32
		n_uncorrected_symbols pcs_u32
		n_corrected_bits      pcs_u32
	}

	_ [0xa000 - 0x92b8]pad_u16

	tx_x2 struct {
		mld_swap_count pcs_u16

		_ [0xa002 - 0xa001]pad_u16

		cl82_control pcs_u16

		_ [0xa011 - 0xa003]pad_u16

		cl82_status [2]pcs_u16
	}

	_ [0xa023 - 0xa013]pad_u16

	rx_x2 struct {
		misc_control [2]pcs_u16
	}

	_ [0xa080 - 0xa025]pad_u16

	rx_cl82 struct {
		live_deskew_decoder_status    pcs_u16
		latched_deskew_decoder_status pcs_u16

		ber_count [2]pcs_u16

		errored_block_count pcs_u16
	}

	_ [0xc010 - 0xa085]pad_u16

	pmd_x4 pmd_x4_common

	_ [0xc050 - 0xc01a]pad_u16

	speed_change_x4 speed_change_x4_common

	_ [0xc070 - 0xc062]pad_u16

	speed_change_x4_config struct {
		final_speed_config pcs_lane_u16

		_ [0xc072 - 0xc071]pad_u16

		final_speed_config1 [6]pcs_u16
		final_speed_fec     pcs_u16
	}

	_ [0xc100 - 0xc079]pad_u16

	tx_x4 struct {
		mac_credit_clock_count  [2]pcs_lane_u16
		mac_credit_loop_count01 pcs_lane_u16
		mac_credit_gen_count    pcs_lane_u16

		_ [0xc111 - 0xc104]pad_u16

		encode_control pcs_lane_u16

		_ [0xc113 - 0xc112]pad_u16

		control pcs_lane_u16

		cl36_tx_control pcs_lane_u16

		_ [0xc120 - 0xc115]pad_u16

		encode_status      [2]pcs_lane_u16
		pcs_status_live    pcs_lane_u16
		pcs_status_latched pcs_lane_u16

		pma_underflow_overflow_status pcs_lane_u16
	}

	_ [0xc130 - 0xc125]pad_u16

	rx_x4 struct {
		pcs_control pcs_lane_u16

		_ [0xc134 - 0xc131]pad_u16

		decoder_control   pcs_lane_u16
		block_sync_config pcs_lane_u16

		_ [0xc137 - 0xc136]pad_u16

		pma_control pcs_lane_u16

		_ [0xc139 - 0xc138]pad_u16

		link_status_control pcs_lane_u16

		deskew_memory_control           pcs_lane_u16
		fec_memory_control              pcs_lane_u16
		cl36_control                    pcs_lane_u16
		synce_fractional_divisor_config pcs_lane_u16

		_ [0xc140 - 0xc13e]pad_u16

		fec_control [4]pcs_lane_u16

		_ [0xc150 - 0xc144]pad_u16

		block_sync_status       pcs_lane_u16
		block_sync_sm_debug     [3]pcs_lane_u16
		block_lock_latch_status pcs_lane_u16
		am_lock_latch_status    pcs_lane_u16

		am_lock_live_status  pcs_lane_u16
		cl82_bip_error_count [3]pcs_lane_u16

		pseudo_logical_lane_to_virtual_lane_mapping [2]pcs_lane_u16

		pseudo_random_test_pattern_errors    pcs_lane_u16
		pseudo_random_test_pattern_is_locked pcs_lane_u16

		_ [0xc160 - 0xc15e]pad_u16

		pcs_latched_status pcs_lane_u16

		pcs_live_status pcs_lane_u16

		decoder_status [4]pcs_lane_u16

		cl91_per_lane_n_corrected_symbols [2]pcs_lane_u16

		cl36_sync_acquisition_next_state pcs_lane_u16
		cl36_sync_acquisition_state      pcs_lane_u16

		cl36_ber_count pcs_lane_u16

		_ [0xc170 - 0xc16b]pad_u16

		cl82_am_latched_status [5]pcs_lane_u16
		cl82_am_live_status    [5]pcs_lane_u16

		cl91_sync_status pcs_lane_u16

		cl91_sync_fsm_state pcs_lane_u16

		_ [0xc180 - 0xc17c]pad_u16

		fec_debug                 [2][5]pcs_lane_u16
		rx_fec_burst_error_lo     [5]pcs_lane_u16
		_                         [0xc190 - 0xc18f]pad_u16
		rx_fec_burst_error_hi     [5]pcs_lane_u16
		rx_fec_corrected_blocks   [2][5]pcs_lane_u16
		_                         [0xc1A0 - 0xc19f]pad_u16
		rx_fec_uncorrected_blocks [2][5]pcs_lane_u16
	}

	_ [0xc1b0 - 0xc1aa]pad_u16

	test1 struct {
		tx_packet_count [2]pcs_u16
		rx_packet_count [2]pcs_u16
	}

	_ [0xc1c0 - 0xc1b4]pad_u16

	an_x4 struct {
		cl73_auto_negotiation_control pcs_lane_u16

		cl73_auto_negotiation_local_up1_abilities [2]pcs_lane_u16

		cl73_auto_negotiation_local_base_abilities [2]pcs_lane_u16

		cl73_auto_negotiation_local_bam_abilities pcs_lane_u16

		cl73_auto_negotiation_misc_control pcs_lane_u16

		_ [0xc1d0 - 0xc1c7]pad_u16

		cl73_r_status    pcs_lane_u16
		cl73_pxng_status pcs_lane_u16

		cl73_pseq_status pcs_lane_u16

		cl73_pseq_remote_fault_status pcs_lane_u16
		cl73_unexpected_page          pcs_lane_u16
		cl73_pseq_base_pages          [3]pcs_lane_u16
		cl73_pseq_link_partner_oui    [5]pcs_lane_u16
		cl73_resolution_error         pcs_lane_u16

		_ [0xc1e0 - 0xc1de]pad_u16

		cl73_local_device_sw_control_pages       [3]pcs_lane_u16
		cl73_link_partner_sw_control_pages       [3]pcs_lane_u16
		cl73_sw_status                           pcs_lane_u16
		cl73_local_device_control                pcs_lane_u16
		cl73_auto_negotiation_ability_resolution pcs_lane_u16

		cl73_auto_negotiation_misc_status pcs_lane_u16

		cl73_tla_sequencer_status pcs_lane_u16
	}

	_ [0xc330 - 0xc1eb]pad_u16

	interlaken interlaken_common

	_ [0xd000 - 0xc341]pad_u16

	dsc_afe3 struct {
		rx_peak_filter_control pmd_lane_u16

		rx_slicer struct {
			a_offset_adjust_data  pmd_lane_u16
			a_offset_adjust_phase pmd_lane_u16
			ab_offset_adjust_lms  pmd_lane_u16
			b_offset_adjust_data  pmd_lane_u16
			b_offset_adjust_phase pmd_lane_u16
			c_offset_adjust_data  pmd_lane_u16
			c_offset_adjust_phase pmd_lane_u16
			cd_offset_adjust_lms  pmd_lane_u16
			d_offset_adjust_data  pmd_lane_u16
			d_offset_adjust_phase pmd_lane_u16
		}
		rx_phase_lms_threshold pmd_lane_u16

		_ [0xd010 - 0xd00c]pad_u16

		rx_dfe_tap2_abcd        [2]pmd_lane_u16
		rx_dfe_tap3_abcd        [2]pmd_lane_u16
		rx_dfe_tap4_9_abcd      [6]pmd_lane_u16
		_                       [0xd020 - 0xd01a]pad_u16
		rx_dfe_tap10_14_abcd    [5]pmd_lane_u16
		rx_dfe_tap7_14_mux_abcd [4]pmd_lane_u16
		load_presets            pmd_lane_u16
	}

	_ [0xd03d - 0xd02a]pad_u16

	uc_cmd uc_cmd_controller

	_ [0xd040 - 0xd03f]pad_u16

	dsc_b struct {
		training_sum_interleave_abcd [4][2]pmd_lane_u16

		training_sum_result_abcd [2]pmd_lane_u16

		_ [0xd04c - 0xd04a]pad_u16

		dc_offset pmd_lane_u16

		vga_status pmd_lane_u16
	}

	_ [0xd050 - 0xd04e]pad_u16

	dsc_c struct {
		cdr_control [3]pmd_lane_u16
		pi_control  pmd_lane_u16

		_ [0xd055 - 0xd054]pad_u16

		training_sum_control         pmd_lane_u16
		training_sum_pattern_control [2]pmd_lane_u16
		training_sum_tap_control     pmd_lane_u16
		training_sum_tdt_control     pmd_lane_u16
		training_sum_misc            pmd_lane_u16

		_ [0xd05c - 0xd05b]pad_u16

		vga_control            pmd_lane_u16
		data_slicer_th_control pmd_lane_u16
		dc_offset_control      pmd_lane_u16
	}

	_ [0xd060 - 0xd05f]pad_u16

	dsc_d struct {
		state_machine struct {
			control [10]pmd_lane_u16

			lock_status pmd_lane_u16

			status_one_hot pmd_lane_u16

			status_eee_one_hot pmd_lane_u16

			restart_status pmd_lane_u16

			status pmd_lane_u16
		}
	}

	_ [0xd070 - 0xd06f]pad_u16

	dsc_e struct {
		rx_phase_slicer_counter  pmd_lane_u16
		rx_lms_slicer_counter    pmd_lane_u16
		rx_data                  [2]pmd_lane_u16
		cdr_phase_error_status   pmd_lane_u16
		rx_data_slicer_counter   pmd_lane_u16
		rx_phase_slicer_counter1 pmd_lane_u16
		rx_lms_slicer_counter1   pmd_lane_u16
		cdr_integration          pmd_lane_u16
		cdr_misc_status          pmd_lane_u16
		cdr_1g_status            pmd_lane_u16
		_                        [0xd07e - 0xd07b]pad_u16
		preset                   pmd_lane_u16
	}

	_ [0xd080 - 0xd07f]pad_u16

	cl93n72_rx struct {
		control [3]pmd_lane_u16

		status pmd_lane_u16

		micro_interrupt_enable0 pmd_lane_u16

		micro_interrupt_status0 pmd_lane_u16

		micro_status1 pmd_lane_u16
	}

	_ [0xd090 - 0xd087]pad_u16

	cl93n72_tx struct {
		local_update_to_link_partner pmd_lane_u16
		local_status_to_link_partner pmd_lane_u16

		control [4]pmd_lane_u16

		status pmd_lane_u16
	}

	_ [0xd0a0 - 0xd097]pad_u16

	tx_phase_interpolator struct {
		control pmd_lane_u16

		frequecy_override pmd_lane_u16
		jitter_control    pmd_lane_u16
		control3          pmd_lane_u16
		control4          pmd_lane_u16

		_ [0xd0a8 - 0xd0a5]pad_u16

		status [4]pmd_lane_u16
	}

	_ [0xd0b0 - 0xd0ac]pad_u16

	clock_and_reset clock_and_reset_common

	_ [0xd0c0 - 0xd0bf]pad_u16

	ams_rx struct {
		control [10]pmd_lane_u16
		_       [0xd0cb - 0xd0ca]pad_u16
		status  pmd_lane_u16
	}

	_ [0xd0d0 - 0xd0cc]pad_u16

	ams_tx struct {
		control [4]pmd_lane_u16

		_ [0xd0d8 - 0xd0d4]pad_u16

		status pmd_lane_u16
	}

	_ [0xd0e0 - 0xd0d9]pad_u16

	sigdet sigdet_common

	_ [0xd100 - 0xd0e9]pad_u16

	dig dig_common

	_ [0xd110 - 0xd10f]pad_u16

	ams_pll struct {
		control [8]pmd_lane_u16
		_       [0xd119 - 0xd118]pad_u16
		status  pmd_lane_u16
	}

	_ [0xd120 - 0xd11a]pad_u16

	tx_pattern tx_pattern_common

	_ [0xd130 - 0xd12f]pad_u16

	tx_equalizer struct {
		control [3]pmd_lane_u16

		status [5]pmd_lane_u16

		tx_uc_control pmd_lane_u16

		misc_control pmd_lane_u16
	}

	_ [0xd140 - 0xd13a]pad_u16

	pll pll_common

	_ [0xd150 - 0xd14b]pad_u16

	tx_common struct {
		control [4]pmd_lane_u16
	}

	_ [0xd160 - 0xd154]pad_u16

	tlb_rx struct {
		tlb_rx_common
		pseudo_random_bitstream_burst_error_live_length pmd_lane_u16
		pseudo_random_bitstream_burst_error_max_length  pmd_lane_u16
	}

	_ [0xd170 - 0xd16f]pad_u16

	tlb_tx struct {
		tlb_tx_common

		_ [0xd178 - 0xd174]pad_u16

		remote_loopback_status pmd_lane_u16
	}

	_ [0xd200 - 0xd179]pad_u16

	uc struct {
		clock_control pmd_lane_u16

		reset_control pmd_lane_u16

		ahb_control pmd_lane_u16

		ahb_status pmd_lane_u16

		write_address                pmd_lane_u32
		write_data                   pmd_lane_u32
		read_address                 pmd_lane_u32
		read_data                    pmd_lane_u32
		program_ram_interface_enable pmd_lane_u16
		program_ram_write_address    pmd_lane_u32

		_ [0xd210 - 0xd20f]pad_u16

		temperature_status pmd_lane_u16

		tx_mailbox      pmd_lane_u32
		rx_mailbox      pmd_lane_u32
		mailbox_control pmd_lane_u16
		ahb_control1    pmd_lane_u16
		ahb_status1     pmd_lane_u16

		ahb_next_auto_increment_write_address             pmd_lane_u16
		ahb_next_auto_increment_read_address              pmd_lane_u16
		ahb_next_auto_increment_program_ram_write_address pmd_lane_u16
		temperature_control                               pmd_lane_u16

		_ [0xd220 - 0xd21c]pad_u16

		program_ram_ecc_control       [2]pmd_lane_u16
		program_ram_ecc_error_address pmd_lane_u16
		program_ram_ecc_error_data    pmd_lane_u16
		program_ram_test_control      pmd_lane_u16

		ram_config pmd_lane_u16

		interrupt_enable pmd_lane_u16
		interrupt_status pmd_lane_u16
	}

	_ [0xffdb - 0xd228]pad_u16

	mdio mdio_common

	_ [0xffff - 0xffe0]pad_u16
}

type tscf_over_sampling_divider int

const (
	tscf_over_sampling_divider_1      tscf_over_sampling_divider = 0
	tscf_over_sampling_divider_2      tscf_over_sampling_divider = 1
	tscf_over_sampling_divider_4      tscf_over_sampling_divider = 2
	tscf_over_sampling_divider_16_5   tscf_over_sampling_divider = 8
	tscf_over_sampling_divider_20_625 tscf_over_sampling_divider = 12
)

type tscf_pll_multipler int

const (
	tscf_pll_multipler_64 tscf_pll_multipler = iota
	tscf_pll_multipler_66
	tscf_pll_multipler_80
	tscf_pll_multipler_128
	tscf_pll_multipler_132
	tscf_pll_multipler_140
	tscf_pll_multipler_160
	tscf_pll_multipler_165
	tscf_pll_multipler_168
	tscf_pll_multipler_170
	tscf_pll_multipler_175
	tscf_pll_multipler_180
	tscf_pll_multipler_184
	tscf_pll_multipler_200
	tscf_pll_multipler_224
	tscf_pll_multipler_264
)

var tscf_pll_multiplers = [...]uint16{
	64, 66, 80, 128, 132, 140, 160, 165,
	168, 170, 175, 180, 184, 200, 224, 264,
}

type tscf_speed int

const (
	tscf_speed_cr     tscf_speed = 0 << 0
	tscf_speed_kr     tscf_speed = 1 << 0
	tscf_speed_optics tscf_speed = 2 << 0
	tscf_speed_hg2    tscf_speed = 1 << 2

	tscf_speed_10g_x1  tscf_speed = 0 << 3
	tscf_speed_20g_x1  tscf_speed = 1 << 3
	tscf_speed_25g_x1  tscf_speed = 2 << 3
	tscf_speed_20g_x2  tscf_speed = 3 << 3
	tscf_speed_40g_x2  tscf_speed = 4 << 3
	tscf_speed_40g_x4  tscf_speed = 5 << 3
	tscf_speed_50g_x2  tscf_speed = 6 << 3
	tscf_speed_50g_x4  tscf_speed = 7 << 3
	tscf_speed_100g_x4 tscf_speed = 8 << 3

	tscf_speed_cl73_20g_vco tscf_speed = 9 << 3
	tscf_speed_cl73_25g_vco tscf_speed = 10 << 3
	tscf_speed_cl36_20g_vco tscf_speed = 11 << 3
	tscf_speed_cl36_25g_vco tscf_speed = 12 << 3
)

func (x tscf_speed) String() (s string) {
	if x == 0 {
		return
	}
	switch x & (0xf << 3) {
	case tscf_speed_10g_x1:
		s = "10G X1"
	case tscf_speed_20g_x1:
		s = "20G X1"
	case tscf_speed_25g_x1:
		s = "25G X1"
	case tscf_speed_20g_x2:
		s = "20G X2"
	case tscf_speed_40g_x2:
		s = "40G X2"
	case tscf_speed_40g_x4:
		s = "40G X4"
	case tscf_speed_50g_x2:
		s = "50G X2"
	case tscf_speed_50g_x4:
		s = "50G X4"
	case tscf_speed_100g_x4:
		s = "100G X4"
	case tscf_speed_cl73_20g_vco:
		s = "1G CL73 20g"
	case tscf_speed_cl73_25g_vco:
		s = "1G CL73"
	case tscf_speed_cl36_20g_vco:
		s = "1G CL36 20g"
	case tscf_speed_cl36_25g_vco:
		s = "1G CL36"
	default:
		s = "invalid "
	}
	switch x & 3 {
	case tscf_speed_cr:
		s += " CR"
	case tscf_speed_kr:
		s += " KR"
	case tscf_speed_optics:
		s += " OPTICS"
	case 3 << 0:
		s += " INVALID"
	}
	if x&tscf_speed_hg2 != 0 {
		s += " HG2"
	}
	return
}
