// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/firmware/fe1a"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/port"

	"fmt"
	"time"
)

type HundredGig struct {
	Common
	core_config  uc_core_config_word
	lane_configs [4]uc_lane_config_word
}

func (r *hundred_gig_controller) set_master_port_number(q *DmaRequest, n uint) {
	if n >= N_lane {
		panic(n)
	}
	r.main.setup.Modify(q, uint16(n)<<14, 3<<14)
}

func (phy *HundredGig) apply_lane_map(q *DmaRequest, r *hundred_gig_controller) {
	// Inverse logical -> physical mappings.
	var (
		rx_phys_by_logical, tx_phys_by_logical [N_lane]uint16
	)

	for p := range phy.Rx_logical_lane_by_phys_lane {
		rx_phys_by_logical[phy.Rx_logical_lane_by_phys_lane[p]] = uint16(p)
		tx_phys_by_logical[phy.Tx_logical_lane_by_phys_lane[p]] = uint16(p)
	}

	// Rx side: only PCS swap.
	// Tx side: first PCS swap then PMD swap.
	// PCS swap.
	rxm := uint16(0)
	for l := range rx_phys_by_logical {
		rxm |= rx_phys_by_logical[l] << uint(2*l)
	}
	r.main.rx_lane_swap.Modify(q, rxm<<0, 0xff<<0)

	var pmd_map [N_lane]uint16
	for p := range tx_phys_by_logical {
		pmd_map[p] = uint16(tx_phys_by_logical[phy.Rx_logical_lane_by_phys_lane[p]])
	}

	var pmd_tx [N_lane]uint16
	for p := range pmd_tx {
		pmd_tx[pmd_map[p]] = uint16(p)
	}

	laneMask := m.LaneMask(1 << 0)
	r.dig.tx_lane_map_012.Set(q, laneMask,
		pmd_tx[0]<<(5*0)|pmd_tx[1]<<(5*1)|pmd_tx[2]<<(5*2))
	r.dig.tx_lane_map_3_lane_address_01.Set(q, laneMask,
		pmd_tx[3]<<(5*0)|
			uint16(phy.Rx_logical_lane_by_phys_lane[0])<<(5*1)|
			uint16(phy.Rx_logical_lane_by_phys_lane[1])<<(5*2))
	r.dig.tx_lane_address_23.Set(q, laneMask,
		uint16(phy.Rx_logical_lane_by_phys_lane[2])<<0|
			uint16(phy.Rx_logical_lane_by_phys_lane[3])<<8)
	q.Do()
}

func (phy *HundredGig) setPortMode(mode port.PortBlockMode) {
	r := get_hundred_gig_controller()
	q := phy.dmaReq()

	const (
		m4x1     = iota // 4 x 25G
		m2x1_1x2        // 2 x 25G + 1 x 50G
		m1x2_2x1        // 1 x 50G + 2 x 25G
		m2x2            // 2 x 50G
		m1x4            // 1 x 100G
	)
	v := uint16(0)
	switch mode {
	case port.PortBlockMode1x4:
		v = m1x4
	case port.PortBlockMode2x2:
		v = m2x2
	case port.PortBlockMode2x1_1x2:
		v = m2x1_1x2
	case port.PortBlockMode1x2_2x1:
		v = m1x2_2x1
	case port.PortBlockMode4x1:
		v = m4x1
	default:
		panic("port mode")
	}

	r.main.setup.Modify(q, v<<4, 7<<4)
	q.Do()
}

func (r *pmd_lane_u16) GetDoForeach(q *DmaRequest, laneMask m.LaneMask) uint16 {
	v := [4]uint16{0xffff, 0xffff, 0xffff, 0xffff}
	laneMask.Foreach(func(l m.LaneMask) {
		r.Get(q, m.LaneMask(1<<l), &v[l])
	})
	q.Do()
	return v[0] & v[1] & v[2] & v[3]
}

func (r *pcs_lane_u16) GetDoForeach(q *DmaRequest, laneMask m.LaneMask) uint16 {
	v := [4]uint16{0xffff, 0xffff, 0xffff, 0xffff}
	laneMask.Foreach(func(l m.LaneMask) {
		r.Get(q, m.LaneMask(1<<l), &v[l])
	})
	q.Do()
	return v[0] & v[1] & v[2] & v[3]
}

