// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ixge

import (
	"fmt"
)

type dev_x540 struct {
	dev
}

func (d *dev_x540) get_put_semaphore(is_put bool) (x reg) {
	const (
		driver   = 1 << 0
		register = 1 << 31
	)
	if is_put {
		x = d.regs.software_semaphore.put_semaphore(&d.dev, driver|register)
	} else {
		d.regs.software_semaphore.get_semaphore(&d.dev, "sw", driver)
		x = d.regs.software_semaphore.get_semaphore(&d.dev, "reg", register)
	}
	return
}

func (d *dev_x540) get_semaphore() { d.get_put_semaphore(false) }
func (d *dev_x540) put_semaphore() { d.get_put_semaphore(true) }

func (d *dev_x540) phy_init() {
	// Pci function selects unit 0 or 1.
	d.phy_index = uint(d.p.Addr.Fn)
	id := d.get_dev_id()
	switch id {
	case dev_id_x550em_x_kr:
		d.kr_phy_init()
	default:
		panic(fmt.Errorf("unsupported phy for device %s", id))
	}
}

func (d *dev_x540) kr_phy_init() {
	addr := reg((kr_phy_reg_phy_0 << d.phy_index) | kr_phy_link_control_1)
	v := d.read_iphy_reg(kr_phy_dev_type, addr)

	v |= kr_phy_link_control_1_an_enable

	v &^= kr_phy_link_control_1_an_fec_req | kr_phy_link_control_1_an_cap_fec

	// Advertise both 10G and 1G speeds.
	v |= kr_phy_link_control_1_an_cap_kr | kr_phy_link_control_1_an_cap_kx

	// Restart auto negotiation.  Self-clearing bit.
	v |= kr_phy_link_control_1_an_restart

	d.write_iphy_reg(kr_phy_dev_type, addr, v)
}

// Internal KR PHY registers
const (
	kr_phy_dev_type = 0

	// Device select 0/1
	kr_phy_reg_phy_0 = 1 << 14
	kr_phy_reg_phy_1 = 1 << 15

	// Registers
	kr_phy_port_car_gen_control = 0x0010
	kr_phy_link_control_1       = 0x020c
	kr_phy_dsp_txffe_state_4    = 0x0634
	kr_phy_dsp_txffe_state_5    = 0x0638
	kr_phy_rx_trn_linkup_ctrl   = 0x0b00
	kr_phy_pmd_dfx_burnin       = 0x0e00
	kr_phy_tx_coeff_control_1   = 0x1520
	kr_phy_rx_ana_ctl           = 0x1a00

	// kr_phy_link_control_1 register
	kr_phy_link_control_1_force_speed_mask = 0x7 << 8
	kr_phy_link_control_1_force_speed_1g   = 2 << 8
	kr_phy_link_control_1_force_speed_10g  = 4 << 8
	kr_phy_link_control_1_an_fec_req       = 1 << 14
	kr_phy_link_control_1_an_cap_fec       = 1 << 15
	kr_phy_link_control_1_an_cap_kx        = 1 << 16
	kr_phy_link_control_1_an_cap_kr        = 1 << 18
	kr_phy_link_control_1_eee_cap_kx       = 1 << 24
	kr_phy_link_control_1_eee_cap_kr       = 1 << 26
	kr_phy_link_control_1_an_enable        = 1 << 29
	kr_phy_link_control_1_an_restart       = 1 << 31
)
