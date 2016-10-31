// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsc

/* IEEE PHY ID[01] registers (reset values 0x600d 0x8770) */
type phyid_common struct {
	id [2]pcs_reg
}

type acc_mdio_common struct {
	/* [15:14] 0 => address 1 => data no post increment
	   2 => data post increment on read/write
	   3 => data post increment on write only
	 [4:0] DEVAD device address. */
	mdio_control pmd_lane_reg

	/* [15:0] mdio address or data */
	mdio_address_or_data pmd_lane_reg
}

/* Clause 93 (PMD for 100G-KR) / Clause 72 (PMD for 10G-KR)
   TX link training. */
type cl93n72_common struct {
	/* [1] enable 10GBASE-KR start-up protocol
	   [0] restart 10GBASE-KR clause 93 link training (self-clearing bit). */
	control pmd_lane_reg

	/* [3] training failure
	   [2] start-up protocol in progress
	   [1] training frame lock/delineation detected
	   [0] 1 => reciever trained and ready to receive data */
	status pmd_lane_reg

	/* [13] 1 => preset coefficents 0 => normal operation
	   [12] 1 => initialize coef, 0 => normal operation
	   [5:4] +1 coef update: 1 => increment, 2 => decrement, 0 => hold
	   [3:2]  0 coef update: 1 => increment, 2 => decrement, 0 => hold
	   [1:0] -1 coef update: 1 => increment, 2 => decrement, 0 => hold. */
	last_coefficient_update_received_from_link_partner pmd_lane_reg

	/* [15] training complete; rx ready
	   [5:4] +1 coef status: 0 => not updated, 1 => updated, 2 => minimum, 3 => maximum
	   [3:2]  0 coef status: 0 => not updated, 1 => updated, 2 => minimum, 3 => maximum
	   [1:0] -1 coef status: 0 => not updated, 1 => updated, 2 => minimum, 3 => maximum */
	last_coefficient_status_received_from_link_partner pmd_lane_reg

	// as above.
	last_coefficient_update_sent_to_link_partner pmd_lane_reg
	last_coefficient_status_sent_to_link_partner pmd_lane_reg
}

type rx_cl82_am_common struct {
	/* Clause 82: 64/66b 40G encoding */
	rx_cl82_alignment_marker_timer pcs_reg
	_                              [0x9130 - 0x9124]pad_reg

	rx_cl82_per_lane_alignment_marker_bytes [3]pcs_reg /* 3 AM bytes for lanes 0-1 (3 regs) */
}

type pmd_x4_common struct {
	// [12:9] osr mode
	// [8] tx disable
	// [3] lane rx power down
	// [2] lane tx power down
	// [1] reset lane data path and registers
	// [0] lane datapath reset override
	lane_reset_control pcs_lane_reg

	// Only used for speed override.
	lane_mode_config pcs_lane_reg

	/* [2] rx clock valid from pmd
	   [1] signal detect from pmd
	   [0] rx lock from pmd */
	lane_status pcs_lane_reg

	latched_lane_status pcs_lane_reg

	// [0] rx lock override
	// [1] signal detect override
	// [2] rx clock valid override
	// [3] lane mode override
	// [4] osr mode override
	// [5] tx disable override enable
	// [6] lane datapath reset override enable
	lane_override pcs_lane_reg

	_ [0xc018 - 0xc015]pad_reg

	eee_control pcs_lane_reg
	eee_status  pcs_lane_reg
}

type speed_change_x4_common struct {
	/* [8] start sw speed change (to start write 0 then write 1)
	   [7:0] speed */
	control pcs_lane_reg

	// [0] done read-to-clear
	// [1] final speed config regs valid read-to-clear
	status_read_to_clear pcs_lane_reg

	// [0] pll lock time out
	// [1] pmd lock time out
	error pcs_lane_reg

	_ [0xc054 - 0xc053]pad_reg
	/*R/W Speed Control logic FSM debug info
	16'h8000 - START
	16'h4000 - RESET_PCS
	16'h2000 - RESET_PMD_LANE
	16'h1000 - RESET_PMD_PLL
	16'h0800 - APPLY_SPEED_CFG
	16'h0400 - WAIT_CFG_DONE
	16'h0200 - ACTIVATE_PMD
	16'h0100 - WAIT_PLL_RESET
	16'h0080 - PLL_LOCK_FAIL
	16'h0040 - ACTIVATE_TX
	16'h0020 - WAIT_PMD_LOCK
	16'h0010 - ACTIVATE_RX
	16'h0008 - PMD_LOCK_FAIL
	16'h0004 - DONE
	16'h0002 - STOP
	16'h0001 - BYPASS
	*/
	debug            pcs_lane_reg
	n_lanes_override pcs_lane_reg
	_                [0xc058 - 0xc056]pad_reg
	bypass           pcs_lane_reg
	_                [0xc060 - 0xc059]pad_reg
	enable_override  [2]pcs_lane_reg
}