func (phy *HundredGig) Init() {

	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	laneMask := m.LaneMask(1 << 0)
	allLanes := m.LaneMask(0xf)

	// Take Core and data path out of reset.
	r.pmd_x1.reset.Set(q, 1<<0|1<<1)
	q.Do()

	// uC master clock enable.
	r.uc.clock_control.Modify(q, laneMask, 1<<0, 1<<0)
	q.Do()

	// uC out of master reset.
	r.uc.reset_control.Modify(q, laneMask, 1<<0, 1<<0)
	q.Do()

	// Uc zero/init code ram
	r.uc.ahb_control.Modify(q, laneMask, 1<<8, 1<<8)
	q.Do()

	// Wait for code ram init done.
	start := time.Now()
	for r.uc.ahb_status.GetDo(q, laneMask)&(1<<0) == 0 {
		if time.Since(start) > 100*time.Millisecond {
			panic("tscf wait for code ram init timeout")
		}
		time.Sleep(100 * time.Microsecond)
	}

	// reset code ram init
	//r.uc.ahb_control.Modify(q, laneMask, 0<<8, 1<<8)
	//q.Do()

	// program ram out of reset
	r.uc.program_ram_write_address.Set(q, laneMask, 0)
	q.Do()
	r.uc.reset_control.Modify(q, laneMask, 1<<3, 1<<3)
	q.Do()
	r.uc.program_ram_interface_enable.Modify(q, laneMask, 1<<0, 1<<0)
	q.Do()

	phy.PortBlock.LoadFirmware(&q.DmaRequest, fe1a.Fucode.Data)
	q.Do()
	if q.Err != nil {
		panic(q.Err)
	}

	r.uc.program_ram_interface_enable.Modify(q, laneMask, 0<<0, 1<<0)
	q.Do()

	// Disable pmd_ln_h_rstb input pin.
	r.clock_and_reset.lane_reset_and_powerdown_pin_disable.Modify(q, laneMask, 1<<0, 1<<0)
	q.Do()

	// Set uc active bit.
	r.dig.top_user_control.Modify(q, laneMask, 1<<15, 1<<15)
	q.Do()

	// core clock enable
	r.uc.clock_control.Modify(q, laneMask, 1<<1, 1<<1)
	q.Do()

	// core out of reset
	r.uc.reset_control.Modify(q, laneMask, 1<<1, 1<<1)
	q.Do()

	// Wait for microcode to be ready.
	{
		start := time.Now()
		for r.uc_cmd.control.GetDo(q, laneMask)&(1<<7) == 0 {
			time.Sleep(100 * time.Microsecond)
			if time.Since(start) > 100*time.Millisecond {
				panic("ucode not ready")
			}
		}
	}

	// Have microcode verify crc. Works but is slow so disabled.
	if false {
		crc := uint16(0)
		len := uint16(len(fe1a.Fucode.Data))
		c := uc_cmd{command: uc_cmd_compute_ucode_crc, in: &len, out: &crc}
		err := c.do(q, laneMask, &r.uc_cmd)
		if err != nil {
			panic(err)
		}
		if crc != fe1a.Fucode.Crc {
			panic(fmt.Errorf("uc ucode crc does not match got %x != want %x", crc, fe1a.Fucode.Crc))
		}
	}

	uc_mem := phy.get_uc_mem()

	// Re-enable pmd_ln_h_rstb input pin.
	r.clock_and_reset.lane_reset_and_powerdown_pin_disable.Modify(q, laneMask, 0<<0, 1<<0)
	q.Do()

	// Override default pll charge pump current.
	r.ams_pll.control[2].Modify(q, laneMask, 0x5<<1, 0xf<<1)
	q.Do()

	phy.core_config = uc_core_config_word{
		vco_rate_in_hz: 175 * 156.25e6,
	}

	r.pll.multiplier.Modify(q, laneMask, uint16(tscf_pll_multipler_175)<<0, 0xf<<0)
	q.Do()

	phy.apply_lane_map(q, r)

	// Setup auto-negotiation timers.
	{
		r.an_x1.cl73.ignore_link_timer.Set(q, 0x29a)
		r.an_x1.cl73.break_link_timer.Set(q, 0x10ed)
		r.an_x1.cl73.parallel_detect_dme_lock_timer.Set(q, 0x14d4)
		r.an_x1.cl73.parallel_detect_signal_detect_timer.Set(q, 0xa6a)
		r.an_x1.cl73.qualify_link_timer_no_cl72_training.Set(q, 0x3080)
		r.an_x1.cl73.qualify_link_timer_yes_cl72_training.Set(q, 0x8382)
	}

	r.set_master_port_number(q, 0)

	r.dig.top_user_control.Modify(q, laneMask, 0<<13, 1<<13)
	q.Do()

	// Write ucode config word.
	uc_mem.config_word.Set(q, laneMask, &phy.core_config)

	// Set FEC zero counting mode.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.rx_x4.fec_control[1].Set(q, lm, 0)
	})

	// Core out of reset.
	r.dig.top_user_control.Modify(q, laneMask, 1<<13, 1<<13)
	q.Do()

	// All lanes out of reset.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.pmd_x4.lane_reset_control.Modify(q, lm, 1<<0|1<<1, 1<<0|1<<1)
	})
	q.Do()

	// Data path out of reset.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.clock_and_reset.lane_reset_and_powerdown.Modify(q, lm, 1<<1, 1<<1)
	})
	q.Do()

	// Set rx/tx lane polarity.
	allLanes.Foreach(func(l m.LaneMask) {
		var v [2]uint16
		if int(l) < len(phy.Rx_invert_lane_polarity) && phy.Rx_invert_lane_polarity[l] {
			v[0] = 1
		}
		if int(l) < len(phy.Tx_invert_lane_polarity) && phy.Tx_invert_lane_polarity[l] {
			v[1] = 1
		}
		lm := m.LaneMask(1 << l)
		r.tlb_rx.misc_control.Modify(q, lm, v[0]<<0, 1<<0)
		r.tlb_tx.misc_control.Modify(q, lm, v[1]<<0, 1<<0)
	})
	q.Do()

	// TX FIR coefficient init.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		setupHundredGigTxSettings(q, lm)
	})

	// Initialize uC lane config word to all zeros.
	{
		allLanes.Foreach(func(lane m.LaneMask) {
			lm := m.LaneMask(1 << lane)
			uc_mem.lanes[lane].config_word.Set(q, lm, &phy.lane_configs[lane])
		})
	}

	// Initialize port-mode from configuration
	phy.setPortMode(phy.PortBlock.GetMode())

	// Take all RXP lanes out of reset.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.rx_x4.pma_control.Modify(q, lm, 1<<0, 1<<0)
	})
	q.Do()

	// TXP out of reset + enable.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.tx_x4.control.Modify(q, lm, 1<<0|1<<1, 1<<0|1<<1)
	})
	q.Do()

	if false {
		// default is level 1; level 0 gives not much.
		logLevel := uint8(10)
		allLanes.Foreach(func(l m.LaneMask) {
			lm := m.LaneMask(1 << l)
			uc_mem.lanes[l].event_log_level.Set(q, lm, logLevel)
		})
		uc_mem.event_log_level.Set(q, laneMask, logLevel)
	}
}

