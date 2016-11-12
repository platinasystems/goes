// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

type phyid_common struct {
	id [2]pcs_u16
}

type acc_mdio_common struct {
	mdio_control pmd_lane_u16

	mdio_address_or_data pmd_lane_u16
}

type cl93n72_common struct {
	control pmd_lane_u16

	status pmd_lane_u16

	last_coefficient_update_received_from_link_partner pmd_lane_u16

	last_coefficient_status_received_from_link_partner pmd_lane_u16

	last_coefficient_update_sent_to_link_partner pmd_lane_u16
	last_coefficient_status_sent_to_link_partner pmd_lane_u16
}

type rx_cl82_am_common struct {
	rx_cl82_alignment_marker_timer pcs_u16

	_ [0x9130 - 0x9124]pad_u16

	rx_cl82_per_lane_alignment_marker_bytes [3]pcs_u16
}

type pmd_x4_common struct {
	lane_reset_control pcs_lane_u16

	lane_mode_config pcs_lane_u16

	lane_status pcs_lane_u16

	latched_lane_status pcs_lane_u16

	lane_override pcs_lane_u16

	_ [0xc018 - 0xc015]pad_u16

	eee_control pcs_lane_u16
	eee_status  pcs_lane_u16
}

type speed_change_x4_common struct {
	control pcs_lane_u16

	status_read_to_clear pcs_lane_u16

	error pcs_lane_u16

	_ [0xc054 - 0xc053]pad_u16

	debug pcs_lane_u16

	n_lanes_override pcs_lane_u16

	_ [0xc058 - 0xc056]pad_u16

	bypass pcs_lane_u16

	_ [0xc060 - 0xc059]pad_u16

	enable_override [2]pcs_lane_u16
}

type interlaken_common struct {
	control pcs_u16

	_ [0xc340 - 0xc331]pad_u16

	status pcs_u16
}

type clock_and_reset_common struct {
	over_sampling_mode_control pmd_lane_u16

	lane_reset_and_powerdown pmd_lane_u16

	lane_afe_reset_and_powerdown pmd_lane_u16

	lane_reset_and_powerdown_pin_disable pmd_lane_u16

	lane_debug_reset_control pmd_lane_u16

	uc_ack_lane_control pmd_lane_u16

	lane_reset_occurred pmd_lane_u16

	clock_reset_debug_control pmd_lane_u16

	pmd_lane_mode_status pmd_lane_u16

	lane_data_path_reset_status pmd_lane_u16

	lane_is_masked_from_multicast_writes pmd_lane_u16

	over_sampling_status          pmd_lane_u16
	over_sampling_status_from_pin pmd_lane_u16
	_                             [0xd0be - 0xd0bd]pad_u16
	lane_reset_active_low         pmd_lane_u16
}

type sigdet_common struct {
	control [3]pmd_lane_u16

	_ [0x8 - 0x3]pad_u16

	status pmd_lane_u16
}

type dig_common struct {
	revision_id0 pmd_lane_u16

	reset_control_pmd_active_low pmd_lane_u16

	reset_control_datapath pmd_lane_u16

	lane_is_masked_from_multicast_writes pmd_lane_u16

	top_user_control pmd_lane_u16

	uc_ack_core_control pmd_lane_u16

	core_reset_occurred pmd_lane_u16

	reset_seq_timer pmd_lane_u16

	core_datapath_reset_status pmd_lane_u16

	pmd_core_mode pmd_lane_u16

	revision_id1 pmd_lane_u16

	tx_lane_map_012 pmd_lane_u16

	tx_lane_map_3_lane_address_01 pmd_lane_u16

	tx_lane_address_23 pmd_lane_u16

	revision_id2 pmd_lane_u16
}

type tx_pattern_common struct {
	data [15]pmd_lane_u16
}

type pll_common struct {
	calibration_control [6]pmd_lane_u16

	reference_clock pmd_lane_u16

	multiplier pmd_lane_u16

	status [2]pmd_lane_u16

	debug_status pmd_lane_u16
}

type tlb_rx_common struct {
	pseudo_random_bitstream_checker_count_control pmd_lane_u16
	pseudo_random_bitstream_checker_control       pmd_lane_u16

	tx_to_rx_loopback pmd_lane_u16

	misc_control pmd_lane_u16

	pseudo_random_bitstream_checker_enable_timer_control pmd_lane_u16
	_                                                    [0xd168 - 0xd165]pad_u16
	tx_to_rx_loopback_status                             pmd_lane_u16
	pseudo_random_bitstream_checker_lock_status          pmd_lane_u16
	pseudo_random_bitstream_error_counts                 [2]pmd_lane_u16

	pmd_lock_status pmd_lane_u16
}

type tlb_tx_common struct {
	pattern_gen_config                        pmd_lane_u16
	pseudo_random_bitstream_generator_control pmd_lane_u16

	remote_loopback pmd_lane_u16

	misc_control pmd_lane_u16
}

type mdio_common struct {
	mask_data         pmd_lane_u16
	broadcast_address pmd_lane_u16

	mmd_select pmd_lane_u16

	aer pmd_lane_u16

	block_address pmd_lane_u16
}

type uc_cmd_controller struct {
	control pmd_lane_u16

	data pmd_lane_u16
}
