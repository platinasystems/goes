// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
)

func get_tsce_regs() *tsce_regs { return (*tsce_regs)(m.RegsBasePointer) }

// Register Map for TSCE core
type tsce_regs struct {
	_ [0x0002 - 0x0000]pad_reg

	phyid_common

	_ [0x000d - 0x0004]pad_reg

	acc_mdio_common

	_ [0x0096 - 0x000f]pad_reg

	cl93n72_common

	_ [0x9000 - 0x009c]pad_reg

	main struct {
		setup pcs_reg

		_ [0x9002 - 0x9001]pad_reg

		synce_control pcs_reg

		rx_lane_swap pcs_reg

		devices_in_package pcs_reg

		misc pcs_reg

		_ [0x9007 - 0x9006]pad_reg

		tick_generation_control [2]pcs_reg

		loopback_control pcs_reg

		mdio_broadcast pcs_reg

		_ [0x900e - 0x900b]pad_reg

		serdes_id pcs_reg
	}

	_ [0x9010 - 0x900f]pad_reg

	pmd_x1 struct {
		reset pcs_reg

		mode pcs_reg

		status pcs_reg

		latched_status pcs_reg

		override pcs_reg
	}

	_ [0x9030 - 0x9015]pad_reg

	packet_generator struct {
		control [2]pcs_reg

		pseudo_random_test_pattern_control pcs_reg

		rx_crc_errors       pcs_reg
		rx_prtp_errors      pcs_reg
		rx_prtp_lock_status pcs_reg

		_ [0x9037 - 0x9036]pcs_reg

		testpatt_seed [2][4]pcs_reg

		_ [0x9040 - 0x903f]pcs_reg

		repeated_payload_bytes pcs_reg

		error_mask   [5]pcs_reg
		error_inject [2]pcs_reg
	}

	_ [0x9050 - 0x9048]pad_reg

	mem_ecc struct {
		twobit_ecc_error_interrupt_enable pcs_reg
		disable_ecc_check_generate        pcs_reg
		inject_ecc_errors                 pcs_reg
		ecc_error_status                  [3]pcs_reg
	}

	_ [0x9060 - 0x9056]pad_reg

	mem_debug struct {
		tm_deskew_memory pcs_reg
		tm_rfec0_memory  pcs_reg
		tm_rfec1_memory  pcs_reg
	}

	_ [0x90b1 - 0x9063]pad_reg

	misc struct {
		xgxs_status            pcs_reg
		_                      [0x90b3 - 0x90b2]pad_reg
		tx_rate_mismatch       pcs_reg
		scrambler_enable_6Gbps pcs_reg
		cl72_enable_control    pcs_reg
	}

	_ [0x9123 - 0x90b6]pad_reg

	rx_cl82_am_common

	_ [0x9140 - 0x9133]pad_reg

	rx_cl82_am_tsc struct {
		rx_cl82_tsc_per_lane_alignment_marker_bytes [3]pcs_reg
	}

	_ [0x9220 - 0x9143]pad_reg

	rx_x1 struct {
		sync_state_machine pcs_reg
		decode_control1    pcs_reg
		deskew_windows     pcs_reg

		cl49_sync_header_counters [3]pcs_reg

		sync_code_word [5]pcs_reg

		sync_code_word_mask [5]pcs_reg
	}

	_ [0x9240 - 0x9230]pad_reg

	an_x1 struct {
		oui pcs_reg_32

		priority_remap [7]pcs_reg

		_ [0x9250 - 0x9249]pad_reg

		cl37 struct {
			restart_timer       pcs_reg
			complete_ack_timer  pcs_reg
			timeout_error_timer pcs_reg
		}

		cl73 struct {
			break_link_timer               pcs_reg
			timeout_error_timer            pcs_reg
			parallel_detect_dme_lock_timer pcs_reg
			link_up_timer                  pcs_reg

			qualify_link_status_timer [2]pcs_reg

			parallel_detect_signal_detect_timer pcs_reg
			ignore_cl37_sync_status_down_timer  pcs_reg
			period_to_wait_for_link_before_cl37 pcs_reg
			ignore_link_timer                   pcs_reg
			dme_page_timers                     pcs_reg
			sgmii_timer                         pcs_reg
		}
	}

	_ [0x9260 - 0x925f]pad_reg

	speed_change struct {
		pll_lock_timer_period       pcs_reg
		pmd_rx_lock_timer_period    pcs_reg
		pipeline_reset_timer_period pcs_reg
		tx_pipeline_reset_count     pcs_reg
		sc_status                   pcs_reg

		_ [0x9270 - 0x9265]pad_reg

		lanes [4]struct {
			main_override pcs_reg
			_             [0x02 - 0x01]pad_reg

			speed [9]pcs_reg
			_     [0x10 - 0xb]pad_reg
		}
	}

	_ [0xa000 - 0x92b0]pad_reg

	tx_x2 struct {
		mld_swap_count pcs_lane_reg

		cl48_control pcs_lane_reg

		cl82_control      pcs_reg
		acol_insert_count pcs_reg
		_                 [0xa011 - 0xa004]pad_reg
		cl82_status       pcs_reg
	}

	_ [0xa020 - 0xa012]pad_reg

	rx_x2 struct {
		qreserved   [3]pcs_reg
		rx_x2_misc  [2]pcs_reg
		_           [0xa031 - 0xa025]pad_reg
		skew_status [2]pcs_reg
	}

	_ [0xa080 - 0xa033]pad_reg

	rx_cl82 struct {
		rx_decoder_status pcs_reg

		rx_deskew pcs_reg

		_ [0xa085 - 0xa082]pad_reg

		ber_high_order                  pcs_reg
		error_blocks_high_order_40_100G pcs_reg
	}

	_ [0xc010 - 0xa087]pad_reg

	pmd_x4 pmd_x4_common

	_ [0xc040 - 0xc01a]pad_reg

	patgen1 struct {
		tx_packet_count [2]pcs_reg
		rx_packet_count [2]pcs_reg
	}

	_ [0xc050 - 0xc044]pad_reg

	speed_change_x4 speed_change_x4_common

	_ [0xc070 - 0xc062]pad_reg

	speed_change_x4_config struct {
		final_speed_config pcs_lane_reg

		_ [0xc072 - 0xc071]pad_reg

		final_speed_config1 [9]pcs_lane_reg
	}

	_ [0xc100 - 0xc07b]pad_reg

	tx_x4 struct {
		mac_credit_clock_count  [2]pcs_lane_reg
		mac_credit_loop_count01 pcs_lane_reg
		mac_credit_gen_count    pcs_lane_reg
		pcs_clock_count         pcs_lane_reg
		pcs_credit_gen_count    pcs_lane_reg

		_ [0xc111 - 0xc106]pad_reg

		encode_control [2]pcs_lane_reg

		misc pcs_lane_reg

		_ [0xc120 - 0xc114]pad_reg

		encode_status pcs_lane_reg
		pcs_status    pcs_lane_reg
	}

	_ [0xc130 - 0xc122]pad_reg

	rx_x4 struct {
		pcs_control pcs_lane_reg

		fec_control     [3]pcs_lane_reg
		decoder_control [2]pcs_lane_reg
		cl36_rx0        pcs_lane_reg
		pma_control0    pcs_lane_reg

		_ [0xc139 - 0xc138]pad_reg

		link_status_control pcs_lane_reg

		_ [0xc140 - 0xc13a]pad_reg

		user_fec_debug_read_data [2]pcs_lane_reg
		fec_burst_error_status   [2]pcs_lane_reg
		barrel_shifter_state     pcs_lane_reg
		cl49_lock_fsm_state      pcs_lane_reg
		decode_status            [6]pcs_lane_reg
		syncacq_status           [2]pcs_lane_reg
		bercnt                   pcs_lane_reg

		_ [0xc152 - 0xc14f]pad_reg

		latched_pcs_status [2]pcs_lane_reg

		pcs_live_status pcs_lane_reg

		cl82_am_lock_sm_latched_status pcs_lane_reg
		cl82_am_lock_sm_live_status    pcs_lane_reg
		fec_corrected_blocks_counter   [2]pcs_lane_reg
		fec_uncorrected_blocks_counter [2]pcs_lane_reg
		rx_gbox_error_status           pcs_lane_reg

		_ [0xc161 - 0xc15c]pad_reg

		t12_fec_control [3]pcs_lane_reg

		_ [0xc170 - 0xc164]pad_reg

		t12_fec_debug                  [2]pcs_lane_reg
		t12_fec_burst_error_status     [2]pcs_lane_reg
		t12_bercnt                     pcs_lane_reg
		t12_cl49_lock_status           pcs_lane_reg
		t12_virtual_lane_mapping       pcs_lane_reg
		t12_barrel_shifter_state       pcs_lane_reg
		t12_corrected_blocks_counter   [2]pcs_lane_reg
		t12_uncorrected_blocks_counter [2]pcs_lane_reg
		t12_cl82_latched_status        pcs_lane_reg
		t12_cl82_live_status           pcs_lane_reg
	}

	_ [0xc180 - 0xc17e]pad_reg

	an_x4 struct {
		enables pcs_lane_reg

		cl37_base_page_abilities pcs_lane_reg
		cl37_bam_abilities       pcs_lane_reg
		cl37_over_1g_abilities   [2]pcs_lane_reg

		cl73_base_page_abilities [2]pcs_lane_reg

		cl73_bam_abilities pcs_lane_reg

		misc_controls pcs_lane_reg

		_ [0xc190 - 0xc189]pad_reg

		link_partner_message_page_5    [4]pcs_lane_reg
		link_partner_message_page_1024 [4]pcs_lane_reg
		link_partner_base_page         [3]pcs_lane_reg

		_ [0xc1a0 - 0xc19b]pad_reg

		local_device_sw_pages [3]pcs_lane_reg
		link_partner_sw_pages [3]pcs_lane_reg

		sw_control_status pcs_lane_reg

		local_device_controls pcs_lane_reg

		page_sequencer_status pcs_lane_reg

		page_exchanger_status pcs_lane_reg

		page_decoder_status pcs_lane_reg

		ability_resolution pcs_lane_reg

		misc_status pcs_lane_reg

		tla_sequencer_status pcs_lane_reg

		sequencer_unexpected_page pcs_lane_reg
	}

	_ [0xc253 - 0xc1af]pad_reg

	cl72_link struct {
		control pcs_reg
	}

	_ [0xc301 - 0xc254]pad_reg

	digital_control struct {
		ctl_1000x pcs_reg

		_ [0xc30a - 0xc302]pad_reg

		spare [2]pcs_reg
	}

	_ [0xc330 - 0xc30c]pad_reg

	interlaken interlaken_common

	_ [0xd001 - 0xc341]pad_reg

	dsc_a struct {
		cdr_control [3]pmd_lane_reg

		rx_pi_control          pmd_lane_reg
		cdr_integration_status pmd_lane_reg
		cdr_phase_error_status pmd_lane_reg
		rx_pi_d_counter        pmd_lane_reg
		rx_pi_p_counter        pmd_lane_reg
		rx_pi_m_counter        pmd_lane_reg
		rx_pi_differential     pmd_lane_reg
		training_sum           pmd_lane_reg

		_ [0xd00d - 0xd00c]pad_reg
	}

	uc_cmd uc_cmd_regs

	_ [0xd010 - 0xd00f]pad_reg

	dsc_b struct {
		state_machine struct {
			control            [10]pmd_lane_reg
			lock_status        pmd_lane_reg
			status_one_hot     pmd_lane_reg
			status_eee_one_hot pmd_lane_reg
			restart_status     pmd_lane_reg
			dsc_c_sm_status    pmd_lane_reg
		}
	}

	_ [0xd020 - 0xd01f]pad_reg

	dsc_c struct {
		dfe_common_control   pmd_lane_reg
		dfe_control          [5][2]pmd_lane_reg
		dfe_override         pmd_lane_reg
		vga_control          pmd_lane_reg
		vga_pat_eyediag      pmd_lane_reg
		p1_fractional_offset pmd_lane_reg
	}

	_ [0xd030 - 0xd02f]pad_reg

	dsc_d struct {
		training_sum_control [4]pmd_lane_reg
		training_sum_status  [6]pmd_lane_reg
		vga_status           pmd_lane_reg

		dfe_status     [3]pmd_lane_reg
		vga_tap_values pmd_lane_reg
	}

	_ [0xd040 - 0xd03f]pad_reg

	dsc_e struct {
		control pmd_lane_reg

		peak_filter_control [2]pmd_lane_reg

		adj_data [3][2]pmd_lane_reg

		dc_offset pmd_lane_reg
	}

	_ [0xd050 - 0xd04a]pad_reg

	cl72_rx struct {
		receive_status pmd_lane_reg
		misc_control   pmd_lane_reg
		debug2         pmd_lane_reg
		lp_control     pmd_lane_reg
		status1        pmd_lane_reg
	}

	_ [0xd060 - 0xd055]pad_reg

	cl72_tx struct {
		coefficient_update           pmd_lane_reg
		misc2_control                pmd_lane_reg
		debug3                       pmd_lane_reg
		pcs_control                  pmd_lane_reg
		local_device_status          pmd_lane_reg
		ready_for_command            pmd_lane_reg
		kr_default                   [2]pmd_lane_reg
		misc_coefficient_control     pmd_lane_reg
		local_device_status_override pmd_lane_reg
		debug_status                 pmd_lane_reg
	}

	_ [0xd070 - 0xd06b]pad_reg

	tx_phase_interpolator struct {
		control   [5]pmd_lane_reg
		_         [0xd076 - 0xd075]pad_reg
		control_6 pmd_lane_reg
		_         [0xd078 - 0xd077]pad_reg
		status    [4]pmd_lane_reg
	}

	_ [0xd080 - 0xd07c]pad_reg

	clock_and_reset clock_and_reset_common

	_ [0xd090 - 0xd08f]pad_reg

	ams_rx struct {
		control          [5]pmd_lane_reg
		_                [0xd098 - 0xd095]pad_reg
		internal_control pmd_lane_reg
		status           pmd_lane_reg
	}

	_ [0xd0a0 - 0xd09a]pad_reg

	ams_tx struct {
		control          [3]pmd_lane_reg
		_                [0xd0d8 - 0xd0d3]pad_reg
		internal_control pmd_lane_reg
		status           pmd_lane_reg
	}

	_ [0xd0b0 - 0xd0aa]pad_reg

	ams_com struct {
		pll_control [10]pmd_lane_reg
		status      pmd_lane_reg
	}

	_ [0xd0c0 - 0xd0bb]pad_reg

	sigdet sigdet_common

	_ [0xd0d0 - 0xd0c9]pad_reg

	tlb_rx tlb_rx_common

	_ [0xd0e0 - 0xd0dd]pad_reg

	tlb_tx struct {
		tlb_tx_common

		tx_pi_loop_timing_config pmd_lane_reg

		_ [0xd0e8 - 0xd0e5]pad_reg

		remote_loopback_pd_status pmd_lane_reg
	}

	_ [0xd0f0 - 0xd0e9]pad_reg

	dig dig_common

	_ [0xd100 - 0xd0ff]pad_reg

	tx_pattern tx_pattern_common

	_ [0xd110 - 0xd10f]pad_reg

	tx_equalizer struct {
		control [3]pmd_lane_reg

		status [4]pmd_lane_reg

		tx_uc_control pmd_lane_reg

		misc_control pmd_lane_reg

		control4 pmd_lane_reg
	}

	_ [0xd120 - 0xd11a]pad_reg

	pll pll_common

	_ [0xd130 - 0xd12b]pad_reg

	tx_common struct {
		cl72_tap_limit_control   [2]pmd_lane_reg
		cl72_tap_present_control pmd_lane_reg
		cl72_debug1              pmd_lane_reg
		cl72_max_wait_timer      pmd_lane_reg
		cl72_wait_timer          pmd_lane_reg
	}

	_ [0xd200 - 0xd136]pad_reg

	uc struct {
		ram_word pmd_lane_reg
		address  pmd_lane_reg

		command1 pmd_lane_reg

		write_data pmd_lane_reg
		read_data  pmd_lane_reg

		mdio_8051_fsm_status pmd_lane_reg

		status1                         pmd_lane_reg
		external_station_to_uc_mbox     [2]pmd_lane_reg
		uc_to_external_station_mbox_lsw pmd_lane_reg
		command2                        pmd_lane_reg
		uc_to_external_station_mbox_msw pmd_lane_reg

		command3 pmd_lane_reg

		command4 pmd_lane_reg

		temperature_status pmd_lane_reg

		_ [0xd210 - 0xd20f]pad_reg

		program_ram_control1 pmd_lane_reg

		_ [0xd214 - 0xd211]pad_reg

		data_ram_control1 pmd_lane_reg

		_ [0xd218 - 0xd215]pad_reg

		internal_ram_control pmd_lane_reg
	}

	_ [0xffdb - 0xd219]pad_reg

	mdio mdio_common

	_ [0xffff - 0xffe0]pad_reg
}