func (phy *HundredGig) SetEnable(p m.Porter, enable bool) {
	laneMask := p.GetLaneMask()
	r := get_hundred_gig_controller()
	q := phy.dmaReq()

	disable := uint16(1)
	if enable {
		disable = 0
	}
	laneMask.ForeachMask(func(lm m.LaneMask) {
		r.tx_equalizer.misc_control.Modify(q, lm, disable<<0, 1<<0)
	})
	q.Do()

	// When disabled, force zero for pmd signal detect.
	{
		const (
			force_pmd_signal_detect          = 1 << 7
			force_pmd_signal_detect_value_hi = 1 << 8
			force_pmd_signal_detect_value_lo = 0 << 8
		)
		v, mask := uint16(0), uint16(force_pmd_signal_detect|force_pmd_signal_detect_value_hi)
		if !enable {
			v = force_pmd_signal_detect | force_pmd_signal_detect_value_lo
		}
		laneMask.ForeachMask(func(lm m.LaneMask) {
			r.sigdet.control[1].Modify(q, lm, v, mask)
		})
		q.Do()
	}
}

func (phy *HundredGig) SetAutoneg(port m.Porter, enable bool) {
	r := get_hundred_gig_controller()

	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLane := laneMask.FirstLane()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())
	uc_mem := phy.get_uc_mem()

	co := &phy.core_config
	la := &phy.lane_configs[firstLane]
	if enable {
		co.core_config_from_pcs = true
		la.lane_config_from_pcs = true
		la.autoneg_enable = true
		la.cl72_restart_timeout_enable = false
		//FIXME media_type needs to be set to match media type read from QSFP EEPROM, hard code to cable for now
		la.media_type = 0x1
	} else {
		co.core_config_from_pcs = false
		la.lane_config_from_pcs = false
		la.autoneg_enable = false
		la.cl72_restart_timeout_enable = true
	}

	// Unconditionally shut off speed-change enable bit
	laneMask.ForeachMask(func(lm m.LaneMask) {
		r.speed_change_x4.control.Modify(q, lm, 0<<8, 1<<8)
	})
	q.Do()

	// Set single-port mode when enabling i.e. causes reset of pll when an done
	const (
		single_port_mode_on = 1 << 3
	)
	var setup uint16
	if enable {
		setup |= single_port_mode_on
	} else {
		setup = 0
	}
	r.main.setup.Modify(q, setup, single_port_mode_on)
	q.Do()

	// TODO In the an disable case check to see if any sister lane
	// has An enabled - if so it'll result in a reconfig of
	// core firmware (see tscf.c<tier2>

	// Put core data path into reset.
	r.dig.top_user_control.Modify(q, firstLaneMask, 0<<13, 1<<13)
	q.Do()

	// Update core config
	uc_mem.config_word.Set(q, firstLaneMask, co)
	q.Do()

	// Take core data path out of reset.
	r.dig.top_user_control.Modify(q, firstLaneMask, 1<<13, 1<<13)
	q.Do()

	// Put lane data path into reset.
	laneMask.ForeachMask(func(l m.LaneMask) {
		r.clock_and_reset.lane_reset_and_powerdown.Modify(q, l, 0<<1, 1<<1)
	})
	q.Do()

	// Update firmware structure based on new values per lane
	laneMask.Foreach(func(lane m.LaneMask) {
		lm := m.LaneMask(1 << lane)
		uc_mem.lanes[lane].config_word.Set(q, lm, la)
	})
	q.Do()

	// Unreset lane data path.
	laneMask.ForeachMask(func(l m.LaneMask) {
		r.clock_and_reset.lane_reset_and_powerdown.Modify(q, l, 1<<1, 1<<1)
	})
	q.Do()

	// Set master port number to start lane
	// NB: why is this needed
	r.set_master_port_number(q, firstLane)
	q.Do()

	phy.setLocalAdvert(port)

	phy.setCL72(port, enable)

	phy.setAnEnable(port, enable)
}