type interlaken_common struct {
	control pcs_reg
	_       [0xc340 - 0xc331]pad_reg
	status  pcs_reg
}

type clock_and_reset_common struct {
	// [0-3]  oversample mode value
	// [15]   oversample mode force. If set, value used otherwise pin-input used.
	over_sampling_mode_control pmd_lane_reg

	/* [1] active low lane data path soft reset.  This soft reset is equivalent to the hard reset input pin pmd_ln_dp_h_rstb_i.
	   [2] 0 => lane rx power up, 1 => lane rx power down
	   [3] 0 => lane tx power up, 1 => lane tx power down
	   [4] power down for afe signal detect; 1 => power down, 0 => power up */
	lane_reset_and_powerdown pmd_lane_reg

	lane_afe_reset_and_powerdown pmd_lane_reg

	/* [0] disable ln_h_rstb input pin
	   [1] disable ln_dp_h_rstb input pin
	   [2] disable rx h pwrdn input pin
	   [3] disable tx h pwrdn input pin */
	lane_reset_and_powerdown_pin_disable pmd_lane_reg

	lane_debug_reset_control pmd_lane_reg

	// [1] uC will write this to 1 to acknowledge a reset event after seeing "lane_dp_reset_coccured"
	// [0] uC will write this to 1 to indicate it's configuration of the lane is complete.
	//     Writing to 1'b1 will release internal hold on lane_dp_reset, only if lane_dp_reset_state is 3'b001.
	uc_ack_lane_control pmd_lane_reg

	// [0] Set to 1'b1 upon lane level register reset and remains so until cleared by register write from uC.
	lane_reset_occurred pmd_lane_reg

	clock_reset_debug_control pmd_lane_reg

	pmd_lane_mode_status pmd_lane_reg

	// [2] lane data path reset active via register or pin control.
	// [1] lane data path reset occurred (latched high)
	// [0] lane data path reset held.
	// Reset value: all ones.
	lane_data_path_reset_status pmd_lane_reg

	lane_is_masked_from_multicast_writes pmd_lane_reg

	/* [3:0] oversampling mode
	   OSx1          4'd0
	   OSx2          4'd1
	   OSx4          4'd2
	   OSx16P5       4'd8
	   OSx20P625     4'd12 */
	over_sampling_status          pmd_lane_reg
	over_sampling_status_from_pin pmd_lane_reg
	_                             [0xd0be - 0xd0bd]pad_reg
	// [0] active low lane soft reset.
	// Default value: 1
	lane_reset_active_low pmd_lane_reg
}

type sigdet_common struct {
	// [1] default value 0xa008
	//   [0] Disable the signal_detect from AFE.
	//   [1] Enable the external (optical) LOS path into the sigdet filter.
	//   [2] Invert the polarity of the pmd_ext_los pin.
	//   [3] Ignore the pmd_rx_mode (low power mode) input pin. Set to 1'b0 if EEE mode is supported by the PCS
	//   [4] 1'b1 will use 1us heartbeat for los_count, signal_detect_count and mask_count counters instead of comclk.
	//   [5] pmd_energy_detect force.
	//   [6] pmd_energy_detect Force Value.
	//   [7] pmd_signal_detect Force.
	//   [8] pmd_signal_detect Force Value.
	//   [15:11] Defines the mask_count timer for energy_detect. Valid range is 0 to 31 which maps to 0 to 448 clock cycles.
	//      Refer PMD spec for more details about the mapping.
	control [3]pmd_lane_reg

	_ [0x8 - 0x3]pad_reg

	/* [10:8] live status of signal detect threshold from AFE
	   [5] latched signal detect raw change (clear on read)
	   [4] raw signal detect
	   [3] latched filtered energy detect raw change (clear on read)
	   [2] filtered energy detect
	   [1] latched filtered signal detect change (clear on read)
	   [0] filtered signal detect. */
	status pmd_lane_reg
}