type tsce_over_sampling_divider int

const (
	tsce_over_sampling_divider_1 tsce_over_sampling_divider = iota
	tsce_over_sampling_divider_2
	tsce_over_sampling_divider_3
	tsce_over_sampling_divider_3_3
	tsce_over_sampling_divider_4
	tsce_over_sampling_divider_5
	tsce_over_sampling_divider_7_25
	tsce_over_sampling_divider_8
	tsce_over_sampling_divider_8_25
	tsce_over_sampling_divider_10
)

var tsce_over_sampling_dividers = [...]float64{
	tsce_over_sampling_divider_1:    1,
	tsce_over_sampling_divider_2:    2,
	tsce_over_sampling_divider_3:    3,
	tsce_over_sampling_divider_3_3:  3.3,
	tsce_over_sampling_divider_4:    4,
	tsce_over_sampling_divider_5:    5,
	tsce_over_sampling_divider_7_25: 7.25,
	tsce_over_sampling_divider_8:    8,
	tsce_over_sampling_divider_8_25: 8.25,
	tsce_over_sampling_divider_10:   10,
}

type tsce_pll_multiplier int

const (
	tsce_pll_multiplier_46 tsce_pll_multiplier = iota
	tsce_pll_multiplier_72
	tsce_pll_multiplier_40
	tsce_pll_multiplier_42
	tsce_pll_multiplier_48
	tsce_pll_multiplier_50
	tsce_pll_multiplier_52
	tsce_pll_multiplier_54
	tsce_pll_multiplier_60
	tsce_pll_multiplier_64
	tsce_pll_multiplier_66
	tsce_pll_multiplier_68
	tsce_pll_multiplier_70
	tsce_pll_multiplier_80
	tsce_pll_multiplier_92
	tsce_pll_multiplier_100
)