func (phy *HundredGig) setCL72(port m.Porter, enable bool) {
	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	const (
		enable_training = 1 << 1
		sc_x4_enabled   = 1 << 8
	)

	if enable {
		laneMask.ForeachMask(func(l m.LaneMask) {
			r.cl93n72_common.control.Modify(q, l,
				enable_training, enable_training)
		})
	} else {
		laneMask.ForeachMask(func(l m.LaneMask) {
			r.cl93n72_common.control.Modify(q, l,
				0, enable_training)
		})
	}
	q.Do()

	sc_x4_control := r.speed_change_x4.control.GetDo(q, firstLaneMask)
	if sc_x4_control&sc_x4_enabled == sc_x4_enabled {
		laneMask.ForeachMask(func(lm m.LaneMask) {
			r.speed_change_x4.control.Modify(q, lm, 0<<8, 1<<8)
		})
		q.Do()
		laneMask.ForeachMask(func(lm m.LaneMask) {
			r.speed_change_x4.control.Modify(q, lm, 1<<8, 1<<8)
		})
		q.Do()
	}
}

// NB Drive from redis db data
func (phy *HundredGig) setLocalAdvert(port m.Porter) {

	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	const (
		b_FEC      = 1 << 8
		b_PAUSE    = 1 << 6
		b_1G_KX1   = 1 << 5
		b_100G_CR4 = 1 << 4
		b_100G_KR4 = 1 << 3
		b_40G_CR4  = 1 << 2
		b_40G_KR4  = 1 << 1
		b_10G_KR1  = 1 << 0
	)

	var an_base uint16
	if laneMask.NLanes() == 1 {
		an_base = b_FEC | b_PAUSE | b_10G_KR1 | b_1G_KX1
	} else {
		an_base = b_FEC | b_PAUSE | b_100G_CR4 | b_40G_CR4 | b_100G_KR4 | b_40G_KR4
	}
	r.an_x4.cl73_auto_negotiation_local_base_abilities[1].Set(q, firstLaneMask, an_base)
	q.Do()

	const (
		b_tx_nonce      = 0x15 << 5
		b_base_page_sel = 1 << 0
		b_retry_count   = 0
	)
	an_base = b_base_page_sel | b_tx_nonce
	r.an_x4.cl73_auto_negotiation_local_base_abilities[0].Set(q, firstLaneMask, an_base)
	q.Do()

	r.an_x4.cl73_auto_negotiation_misc_control.Modify(q, firstLaneMask, b_retry_count,
		b_retry_count)
	q.Do()
}

