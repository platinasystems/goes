// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/firmware/fe1a"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/port"

	"fmt"
	"time"
)

type Tsce struct {
	Common
	core_config  uc_core_config_word
	lane_configs [4]uc_lane_config_word
}

func (r *tsce_regs) set_master_port_number(q *DmaRequest, n uint) {
	if n >= N_lane {
		panic(n)
	}
	r.main.setup.Modify(q, uint16(n)<<8, 3<<8)
}

func (phy *Tsce) apply_lane_map(q *DmaRequest, r *tsce_regs, laneMask m.LaneMask) {
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
	{
		m := uint16(0)
		for l := range rx_phys_by_logical {
			m |= rx_phys_by_logical[l] << uint(2*l)
		}
		r.main.rx_lane_swap.Set(q, m)
	}

	var pmd_map [N_lane]uint16
	for p := range tx_phys_by_logical {
		pmd_map[p] = uint16(tx_phys_by_logical[phy.Rx_logical_lane_by_phys_lane[p]])
	}

	laneMask.ForeachMask(func(l m.LaneMask) {
		r.dig.tx_lane_map_012.Set(q, l,
			pmd_map[0]<<(5*0)|pmd_map[1]<<(5*1)|pmd_map[2]<<(5*2))
		r.dig.tx_lane_map_3_lane_address_01.Set(q, l,
			pmd_map[3]<<(5*0)|
				uint16(phy.Rx_logical_lane_by_phys_lane[0])<<(5*1)|
				uint16(phy.Rx_logical_lane_by_phys_lane[1])<<(5*2))
		r.dig.tx_lane_address_23.Set(q, l,
			uint16(phy.Rx_logical_lane_by_phys_lane[2])<<(5*0)|
				uint16(phy.Rx_logical_lane_by_phys_lane[3])<<8)
	})
}