var tsce_pll_multipliers = [...]float64{
	tsce_pll_multiplier_46:  46,
	tsce_pll_multiplier_72:  72,
	tsce_pll_multiplier_40:  40,
	tsce_pll_multiplier_42:  42,
	tsce_pll_multiplier_48:  48,
	tsce_pll_multiplier_50:  50,
	tsce_pll_multiplier_52:  52,
	tsce_pll_multiplier_54:  54,
	tsce_pll_multiplier_60:  60,
	tsce_pll_multiplier_64:  64,
	tsce_pll_multiplier_66:  66,
	tsce_pll_multiplier_68:  68,
	tsce_pll_multiplier_70:  70,
	tsce_pll_multiplier_80:  80,
	tsce_pll_multiplier_92:  92,
	tsce_pll_multiplier_100: 100,
}

type tsce_speed int

const (
	tsce_speed_invalid         tsce_speed = 0
	tsce_speed_10m             tsce_speed = 1
	tsce_speed_100m            tsce_speed = 2
	tsce_speed_1000m           tsce_speed = 3
	tsce_speed_1g_cx1          tsce_speed = 4
	tsce_speed_1g_kx1          tsce_speed = 5
	tsce_speed_2p5g_x1         tsce_speed = 6
	tsce_speed_5g_x1           tsce_speed = 7
	tsce_speed_10g_cx4         tsce_speed = 8
	tsce_speed_10g_kx4         tsce_speed = 9
	tsce_speed_10g_x4          tsce_speed = 10
	tsce_speed_13g_x4          tsce_speed = 11
	tsce_speed_15g_x4          tsce_speed = 12
	tsce_speed_16g_x4          tsce_speed = 13
	tsce_speed_20g_cx4         tsce_speed = 14
	tsce_speed_10g_cx2         tsce_speed = 15
	tsce_speed_10g_x2          tsce_speed = 16
	tsce_speed_20g_x4          tsce_speed = 17
	tsce_speed_10p5g_x2        tsce_speed = 18
	tsce_speed_21g_x4          tsce_speed = 19
	tsce_speed_12p7g_x2        tsce_speed = 20
	tsce_speed_25p45g_x4       tsce_speed = 21
	tsce_speed_15p75g_x2       tsce_speed = 22
	tsce_speed_31p5g_x4        tsce_speed = 23
	tsce_speed_31p5g_kr4       tsce_speed = 24
	tsce_speed_20g_cx2         tsce_speed = 25
	tsce_speed_20g_x2          tsce_speed = 26
	tsce_speed_40g_x4          tsce_speed = 27
	tsce_speed_10g_kr1         tsce_speed = 28
	tsce_speed_10p6_x1         tsce_speed = 29
	tsce_speed_20g_kr2         tsce_speed = 30
	tsce_speed_20g_cr2         tsce_speed = 31
	tsce_speed_21g_x2          tsce_speed = 32
	tsce_speed_40g_kr4         tsce_speed = 33
	tsce_speed_40g_cr4         tsce_speed = 34
	tsce_speed_42g_x4          tsce_speed = 35
	tsce_speed_100g_cr10       tsce_speed = 36
	tsce_speed_107g_x10        tsce_speed = 37
	tsce_speed_120g_x12        tsce_speed = 38
	tsce_speed_127g_x12        tsce_speed = 39
	tsce_speed_5g_kr1          tsce_speed = 49
	tsce_speed_10p5g_x4        tsce_speed = 50
	tsce_speed_10m_10p3125     tsce_speed = 53
	tsce_speed_100m_10p3125    tsce_speed = 54
	tsce_speed_1000m_10p3125   tsce_speed = 55
	tsce_speed_2p5g_x1_10p3125 tsce_speed = 56
)

