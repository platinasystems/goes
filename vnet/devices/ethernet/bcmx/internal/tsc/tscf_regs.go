package tsc

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
)

func get_tscf_regs() *tscf_regs { return (*tscf_regs)(m.RegsBasePointer) }

// Register Map for TSCF core
type tscf_regs struct {
	_ [0x0002 - 0x0000]pad_reg

	phyid_common

	_ [0x000d - 0x0004]pad_reg

	acc_mdio_common

	_ [0x0096 - 0x000f]pad_reg

	cl93n72_common

	_ [0x9000 - 0x009c]pad_reg

	main struct {
		// [15:14] master port number
		// [13:10] per logical lane cl72 link-training enable
		// [9:7] refclk frequency (0 => 25Mhz, 1 => 100Mhz, 2 => 125 Mhz, 3 => 156.25Mhz,
		//                         4 => 187.5Mhz, 5 => 161.25Mhz, 6 => 50Mhz, 7 => 106.25Mhz)
		// [6:4] port mode select
		//   0 => 4 1 lane ports: 4x1
		//   1 => 2 1 lane ports + 1 2 lane port: 2x1 + 1x2
		//   2 => 1 2 lane port  + 2 1 lane ports: 1x2 + 2x1
		//   3 => 2 2 lane ports: 2x2
		//   4 => 1 4 lane port: 1x4
		// [3] single port mode (reset pll when autonegotiation completes)
		//   Used by AN logic to determine whether to reset the PLL after AN completes.
		//   If set, when AN completes, the PLL will be reset to operate consistent with the resolved AN speed.
		//   If not set, the PLL will not change once AN completes.
		// [2] standalone mode (no mac hooked up)
		// [1] cl73 vco frequency 0 => 20.625GHz, 1 => 25.78125GHz
		// [0] tsc_clk_frequency 0 => pll/40, 1 => pll/48
		// Default: 0x180 => refclk = 156.25MHz all else 0
		setup pcs_reg

		/* 2 bits for each of 4 lanes. */
		synce_control [2]pcs_reg

		/* [10:9] 100G forward error correction status.  Valid only when autonegotiation resolves to 100G speeds.
		   [8] pll reset enable by speed control logic
		   PCS lane swap: 2 bit physical lane (pin) for each of 4 logical lanes. */
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

		_ [0x9007 - 0x9005]pad_reg

		/* [0] [15] enable tick counts instead of ref clock sel, [14:0] tick numerator [18:4]
		   [1] [15:12] tick numerator [3:0], [11:0] tick denominator */
		tick_generation_control [2]pcs_reg

		/* [7:4] per-lane remote PCS loopback enable
		   [3:0] per-lane local PCS loopback enable. */
		loopback_control pcs_reg

		mdio_broadcast      pcs_reg
		mdio_timeout        pcs_reg
		mdio_timeout_status pcs_reg
		_                   [0x900e - 0x900d]pad_reg

		// [5:0] model (tscf = 0x14), [15:14] revision (0 => tscf a0, 1 => tscf b0)
		serdes_id pcs_reg
	}

	_ [0x9010 - 0x900f]pad_reg

	pmd_x1 struct {
		/* [1] core power on reset active low
		   [0] core data path reset active low. */
		reset pcs_reg

		/* Used only when speed control is bypassed. [11:8] otp options [7:0] speed id. */
		mode pcs_reg

		/* [1] rx clock valid from pmd
		   [0] pmd pll lock status. */
		status pcs_reg

		/* For speed control bypass
		   [3] override enable for core data path reset
		   [2] override enable for core mode
		   [1] tx clock valid override
		   [0] pll lock override. */
		override pcs_reg
	}

	_ [0x9030 - 0x9014]pad_reg

	packet_generator struct {
		// [0] [15:12] 0 idles, 1 => 1 packet, 2 => unlimitied packets
		//   [11] rx packet crc check enable
		//   [10] rx msbus type 1: MII/GMII type octet, 0: XGMII/XLGMII type octet
		//   [9] clear crc error count in packet checker
		//   [8:7] Port for which Packet checker or PRTP checker should be enabled.
		//     Packet checker and PRTP checker are shared across all port, so only one port can have either of them enabled
		//   [6:3] per lane prtp data pattern sel 0 local fault pattern, 1 zero data pattern
		//   [2] replace idles with eee lpi
		//   [1:0] Port for which Packet generator or PRTP generator should be enabled.
		//     Packet generator and PRTP generator are shared across all port, so only one port can have either of them enabled
		//
		// [1] [4:0] IPG_SIZE number of bytesrange is 0 to 32 bytes
		//    [10:5] PKT_SIZE number of 256 bytesrange is 256 to 16384 bytes.
		//      there is a dependency with respect to ipg_size programming in non GMII mode because of SOP alignment
		//      need to meet this requirement: (pkt_size+ipg_size) / 8 = 4
		//    [13:11] PAYLOAD_TYPE
		//    [14] TX_MSBUS_TYPE    1: MII/GMII type octet, 0: XGMII/XLGMII type octet
		//    [15] PKTGEN_EN        packet gen enable
		//
		// [2] [8:5] clear rx counters [4:1] clear tx counters
		control [3]pcs_reg

		pseudo_random_test_pattern_control pcs_reg
		rx_crc_errors                      pcs_reg

		_ [0x9037 - 0x9035]pad_reg

		testpatt_seed [2][4]pcs_reg /* 58 bits A then B */

		_ [0x9040 - 0x903f]pad_reg

		// 2 repeated payload bytes
		repeated_payload_bytes pcs_reg
		error_mask             [5]pcs_reg /* 80 bits high first */
		error_inject           [2]pcs_reg /* 2 regs */
	}

	_ [0x9123 - 0x9048]pad_reg

	rx_cl82_am_common

	_ [0x9200 - 0x9133]pad_reg

	tx_x1 struct {
		// [11:8] five lane bitmux watermark (units of 66 bit blocks)
		// [7:4] two lane bitmux watermark
		// [3:0] single lane bitmux watermark */
		pma_fifo_watermark        pcs_reg
		pma_delay_after_watermark pcs_reg /* units of cycles */

		/* [0] fec enable for tx pipeline */
		cl91_fec_enable pcs_reg
	}

	_ [0x9221 - 0x9203]pad_reg

	rx_x1 struct {
		/* [7:2] CL49 number of error blocks before high BER is determined
		   [14:8] as above for CL82
		   [15] 1 => BER measurement window set to 512 blocks, 0 => IEEE definition */
		decode_control pcs_reg
		deskew_windows pcs_reg

		/* Clause 91 = Reed-Solomon FEC for 100GBASE-R
		   [5] alignment mark spacing 0 => 16k 1 => 1k
		   [4] disable error correction independent of data path
		   [3] disable error code word marking
		   [2] disable reed-solomon error correction (not coding)
		   [1] check for symbol errors over (1 => 128 0 => 8k) code word window
		   [0] fec enable in rx pipeline. */
		cl91_config pcs_reg

		/* When symbol error count in a window of 8k or 128 code words exceeds threshold sync headers are seen as corrupt. */
		cl91_symbol_error_threshold pcs_reg
		cl91_symbol_error_timer     pcs_reg /* units of 15us ticks */
		_                           [0x9230 - 0x9226]pad_reg

		forward_error_correction_mem_ecc_status [4]pcs_reg /* lanes 0-3 */
		deskew_mem_ecc_status                   [4]pcs_reg /* lanes 0-3 */

		/* 4 lanes, deskew mem, forward-error-correction mem */
		interrupt_status [2]pcs_reg // 1 bit/2-bit errors
		interrupt_enable [2]pcs_reg
		ecc_disable      pcs_reg
		ecc_error_inject pcs_reg
	}

	_ [0x9240 - 0x923e]pad_reg

	an_x1 struct {
		oui pcs_reg_32

		/* 2 bits for each serdes speed. */
		priority_remap [5]pcs_reg
		_              [0x9250 - 0x9247]pad_reg

		/* Clause 73: auto negotiation for backplane. */
		cl73 struct {
			break_link_timer                    pcs_reg
			auto_negotiation_error_timer        pcs_reg
			parallel_detect_dme_lock_timer      pcs_reg
			parallel_detect_signal_detect_timer pcs_reg

			// Period in ticks to ignore link while CL73 and/or CL72 are running.
			ignore_link_timer pcs_reg

			qualify_link_timer_yes_cl72_training pcs_reg
			qualify_link_timer_no_cl72_training  pcs_reg
			page_timers                          pcs_reg
		}
	}

	_ [0x9260 - 0x9258]pad_reg

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
			speed [4]pcs_reg

			credit_clock_count          [2]pcs_reg
			credit_loop_count_01        pcs_reg
			credit_mac_generation_count pcs_reg

			_ [0x10 - 0x08]pad_reg
		}
	}

	rx_x1a struct {
		// [9:3] FEC alignment FSM latched state bits (clear on read)
		// [2:0] alignment FSM current state.
		forward_error_correction_alignment_status pcs_reg

		/* [10] ability to bypass error indication
		   [9] ability to perform error detection w/o correction
		   [8] latched hi version of (*)
		   [7] latched lo version of (*)
		   [6] (*) live symbol error over threshold.
		       set when number of RS FEC symbol errors in a window of 8k or 128 exceeds threshold.
		   [5] latched hi alignment status
		   [4] latched lo alignment status
		   [3] live alignment (deskew) status
		   [2] latched hi restart lock
		   [1] latched lo restart lock
		   [0] live restart lock. */
		cl91_status pcs_reg

		/* Single instance of each counter. */
		n_corrected_symbols   pcs_reg_32
		n_uncorrected_symbols pcs_reg_32
		n_corrected_bits      pcs_reg_32
	}

	_ [0xa000 - 0x92b8]pad_reg

	tx_x2 struct {
		mld_swap_count pcs_reg
		_              [0xa002 - 0xa001]pad_reg
		cl82_control   pcs_reg
		_              [0xa011 - 0xa003]pad_reg
		cl82_status    [2]pcs_reg
	}

	_ [0xa023 - 0xa013]pad_reg

	rx_x2 struct {
		misc_control [2]pcs_reg
	}

	_ [0xa080 - 0xa025]pad_reg

	rx_cl82 struct {
		live_deskew_decoder_status    pcs_reg
		latched_deskew_decoder_status pcs_reg
		ber_count                     [2]pcs_reg // lo 8 bits + hi 14 bits
		errored_block_count           pcs_reg
	}

	_ [0xc010 - 0xa085]pad_reg

	pmd_x4 pmd_x4_common

	_ [0xc050 - 0xc01a]pad_reg

	speed_change_x4 speed_change_x4_common

	_ [0xc070 - 0xc062]pad_reg

	speed_change_x4_config struct {
		final_speed_config  pcs_lane_reg /* see sc_speed_override + 0-1 */
		_                   [0xc072 - 0xc071]pad_reg
		final_speed_config1 [6]pcs_reg
		final_speed_fec     pcs_reg
	}

	_ [0xc100 - 0xc079]pad_reg

	tx_x4 struct {
		/* [14] enable
		   [13:0] clock count */
		mac_credit_clock_count  [2]pcs_lane_reg
		mac_credit_loop_count01 pcs_lane_reg
		mac_credit_gen_count    pcs_lane_reg

		_ [0xc111 - 0xc104]pad_reg

		// [12] enable hg2 extensions
		// [11] hg2 invalid message code support
		// [10] hg2 enable
		// [9] bypass cl49 tx state machine
		// [6:5] 1 => force encoder output to local faults
		//       2 => force output to idles
		// [1:0] encode mode (undocumented).
		//   0 => normal, 1 => xfi, 2 => mld (? guessing from sdk)
		encode_control pcs_lane_reg

		_ [0xc113 - 0xc112]pad_reg

		/* [15:13] scrambler mode (0 => bypass scrambler, 1 64b66b (all 66), 2 => 8b10b 3 => 64b66b (sync bits not scrambled)
		   [12:11] # PCS bit lanes bit muxed
		   [10] tx fec enable
		   [9] 0 => 66 bit, 1 => 40 bit write in t_pma
		   [8] link interrupt passed thru to the RS layer.
		   [7] link fault from RS layer to PCS
		   [6] remote fault from RS layer to PCS
		   [5:2] os mode
		   [1] active low reset for txp lanes
		   [0] per lane enable to allow DVs from MAC to enter TXP */
		control pcs_lane_reg

		/* [2] cl36 tx catch_all disable
		   [1] lpi (lpi = low power idle) enable
		   [0] cl36 tx pipeline enable. */
		cl36_tx_control pcs_lane_reg

		_ [0xc120 - 0xc115]pad_reg

		encode_status      [2]pcs_lane_reg
		pcs_status_live    pcs_lane_reg
		pcs_status_latched pcs_lane_reg

		/* per lane [1] underflow [0] overflow */
		pma_underflow_overflow_status pcs_lane_reg
	}

	_ [0xc130 - 0xc125]pad_reg

	rx_x4 struct {
		/* 15:14 descr mode
		   13:12 dec tl mode
		   10:8 deskew mode
		   7:6 dec fsm mode
		   1 cl74 fec enable (clause 74 = FEC for 10GBASE-R)
		   0 lpi enable */
		pcs_control pcs_lane_reg

		_ [0xc134 - 0xc131]pad_reg

		/* [15] bypass cl49 rx state machine
		   [14] test mode enable for cl49 and cl82
		   [13] hg2 invalid message code support
		   [12] hg2 pcs support enable
		   [9] hg2 codec enable
		   [8] disable cl49 ber monitor
		*/
		decoder_control   pcs_lane_reg
		block_sync_config pcs_lane_reg
		_                 [0xc137 - 0xc136]pad_reg

		// [0] Active low per lane reset for rxp
		pma_control pcs_lane_reg

		_ [0xc139 - 0xc138]pad_reg

		// 1'b1 - If the link status transitions from UP (1) to DOWN
		// (0), this bit maintains the DOWN (0) value of the link
		// status until the SW clears this bit.  1'b0 (default value) - The link
		// status information is passed directly from the PCS to the
		// MAC and status registers without modification
		link_status_control pcs_lane_reg

		deskew_memory_control           pcs_lane_reg
		fec_memory_control              pcs_lane_reg
		cl36_control                    pcs_lane_reg
		synce_fractional_divisor_config pcs_lane_reg

		_ [0xc140 - 0xc13e]pad_reg

		// [1]
		//   gap counting mode: 0 => count zeros, 1 => count gaps
		//
		fec_control [4]pcs_lane_reg

		_ [0xc150 - 0xc144]pad_reg

		/* [5] pmd lock
		   [4:0] block lock status per pseudo-logical lane status */
		block_sync_status       pcs_lane_reg
		block_sync_sm_debug     [3]pcs_lane_reg
		block_lock_latch_status pcs_lane_reg
		am_lock_latch_status    pcs_lane_reg

		/* [4:0] alignment_mark lock pseudo-logical lanes 0-4
		   [9:5] block lock pseudo-logical lanes 0-4 */
		am_lock_live_status                         pcs_lane_reg
		cl82_bip_error_count                        [3]pcs_lane_reg /* 0-2 for pseudo-logical lanes 0-4 */
		pseudo_logical_lane_to_virtual_lane_mapping [2]pcs_lane_reg /* 0-1 5 bits for pseudo-logical langes 0-4 */
		pseudo_random_test_pattern_errors           pcs_lane_reg
		pseudo_random_test_pattern_is_locked        pcs_lane_reg

		_ [0xc160 - 0xc15e]pad_reg

		/* [9] local fault lo->hi since last read
		   [8] remote fault lo->hi since last read
		   [7] link interrupt lo->hi since last read
		   [6] lpi received lo->hi since last read
		   [5] hi BER lo->hi since last read
		   [4] hi BER hi->lo since last read
		   [3] link status lo->hi since last read
		   [2] link status hi->lo since last read
		   [1] deskew latched hi since last read
		   [0] deskew latched lo since last read */
		pcs_latched_status pcs_lane_reg

		/* [6] local fault
		   [5] remote fault
		   [4] link interrupt
		   [3] lpi received
		   [2] hi BER
		   [1] link status
		   [0] deskew achieved */
		pcs_live_status pcs_lane_reg

		/* [0] BER monitor current FSM state. */
		/* [1] Clause 49 = 10GBASE-10KR 64/66B encoding.
		   [12:5] cl49 rx fsm latched state
		   [4:0] BER monitor fsm latched state */
		decoder_status [4]pcs_lane_reg

		/* Per lane counters. */
		cl91_per_lane_n_corrected_symbols [2]pcs_lane_reg

		/* sync acquisition state machine next state, current state */
		cl36_sync_acquisition_next_state pcs_lane_reg
		cl36_sync_acquisition_state      pcs_lane_reg /* clear on read bitmap */

		/* Clear on read symbol error count (saturates 0xff). */
		cl36_ber_count pcs_lane_reg

		_ [0xc170 - 0xc16b]pad_reg

		cl82_am_latched_status [5]pcs_lane_reg /* 5 regs for pseudo-logical lanes 0-4 */
		cl82_am_live_status    [5]pcs_lane_reg /* 5 regs for pseudo-logical lanes 0-4 */

		/* [5] fec lane [4:3] valid
		   [4:3] fec lane
		   [2] latched hi AMPS lock live
		   [1] latched lo AMPS lock live
		   [0] AMPS lock live detected
		   AMPS = alignment marker payload sequence. */
		cl91_sync_status pcs_lane_reg

		/* fec sync fsm state
		   [5:2] latched state bitmap
		   [1:0] current state */
		cl91_sync_fsm_state pcs_lane_reg

		_ [0xc180 - 0xc17c]pad_reg

		fec_debug                 [2][5]pcs_lane_reg /* streams 0-4, lo/hi */
		rx_fec_burst_error_lo     [5]pcs_lane_reg
		_                         [0xc190 - 0xc18f]pad_reg
		rx_fec_burst_error_hi     [5]pcs_lane_reg
		rx_fec_corrected_blocks   [2][5]pcs_lane_reg
		_                         [0xc1A0 - 0xc19f]pad_reg
		rx_fec_uncorrected_blocks [2][5]pcs_lane_reg
	}

	_ [0xc1b0 - 0xc1aa]pad_reg

	test1 struct {
		tx_packet_count [2]pcs_reg
		rx_packet_count [2]pcs_reg
	}

	_ [0xc1c0 - 0xc1b4]pad_reg

	an_x4 struct {
		/* [13] disable remote fault
		   [12:11] log2 number of lanes (software must set before starting AN)
		   [10] Broadcom autoneg mode (BAM) enable
		   [9] HP autoneg mode (HPAM) enable
		   [8] CL73 autoneg mode enable
		   [6] cl73 nonce match override
		   [5] cl73 nonce match value
		   [3] BAM -> HPAM enable
		   [2] HPAM -> CL73 enable
		   [0] AN restart (zero to one transition starts AN). */
		cl73_auto_negotiation_control pcs_lane_reg

		/* [0] [15]BAM_HG2          Hi Gig Mode
		       [9] BAM_50G_CR4      UP1 Page Bit 33
		       [8] BAM_50G_KR4      UP1 Page Bit 32
		       [7] BAM_50G_CR2      UP1 Page Bit 25
		       [6] BAM_50G_KR2      UP1 Page Bit 24
		       [3] BAM_40G_CR2      UP1 Page Bit 23
		       [2] BAM_40G_KR2      UP1 Page Bit 22
		       [1] BAM_20G_CR2      UP1 Page Bit 17
		       [0] BAM_20G_KR2      UP1 Page Bit 16
		   [1] [4] BAM_25G_CR1      UP1 Page Bit 21
		       [3] BAM_25G_KR1      UP1 Page Bit 20
		       [2] BAM_20G_CR1      UP1 Page Bit 19
		       [1] BAM_20G_KR1      UP1 Page Bit 18
		*/
		cl73_auto_negotiation_local_up1_abilities [2]pcs_lane_reg /* 0-1 */

		/*
			[0]
			[9-5] TX_NONCE       First CL73 nonce to be transmitted
			[4-0] CL73_BASE_SELECTOR IEEE Annex 28A Message Selector
			[1]
			[11] CL73_REMOTE_FAULT 0-no remote fault field 1-yes
			[10] NEXT_PAGE       Next page ability
			[9-8] FEC_REQ  Forward Error Correction request
			[7-6] CL73_PAUSE Pause Ability
			[5] BASE_1G_KX1      Base Page Bit A0
			[4] BASE_100G_CR4    Base Page Bit A8
			[3] BASE_100G_KR4    Base Page Bit A7
			[2] BASE_40G_CR4     Base Page Bit A4
			[1] BASE_40G_KR4     Base page Bit A3
			[0] BASE_10G_KR1     Base Page Bit A2
		*/
		cl73_auto_negotiation_local_base_abilities [2]pcs_lane_reg /* 0-1 */

		/*
		   [8-0] CL73_BAM_CODE BAM code
		*/
		cl73_auto_negotiation_local_bam_abilities pcs_lane_reg

		/* [15] advertise OUI in cl73 BAM
		   [14] require OUI in cl73 BAM to detect
		   [13] advertise OUI in cl73 HPAM
		   [12] require OUI in cl73 HPAM
		   [9:6] number of retries before declaring failure
		   [5] wait for link fail inhibit timer to expire
		   [4] ignore link fail inhibit timer
		   [3] trap sequencer in GOOD_CHECK state
		   [2] trap sequencer in GOOD state
		   [1] enable 1g kx parallel detect */
		cl73_auto_negotiation_misc_control pcs_lane_reg

		_ [0xc1d0 - 0xc1c7]pad_reg

		cl73_r_status    pcs_lane_reg
		cl73_pxng_status pcs_lane_reg

		/*[10] CL73_AN_COMPLETE PSEQ CL73 auto-neg is complete.
		  [9]  RX_NP_TOGGLE_ERR PSEQ Received NP without T toggling.
		  [8]  RX_INVALID_SEQ   PSEQ Received invalid page sequence.
		  [7]  RX_UP_OUI_MATCH  PSEQ Received MPS-5 OUI match.
		  [6]  RX_UP_OUI_MISMATCH PSEQ Received MPS-5 OUI mismatch.
		  [5]  RX_MP_MISMATCH   PSEQ Received mismatching message page.
		  [4]  RX_MP_OUI        PSEQ Received message page 5.
		  [3]  RX_MP_NULL       PSEQ Received message page 1.
		  [2]  RX_NP            PSEQ Received next page.
		  [1]  RX_BP            PSEQ Received base page.
		  [0]  HP_MODE          PSEQ Entered Hewlett-Packard mode.*/
		cl73_pseq_status pcs_lane_reg

		cl73_pseq_remote_fault_status pcs_lane_reg
		cl73_unexpected_page          pcs_lane_reg
		cl73_pseq_base_pages          [3]pcs_lane_reg /* 3 regs pages 1-3 */
		cl73_pseq_link_partner_oui    [5]pcs_lane_reg /* 5 regs */
		cl73_resolution_error         pcs_lane_reg

		_ [0xc1e0 - 0xc1de]pad_reg

		cl73_local_device_sw_control_pages       [3]pcs_lane_reg /* pages 2 1 0 */
		cl73_link_partner_sw_control_pages       [3]pcs_lane_reg /* pages 2 1 0 */
		cl73_sw_status                           pcs_lane_reg
		cl73_local_device_control                pcs_lane_reg
		cl73_auto_negotiation_ability_resolution pcs_lane_reg

		/* [15] complete
		   [14:9] retry count clear on read
		   [8] force speed
		   [7] parallel detect complete
		   [6] is active
		   [5:2] fail count
		   [1] parallel detect in progress */
		cl73_auto_negotiation_misc_status pcs_lane_reg

		cl73_tla_sequencer_status pcs_lane_reg
	}

	_ [0xc330 - 0xc1eb]pad_reg

	interlaken interlaken_common

	_ [0xd000 - 0xc341]pad_reg

	/* DSC = digital signal conditioning. */
	dsc_afe3 struct {
		/* [14:11] main peak filter
		   [10:8] low frequency peak filter */
		rx_peak_filter_control pmd_lane_reg

		/* [13:8] data negative
		   [5:0] data positive. */
		rx_slicer struct {
			a_offset_adjust_data  pmd_lane_reg
			a_offset_adjust_phase pmd_lane_reg
			ab_offset_adjust_lms  pmd_lane_reg
			b_offset_adjust_data  pmd_lane_reg
			b_offset_adjust_phase pmd_lane_reg
			c_offset_adjust_data  pmd_lane_reg
			c_offset_adjust_phase pmd_lane_reg
			cd_offset_adjust_lms  pmd_lane_reg
			d_offset_adjust_data  pmd_lane_reg
			d_offset_adjust_phase pmd_lane_reg
		}
		rx_phase_lms_threshold pmd_lane_reg

		_ [0xd010 - 0xd00c]pad_reg

		rx_dfe_tap2_abcd        [2]pmd_lane_reg
		rx_dfe_tap3_abcd        [2]pmd_lane_reg
		rx_dfe_tap4_9_abcd      [6]pmd_lane_reg
		_                       [0xd020 - 0xd01a]pad_reg
		rx_dfe_tap10_14_abcd    [5]pmd_lane_reg
		rx_dfe_tap7_14_mux_abcd [4]pmd_lane_reg /* taps 7-14 2 x 4 x 2 bit mux */
		load_presets            pmd_lane_reg
	}

	_ [0xd03d - 0xd02a]pad_reg

	uc_cmd uc_cmd_regs

	_ [0xd040 - 0xd03f]pad_reg

	dsc_b struct {
		/* [7:0] lo [23:8] hi x streams abcd */
		training_sum_interleave_abcd [4][2]pmd_lane_reg

		/* [9:0] lo [25:10] hi */
		training_sum_result_abcd [2]pmd_lane_reg
		_                        [0xd04c - 0xd04a]pad_reg
		dc_offset                pmd_lane_reg

		/* [15:8] vga gain
		   [7:0] data threshold. */
		vga_status pmd_lane_reg
	}

	_ [0xd050 - 0xd04e]pad_reg

	dsc_c struct {
		cdr_control [3]pmd_lane_reg
		pi_control  pmd_lane_reg

		_ [0xd055 - 0xd054]pad_reg

		training_sum_control         pmd_lane_reg
		training_sum_pattern_control [2]pmd_lane_reg
		training_sum_tap_control     pmd_lane_reg
		training_sum_tdt_control     pmd_lane_reg
		training_sum_misc            pmd_lane_reg

		_ [0xd05c - 0xd05b]pad_reg

		vga_control            pmd_lane_reg
		data_slicer_th_control pmd_lane_reg
		dc_offset_control      pmd_lane_reg
	}

	_ [0xd060 - 0xd05f]pad_reg

	dsc_d struct {
		/* [0]
		   [15] set measure imcomplete & force new measurement in EEE mode.
		   [14] enable CONFIG -> WAIT_FOR_SIGNAL transition
		   [13] enable RESTART -> CONFIG transition
		   [11] enable EEE_DONE -> DONE transition
		   [8] enable measurement during EEE_MEASURE state
		   [6] hw_tune_en
		   [5] uc_tune_en
		   [4] timer enable
		   [3] ignore rx mode
		   [2] enable rx afe powerdown in eee quiet mode
		   [1] eee mode enable. */
		/* [9]
		   [0] restart DSC state machine in restart state (self-clearing)
		  [1] set DSC state machine in restart state and hold there later until set to 0. */
		state_machine struct {
			control [10]pmd_lane_reg

			// [0] rx dsc locked
			// [1] eee measurement incomplete
			// [15:7] eee measurement
			lock_status pmd_lane_reg

			// bitmap of fsm states entered since last read.  clear on read.
			// cdr = clock and data recovery.
			// [0] reset, [1] restart, [2] config, [3] wait_for_signal,
			// [4] acquire_cdr, [5] cdr_settle
			// [6] hw_tune, [7] uc_tune, [8] measure, [9] done */
			status_one_hot pmd_lane_reg

			/* quiet, ana_power, acquire_cdr, cdr_settle, hw_tune, measure, done */
			status_eee_one_hot pmd_lane_reg

			restart_status pmd_lane_reg

			/* [15:11] live DSC state
			   [10:5] uc request
			   [4] uc ready for request
			   [3:0] dsc state machine scratch */
			status pmd_lane_reg
		}
	}

	_ [0xd070 - 0xd06f]pad_reg

	dsc_e struct {
		rx_phase_slicer_counter pmd_lane_reg
		rx_lms_slicer_counter   pmd_lane_reg
		/* bits [15:0] [35:20] of 40 bit data */
		rx_data                  [2]pmd_lane_reg
		cdr_phase_error_status   pmd_lane_reg
		rx_data_slicer_counter   pmd_lane_reg
		rx_phase_slicer_counter1 pmd_lane_reg
		rx_lms_slicer_counter1   pmd_lane_reg
		cdr_integration          pmd_lane_reg
		cdr_misc_status          pmd_lane_reg
		cdr_1g_status            pmd_lane_reg
		_                        [0xd07e - 0xd07b]pad_reg
		preset                   pmd_lane_reg
	}

	_ [0xd080 - 0xd07f]pad_reg

	cl93n72_rx struct {
		/* [0]
		[2] 1 => rx training complete 0 => in progress
		   [1] coarse lock to recovered clock has occurred
		   [0] rx training enable. */
		/* [1] [6:4] number of bad alignment markers to lose frame lock
		   [1:0] number of good alignment markers to declare frame lock. */
		/* [2] [9] frame consistency check enable (3 consecutive frames from link partner with same coef status/update
		       before setting latched status bits).
		   [8] rx data path lane clock enable
		   [7] ppm offset enable
		   [6] strict (e.g. standard) marker check
		   [5] check for standard specified DME
		   [4] check for DME cell boundary transitions
		   [3:0] control frame delay 0 => disable, -7 early - +7 late */
		control [3]pmd_lane_reg

		/* [0] rx frame locked. */
		status pmd_lane_reg

		/* [2] rx frame lock change micro interrupt enable
		   [1] rx status field change interrupt enable (local device request TX FIR coef change, link partner sends response)
		   [0] rx link partner requests update tx fir coef interrupt enable. */
		micro_interrupt_enable0 pmd_lane_reg

		/* [0] link partner update. */
		micro_interrupt_status0 pmd_lane_reg

		micro_status1 pmd_lane_reg
	}

	_ [0xd090 - 0xd087]pad_reg

	cl93n72_tx struct {
		/* Last update/status sent to link partner. */
		local_update_to_link_partner pmd_lane_reg
		local_status_to_link_partner pmd_lane_reg

		/* [0] [8] PRBS seed random (1 random, 0 per IEEE spec)
		   [7] seed order S10 -> S0 else S0 -> S10
		   [6] 1 by lane, 0 random seed every frame
		   [5:4] PRBS lane select
		   [2] remote is trained; rx ready.  set by training fsm when link partner rx ready bit is set.
		   [1] training fsm frame lock achieved
		   [0] remote & local tx equalizers have been trained; normal data transmission may commence. */

		/* [1] [2] tx data path lane clock enable
		   [1] disable max wait timer
		   [0] PRBS ring oscillator disable */

		/* [2] [13:8] TX FIR post (+1) cursor tap coef. value
		   [4:0] TX FIR pre (-1) cursor tap coef. value */

		/* [3] [6:0] TX FIR main (+0) cursor tap coef. value. */
		control [4]pmd_lane_reg

		/* [1] training fsm sigal detect: 1 => send data state, 0 => training state.
		   [0] 1 => remote tx and local equalizers have been optimized & normal data transmission may commence. */
		status pmd_lane_reg
	}

	_ [0xd0a0 - 0xd097]pad_reg

	tx_phase_interpolator struct {
		/* [0] phase interpolator enable
		   [3] phase interpolator frequency override enable */
		control pmd_lane_reg

		/* [15:0] -8k to 8k.  ppm = (790 * v / 8192) */
		frequecy_override pmd_lane_reg
		jitter_control    pmd_lane_reg
		control3          pmd_lane_reg
		control4          pmd_lane_reg
		_                 [0xd0a8 - 0xd0a5]pad_reg
		status            [4]pmd_lane_reg
	}

	_ [0xd0b0 - 0xd0ac]pad_reg

	clock_and_reset clock_and_reset_common

	_ [0xd0c0 - 0xd0bf]pad_reg

	ams_rx struct {
		control [10]pmd_lane_reg
		_       [0xd0cb - 0xd0ca]pad_reg
		status  pmd_lane_reg
	}

	_ [0xd0d0 - 0xd0cc]pad_reg

	ams_tx struct {
		/* [2] [15] force electrical idle
		   [14:13] tx driver mode
		   [12] sign for 2nd post cursor tap
		   [11:8] 2nd post cursor coefficent
		   [7] sign for 3rd post cursor tap
		   [6:4] 3rd post cursor coefficent
		   [3:0] master amplitude control. */
		control [4]pmd_lane_reg

		_ [0xd0d8 - 0xd0d4]pad_reg

		status pmd_lane_reg
	}

	_ [0xd0e0 - 0xd0d9]pad_reg

	sigdet sigdet_common

	_ [0xd100 - 0xd0e9]pad_reg

	dig dig_common

	_ [0xd110 - 0xd10f]pad_reg

	ams_pll struct {
		// [2] [0] 0 => Rz 2.4K, 1 => 4.8K
		//     [4:1] charge pump current (value+1)*50 uAmps (50, 100, 150, ... 800 uAmps) */
		control [8]pmd_lane_reg
		_       [0xd119 - 0xd118]pad_reg
		status  pmd_lane_reg
	}

	_ [0xd120 - 0xd11a]pad_reg

	tx_pattern tx_pattern_common

	_ [0xd130 - 0xd12f]pad_reg

	tx_equalizer struct {
		/* [0] [11:8] TX FIR post (+1) tap offset (-8 to 7 2s complement)
		   [7:4]  TX FIR main (+0) tap offset
		   [3:0]  TX FIR pre  (-1) tap offset */
		/* [1] [11:8] TX FIR post2 (+2) tap offset
		   [4:0]  TX FIR post2 (+2) tap value (-16 to 15 2s complement) */
		/* [2] [11:8] TX FIR post3 (+3) tap offset
		   [3:0]  TX FIR post3 (+3) tap value */
		control [3]pmd_lane_reg

		/* [0] [13:8] TX FIR post tap (+1) value after override mux
		   [4:0]  TX FIR pre  tap (-1) value after override mux */
		/* [1] [6:0]  TX FIR main tap (+0) value after override mux */
		/* [2] [13:8] TX FIR post tap (+1) value after offset adjustment
		   [4:0]  TX FIR pre  tap (-1) value after offset adjustment */
		/* [3] [12:8] TX FIR post2 tap (+2) value after offset adjustment
		   [6:0]  TX FIR main tap (+0) value after offset adjustment */
		/* [4] [3:0]  TX FIR post3 tap (+3) value after offset adjustment */
		status [5]pmd_lane_reg

		// [0] uc tx disable.  Used by uc for tx disable control during cl72 forced mode.
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
	}

	_ [0xd140 - 0xd13a]pad_reg

	pll pll_common

	_ [0xd150 - 0xd14b]pad_reg

	tx_common struct {
		control [4]pmd_lane_reg
	}

	_ [0xd160 - 0xd154]pad_reg

	tlb_rx struct {
		tlb_rx_common
		pseudo_random_bitstream_burst_error_live_length pmd_lane_reg
		pseudo_random_bitstream_burst_error_max_length  pmd_lane_reg
	}

	_ [0xd170 - 0xd16f]pad_reg

	tlb_tx struct {
		tlb_tx_common

		_ [0xd178 - 0xd174]pad_reg

		remote_loopback_status pmd_lane_reg
	}

	_ [0xd200 - 0xd179]pad_reg

	uc struct {
		// [0] master clock enable
		// [1] core clock enable
		clock_control pmd_lane_reg

		// [0] active low master reset
		// [1] active low core reset
		// [3] active low program ram reset.
		reset_control pmd_lane_reg

		/* AHB-lite is micro controller (ARM bus).
		   [13] auto increment read  address enable
		   [12] auto increment write address enable
		   [9] zero data ram
		   [8] zero code ram
		   [5:4] log2 n bytes read data size
		   [1:0] log2 n bytes write data size. */
		ahb_control pmd_lane_reg

		/* [0] code/data init done. */
		ahb_status pmd_lane_reg

		write_address                pmd_lane_reg_32
		write_data                   pmd_lane_reg_32 // transaction when lo bits (first 16 bit register) are written
		read_address                 pmd_lane_reg_32 // transaction when lo bits (first 16 bit register) are written
		read_data                    pmd_lane_reg_32
		program_ram_interface_enable pmd_lane_reg
		program_ram_write_address    pmd_lane_reg_32
		_                            [0xd210 - 0xd20f]pad_reg
		temperature_status           pmd_lane_reg

		tx_mailbox                                        pmd_lane_reg_32
		rx_mailbox                                        pmd_lane_reg_32
		mailbox_control                                   pmd_lane_reg
		ahb_control1                                      pmd_lane_reg
		ahb_status1                                       pmd_lane_reg
		ahb_next_auto_increment_write_address             pmd_lane_reg
		ahb_next_auto_increment_read_address              pmd_lane_reg
		ahb_next_auto_increment_program_ram_write_address pmd_lane_reg
		temperature_control                               pmd_lane_reg
		_                                                 [0xd220 - 0xd21c]pad_reg

		program_ram_ecc_control       [2]pmd_lane_reg
		program_ram_ecc_error_address pmd_lane_reg
		program_ram_ecc_error_data    pmd_lane_reg
		program_ram_test_control      pmd_lane_reg

		// [0] enable
		// [13:8] # of kbytes of data ram versus code ram (e.g. 4 => 4k of data ram, 32k of code ram) ram is 36k total.
		// Default value: 0x401
		ram_config pmd_lane_reg

		// [0] mailbox message out, [1] single bit ecc error, [2] multi bit ecc error,
		// etc.
		interrupt_enable pmd_lane_reg
		interrupt_status pmd_lane_reg
	}

	_ [0xffdb - 0xd228]pad_reg

	mdio mdio_common

	_ [0xffff - 0xffe0]pad_reg
}