func (phy *HundredGig) setAnEnable(port m.Porter, enable bool) {
	// Used to signal PLL to restart after AN
	r := get_hundred_gig_controller()

	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	const (
		speed_change_enable = 1 << 8
		an_restart          = 1 << 0
		cl73_an_enable      = 1 << 8
		cl73_enable_1gkx    = 1 << 1
		cl73_nonce_oride    = 1 << 6
		cl73_nonce_match    = 1 << 5
	)

	if enable {
		r.speed_change_x4.control.Modify(q, firstLaneMask, 0<<8, 1<<8)
		q.Do()
	} else {
		r.speed_change_x4.control.Modify(q, firstLaneMask, 1<<8, 1<<8)
		q.Do()
	}

	// Set CL37 HVCO bit
	var setup uint16
	if r.pll.multiplier.GetDo(q, firstLaneMask) == uint16(tscf_pll_multipler_165) {
		setup |= 1 << 1
	} else {
		setup &^= uint16(1) << 1
	}
	r.main.setup.Modify(q, setup, 1<<1)
	q.Do()

	// This stabilizes 4-lane autoneg case
	r.an_x1.cl73.auto_negotiation_error_timer.Set(q, 0x1000)
	q.Do()

	// Turn off an-restart
	r.an_x4.cl73_auto_negotiation_control.Set(q, firstLaneMask, 0)
	q.Do()

	// Turn on cl73 and an_restart and number of lanes if enabled
	var log2NLanes uint16

	nlanes := laneMask.NLanes()
	switch nlanes {
	case 1:
		log2NLanes = 0
	case 2:
		log2NLanes = 1
	case 4:
		log2NLanes = 2
	case 10:
		log2NLanes = 3
	default:
		panic(fmt.Errorf("unkown number of lanes %d", nlanes))
	}

	if enable {
		v := uint16(an_restart | cl73_an_enable | (log2NLanes << 11) |
			cl73_enable_1gkx)
		r.an_x4.cl73_auto_negotiation_control.Set(q, firstLaneMask, v)
		q.Do()

		// clear an bit
		v &^= an_restart
		r.an_x4.cl73_auto_negotiation_control.Set(q, firstLaneMask, v)
		q.Do()
	} else {
		r.an_x4.cl73_auto_negotiation_control.Set(q, firstLaneMask, 0)
		q.Do()

	}
}

