// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsc

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
		// [15:13] ref clock select
		//   0 => 25e6 Hz, 1 => 100e6, 2 => 1.25e6, 3 156.25e6
		//   4 => 187.5e6, 5 => 161.25e6, 6 => 50e6, 7 => 106.25e6
		// [12] cl37 high vco
		// [11] cl73 low vco
		// [10] pll reset enable
		// [9:8] master port number
		// [6:4] port mode select
		//   0 => 4 1 lane ports: 4x1
		//   1 => 2 1 lane ports + 1 2 lane port: 2x1 + 1x2
		//   2 => 1 2 lane port  + 2 1 lane ports: 1x2 + 2x1
		//   3 => 2 2 lane ports: 2x2
		//   4 => 1 4 lane port: 1x4
		// [3] single port mode (reset pll when autonegotiation completes)
		// [2] standalone mode (no mac hooked up)
		setup pcs_reg

		_ [0x9002 - 0x9001]pad_reg

		/* 2 bits for each of 4 lanes. */
		synce_control pcs_reg

		// PCS lane swap: 2 bit physical lane (pin) for each of 4 logical lanes.
		rx_lane_swap pcs_reg

		/* [7] auto negotiation
		   [6] TC
		   [5] DTE_XS
		   [4] PHY_XS
		   [3] PCS_XS
		   [2] WIS
		   [1] PMA_PMD
		   [0] Clause 22. */
		devices_in_package pcs_reg

		misc pcs_reg

		_ [0x9007 - 0x9006]pad_reg

		/* [0] [15] enable tick counts instead of ref clock sel, [14:0] tick numerator [18:4]
		   [1] [15:12] tick numerator [3:0], [11:0] tick denominator */
		tick_generation_control [2]pcs_reg

		/* [7:4] per-lane remote PCS loopback enable
		   [3:0] per-lane local PCS loopback enable. */
		loopback_control pcs_reg

		mdio_broadcast pcs_reg

		_ [0x900e - 0x900b]pad_reg

		// [5:0] model (tscf = 0x14), [15:14] revision (0 => tscf a0, 1 => tscf b0)
		serdes_id pcs_reg
	}

	_ [0x9010 - 0x900f]pad_reg

	pmd_x1 struct {
		// [8] enable direct uc pram interface writes
		// [1] core power on reset active low
		// [0] core data path reset active low.
		reset pcs_reg

		/* Used only when speed control is bypassed. [11:8] otp options [7:0] speed id. */
		mode pcs_reg

		/* [1] rx clock valid from pmd
		   [0] pmd pll lock status. */
		status pcs_reg

		latched_status pcs_reg

		/* For speed control bypass
		   [3] override enable for core data path reset
		   [2] override enable for core mode
		   [1] tx clock valid override
		   [0] pll lock override. */
		override pcs_reg
	}

	_ [0x9030 - 0x9015]pad_reg

	packet_generator struct {
		control [2]pcs_reg

		pseudo_random_test_pattern_control pcs_reg
		rx_crc_errors                      pcs_reg
		rx_prtp_errors                     pcs_reg
		rx_prtp_lock_status                pcs_reg
		_                                  [0x9037 - 0x9036]pcs_reg
		testpatt_seed                      [2][4]pcs_reg /* 58 bits A then B */

		_ [0x9040 - 0x903f]pcs_reg
		// 2 repeated payload bytes
		repeated_payload_bytes pcs_reg
		error_mask             [5]pcs_reg /* 80 bits high first */
		error_inject           [2]pcs_reg /* 2 regs */
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

		// [0] - valid/invalid counters
		// [1] - valid counters
		// [2] - invalid counters
		cl49_sync_header_counters [3]pcs_reg

		// [0] - bit 65:50
		// [1] - bit 49:34
		// [2] - bit 33:18
		// [3] - bit 17:2
		// [4] - bit 1:0
		sync_code_word [5]pcs_reg

		// [0] - bit 65:50
		// [1] - bit 49:34
		// [2] - bit 33:18
		// [3] - bit 17:2
		// [4] - bit 1:0
		sync_code_word_mask [5]pcs_reg
	}

	_ [0x9240 - 0x9230]pad_reg

	an_x1 struct {
		oui pcs_reg_32

		// [0] pri 5-0
		// [1] pri 11-6
		// [2] pri 17-12
		// [3] pri 23-18
		// [4] pri 29-24
		// [5] pri 35-30
		// [6] config
		priority_remap [7]pcs_reg

		_ [0x9250 - 0x9249]pad_reg

		/* Clause 37: auto negotiation for backplane. */
		cl37 struct {
			restart_timer       pcs_reg
			complete_ack_timer  pcs_reg
			timeout_error_timer pcs_reg
		}

		cl73 struct {
			// Timer for the amount of time to disable transmission in order to assure that the link parner enters a Link Fail state.
			break_link_timer pcs_reg
			// Timer for the amout ot time to wait to receive a page from the link partner.
			timeout_error_timer            pcs_reg
			parallel_detect_dme_lock_timer pcs_reg
			link_up_timer                  pcs_reg
			// [0] cl72, [1] not cl72
			// Timer for qualifying a link_status==FAIL indication or a link_status==OK indication when a link is first being
			// established and cl72 training is/is-not being run.
			qualify_link_status_timer           [2]pcs_reg
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

		/* 0x9270 + (lane * 0x10) for lanes 0-3 */
		lanes [4]struct {
			/* [0] [15:8] speed
			   [5] cl36 rx pipeline enable
			   [4] cl36 tx pipeline enable
			   [3] 0 => 66 bit data to t_pma, 1 => 40 bit
			   [2:0] log2 number of lanes. */
			/* [2] [15] cl72 enable
			   [14:11] over-sampler mode
			   [10:9] tx fifo mode
			   [8:7] tx encoder mode
			   [6] tx higig2 enable
			   [5:4] number of tx lanes bit muxed
			   [3:1] scr mode */
			main_override pcs_reg
			_             [0x02 - 0x01]pad_reg

			speed [9]pcs_reg
			_     [0x10 - 0xb]pad_reg
		}
	}

	_ [0xa000 - 0x92b0]pad_reg

	tx_x2 struct {
		mld_swap_count pcs_lane_reg

		// [3:0] CL48_TX_QRSVDCTRL For each bit set in this control swap the ordered set byte with the RX_X2_Control0_qrsvd_0.QrsvdSwap byte, for the TX PCS.For CL48 only
		// [4] CL48_TX_RF_ENABLE If this bit is a one, RFs are passed from the RS LAYER to the PCS.If this bit is a zero, RFs are replaced by IDLEs which are then passedfrom the RS LAYER to the PCS.For CL48 only.
		// 	[5] CL48_TX_LF_ENABLE If this bit is a one, LFs are passed from the RS LAYER to the PCS.If this bit is a zero, LFs are replaced by IDLEs which are then passedfrom the RS LAYER to the PCS.For CL48 only.
		// 	[6] CL48_TX_LI_ENABLE If this bit is a one, LIs (Link Interrupt) are passed from the RS LAYER to the encoder.If this bit is a zero, LIs are replaced by IDLEs which are then passedfrom the RS LAYER to the encoder.For CL48 only.
		// 	[15] BRCM_MODE_USE_K20PT5 Enable inserting /K20.5/ instead of /D20.5/ when LPI are encoded/decoded
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
		rx_decoder_status               pcs_reg
		rx_deskew                       pcs_reg
		_                               [0xa085 - 0xa082]pad_reg
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
		final_speed_config  pcs_lane_reg /* see sc_speed_override + 0-1 */
		_                   [0xc072 - 0xc071]pad_reg
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

		/* [15:14] scr mode (0-bypass; 1-64b66b all bits 2-8b10b all bits 3-64b66b no syncbits
		   [10] FEC enable
		   [8] cl49 tx RF enable (passed from RS to PCS)
		   [7] cl49 tx LF enable (passed from RS to PCS)
		   [6] cl49 tx LI (Link Interrupt) enable (passed from RS to PCS)
		   [2:3] tx fifo watermark
		   [1] tx lane reset (active low)
		   [0] tx lane enable. */
		misc pcs_lane_reg

		_ [0xc120 - 0xc114]pad_reg

		encode_status pcs_lane_reg
		pcs_status    pcs_lane_reg
	}

	_ [0xc130 - 0xc122]pad_reg

	rx_x4 struct {
		// [0] BRCM64B66_DESCRAMBLER_ENABLE If asserted, the data sent to the the brcm64b66 decoder is scrambled.
		//        Sync headers are not scrambled.
		// [2] LPI_ENABLE If off (0), LPIs are converted to IDLEs.  NOTE: LPI_ENABLE APPLIES TO BOTH TX AND RX pipelines
		// [4:3] CL36BYTEDELETEMODE
		//   2'b00 - 100M mode (Delete 9 out of every 10 bytes)
		//   2'b01 - 10M mode (Delete 99 out of every 100 bytes)
		//   2'b10 - Passthrough (No deletion)
		// [7:5] DESC2_MODE {000:NONE}, {001:CL49}, {010:BRCM}, {011:}, {100:CL48}, {101:CL36}, {110:CL82}, {111:NONE}
		// [10:8] DESKEWMODE 3'b000 - None 3'b001 - byte based deskew for 8b10b mode
		//    3'b010 - block based deskew for BRCM 64b66b mode
		//    3'b011 - block based deskew for IEEE CL82 mode
		//    3'b100 - cl36 mode enable
		// [13:11] DECODERMODE  3'b000 - None 3'b001 - cl49 64b66b mode 3'b010 - BRCM 64b66b mode
		//    3'b011 -  - cl49/BRCM 64b66b mode 3'b100 - 8b10b mode - cl48 mode 3'b101 - 8b10b mode - cl36 mode
		// [15:14] DESCRAMBLERMODE  r_descr1 mode00 bypass descrambler01 64b66b descrambler10 8b10b descrambler11 reserved
		pcs_control pcs_lane_reg

		fec_control              [3]pcs_lane_reg
		decoder_control          [2]pcs_lane_reg
		cl36_rx0                 pcs_lane_reg
		pma_control0             pcs_lane_reg
		_                        [0xc139 - 0xc138]pad_reg
		link_status_control      pcs_lane_reg
		_                        [0xc140 - 0xc13a]pad_reg
		user_fec_debug_read_data [2]pcs_lane_reg
		fec_burst_error_status   [2]pcs_lane_reg
		barrel_shifter_state     pcs_lane_reg
		cl49_lock_fsm_state      pcs_lane_reg
		decode_status            [6]pcs_lane_reg
		syncacq_status           [2]pcs_lane_reg
		bercnt                   pcs_lane_reg
		_                        [0xc152 - 0xc14f]pad_reg

		// [0]
		//   [13] link interrupt has transitioned high since last read.
		//   [14] remote fault has transitioned high since last read.
		//   [15] local fault has transitioned high since last read.
		// [1]
		//   All bits are clear on read.
		//   [6] Sync/block lock status indicator has transitioned low since last read.
		//       Block alignment for 64/66, sync_status FSM status for 8b10b = comma_align with no Errors
		//   [7] Sync status indicator has transitioned high since last read.
		//   [8] Link status has transitioned low since last read.
		//   [9] Link status indicator has transitioned high since last read.
		//   [10] High bit error rate has transitioned low since last read.
		//   [11] High bit error rate has transitioned high since last read.
		//   [12] Deskew status has transitioned low since last read.
		//   [13] Deskew status has transitioned high since last read.
		//   [14] Alignment marker lock has transitioned low since last read.
		//   [15] Alignment marker lock has transitioned high since last read.
		latched_pcs_status [2]pcs_lane_reg

		// [0] live sync status indicator for cl36, cl48, brcm mode
		// [1] live link status indicator
		// [2] high bit error rate indicator
		// [3] deskew achieved
		// [4] alignment marker lock status
		// [5] low power indicator
		// [6] link interrupt
		// [7] remote fault
		// [8] local fault
		pcs_live_status pcs_lane_reg

		cl82_am_lock_sm_latched_status pcs_lane_reg
		cl82_am_lock_sm_live_status    pcs_lane_reg
		fec_corrected_blocks_counter   [2]pcs_lane_reg
		fec_uncorrected_blocks_counter [2]pcs_lane_reg
		rx_gbox_error_status           pcs_lane_reg
		_                              [0xc161 - 0xc15c]pad_reg

		t12_fec_control                [3]pcs_lane_reg
		_                              [0xc170 - 0xc164]pad_reg
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
		/*[13:12] log2 number of lanes (software must set before starting AN)
		  [11] CL37 BAM enable
		  [10] CL73 BAM enable
		  [9] CL73 HPAM enable
		  [8] CL73 enable
		  [7] CL37 SGMII enable
		  [6] CL37 enable
		  [5] CL37 BAM to SGMII auto enable
		  [4] CL37 SGMII to CL37 auto enable
		  [3] CL73 BAM to HPAM auto enable
		  [2] HPAM to CL73 auto enable
		  [1] CL37 an restart
		  [0] CL73 an restart */
		enables pcs_lane_reg

		cl37_base_page_abilities pcs_lane_reg
		cl37_bam_abilities       pcs_lane_reg
		cl37_over_1g_abilities   [2]pcs_lane_reg

		/*[0][11] TX_NONCE_MATCH override
		     [10] TX_NONCE_MATCH value force
		     [9-5] TX_NONCE First CL73 nonce to be transmitted
		     [4-0] CL73_BASE_SELECTOR IEEE Annex 28A Message Selector
		  [1]
		     [11] CL73_REMOTE_FAULT 0-no remote fault field 1-yes
		     [10] NEXT_PAGE       Next page ability
		     [9-8] FEC_REQ  Forward Error Correction request
		     [7-6] CL73_PAUSE Pause Ability
		     [5] BASE_1G_KX1      MP5 ability bit 0
		     [4] BASE_10G_KX4     MP5 ability bit 1
		     [3] BASE_10G_KR      MP5 ability bit 2
		     [2] BASE_40G_KR4     MP5 ability bit 3
		     [1] BASE_40G_CR4     MP5 ability bit 4
		     [0] BASE_100G_CR10   MP5 ability bit 5
		*/
		cl73_base_page_abilities [2]pcs_lane_reg

		cl73_bam_abilities pcs_lane_reg

		// [0] PD_KX4_EN
		// [1] PD_KX_EN
		// [2] AN_GOOD_TRAP
		// [3] AN_GOOD_CHECK_TRAP
		// [4] LINKFAILTIMER_DIS
		// [5] LINKFAILTIMERQUAL_EN
		// [9:6] AN_FAIL_COUNT_LIMIT
		//     Number of times AN may retry after AN failure
		// [15:10]  OUI_CONTROL
		//   bit 5: require programmable OUI to detect CL73 HP modebit
		//   4: advertise programmable OUI for CL73 HP modebit
		//   3: require programmable OUI to detect CL73 BAMbit
		//   2: advertise programmable OUI in CL73 BAMbit
		//   1: require programmable OUI to detect CL37 BAMbit
		//   0: advertise programmable OUI in CL37 BAM
		misc_controls pcs_lane_reg

		_ [0xc190 - 0xc189]pad_reg

		link_partner_message_page_5    [4]pcs_lane_reg
		link_partner_message_page_1024 [4]pcs_lane_reg
		link_partner_base_page         [3]pcs_lane_reg

		_ [0xc1a0 - 0xc19b]pad_reg

		local_device_sw_pages [3]pcs_lane_reg
		link_partner_sw_pages [3]pcs_lane_reg

		// [7:0] TLA_LN_SEQUENCER_FSM_STATUS1 TLA Lane sequencer fsm latched status cont.  Clear on read of tla_ln_seq_status register
		// [8] PD_CL37_COMPLETED Parallel detect process has selected cl37 and it was completed.
		// [13] LP_STATUS_VALID  Set by HW, Clear on Read of lp_page_0
		// [14] LD_CONTROL_VALID Set by SW write to ld_page_0, Cleared when HW transfers the ld_page's
		// [15] AN_COMPLETED_SW  Software control page sequence. All page exchanges have completed
		sw_control_status pcs_lane_reg

		local_device_controls pcs_lane_reg

		// [0] CL73 auto-neg is complete
		// [1] CL37 auto-neg is complete; clear on read
		// [2] Received auto-neg next page without T toggling; clear on read
		// [3] Received invalid auto-neg page sequence; clear on read
		// [4] Received auto-neg MPS-5 OUI match; clear on read
		// [5] Received auto-neg MPS-5 OUI mismatch; clear on read
		// [6] Received auto-neg unformatted page 3; clear on read
		// [7] Received mismatching auto-neg message page; clear on read
		// [8] Received auto-neg message page 1024 (Over1G Message); clear on read
		// [9] Received auto-neg message page 5 (Organizationally Unique Identifier Message); clear on read
		// [10] Received auto-neg message page 1 (Null Message); clear on read
		// [11] Received auto-neg next page; clear on read
		// [12] Received auto-neg base page; clear on read
		// [13] Received non-SGMII page when in SGMII auto-neg mode; clear on read
		// [14] In Hewlett-Packard auto-neg mode; clear on read
		// [15] SGMII mode
		page_sequencer_status pcs_lane_reg

		// [0] Received auto-neg restart (0) page
		// [1] Entered auto-neg IDLE_DETECT state; clear on read
		// [2] Entered auto-neg DISABLE_LINK state; clear on read
		// [3] Entered auto-neg ERROR state; clear on read
		// [4] Entered auto-neg AN_ENABLE state; clear on read
		// [5] Entered auto-neg ABILITY_DETECT state; clear on read
		// [6] Entered auto-neg ACKNOWLEDGE_DETECT state; clear on read
		// [7] Entered auto-neg COMPLETE_ACKNOWLEDGE state; clear on read
		// [8] Auto-neg consistency mismatch detected; clear on read
		// [9] Page Exchanger Received non-zero configuration ordered set; clear on read
		// [10] Page Exchanger entered AN_RESTART state; clear on read
		// [11] Page Exchanger entered AN_GOOD_CHECK state; clear on read
		// [12] Page Exchanger entered LINK_OK state; clear on read
		// [13] Page Exchanger entered NEXT_PAGE_WAIT state; clear on read
		page_exchanger_status pcs_lane_reg

		// [1:0] DME Receive State
		// [2] Received invalid ordered set; clear on read
		// [3] Received configuration ordered set; clear on read
		// [4] Received idle ordered set; clear on read
		// [5] Valid DME page received; clear on read
		// [6] DME Delimiter detected; clear on read
		// [7] Missing DME clock transition detected; clear on read
		// [8] A CL73 DME page longer than the maximum specified by cl73_page_test_max_timer was detected; clear on read
		// [9] A CL73 DME page shorter than the  minimum specified by cl73_page_test_min_timer was detected; clear on read
		// [10] Too long DME pulse detected, duration - minimum 35 samples. Each sample 0.4ns; clear on read
		// [11] Too short DME pulse detected, duration - 2 to 4 samples. Each sample 0.4ns; clear on read
		// [12] Too moderate DME pulse detected, duration - 19 to 29 samples. Each sample 0.4ns; clear on read
		page_decoder_status pcs_lane_reg

		// [0] an hcd switch to cl37
		// [1] HCD Hi-Gig II ability
		// [2] HCD CL72 training ability
		// [3] HCD forward-error correction ability
		// [11:4] HCD speed
		// [12] tx pause ability
		// [13] rx pause ability
		// [14] full duplex
		// [15] error: no common speed
		ability_resolution pcs_lane_reg

		// [0] Speed status for parallel detect attempt 0: KX4, 1: KX
		// [1] Parallel detect is active
		// [5:2] Number of AN retries due to AN failure.  Saturates, clear on read.
		// [6] Auto-neg in progress
		// [7] Parallel detect complete
		// [8] Remote fault indicated in AN base page
		// [14:9] Number of AN retries for any reason. Saturates, clear on Read.
		// [15] AN complete
		misc_status pcs_lane_reg

		// [0] init
		// [1] set an speed
		// [2] wait an speed
		// [3] wait an done
		// [4] set rs speed
		// [5] wait rs cl72 enable
		// [7] an ignore link
		// [8] an good check (cl73)
		// [9] an good (cl73)
		// [10] an fail
		// [11] an dead
		// [12] set pd speed
		// [13] wait pd speed
		// [14] pd ignore link
		// [15] pd good check
		tla_sequencer_status pcs_lane_reg

		// First unexpected page received.
		sequencer_unexpected_page pcs_lane_reg
	}

	_ [0xc253 - 0xc1af]pad_reg

	cl72_link struct {
		control pcs_reg
	}

	_ [0xc301 - 0xc254]pad_reg

	digital_control struct {
		ctl_1000x pcs_reg
		_         [0xc30a - 0xc302]pad_reg
		spare     [2]pcs_reg
	}

	_ [0xc330 - 0xc30c]pad_reg

	interlaken interlaken_common

	_ [0xd001 - 0xc341]pad_reg

	dsc_a struct {
		cdr_control            [3]pmd_lane_reg
		rx_pi_control          pmd_lane_reg
		cdr_integration_status pmd_lane_reg
		cdr_phase_error_status pmd_lane_reg
		rx_pi_d_counter        pmd_lane_reg
		rx_pi_p_counter        pmd_lane_reg
		rx_pi_m_counter        pmd_lane_reg
		rx_pi_differential     pmd_lane_reg
		training_sum           pmd_lane_reg
		_                      [0xd00d - 0xd00c]pad_reg
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
		dfe_common_control pmd_lane_reg
		// 2 register blocks; control/pattern-control
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
		// [0] dfe1
		// [1] dfe2
		// [2] dfe3,4,5
		dfe_status     [3]pmd_lane_reg
		vga_tap_values pmd_lane_reg
	}

	_ [0xd040 - 0xd03f]pad_reg

	dsc_e struct {
		control             pmd_lane_reg
		peak_filter_control [2]pmd_lane_reg
		// [0] adj-data-odd/adj-data-even
		// [1] adj-p1-odd/adj-p1-even
		// [2] adj-m1-odd/adj-m1-even
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
		/* [2] [15] force electrical idle
		   [14:13] tx driver mode
		   [12] sign for 2nd post cursor tap
		   [11:8] 2nd post cursor coefficent
		   [7] sign for 3rd post cursor tap
		   [6:4] 3rd post cursor coefficent
		   [3:0] master amplitude control. */
		control          [3]pmd_lane_reg
		_                [0xd0d8 - 0xd0d3]pad_reg
		internal_control pmd_lane_reg
		status           pmd_lane_reg
	}

	_ [0xd0b0 - 0xd0aa]pad_reg

	ams_com struct {
		// [0-8] pll
		// [9]   internal-pll
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
		// [0]
		// [4:0] TXFIR_PRE_OVERRIDE tx fir pre tap override value.when txfir_override_en field of the txfir_control2_registeris set 1'b1, then the txfir_pre_override field is used tooverride the cl72 pre tap value
		// [11:5] TXFIR_POST_OVERRIDE tx fir post tap override value.when txfir_override_en field of the txfir_control2_registeris set 1'b1, then the txfir_post_override field is used tooverride the cl72 post tap value

		// [1]
		// [6:0] TXFIR_MAIN_OVERRIDE tx fir main tap override value.when txfir_override_en field of the txfir_control2_registeris set 1'b1, then the txfir_main_override field is used tooverride the cl72 main tap value
		// [11:7] TXFIR_POST2      tx fir post2 tap value.Post2 tap value only driven from a registerThe value range is -16 ..+15 and it is in 2's complement format
		// [15] TXFIR_OVERRIDE_EN txfir override enablewhen txfir_override_en field to 1'b1, then txfir_main_override,txfir_post_override and  txfir_pre_override filed are used tooverride the cl72 main/post/pre tap value

		// [2]
		// [3:0] TXFIR_PRE_OFFSET tx fir pre tap offset value -8 to +7This field is used to adjust the Pre tap valuesThe mapping is not 2's complement, it isregister value  = tap adjusted by0  = -81  = -72  = -63  = -54  = -45  = -36  = -27  = -18  =  09  = +110 = +211 = +312 = +413 = +514 = +615 = +7
		// [7:4] TXFIR_MAIN_OFFSET tx fir main tap offset value -8 to +7This field is used to adjust the Main tap valuesThe mapping is not 2's complement, please txfir_pre_offset field description
		// [11:8] TXFIR_POST_OFFSET tx fir post tap offset value -8 to +7This field is used to adjust the Post tap valuesThe mapping is not 2's complement, please txfir_pre_offset field description
		// [15:12] TXFIR_POST2_OFFSET tx fir post2 tap offset value -8 to +7This field is used to adjust the Post2 tap valuesThe mapping is not 2's complement, please txfir_pre_offset field description
		control [3]pmd_lane_reg

		/* [0] [13:8] TX FIR post tap (+1) value after override mux
		   [4:0]  TX FIR pre  tap (-1) value after override mux */
		/* [1] [6:0]  TX FIR main tap (+0) value after override mux */
		/* [2] [13:8] TX FIR post tap (+1) value after offset adjustment
		   [4:0]  TX FIR pre  tap (-1) value after offset adjustment */
		/* [3] [12:8] TX FIR post2 tap (+2) value after offset adjustment
		   [6:0]  TX FIR main tap (+0) value after offset adjustment */
		/* [4] [3:0]  TX FIR post3 tap (+3) value after offset adjustment */
		status [4]pmd_lane_reg

		tx_uc_control pmd_lane_reg

		/* [12] data path reset tx disable (0 => enable, 1 => disable)
		   [11:10] tx disable output (0 => electrical idles, 1 => send power down 2 => send ones 3 => send zeros)
		   [9] enable EEE alert mode
		   [8] enable EEE quiet mode
		   [7] tx disable timer units (0 => 2usec, 1 => 1msec)
		   [6:2] number of units
		   [1] tx disable from pmd tx disable pin
		   [0] sw tx disable. */
		misc_control pmd_lane_reg

		// [3:0] TXFIR_POST3 tx fir post3 tap value.Post3 tap value only driven from a registerThe value range is -8 ..+7 and it is in 2's complement format
		// [11:8] TXFIR_POST3_OFFSET tx fir post3 tap offset value -8 to +7This field is used to adjust the Post3 tap valuesThe mapping is not 2's complement, please txfir_pre_offset field description
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

		// [0] Start read/write to program RAM
		// [1] Stop read/write to program RAM
		// [2] Read program RAM to readback the microcode
		// [3] Write program RAM to load the microcode
		// [4] MDIO DW8051 reset active low
		// [6] Enable auto-increment of RAM address after read to the ram_rddata registers
		// [8:7] Ram access through the mdio (register interface) mode
		//   0 => Program RAM load while DW8051 in reset
		//   1 => Program RAM access while dw8051 could be running
		//   2 => Data memory access while dw8051 could be running
		// [9] Program/Data RAM access width 1 => 8 bits (byte), 0 => 16 bits (word)
		// [10] uc will automatically out of reset after the download is complete
		// [11] Force chip select to program RAM during mdio program write.
		//   0 = chip select to program RAM from mdio_to_program.
		//   1 = chip select to program RAM  is from mdio_prog_ram_cs_frc_val
		// [12] Force chip select to program RAM value. See mdio_prog_ram_cs_frc description.
		// [13] Force pmi_hp_ack.0 = pmi_hp_ack to the DW8051_to_pmi fsm is from pmi_hp_ack pins.1 = pmi_hp_ack to the DW8051_to_pmi fsm is pmi_hp_ack_frc_val.
		// [14] Force pmi_hp_ack value. See pmi_hp_ack_frc description.
		// [15] Zero all program RAM command.  This operation is normally required before the microcode is loaded to calculate the checksum
		command1 pmd_lane_reg

		write_data pmd_lane_reg
		read_data  pmd_lane_reg

		// [0] Set when user changes the start_address during burst read/write.
		// [1] Set when user set stop during burst read/write.
		// [5:2] FSM value
		// [15] Program RAM initialization done
		mdio_8051_fsm_status pmd_lane_reg

		status1                         pmd_lane_reg
		external_station_to_uc_mbox     [2]pmd_lane_reg
		uc_to_external_station_mbox_lsw pmd_lane_reg
		command2                        pmd_lane_reg
		uc_to_external_station_mbox_msw pmd_lane_reg

		// [0] Enable parallel interface to load program RAM
		// [1] Bypass the flops used to capature data from the pram_interface
		// [2] Parallel interface reset from Program RAM load 0 - asserted 1 - de-asserted
		// [10] program ram in rush current input(s) force
		// [11] program ram in rush current input(s) force value
		// [12] Disable program RAM ECC
		// [15:13] MICRO_GEN_STATUS_SEL
		command3 pmd_lane_reg

		// [1] uc reset active low
		// [0] uc clock enable
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