type dig_common struct {
	/* [15:14] all layer rev id (A = 0 B = 1...)
	   [13:11] metal mask rev
	   [5:0] model number. (0x1b falcon) */
	revision_id0 pmd_lane_reg

	// [0] core reset active low (default: 1)
	reset_control_pmd_active_low pmd_lane_reg

	// [1] disable pmd_core_dp_h_rstb pin
	reset_control_datapath pmd_lane_reg

	lane_is_masked_from_multicast_writes pmd_lane_reg

	// [15] micro-controller (uc) active
	//   When set to 1'b1 then Hardware should wait for uC handshakes to wake up from datapath reset.
	//   When set to 1'b0 then Hardware can internally assume that uc_ack_* = 1.
	// [14] active high PLL power down
	// [13] active low core level soft reset (resets data path for all lanes)
	// [9:0] 1usec heartbeat clock count (set to 4usec in units of comclk).  default: 625 = 1usec with 156.25Mhz clock.
	top_user_control pmd_lane_reg

	// [0] uC will write this to 1 to indicate it's configuration of the core is complete.
	//   Writing to 1'b1 will release internal hold on core_dp_reset, only if core_dp_reset_state is 3'b001.
	// [1] uC will write this to 1 to acknowledge a reset event after seeing "core_dp_reset_coccured".
	uc_ack_core_control pmd_lane_reg

	// [0] Set to 1'b1 upon core level register reset and remains so until cleared by register write from uC.
	core_reset_occurred pmd_lane_reg

	reset_seq_timer pmd_lane_reg

	// [14] lane reset released
	// [12:8] lane reset released index
	// [2] Set to 1'b1 whenenver core_dp_reset is currently requested through any register or pin controls.
	// [1] Set to 1'b1 whenenver core_dp_reset is currently requested through any register or pin controls and is latched high.
	// [0] Set to 1'b1 whenenver core_dp_reset is internally held.
	//   Cleared to 1'b0, only if core_dp_reset_state==001 and uc_ack_core_cfg_done == 1.
	// default: 0x7
	core_datapath_reset_status pmd_lane_reg

	pmd_core_mode pmd_lane_reg

	/* [15:12] # lanes
	   [5] mdio present
	   [4] microcontroller present
	   [3] clause 72 present
	   [2] pcs interface retiming flops present
	   [1] ultra low latency path present
	   [0] EEE support present. */
	revision_id1 pmd_lane_reg

	/* Tx swap = PCS (rx) swap and then PMD (tx) swap.
	     [14:10] phys PMD pin index mapped to physical analog front end lane 2
		 [9:5]   phys PMD pin index mapped to physical analog front end lane 1
		 [4:0]   phys PMD pin index mapped to physical analog front end lane 0 */
	tx_lane_map_012 pmd_lane_reg

	/* [14:10] logical address of lane with PMD physical pin 1
	   [9:5]   logical address of lane with PMD physical pin 0 (needed for pmd loopback?)
	   [4:0]   phys PMD pin index mapped to physical analog front end lane 3. */
	tx_lane_map_3_lane_address_01 pmd_lane_reg

	/* [12:8] logical address of lane with PMD physical pin 3
	   [4:0]  logical address of lane with PMD physical pin 2. */
	tx_lane_address_23 pmd_lane_reg

	revision_id2 pmd_lane_reg
}

type tx_pattern_common struct {
	data [15]pmd_lane_reg
}

type pll_common struct {
	// [4] [15] start pll sequence (write 0 then write 1 to start)
	// 	   [4] force cap done
	// 	   [3] force pll cap pass
	// 	   [1] force pll lock.
	// [5] [13:0] refclk_divcnt
	calibration_control [6]pmd_lane_reg

	//  [2:0] REFCLK_DIVCNT_SEL Refclk Divider Mode Select.
	//       refclk_divcnt value to generate 25 Khz signal   refclk_divcnt_sel[2:0]
	// 390.625 Mhz               15625                             3'd0
	// 161.1328185 Mhz           6445.31274 = ~6445                3'd1
	// 156.25 Mhz                6250                              3'd2   (default)
	// 125.00 Mhz                5000                              3'd3
	// 106.25 Mhz                4250                              3'd4
	// 78.125 Mhz                3125                              3'd5
	// -                         -                                 3'd6   (rsvd for future use)-
	//                          refclk_divcnt[13:0]               3'd7   (programmable with max value of 16383)
	// defaulted to 0x7 for faster pll lock time (0x2 before)
	reference_clock pmd_lane_reg

	// [13:9] vco range adjust
	// [8] rescale force
	// [7:4] rescale force val
	// [3:0] multiplier.
	// 64(0000), 66(0001), 80(0010), 128(0011),
	// 132(0100), 140(0101), 160(0110), 165(0111),
	// 168(1000), 170(1001), 175(1010), 180(1011),
	// 184(1100), 200(1101), 224(1110), 264(1111)
	// Default value for tscf: pll multiplier 165; all else 0
	// Default value for tsce: pll multiplier 66; all else 0
	multiplier pmd_lane_reg

	/* [0] [15] pll lock lost (clear on read )
	   [12] frequency det done
	   [11] frequency pass (lock)
	   [10] pll sequencer done
	   [9] pll sequencer finished successfully
	   [8] pll is locked. */
	status       [2]pmd_lane_reg
	debug_status pmd_lane_reg
}