func (phy *HundredGig) SetSpeed(port m.Porter, speed float64, isHiGig bool) {
	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	// Put lane data path into reset.
	laneMask.ForeachMask(func(lm m.LaneMask) {
		r.clock_and_reset.lane_reset_and_powerdown.Modify(q, lm, 0<<1, 1<<1)
	})
	q.Do()

	// tx disable from pmd tx disable pin
	laneMask.ForeachMask(func(lm m.LaneMask) {
		r.tx_equalizer.misc_control.Modify(q, lm, 1<<1, 1<<1)
	})
	q.Do()

	// Set port mode based on lane mask.
	phy.setPortMode(phy.PortBlock.GetMode())

	// Based on input speed, set
	ts, ts_ok := phy.speedConfig(laneMask, port.GetPortCommon().GetPhyInterface(), speed, isHiGig)
	if !ts_ok {
		panic(fmt.Errorf("unsupported speed/lane mask combination %e 0x%x", speed, laneMask))
	}

	// Set over-sampling mode.
	{
		const force = 1 << 15
		laneMask.ForeachMask(func(lm m.LaneMask) {
			r.clock_and_reset.over_sampling_mode_control.Set(q, lm, force|uint16(ts.div))
		})
		q.Do()
	}

	v := tscf_pll_multipler(r.pll.multiplier.GetDo(q, firstLaneMask))
	uc_mem := phy.get_uc_mem()
	if v != ts.mul || true {
		// Put core data path into reset.
		r.dig.top_user_control.Modify(q, firstLaneMask, 0<<13, 1<<13)

		// Set pll multiplier for this speed.
		laneMask.ForeachMask(func(lm m.LaneMask) {
			r.pll.multiplier.Modify(q, lm, uint16(ts.mul)<<0, 0xf<<0)
		})

		// PLL reset enable by speed control logic
		r.main.rx_lane_swap.Modify(q, 1<<8, 1<<8)
		q.Do()

		// Update firmware_core_config
		uc_mem.config_word.Set(q, firstLaneMask, &phy.core_config)

		// Take core data path out of reset.
		r.dig.top_user_control.Modify(q, firstLaneMask, 1<<13, 1<<13)
		q.Do()
	}

	// Set the speed-id and enable the port; then wait for valid status before continuing
	// ts.speed should be 0x42 for 100e9
	// NB: firstLaneMask for speed change enable but we don't and it works.
	firstLaneMask.ForeachMask(func(lm m.LaneMask) {
		r.speed_change_x4.control.Set(q, lm, uint16(ts.speed)|0<<8)
	})
	q.Do()
	firstLaneMask.ForeachMask(func(lm m.LaneMask) {
		r.speed_change_x4.control.Set(q, lm, uint16(ts.speed)|1<<8)
	})
	q.Do()

	start := time.Now()
	for r.speed_change_x4.status_read_to_clear.GetDoForeach(q, firstLaneMask)&(1<<1) == 0 {
		if time.Since(start) > 100*time.Millisecond {
			panic("tscf wait for port speed-set")
		}
		time.Sleep(100 * time.Microsecond)
	}

	// Update firmware structure based on new values per lane
	laneMask.Foreach(func(lane m.LaneMask) {
		lm := m.LaneMask(1 << lane)
		uc_mem.lanes[lane].config_word.Set(q, lm, &phy.lane_configs[lane])
	})

	// Unreset lane data path.
	laneMask.ForeachMask(func(lm m.LaneMask) {
		r.clock_and_reset.lane_reset_and_powerdown.Modify(q, lm, 1<<1, 1<<1)
	})
	q.Do()

}

func (phy *HundredGig) SetLoopback(port m.Porter, loopback_type m.PortLoopbackType) {
	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	// get training status and error if enabled
	if r.cl93n72_common.control.GetDo(q, firstLaneMask)&1 == 1 {
		// An error condition
		panic("arg")
	}

	lm := uint16(laneMask)
	modifyMask := uint16(0xffff)
	switch loopback_type {
	case m.PortLoopbackNone, m.PortLoopbackMac:
		// Clear all lanes loopback in phy.
		lm |= lm << 4
		modifyMask = 0
	case m.PortLoopbackPhyRemote:
		lm <<= 4
	}

	r.main.loopback_control.Modify(q, lm&modifyMask, lm)

	// rx-lock-override & signal-detect-override & tx-disable-override-enable
	laneOverride := uint16((1 << 0) | (1 << 1) | (1 << 5))
	laneMask.ForeachMask(func(lmask m.LaneMask) {
		r.pmd_x4.lane_override.Modify(q, lmask, laneOverride&modifyMask, laneOverride)
	})

	// tx disable
	txDisable := uint16(1 << 8)
	laneMask.ForeachMask(func(lmask m.LaneMask) {
		r.pmd_x4.lane_reset_control.Modify(q, lmask, txDisable&modifyMask, txDisable)
	})
	q.Do()
}