func (phy *Tsce) Init() {

	// 2 lanes - one for each mgmt port
	laneMask := m.LaneMask(0x1)
	allLanes := m.LaneMask(0x5)
	r := get_tsce_regs()
	uc_mem := phy.get_uc_mem()

	q := phy.dmaReq()

	// Take Core and data path out of reset.
	r.pmd_x1.reset.Set(q, 1<<0|1<<1)
	q.Do()

	// Set ref clock to 156Mhz
	r.main.setup.Modify(q, 0x3<<13, uint16(0x7)<<13)

	// Bypass the flops used to capature data from the pram_interface
	r.uc.command3.Modify(q, laneMask, 0<<1, 1<<1)

	// Enable UC clock and take out of reset.
	v := uint16(1 << 0)                                    // clock enable + active low reset
	r.uc.command4.Modify(q, laneMask, v|(1<<1), 1<<0|1<<1) // take out of reset
	r.uc.command4.Modify(q, laneMask, 0<<1, 1<<1)
	r.uc.command4.Modify(q, laneMask, 1<<1, 1<<1)

	// Program RAM load while DW8051 in reset
	r.uc.command1.Modify(q, laneMask, 0<<7, 0x3<<7)

	// Write RAM start address of 0
	r.uc.address.Set(q, laneMask, 0)

	// Initiate zero of program ram.
	r.uc.command1.Modify(q, laneMask, 0<<15, 1<<15)
	r.uc.command1.Modify(q, laneMask, 1<<15, 1<<15)

	q.Do()

	time.Sleep(500 * time.Microsecond)
	r.uc.command1.Modify(q, laneMask, 0<<15, 1<<15)
	q.Do()

	// Wait for code ram init done.
	start := time.Now()
	for r.uc.mdio_8051_fsm_status.GetDo(q, laneMask)&(1<<15) == 0 {
		if time.Since(start) > 100*time.Millisecond {
			panic("tsce wait for code ram init timeout")
		}
		time.Sleep(100 * time.Microsecond)
	}

	{
		// Enable direct uc pram interface writes.
		r.pmd_x1.reset.Modify(q, 1<<8, 1<<8)

		// [0] Enable parallel interface to load program RAM
		// [1] Bypass the flops used to capature data from the pram_interface
		// [2] Parallel interface reset from Program RAM load 0 - asserted 1 - de-asserted
		cmd3 := uint16((1 << 2) | (0 << 1) | (1 << 0))
		r.uc.command3.Modify(q, laneMask, cmd3, 0x7)
		q.Do()

		phy.PortBlock.LoadFirmware(&q.DmaRequest, fe1a.Eucode.Data)
		q.Do()
		if q.Err != nil {
			panic(q.Err)
		}

		// Disable direct uc pram interface writes.
		r.uc.command3.Modify(q, laneMask, ^cmd3, 0x7)
		r.pmd_x1.reset.Modify(q, 0<<8, 1<<8)
		q.Do()
	}

	// Disable pmd_ln_hrstb input pin.
	r.clock_and_reset.lane_reset_and_powerdown_pin_disable.Modify(q, laneMask, 1<<0, 1<<0)

	// Set uc active bit.
	r.dig.top_user_control.Modify(q, laneMask, 1<<15, 1<<15)
	q.Do()

	// Take UC out of reset.
	r.uc.command1.Modify(q, laneMask, 1<<4, 1<<4)
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

	// Have microcode verify crc.  CRC check does not reliably work (microcode problem?) so skip it.
	if false {
		crc := uint16(0)
		len := uint16(len(fe1a.Eucode.Data))
		c := uc_cmd{command: uc_cmd_compute_ucode_crc, in: &len, out: &crc}
		err := c.do(q, laneMask, &r.uc_cmd)
		if err != nil {
			panic(err)
		}
		if crc != fe1a.Eucode.Crc {
			panic(fmt.Errorf("uc ucode crc does not match got %x != want %x", crc, fe1a.Eucode.Crc))
		}
	}

	// Re-enable pmd_ln_hrstb input pin.
	r.clock_and_reset.lane_reset_and_powerdown_pin_disable.Modify(q, laneMask, 0<<0, 1<<0)

	// Set PLL multiplier to 66. Removed: 66 is default value.
	// r.pll.multiplier.Modify(q, laneMask, uint16(tsce_pll_multiplier_66), 0xf)
	// q.Do()

	phy.apply_lane_map(q, r, allLanes)

	// Setup auto-negotiation timers.
	{
		r.an_x1.cl37.restart_timer.Set(q, 0x29a)
		r.an_x1.cl37.complete_ack_timer.Set(q, 0x29a)
		//r.an_x1.cl37.timeout_error_timer.Set(q, 0x29a)

		r.an_x1.cl73.break_link_timer.Set(q, 0x10ed)
		r.an_x1.cl73.timeout_error_timer.Set(q, 0x100)
		r.an_x1.cl73.parallel_detect_dme_lock_timer.Set(q, 0x14d4)
		r.an_x1.cl73.link_up_timer.Set(q, 0x29a)
		r.an_x1.cl73.qualify_link_status_timer[0].Set(q, 0x8382)
		r.an_x1.cl73.qualify_link_status_timer[1].Set(q, 0x8382)
		r.an_x1.cl73.parallel_detect_signal_detect_timer.Set(q, 0xa6a)
		r.an_x1.cl73.ignore_cl37_sync_status_down_timer.Set(q, 0x29a)
		r.an_x1.cl73.period_to_wait_for_link_before_cl37.Set(q, 0xa6a)
		r.an_x1.cl73.ignore_link_timer.Set(q, 0x29a)
		r.an_x1.cl73.dme_page_timers.Set(q, 0x3b5f)
		r.an_x1.cl73.sgmii_timer.Set(q, 0x6b)
	}

	// Set MLD counters
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.tx_x2.mld_swap_count.Set(q, lm, 0xfffc)
	})

	r.set_master_port_number(q, 0)
	q.Do()

	// Set CL48 aneg timer related info (see temod_cl48_lfrfli_init())
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.tx_x2.cl48_control.Modify(q, lm, 1<<4|1<<5|1<<6, 1<<4|1<<5|1<<6)
	})
	q.Do()

	// Take core data path out of reset for all lanes.
	r.dig.top_user_control.Modify(q, laneMask, 1<<13, 1<<13)
	q.Do()

	// Write ucode config word.
	{
		phy.core_config = uc_core_config_word{
			vco_rate_in_hz: 10.3125e9,
		}
		mem := phy.get_uc_mem()
		mem.config_word.Set(q, laneMask, &phy.core_config)
	}

	// Take core data path out of reset for all lanes.
	r.dig.top_user_control.Modify(q, laneMask, 1<<13, 1<<13)
	q.Do()

	//
	// Per lane initializations
	//

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
		setupTsceTxSettings(q, lm)
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
		r.rx_x4.pma_control0.Modify(q, lm, 0<<0, 1<<0)
		r.rx_x4.pma_control0.Modify(q, lm, 1<<0, 1<<0)
	})
	q.Do()

	// TXP out of reset + enable.
	allLanes.ForeachMask(func(lm m.LaneMask) {
		r.tx_x4.misc.Modify(q, lm, 0<<0|0<<1, 1<<0|1<<1)
		r.tx_x4.misc.Modify(q, lm, 1<<0|1<<1, 1<<0|1<<1)
		// r.tlb_rx.tx_to_rx_loopback.Modify(q, lm, 1<<0, 1<<0)
	})
	q.Do()
}

