// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsc

type phyid_common struct {
	id [2]pcs_reg
}

type acc_mdio_common struct {
	mdio_control pmd_lane_reg

	mdio_address_or_data pmd_lane_reg
}

type cl93n72_common struct {
	control pmd_lane_reg

	status pmd_lane_reg

	last_coefficient_update_received_from_link_partner pmd_lane_reg

	last_coefficient_status_received_from_link_partner pmd_lane_reg

	last_coefficient_update_sent_to_link_partner pmd_lane_reg
	last_coefficient_status_sent_to_link_partner pmd_lane_reg
}

type rx_cl82_am_common struct {
	rx_cl82_alignment_marker_timer pcs_reg

	_ [0x9130 - 0x9124]pad_reg

	rx_cl82_per_lane_alignment_marker_bytes [3]pcs_reg
}

type pmd_x4_common struct {
	lane_reset_control pcs_lane_reg

	lane_mode_config pcs_lane_reg

	lane_status pcs_lane_reg

	latched_lane_status pcs_lane_reg

	lane_override pcs_lane_reg

	_ [0xc018 - 0xc015]pad_reg

	eee_control pcs_lane_reg
	eee_status  pcs_lane_reg
}

type speed_change_x4_common struct {
	control pcs_lane_reg

	status_read_to_clear pcs_lane_reg

	error pcs_lane_reg

	_ [0xc054 - 0xc053]pad_reg

	debug pcs_lane_reg

	n_lanes_override pcs_lane_reg

	_ [0xc058 - 0xc056]pad_reg

	bypass pcs_lane_reg

	_ [0xc060 - 0xc059]pad_reg

	enable_override [2]pcs_lane_reg
}

type interlaken_common struct {
	control pcs_reg

	_ [0xc340 - 0xc331]pad_reg

	status pcs_reg
}

type clock_and_reset_common struct {
	over_sampling_mode_control pmd_lane_reg

	lane_reset_and_powerdown pmd_lane_reg

	lane_afe_reset_and_powerdown pmd_lane_reg

	lane_reset_and_powerdown_pin_disable pmd_lane_reg

	lane_debug_reset_control pmd_lane_reg

	uc_ack_lane_control pmd_lane_reg

	lane_reset_occurred pmd_lane_reg

	clock_reset_debug_control pmd_lane_reg

	pmd_lane_mode_status pmd_lane_reg

	lane_data_path_reset_status pmd_lane_reg

	lane_is_masked_from_multicast_writes pmd_lane_reg

	over_sampling_status          pmd_lane_reg
	over_sampling_status_from_pin pmd_lane_reg
	_                             [0xd0be - 0xd0bd]pad_reg
	lane_reset_active_low         pmd_lane_reg
}

type sigdet_common struct {
	control [3]pmd_lane_reg

	_ [0x8 - 0x3]pad_reg

	status pmd_lane_reg
}

type dig_common struct {
	revision_id0 pmd_lane_reg

	reset_control_pmd_active_low pmd_lane_reg

	reset_control_datapath pmd_lane_reg

	lane_is_masked_from_multicast_writes pmd_lane_reg

	top_user_control pmd_lane_reg

	uc_ack_core_control pmd_lane_reg

	core_reset_occurred pmd_lane_reg

	reset_seq_timer pmd_lane_reg

	core_datapath_reset_status pmd_lane_reg

	pmd_core_mode pmd_lane_reg

	revision_id1 pmd_lane_reg

	tx_lane_map_012 pmd_lane_reg

	tx_lane_map_3_lane_address_01 pmd_lane_reg

	tx_lane_address_23 pmd_lane_reg

	revision_id2 pmd_lane_reg
}

type tx_pattern_common struct {
	data [15]pmd_lane_reg
}

type pll_common struct {
	calibration_control [6]pmd_lane_reg

	reference_clock pmd_lane_reg

	multiplier pmd_lane_reg

	status [2]pmd_lane_reg

	debug_status pmd_lane_reg
}

type tlb_rx_common struct {
	pseudo_random_bitstream_checker_count_control pmd_lane_reg
	pseudo_random_bitstream_checker_control       pmd_lane_reg

	tx_to_rx_loopback pmd_lane_reg

	misc_control pmd_lane_reg

	pseudo_random_bitstream_checker_enable_timer_control pmd_lane_reg
	_                                                    [0xd168 - 0xd165]pad_reg
	tx_to_rx_loopback_status                             pmd_lane_reg
	pseudo_random_bitstream_checker_lock_status          pmd_lane_reg
	pseudo_random_bitstream_error_counts                 [2]pmd_lane_reg

	pmd_lock_status pmd_lane_reg
}

type tlb_tx_common struct {
	pattern_gen_config                        pmd_lane_reg
	pseudo_random_bitstream_generator_control pmd_lane_reg

	remote_loopback pmd_lane_reg

	misc_control pmd_lane_reg
}

type mdio_common struct {
	mask_data         pmd_lane_reg
	broadcast_address pmd_lane_reg

	mmd_select pmd_lane_reg

	aer pmd_lane_reg

	block_address pmd_lane_reg
}

type uc_cmd_regs struct {
	control pmd_lane_reg

	data pmd_lane_reg
}