func setupHundredGigTxSettings(q *DmaRequest, laneMask m.LaneMask) {
	// Set pre, main, post1, post2, post3, amp
	r := get_hundred_gig_controller()

	// tx fir main +0 tap value
	r.cl93n72_tx.control[3].Modify(q, laneMask, 0x64, 0x7f)
	// tx fir post +1 pre -1 tap values
	r.cl93n72_tx.control[2].Modify(q, laneMask, 0xc, 0x1f)
	// tx fir post +2 value
	r.tx_equalizer.control[1].Modify(q, laneMask, 0, 0x1f)
	// tx fir post +3 value
	r.tx_equalizer.control[2].Modify(q, laneMask, 0, 0xf)
	// amp
	r.ams_tx.control[2].Modify(q, laneMask, 0x8, 0xf)
	q.Do()
}

type tscf_speed_config struct {
	speed tscf_speed
	div   tscf_over_sampling_divider
	mul   tscf_pll_multipler
}

func (phy *HundredGig) ValidateSpeed(port m.Porter, speed float64, isHiGig bool) (ok bool) {
	_, ok = phy.speedConfig(port.GetLaneMask(), port.GetPhyInterface(), speed, isHiGig)
	return
}

func (phy *HundredGig) speedConfig(laneMask m.LaneMask, pi m.PhyInterface, speed float64, isHiGig bool) (sc tscf_speed_config, ok bool) {
	var (
		c uc_core_config_word
		l uc_lane_config_word
	)

	c = phy.core_config
	l = phy.lane_configs[0]

	l.dfe_on = true
	l.cl72_restart_timeout_enable = true

	switch pi {
	case m.PhyInterfaceOptics:
		l.media_type = uc_lane_config_media_optics
		sc.speed = tscf_speed_optics
	case m.PhyInterfaceKR:
		l.media_type = uc_lane_config_media_backplane
		sc.speed = tscf_speed_kr
	case m.PhyInterfaceCR:
		l.media_type = uc_lane_config_media_copper_cable
		sc.speed = tscf_speed_cr
	default:
		panic("phy interface")
	}

	if isHiGig {
		sc.speed |= tscf_speed_hg2
	}

	nLanes := laneMask.NLanes()
	refClock := phy.Switch.GetPhyReferenceClockInHz()
	if refClock != 156.25e6 {
		panic("unsupported ref clock")
	}

	// Rates < 10g use 8/10 encoding; else 64/66 encoding.
	switch speed {
	case 1e9:
		if ok = nLanes == 1; ok {
			sc.speed = tscf_speed_cl36_20g_vco
			// 1e9 = (8/10) * refClock * 132 / 16.5
			sc.mul = tscf_pll_multipler_132
			sc.div = tscf_over_sampling_divider_16_5
			l.dfe_on = false
		}
	case 10e9:
		if ok = nLanes == 1; ok {
			sc.speed |= tscf_speed_10g_x1
			// 10e9 = (64/66) * refClock * 132 / 2
			sc.mul = tscf_pll_multipler_132
			sc.div = tscf_over_sampling_divider_2
			l.dfe_on = (l.media_type == uc_lane_config_media_copper_cable || l.media_type == uc_lane_config_media_backplane)
		}
	case 20e9:
		switch nLanes {
		case 1:
			ok = true
			sc.speed |= tscf_speed_20g_x1
			// 20e9 = (64/66) * refClock * 132 / 1
			sc.mul = tscf_pll_multipler_132
			sc.div = tscf_over_sampling_divider_1
		case 2:
			ok = true
			sc.speed |= tscf_speed_20g_x2
			// 20e9 = 2 lanes x (64/66) * refClock * 132 / 2
			sc.mul = tscf_pll_multipler_132
			sc.div = tscf_over_sampling_divider_2
		}
	case 25e9:
		if ok = (nLanes == 1); ok {
			sc.speed |= tscf_speed_25g_x1
			// 25e9 = (64/66) * refClock * 165 / 1
			sc.mul = tscf_pll_multipler_165
			sc.div = tscf_over_sampling_divider_1
		}
	case 40e9:
		switch nLanes {
		case 2:
			ok = true
			sc.speed |= tscf_speed_40g_x2
			// 40e9 = 2 lanes x (64/66) * refClock * 132 / 1
			sc.mul = tscf_pll_multipler_132
			sc.div = tscf_over_sampling_divider_1
			l.dfe_on = true
			l.dfe_low_power_mode = true
		case 4:
			ok = true
			sc.speed |= tscf_speed_40g_x4
			// 40e9 = 4 lanes x (64/66) * refClock * 132 / 2
			sc.mul = tscf_pll_multipler_132
			sc.div = tscf_over_sampling_divider_2
			l.dfe_on = l.media_type != uc_lane_config_media_optics
			l.dfe_low_power_mode = l.dfe_on
		}
	case 50e9:
		switch nLanes {
		case 2:
			ok = true
			sc.speed |= tscf_speed_50g_x2
			// 50e9 = 2 lanes x (64/66) * refClock * 165 / 1
			sc.mul = tscf_pll_multipler_165
			sc.div = tscf_over_sampling_divider_1
		case 4:
			ok = true
			sc.speed |= tscf_speed_50g_x4
			// 50e9 = 4 lanes x (64/66) * refClock * 165 / 2
			sc.mul = tscf_pll_multipler_165
			sc.div = tscf_over_sampling_divider_2
		}
	case 100e9:
		if ok = nLanes == 4; ok {
			sc.speed |= tscf_speed_100g_x4
			// 100e9 = 4 lanes x (64/66) * refClock * 165 / 1
			sc.mul = tscf_pll_multipler_165
			sc.div = tscf_over_sampling_divider_1
			// test
			//l.dfe_on = l.media_type != uc_lane_config_media_optics
			l.dfe_on = true
		}
	}
	if !ok {
		return
	}

	// Set core-config
	c.vco_rate_in_hz = float64(tscf_pll_multiplers[sc.mul]) * refClock

	phy.core_config = c
	laneMask.Foreach(func(lane m.LaneMask) {
		phy.lane_configs[lane] = l
	})

	return
}