func setupTsceTxSettings(q *DmaRequest, laneMask m.LaneMask) {
	// Set pre, main, post1, post2, post3, amp
	r := get_tsce_regs()

	// pre
	r.tx_equalizer.control[0].Modify(q, laneMask, 0xc, 0x1f)
	// tx fir main +0 tap value
	r.tx_equalizer.control[1].Modify(q, laneMask, 0x64, 0x7f)
	// tx fir post +1 values
	r.tx_equalizer.control[0].Modify(q, laneMask, 0x0, 0x7e0)
	// tx fir post +2 value
	r.tx_equalizer.control[1].Modify(q, laneMask, 0, 0xf80)
	// tx fir post +3 value
	r.tx_equalizer.control4.Modify(q, laneMask, 0, 0xf)
	// amp
	r.ams_tx.control[2].Modify(q, laneMask, 0xc, 0xf)
	q.Do()
}

func (phy *Tsce) SetSpeed(port m.Porter, speed float64, isHiGig bool) {

	r := get_tsce_regs()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLane := laneMask.FirstLane()
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

	v := tsce_pll_multiplier(r.pll.multiplier.GetDo(q, firstLaneMask))
	uc_mem := phy.get_uc_mem()
	if v != ts.mul || true {
		// Put core data path into reset.
		r.dig.top_user_control.Modify(q, firstLaneMask, 0<<13, 1<<13)

		// Release the UC reset
		r.uc.command4.Modify(q, laneMask, 1<<1, 1<<1)

		// Set pll multiplier for this speed.
		laneMask.ForeachMask(func(lm m.LaneMask) {
			r.pll.multiplier.Modify(q, lm, uint16(ts.mul)<<0, 0xf<<0)
		})

		// PLL reset enable by speed control logic
		r.main.setup.Modify(q, 1<<10, 1<<10)
		q.Do()

		// Set master port number to start lane
		r.set_master_port_number(q, firstLane)
		q.Do()

		// Update firmware_core_config
		uc_mem.config_word.Set(q, firstLaneMask, &phy.core_config)

		// Take core data path out of reset.
		r.dig.top_user_control.Modify(q, firstLaneMask, 1<<13, 1<<13)
		q.Do()
	}

	// Set the speed-id and enable the port; then wait for valid status before continuing
	// NB: does firstLaneMask for speed change enable but we don't and it works.
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
			panic("tsce wait for port speed-set")
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

func (phy *Tsce) setPortMode(mode port.PortBlockMode) {
	r := get_tsce_regs()
	q := phy.dmaReq()

	const (
		m4x1 = iota
		m2x1_1x2
		m1x2_2x1
		m2x2
		m1x4
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

type tsce_speed_config struct {
	speed tsce_speed
	div   tsce_over_sampling_divider
	mul   tsce_pll_multiplier
}

func (phy *Tsce) ValidateSpeed(port m.Porter, speed float64, isHiGig bool) (ok bool) {
	_, ok = phy.speedConfig(port.GetLaneMask(), port.GetPhyInterface(), speed, isHiGig)
	return
}

func (phy *Tsce) speedConfig(laneMask m.LaneMask, pi m.PhyInterface, speed float64, isHiGig bool) (sc tsce_speed_config, ok bool) {
	var (
		c uc_core_config_word
		l uc_lane_config_word
	)

	c = phy.core_config
	firstLane := laneMask.FirstLane()
	l = phy.lane_configs[firstLane]

	l.cl72_restart_timeout_enable = true

	{
		switch pi {
		case m.PhyInterfaceOptics:
			l.media_type = uc_lane_config_media_optics
			sc.speed = tsce_speed_10g_x4
		case m.PhyInterfaceKR:
			l.media_type = uc_lane_config_media_backplane
			sc.speed = tsce_speed_10g_kr1
		case m.PhyInterfaceCR:
			l.media_type = uc_lane_config_media_copper_cable
			sc.speed = tsce_speed_10g_cx4
		default:
			panic("eagle phy interface")
		}
	}

	nLanes := laneMask.NLanes()
	refClock := phy.Switch.GetPhyReferenceClockInHz()
	if refClock != 156.25e6 {
		panic("eagle unsupported ref clock")
	}

	// Rates < 10g use 8/10 encoding; else 64/66 encoding.
	switch speed {
	case 10e9:
		if ok = nLanes == 1; ok {
			sc.speed = tsce_speed_10g_kr1
			// 10e9 = (64/66) * refClock * 66
			sc.mul = tsce_pll_multiplier_66
			sc.div = tsce_over_sampling_divider_1
			l.dfe_on = true
			l.scrambling_disable = false
			l.cl72_auto_polarity_enable = true // enable this or no 10g kr1
		}
	case 1e9:
		if ok = nLanes == 1; ok {
			sc.speed = tsce_speed_1g_kx1
			// 1e9 = (8/10) * refClock * 66 / 8.25
			sc.mul = tsce_pll_multiplier_66
			sc.div = tsce_over_sampling_divider_8_25
			l.dfe_on = false
			l.scrambling_disable = true
		}

	}
	if !ok {
		return
	}

	// Set core-config
	c.vco_rate_in_hz = tsce_pll_multipliers[sc.mul] * refClock

	phy.core_config = c
	laneMask.Foreach(func(lane m.LaneMask) {
		phy.lane_configs[lane] = l
	})

	return
}

func (phy *Tsce) SetLoopback(port m.Porter, loopback_type m.PortLoopbackType) {}

func (phy *Tsce) SetAutoneg(port m.Porter, enable bool) {
	r := get_tsce_regs()

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
	} else {
		co.core_config_from_pcs = false
		la.lane_config_from_pcs = false
		la.autoneg_enable = false
	}

	laneMask.ForeachMask(func(lm m.LaneMask) {
		r.clock_and_reset.over_sampling_mode_control.Modify(q, lm, 0, 1<<15)
	})
	q.Do()

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
	}
	r.main.setup.Modify(q, setup, single_port_mode_on)
	q.Do()

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

	if true { // NB. does this..
		// test
		r.speed_change_x4.control.Modify(q, firstLaneMask, 0<<8, 1<<8)
		q.Do()
		r.speed_change_x4.control.Modify(q, firstLaneMask, 1<<8, 1<<8)
		q.Do()
	}

	// Set master port number to start lane
	r.set_master_port_number(q, firstLane)
	q.Do()

	phy.setLocalAdvert(port)

	phy.setCL72(port, enable)

	phy.setAnEnable(port, enable)
}

func (phy *Tsce) setAnEnable(port m.Porter, enable bool) {
	// Used to signal PLL to restart after AN
	r := get_tsce_regs()

	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	const (
		speed_change_enable = 1 << 8
		pd_kx4_en           = 1 << 0
		pd_kx_en            = 1 << 1
	)

	// speed change book-keeping
	if enable {
		r.speed_change_x4.control.Modify(q, firstLaneMask, 0<<8, 1<<8)
		q.Do()
	} else {
		r.speed_change_x4.control.Modify(q, firstLaneMask, 1<<8, 1<<8)
		q.Do()
	}

	// Set CL37 HVCO bit for 10G
	var setup uint16
	if r.pll.multiplier.GetDo(q, firstLaneMask) == uint16(tsce_pll_multiplier_66) {
		setup |= 1 << 12
	} else {
		setup &^= 1 << 12
	}
	r.main.setup.Modify(q, setup|1<<10, 1<<12|1<<10)
	q.Do()

	// Turn off all AN bits
	r.an_x4.enables.Set(q, firstLaneMask, 0)
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

	const (
		cl73_an_restart = 1 << 0
		cl73_an_enable  = 1 << 8
		lanes_mask      = 3 << 12
	)
	if enable {
		r.an_x4.misc_controls.Modify(q, firstLaneMask, pd_kx4_en|pd_kx_en, pd_kx4_en|pd_kx_en)
		q.Do()

		// Restart bit is self clearing.
		v := uint16(cl73_an_restart | cl73_an_enable | (log2NLanes << 12))
		r.an_x4.enables.Modify(q, firstLaneMask, v, 1<<0|1<<8|lanes_mask)
		q.Do()

		r.main.setup.Modify(q, 1<<10, 1<<10)
		q.Do()
	} else {
		r.an_x4.enables.Modify(q, firstLaneMask, 0, cl73_an_enable|lanes_mask)
		q.Do()
	}

}

func (phy *Tsce) setCL72(port m.Porter, enable bool) {
	r := get_tsce_regs()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	const (
		restart_training = 1 << 0
		enable_training  = 1 << 1
		sc_x4_enabled    = 1 << 8
	)

	if enable {
		firstLaneMask.ForeachMask(func(l m.LaneMask) {
			r.cl93n72_common.control.Modify(q, l,
				enable_training|restart_training,
				enable_training|restart_training)
		})
	} else {
		firstLaneMask.ForeachMask(func(l m.LaneMask) {
			r.cl93n72_common.control.Modify(q, l,
				0, enable_training)
		})
	}
	q.Do()

	if true {
		sc_x4_control := r.speed_change_x4.control.GetDo(q, firstLaneMask)
		if sc_x4_control&sc_x4_enabled == sc_x4_enabled {
			firstLaneMask.ForeachMask(func(lm m.LaneMask) {
				r.speed_change_x4.control.Modify(q, lm, 0<<8, 1<<8)
			})
			q.Do()
			firstLaneMask.ForeachMask(func(lm m.LaneMask) {
				r.speed_change_x4.control.Modify(q, lm, 1<<8, 1<<8)
			})
			q.Do()
		}
	}
}

func (phy *Tsce) setLocalAdvert(port m.Porter) {

	r := get_tsce_regs()
	q := phy.dmaReq()
	laneMask := port.GetLaneMask()
	firstLaneMask := m.LaneMask(1 << laneMask.FirstLane())

	const (
		b_FEC       = 3 << 8
		b_PAUSE     = 3 << 6
		b_1G_KX1    = 1 << 5
		b_10G_KX4   = 1 << 4
		b_10G_KR    = 1 << 3
		b_40G_KR4   = 1 << 2
		b_40G_CR4   = 1 << 1
		b_100G_CR10 = 1 << 0
	)

	var an_base uint16
	if laneMask.NLanes() == 1 {
		an_base = b_FEC | b_PAUSE | b_1G_KX1 | b_10G_KR
	} else {
		an_base = b_FEC | b_PAUSE | b_40G_KR4 | b_40G_CR4 | b_100G_CR10
	}
	r.an_x4.cl73_base_page_abilities[1].Set(q, firstLaneMask, an_base)
	q.Do()

	const (
		b_tx_nonce      = 0x15 << 5
		b_base_page_sel = 1 << 0
		b_retry_count   = 0xf << 6
	)
	an_base = b_base_page_sel | b_tx_nonce
	r.an_x4.cl73_base_page_abilities[0].Set(q, firstLaneMask, an_base)
	q.Do()

	r.an_x4.misc_controls.Modify(q, firstLaneMask, b_retry_count, b_retry_count)
	q.Do()
}

func (phy *Tsce) SetEnable(p m.Porter, enable bool) {}

func (phy *Tsce) getMgmtStatus(port m.Porter, lane m.LaneMask) (s portStatus) {
	laneMask := m.LaneMask(1 << lane)
	r := get_tsce_regs()
	q := phy.dmaReq()
	var v [25]uint16
	r.rx_x4.pcs_live_status.Get(q, laneMask, &v[0])
	r.rx_x4.latched_pcs_status[0].Get(q, laneMask, &v[21])
	r.rx_x4.latched_pcs_status[1].Get(q, laneMask, &v[22])
	r.sigdet.status.Get(q, laneMask, &v[1])
	r.tlb_rx.pmd_lock_status.Get(q, laneMask, &v[3])
	r.clock_and_reset.lane_data_path_reset_status.Get(q, laneMask, &v[4])
	r.speed_change_x4_config.final_speed_config.Get(q, laneMask, &v[5])
	r.an_x4.enables.Get(q, laneMask, &v[6])
	r.an_x4.sw_control_status.Get(q, laneMask, &v[7])
	r.speed_change_x4.debug.Get(q, laneMask, &v[8])
	r.pmd_x4.lane_status.Get(q, laneMask, &v[9])
	r.cl93n72_common.status.Get(q, laneMask, &v[10])
	r.pll.multiplier.Get(q, laneMask, &v[11])
	r.clock_and_reset.over_sampling_status.Get(q, laneMask, &v[12])
	r.an_x4.page_sequencer_status.Get(q, laneMask, &v[13])
	r.an_x4.page_exchanger_status.Get(q, laneMask, &v[14])
	r.an_x4.ability_resolution.Get(q, laneMask, &v[15])
	r.an_x4.misc_status.Get(q, laneMask, &v[16])
	r.an_x4.sw_control_status.Get(q, laneMask, &v[17])
	r.an_x4.misc_controls.Get(q, laneMask, &v[18])
	r.an_x4.tla_sequencer_status.Get(q, laneMask, &v[19])
	r.an_x4.page_decoder_status.Get(q, laneMask, &v[21])
	r.main.setup.Get(q, &v[20])
	r.cl93n72_common.status.Get(q, laneMask, &v[24])
	q.Do()

	s.name = port.GetPortName() + fmt.Sprintf(":%d", lane)
	s.live_link = v[0]&(1<<1) != 0
	s.signal_detect = v[1]&(1<<0) != 0
	s.pmd_lock = v[3]&(1<<0) != 0
	s.speed = tsce_speed(v[5] >> 8).String()
	s.Autonegotiate.enable = v[6]&(1<<8) != 0
	s.Autonegotiate.done = v[16]&(1<<15) != 0
	s.sigdet_sts = v[1]
	s.Cl72 = cl72_status(v[10])

	// For autoneg debug - remove eventually
	if false {
		fmt.Printf("%s:\n", s.name)
		fmt.Printf("  pcs live status %x\n", v[0])
		fmt.Printf("  pcs latched status %x %x\n", v[21], v[22])
		fmt.Printf("  over sampling force %x\n", r.clock_and_reset.over_sampling_mode_control.GetDo(q, laneMask))
		fmt.Printf("  an page sequencer status(c1a8): %s\n", tsce_an_x4_page_sequencer_status(v[13]))
		fmt.Printf("  an page exchanger status(c1a9):\n%s", tsce_an_x4_page_exchanger_status(v[14]).Lines().Indent(4))
		fmt.Printf("  an ability resolution(c1ab): %s\n", tsce_an_x4_ability_resolution(v[15]))
		fmt.Printf("  an misc status(c1ac): %s\n", tsce_an_x4_misc_status(v[16]))
		fmt.Printf("  sw_control_status: %s\n", tsce_an_x4_sw_control_status(v[17]))
		fmt.Printf("  tla seq status: %s\n", an_x4_tla_sequencer_status(v[19]))
		fmt.Printf("  an misc controls: %x\n", v[18])
		fmt.Printf("  page decoder status: %x\n", v[21])
		fmt.Printf("  main setup: %x\n", v[20])
		fmt.Printf("  cl72 status %x\n", v[24])
		mul, div := tsce_pll_multipliers[v[11]&0xf], tsce_over_sampling_dividers[v[12]]
		const ref = 156.25e6
		r := ref * mul / div
		fmt.Printf("  pll: %f / %f 8/10 %e 64/66 %e \n", mul, div, (8./10.)*r, (64./66.)*r)
	}

	return s
}

type tsce_an_x4_page_sequencer_status uint16

func (x tsce_an_x4_page_sequencer_status) String() (s string) {
	if x&(1<<0) != 0 {
		s += "cl73 done, "
	}
	if x&(1<<1) != 0 {
		s += "cl37 done, "
	}
	if x&(1<<2) != 0 {
		s += "rx next page without T toggling (clear on read), "
	}
	if x&(1<<3) != 0 {
		s += "rx invalid auto-neg page sequence (clear on read), "
	}
	if x&(1<<4) != 0 {
		s += "rx auto-neg MPS-5 OUI match (clear on read), "
	}
	if x&(1<<5) != 0 {
		s += "rx auto-neg MPS-5 OUI mismatch (clear on read), "
	}
	if x&(1<<6) != 0 {
		s += "rx auto-neg unformatted page 3 (clear on read), "
	}
	if x&(1<<7) != 0 {
		s += "rx mismatching auto-neg message page (clear on read), "
	}
	if x&(1<<8) != 0 {
		s += "rx auto-neg message page 1024 (Over 1G Message) (clear on read), "
	}
	if x&(1<<9) != 0 {
		s += "rx auto-neg message page 5 (OUI Message) (clear on read), "
	}
	if x&(1<<10) != 0 {
		s += "rx auto-neg message page 1 (Null Message) (clear on read), "
	}
	if x&(1<<11) != 0 {
		s += "rx auto-neg next page (clear on read), "
	}
	if x&(1<<12) != 0 {
		s += "rx auto-neg base page (clear on read), "
	}
	if x&(1<<13) != 0 {
		s += "rx non-SGMII page when in SGMII auto-neg mode (clear on read), "
	}
	if x&(1<<14) != 0 {
		s += "In HP auto-neg mode (clear on read), "
	}
	if x&(1<<15) != 0 {
		s += "SGMII mode, "
	}
	return
}

type tsce_an_x4_page_exchanger_status uint16

func (x tsce_an_x4_page_exchanger_status) Lines() (lines elib.Lines) {
	if x&(1<<0) != 0 {
		lines.Add("rx auto-neg restart (0) page")
	}
	if x&(1<<1) != 0 {
		lines.Add("entered IDLE_DETECT state; clear on read")
	}
	if x&(1<<2) != 0 {
		lines.Add("entered DISABLE_LINK state; clear on read")
	}
	if x&(1<<3) != 0 {
		lines.Add("entered auto-neg ERROR state; clear on read")
	}
	if x&(1<<4) != 0 {
		lines.Add("entered auto-neg AN_ENABLE state; clear on read")
	}
	if x&(1<<5) != 0 {
		lines.Add("entered auto-neg ABILITY_DETECT state; clear on read")
	}
	if x&(1<<6) != 0 {
		lines.Add("entered auto-neg ACKNOWLEDGE_DETECT state; clear on read")
	}
	if x&(1<<7) != 0 {
		lines.Add("entered auto-neg COMPLETE_ACKNOWLEDGE state; clear on read")
	}
	if x&(1<<8) != 0 {
		lines.Add("auto-neg consistency mismatch detected; clear on read")
	}
	if x&(1<<9) != 0 {
		lines.Add("page exchanger received non-zero configuration ordered set; clear on read")
	}
	if x&(1<<10) != 0 {
		lines.Add("page exchanger entered AN_RESTART state; clear on read")
	}
	if x&(1<<11) != 0 {
		lines.Add("page exchanger entered AN_GOOD_CHECK state; clear on read")
	}
	if x&(1<<12) != 0 {
		lines.Add("page exchanger entered LINK_OK state; clear on read")
	}
	if x&(1<<13) != 0 {
		lines.Add("page exchanger entered NEXT_PAGE_WAIT state; clear on read")
	}
	return
}

type tsce_an_x4_ability_resolution uint16

func (x tsce_an_x4_ability_resolution) String() (s string) {
	if x&(1<<0) != 0 {
		s += "switch to cl37, "
	}
	if x&(1<<1) != 0 {
		s += "hg2, "
	}
	if x&(1<<2) != 0 {
		s += "CL72 training, "
	}
	if x&(1<<3) != 0 {
		s += "forward-error correction, "
	}
	if x&(1<<12) != 0 {
		s += "tx pause, "
	}
	if x&(1<<13) != 0 {
		s += "rx pause, "
	}
	if x&(1<<14) != 0 {
		s += "full duplex, "
	}
	if x&(1<<15) != 0 {
		s += "error: no common speed, "
	} else {
		s += fmt.Sprintf("speed %s", tsce_speed((x>>4)&0xff))
	}
	return
}

type tsce_an_x4_misc_status uint16

func (x tsce_an_x4_misc_status) String() (s string) {
	s += "Parallel Detect: "
	if x&(1<<0) != 0 {
		s += "kx, "
	} else {
		s += "kx4, "
	}
	if x&(1<<1) != 0 {
		s += "active, "
	}
	if x&(1<<7) != 0 {
		s += "complete, "
	}
	s += "AN: "
	if x&(1<<6) != 0 {
		s += "in progress, "
	}
	if x&(1<<8) != 0 {
		s += "remote fault indicated in base page, "
	}
	if x&(1<<15) != 0 {
		s += "complete, "
	}
	s += fmt.Sprintf("retries: failures %d, any reason: %d", (x>>2)&0xf, (x>>9)&0x3f)
	return
}

type tsce_an_x4_sw_control_status uint16

func (x tsce_an_x4_sw_control_status) String() (s string) {
	s += fmt.Sprintf("TLA lane sequencer fsm status: %x, ", uint32(x>>0)&0xff)
	if x&(1<<8) != 0 {
		s += "PD completed & CL37 selected, "
	}
	if x&(1<<13) != 0 {
		s += "link-partner status valid, "
	}
	if x&(1<<14) != 0 {
		s += "ld control valid, "
	}
	if x&(1<<15) != 0 {
		s += "all AN page exchanges have completed, "
	}
	return
}

type an_x4_tla_sequencer_status uint16

func (x an_x4_tla_sequencer_status) String() (s string) {
	if x&(1<<0) != 0 {
		s += "init, "
	}
	if x&(1<<1) != 0 {
		s += "set an speed, "
	}
	if x&(1<<2) != 0 {
		s += "wait an speed, "
	}
	if x&(1<<3) != 0 {
		s += "wait an done, "
	}
	if x&(1<<4) != 0 {
		s += "set rs speed, "
	}
	if x&(1<<5) != 0 {
		s += "wait rs cl72 enable, "
	}
	if x&(1<<7) != 0 {
		s += "an ignore link, "
	}
	if x&(1<<8) != 0 {
		s += "an cl73 good check, "
	}
	if x&(1<<9) != 0 {
		s += "an cl73 good, "
	}
	if x&(1<<10) != 0 {
		s += "an fail, "
	}
	if x&(1<<11) != 0 {
		s += "an dead, "
	}
	if x&(1<<12) != 0 {
		s += "set pd speed, "
	}
	if x&(1<<13) != 0 {
		s += "wait pd speed, "
	}
	if x&(1<<14) != 0 {
		s += "pd ignore link, "
	}
	if x&(1<<15) != 0 {
		s += "pd good check, "
	}
	return
}