var tsce_speed_strings = [...]string{
	tsce_speed_invalid:         "invalid",
	tsce_speed_10m:             "10M",
	tsce_speed_100m:            "100M",
	tsce_speed_1000m:           "1000M",
	tsce_speed_1g_cx1:          "1G CX1",
	tsce_speed_1g_kx1:          "1G KX1",
	tsce_speed_2p5g_x1:         "2.5G X1",
	tsce_speed_5g_x1:           "5G X1",
	tsce_speed_10g_cx4:         "10G CX4",
	tsce_speed_10g_kx4:         "10G KX4",
	tsce_speed_10g_x4:          "10G X4",
	tsce_speed_13g_x4:          "13G X4",
	tsce_speed_15g_x4:          "15G X4",
	tsce_speed_16g_x4:          "16G X4",
	tsce_speed_20g_cx4:         "20G CX4",
	tsce_speed_10g_cx2:         "10G CX2",
	tsce_speed_10g_x2:          "10G X2",
	tsce_speed_20g_x4:          "20G X4",
	tsce_speed_10p5g_x2:        "10.5G X2",
	tsce_speed_21g_x4:          "21G X4",
	tsce_speed_12p7g_x2:        "12.7G X2",
	tsce_speed_25p45g_x4:       "25.45G X4",
	tsce_speed_15p75g_x2:       "15.75G X2",
	tsce_speed_31p5g_x4:        "31.5G X4",
	tsce_speed_31p5g_kr4:       "31.5G KR4",
	tsce_speed_20g_cx2:         "20G CX2",
	tsce_speed_20g_x2:          "20G X2",
	tsce_speed_40g_x4:          "40G X4",
	tsce_speed_10g_kr1:         "10G KR1",
	tsce_speed_10p6_x1:         "10.6G X1",
	tsce_speed_20g_kr2:         "20G KR2",
	tsce_speed_20g_cr2:         "20G CR2",
	tsce_speed_21g_x2:          "21G X2",
	tsce_speed_40g_kr4:         "40G KR4",
	tsce_speed_40g_cr4:         "40G CR4",
	tsce_speed_42g_x4:          "42G X4",
	tsce_speed_100g_cr10:       "100G CR10",
	tsce_speed_107g_x10:        "107G X10",
	tsce_speed_120g_x12:        "120G X12",
	tsce_speed_127g_x12:        "127G X12",
	tsce_speed_5g_kr1:          "5G KR1",
	tsce_speed_10p5g_x4:        "10.5G X4",
	tsce_speed_10m_10p3125:     "10m 10.3125",
	tsce_speed_100m_10p3125:    "100m 10.3125",
	tsce_speed_1000m_10p3125:   "1000m 10.3125",
	tsce_speed_2p5g_x1_10p3125: "2.5g X1 10.3125",
}

func (x tsce_speed) String() string { return elib.Stringer(tsce_speed_strings[:], int(x)) }
