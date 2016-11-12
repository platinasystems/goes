// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icpu

import (
	"github.com/platinasystems/go/elib/hw"
	"unsafe"
)

type u32 hw.U32

func (r *u32) get() uint32  { return (*hw.U32)(r).Get() }
func (r *u32) set(v uint32) { (*hw.U32)(r).Set(v) }

type addr_map struct {
	lower, upper u32
}

type Controller struct {
	_ [0x2000]byte

	paxb [2]struct {
		clock_control               u32
		rc_power_management_control u32
		rc_power_management_status  u32
		ep_power_management_control u32
		ep_power_management_status  u32
		ep_ltr_control              u32
		ep_ltr_status               u32
		_                           [0x20 - 0x1c]byte
		ep_obff_status              u32
		pcie_error_status           u32
		_                           [0x30 - 0x28]byte

		endianness              u32
		apb_timeout_count       u32
		tx_arbitration_priority u32
		_                       [0x40 - 0x3c]byte

		read_completion_buffer_init_start u32
		read_completion_buffer_init_done  u32
		_                                 [0x100 - 0x48]byte

		pcie_rc_axi_config u32

		pcie_ep_axi_config u32

		pcie_rx_debug_status  u32
		pcie_rx_debug_control u32

		_ [0x120 - 0x110]byte

		config_indirect_address u32
		config_indirect_data    u32

		_ [0x1f8 - 0x128]byte

		config_address    u32
		config_data       u32
		pcie_sys_eq_page  u32
		pcie_sys_msi_page u32

		_ [0x210 - 0x208]byte

		pcie_sys_msi_control [6]u32

		_ [0x250 - 0x228]byte

		pcie_sys_eq_pointers   [6]struct{ head, tail u32 }
		pcie_sys_eq_tail_early [6]u32

		_ [0x2a0 - 0x298]byte

		pcie_sys_eq_overwritten [6]u32

		_ [0x2c0 - 0x2b8]byte

		pcie_sys_eq_page_upper  u32
		pcie_sys_msi_page_upper u32

		_ [0x330 - 0x2c8]byte

		pcie_sys_rc_intx_enable u32
		pcie_sys_rc_intx_status u32

		_ [0x340 - 0x338]byte

		pcie_sys_msi_request           u32
		pcie_sys_host_interrupt_enable u32
		pcie_sys_host_interrupt_status u32

		_ [0x350 - 0x34c]byte

		pcie_sys_host_mailbox [4]u32

		pcie_sys_ep_interrupt_enable [2]u32

		_ [0x370 - 0x368]byte

		pcie_sys_ep_interrupt_status [2]u32
		_                            [0x380 - 0x378]byte

		cmicd_to_pcie_interrupt_enable u32

		_ [0xc00 - 0x384]byte

		imap0 [2][8]u32

		imap0_upper [2][8]u32

		_ [0xcc0 - 0xc80]byte

		imap2 [2]addr_map

		func0_imap0_regs_type u32

		_ [0xd00 - 0xcd4]byte

		iarr_2 [3]addr_map

		_ [0xd20 - 0xd18]byte

		oarr_0 [2]addr_map

		_ [0xd34 - 0xd30]byte

		oarr_msi_page [2]u32

		_ [0xd40 - 0xd3c]byte

		pcie_outbound_64bit_address_mapping_table0 [2]addr_map

		oarr_msi_page_upper [2]u32
		iarr_2_size         [2]u32

		oarr_2 [1]addr_map

		omap_2 [1]addr_map

		imap1 [2][8]addr_map

		_ [0xf00 - 0xdf0]byte

		mem_control                    u32
		mem_ecc_error_log              [2]u32
		pcie_link_status               u32
		strap_status                   u32
		reset_status                   u32
		reset_enable_in_pcie_link_down u32
		misc_interrupt_enable          u32
		tx_debug_config                u32
		misc_config                    u32
		misc_status                    u32

		_ [0xf30 - 0xf2c]byte

		interrupt_enable u32
		interrupt_clear  u32
		interrupt_status u32

		_ [0xf40 - 0xf3c]byte

		apb_error_enable_for_cfg_rd_completions u32

		_ [0x1000 - 0xf44]byte
	}

	_ [0x7000 - 0x4000]byte

	window7 [1024]u32

	smbus0 I2cController

	_ [0xb000 - 0x8050]byte

	smbus1 I2cController

	_ [0xc000 - 0xb050]byte
}

func (r *Controller) iGetSet(x *hw.U32, v *uint32, isSet bool) {
	paxb := &r.paxb[0]
	addr := uint32(uintptr(unsafe.Pointer(x)) - uintptr(unsafe.Pointer(r)))
	page, offset := addr&^0xfff, (addr&0xfff)/4

	paxb.imap0[0][7].set(1<<0 | page)
	if isSet {
		r.window7[offset].set(*v)
	} else {
		*v = r.window7[offset].get()
	}
}

func (r *Controller) IndirectGet32(x *hw.U32) (v uint32) { r.iGetSet(x, &v, false); return }
func (r *Controller) IndirectSet32(x *hw.U32, v uint32)  { r.iGetSet(x, &v, true) }

type U32 hw.U32

func (r *U32) Get(regs *Controller) uint32 {
	x := (*hw.U32)(r)
	if regs != nil {
		return regs.IndirectGet32(x)
	} else {
		return x.Get()
	}
}

func (r *U32) Set(regs *Controller, v uint32) {
	x := (*hw.U32)(r)
	if regs != nil {
		regs.IndirectSet32(x, v)
	} else {
		x.Set(v)
	}
}

func (r *Controller) Init() uint64 {
	paxb := &r.paxb[0]

	pci_num := uint(0)
	if paxb.imap0[0][2].get()&(1<<12) != 0 {
		pci_num = 1
	}

	paxb.pcie_ep_axi_config.set(0)
	paxb.oarr_2[0].upper.set(1 << pci_num)
	paxb.oarr_2[0].lower.set(1 << 0)

	v := paxb.oarr_msi_page[0].get()
	paxb.oarr_msi_page[0].set(v | (1 << 0))

	return uint64(1) << (32 + pci_num)
}