type tlb_rx_common struct {
	pseudo_random_bitstream_checker_count_control pmd_lane_reg
	pseudo_random_bitstream_checker_control       pmd_lane_reg

	/* [0] tx -> rx digital loopback enable */
	tx_to_rx_loopback pmd_lane_reg

	// [0] 1 => invert all the datapath bits of the logical lane (polarity).
	// [1] For debugging, 1 => pmd_rx_lock will be forced to 1'b0 during digital loopback.
	//     0 => pmd_rx_lock will be forced to 1'b1 during digital loopback.
	// [2] Enable the Differential Decoder for pmd_rx_data.  Only applicable to PCS RX data in OS1, 2 and 4 modes.
	//   Write it to 1'b0 for 1G OSR modes 16P5 and 20P625.
	misc_control pmd_lane_reg

	pseudo_random_bitstream_checker_enable_timer_control pmd_lane_reg
	_                                                    [0xd168 - 0xd165]pad_reg
	tx_to_rx_loopback_status                             pmd_lane_reg
	pseudo_random_bitstream_checker_lock_status          pmd_lane_reg
	pseudo_random_bitstream_error_counts                 [2]pmd_lane_reg /* 2 registers msb/lsb */

	/* [1] sticky clear on read lock status change
	   [0] PMD is in locked state; PCS should have acceptable BER rate. */
	pmd_lock_status pmd_lane_reg
}

type tlb_tx_common struct {
	pattern_gen_config                        pmd_lane_reg
	pseudo_random_bitstream_generator_control pmd_lane_reg

	remote_loopback pmd_lane_reg

	// [0] invert data path bits (polarity)
	// [1] PCS interface native analog format enable.
	//   1 => TX PCS sends the over-sampled data in this mode which is sent directly to AFE.
	//   0 => Raw Data Mode where for every data request TX PCS will send 20 bits of valid data.
	// [2]  TX Data MUX Select Priority Order. When 1'b1 then priority of Pattern and PRBS generators are swapped w.r.t. CL72.
	//   0 => TX Data Mux select order from higher to lower priority is {rmt_lpbk, patt_gen, cl72_tx, prbs_gen, tx_pcs}.
	//   1 => TX Data Mux select order from higher to lower priority is {rmt_lpbk, prbs_gen, cl72_tx, patt_gen, tx_pcs}.
	// [3] Enables the Differential Encoder for pmd_tx_data. Only applicable to PCS TX data in OS1, 2 and 4 modes.
	//   Set to 0 for 1G OSR modes 16P5 and 20P625.
	misc_control pmd_lane_reg
}

type mdio_common struct {
	mask_data         pmd_lane_reg
	broadcast_address pmd_lane_reg

	/* [15] multi PRTAD mode
	   [14] multi MMD mode
	   [6] pcs enable
	   [5] dte
	   [4] phy
	   [3] an
	   [2] pmd
	   [0] clause 22. */
	mmd_select pmd_lane_reg

	/* Upper 16 bits of MDIO transaction.
	   [15:11]
	     cl22 0, pma/pmd 1, cl73/auto negotiation 3, PHY 4, DTE 5, PCS 6
	   [10:0] lane number (0x1ff => all lanes) */
	aer pmd_lane_reg

	/* clause 22 */
	block_address pmd_lane_reg
}

type uc_cmd_regs struct {
	/* [15:8] supplemental info
	   [7] ready for command (set to zero to issue command)
	   [6] error found
	   [5:0] uc command */
	control pmd_lane_reg

	/* Data for uc commands. */
	data pmd_lane_reg
}