// Return whether PMD is locked or not.
func (phy *HundredGig) PmdLocked(laneMask m.LaneMask) (locked bool) {
	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	locked = r.tlb_rx.pmd_lock_status.GetDo(q, laneMask)&(1<<0) != 0
	return
}

func (phy *HundredGig) getStatus(port m.Porter, lane m.LaneMask) (s portStatus) {
	laneMask := m.LaneMask(1 << lane)
	r := get_hundred_gig_controller()
	q := phy.dmaReq()
	var v [11]uint16
	r.rx_x4.pcs_live_status.Get(q, laneMask, &v[0])
	r.sigdet.status.Get(q, laneMask, &v[1])
	r.tlb_rx.pmd_lock_status.Get(q, laneMask, &v[3])
	r.clock_and_reset.lane_data_path_reset_status.Get(q, laneMask, &v[4])
	r.speed_change_x4_config.final_speed_config.Get(q, laneMask, &v[5])
	r.an_x4.cl73_auto_negotiation_control.Get(q, laneMask, &v[6])
	r.an_x4.cl73_auto_negotiation_misc_status.Get(q, laneMask, &v[7])
	r.speed_change_x4.debug.Get(q, laneMask, &v[8])
	r.pmd_x4.lane_status.Get(q, laneMask, &v[9])
	r.cl93n72_common.status.Get(q, laneMask, &v[10])
	q.Do()

	s.name = port.GetPortName() + fmt.Sprintf(":%d", lane)
	s.live_link = v[0]&(1<<1) != 0
	s.signal_detect = v[1]&(1<<0) != 0
	s.pmd_lock = v[3]&(1<<0) != 0
	s.speed = tscf_speed(v[5] >> 8).String()
	s.Autonegotiate.enable = v[6]&(1<<8) != 0
	s.Autonegotiate.done = v[7]&(1<<15) != 0
	s.sigdet_sts = v[1]
	s.Cl72 = cl72_status(v[10])

	return s
}