type tscf_over_sampling_divider int

const (
	tscf_over_sampling_divider_1      tscf_over_sampling_divider = 0
	tscf_over_sampling_divider_2      tscf_over_sampling_divider = 1
	tscf_over_sampling_divider_4      tscf_over_sampling_divider = 2
	tscf_over_sampling_divider_16_5   tscf_over_sampling_divider = 8  // divide by 16.5
	tscf_over_sampling_divider_20_625 tscf_over_sampling_divider = 12 // divide by 20.625
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
	// Low bits specify interface (cr, kr, serdes sfi/xfi/mld) and higig2
	tscf_speed_cr     tscf_speed = 0 << 0 // copper
	tscf_speed_kr     tscf_speed = 1 << 0 // backplane
	tscf_speed_optics tscf_speed = 2 << 0 // SFI/XFI for 1 lane, MLD for multilane connections to optics
	tscf_speed_hg2    tscf_speed = 1 << 2

	// xN means N lanes
	tscf_speed_10g_x1       tscf_speed = 0 << 3
	tscf_speed_20g_x1       tscf_speed = 1 << 3
	tscf_speed_25g_x1       tscf_speed = 2 << 3
	tscf_speed_20g_x2       tscf_speed = 3 << 3
	tscf_speed_40g_x2       tscf_speed = 4 << 3
	tscf_speed_40g_x4       tscf_speed = 5 << 3
	tscf_speed_50g_x2       tscf_speed = 6 << 3
	tscf_speed_50g_x4       tscf_speed = 7 << 3
	tscf_speed_100g_x4      tscf_speed = 8 << 3
	tscf_speed_cl73_20g_vco tscf_speed = 9 << 3 // 1g clause 73 auto-negotiation
	tscf_speed_cl73_25g_vco tscf_speed = 10 << 3
	tscf_speed_cl36_20g_vco tscf_speed = 11 << 3 // 1g clause 36
	tscf_speed_cl36_25g_vco tscf_speed = 12 << 3
)

func (x tscf_speed) String() (s string) {
	// invalid speed
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
