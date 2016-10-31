package iproc

import (
	"github.com/platinasystems/go/elib/hw"
	"unsafe"
)

type reg hw.Reg32

func (r *reg) get() uint32  { return (*hw.Reg32)(r).Get() }
func (r *reg) set(v uint32) { (*hw.Reg32)(r).Set(v) }

type addr_map struct {
	// lower [31:1] address [0] valid
	// upper [31:0] address [63:32]
	lower, upper reg
}

type Regs struct {
	_ [0x2000]byte

	// PCIE <-> AXI bridge
	paxb [2]struct {
		clock_control               reg
		rc_power_management_control reg
		rc_power_management_status  reg
		ep_power_management_control reg
		ep_power_management_status  reg
		ep_ltr_control              reg
		ep_ltr_status               reg
		_                           [0x20 - 0x1c]byte
		ep_obff_status              reg
		pcie_error_status           reg
		_                           [0x30 - 0x28]byte

		endianness              reg
		apb_timeout_count       reg
		tx_arbitration_priority reg
		_                       [0x40 - 0x3c]byte

		read_completion_buffer_init_start reg
		read_completion_buffer_init_done  reg
		_                                 [0x100 - 0x48]byte

		pcie_rc_axi_config reg

		// [13:9] awuser config 0xf
		// [4:0] aruser config 0xf
		pcie_ep_axi_config reg

		pcie_rx_debug_status  reg
		pcie_rx_debug_control reg
		_                     [0x120 - 0x110]byte

		config_indirect_address reg
		config_indirect_data    reg
		_                       [0x1f8 - 0x128]byte

		config_address    reg
		config_data       reg
		pcie_sys_eq_page  reg
		pcie_sys_msi_page reg
		_                 [0x210 - 0x208]byte

		pcie_sys_msi_control [6]reg
		_                    [0x250 - 0x228]byte

		pcie_sys_eq_pointers   [6]struct{ head, tail reg }
		pcie_sys_eq_tail_early [6]reg
		_                      [0x2a0 - 0x298]byte

		pcie_sys_eq_overwritten [6]reg
		_                       [0x2c0 - 0x2b8]byte

		pcie_sys_eq_page_upper  reg
		pcie_sys_msi_page_upper reg
		_                       [0x330 - 0x2c8]byte

		pcie_sys_rc_intx_enable reg
		pcie_sys_rc_intx_status reg
		_                       [0x340 - 0x338]byte

		pcie_sys_msi_request           reg
		pcie_sys_host_interrupt_enable reg
		pcie_sys_host_interrupt_status reg
		_                              [0x350 - 0x34c]byte

		pcie_sys_host_mailbox [4]reg

		// functions 0 & 1
		pcie_sys_ep_interrupt_enable [2]reg
		_                            [0x370 - 0x368]byte

		pcie_sys_ep_interrupt_status [2]reg
		_                            [0x380 - 0x378]byte

		cmicd_to_pcie_interrupt_enable reg
		_                              [0xc00 - 0x384]byte

		// 8 maps for 2 pci functions for bar0
		// 31:12 address
		// 11:8 awcache
		// 7:4 arcache
		// 3:1 reserved
		// 0:0 valid
		imap0 [2][8]reg

		// 3:0 address bits [35:32]
		imap0_upper [2][8]reg
		_           [0xcc0 - 0xc80]byte

		imap2 [2]addr_map

		func0_imap0_regs_type reg
		_                     [0xd00 - 0xcd4]byte

		iarr_2 [3]addr_map
		_      [0xd20 - 0xd18]byte

		oarr_0 [2]addr_map
		_      [0xd34 - 0xd30]byte

		// [31:12] address 0x19030 [0] valid
		oarr_msi_page [2]reg
		_             [0xd40 - 0xd3c]byte

		pcie_outbound_64bit_address_mapping_table0 [2]addr_map

		oarr_msi_page_upper [2]reg
		iarr_2_size         [2]reg

		// Outbound 36b iPROC Address Filter provisioned for 4GB CMICd; 4GB CMICd memory page number mapped into the PCIE space
		oarr_2 [1]addr_map

		omap_2 [1]addr_map

		imap1 [2][8]addr_map // mapping table for 8 1MB iproc memory pages for bar 1 functions 0 & 1
		_     [0xf00 - 0xdf0]byte

		mem_control                    reg
		mem_ecc_error_log              [2]reg
		pcie_link_status               reg
		strap_status                   reg
		reset_status                   reg
		reset_enable_in_pcie_link_down reg
		misc_interrupt_enable          reg
		tx_debug_config                reg
		misc_config                    reg
		misc_status                    reg
		_                              [0xf30 - 0xf2c]byte

		interrupt_enable reg
		interrupt_clear  reg
		interrupt_status reg
		_                [0xf40 - 0xf3c]byte

		apb_error_enable_for_cfg_rd_completions reg
		_                                       [0x1000 - 0xf44]byte
	}

	_ [0x7000 - 0x4000]byte

	window7 [1024]reg

	smbus0 I2cRegs
	_      [0xb000 - 0x8050]byte
	smbus1 I2cRegs
	_      [0xc000 - 0xb050]byte
}

// Indirect addressing above 0x8000 limit of PCI BAR.
func (r *Regs) iGetSet(x *hw.Reg32, v *uint32, isSet bool) {
	paxb := &r.paxb[0]
	addr := uint32(uintptr(unsafe.Pointer(x)) - uintptr(unsafe.Pointer(r)))
	page, offset := addr&^0xfff, (addr&0xfff)/4

	// valid bit plus page
	paxb.imap0[0][7].set(1<<0 | page)
	if isSet {
		r.window7[offset].set(*v)
	} else {
		*v = r.window7[offset].get()
	}
}

func (r *Regs) IndirectGet32(x *hw.Reg32) (v uint32) { r.iGetSet(x, &v, false); return }
func (r *Regs) IndirectSet32(x *hw.Reg32, v uint32)  { r.iGetSet(x, &v, true) }

type Reg32 hw.Reg32

func (r *Reg32) Get(regs *Regs) uint32 {
	x := (*hw.Reg32)(r)
	if regs != nil {
		return regs.IndirectGet32(x)
	} else {
		return x.Get()
	}
}

func (r *Reg32) Set(regs *Regs, v uint32) {
	x := (*hw.Reg32)(r)
	if regs != nil {
		regs.IndirectSet32(x, v)
	} else {
		x.Set(v)
	}
}

func (r *Regs) Init() uint64 {
	paxb := &r.paxb[0]

	pci_num := uint(0)
	if paxb.imap0[0][2].get()&(1<<12) != 0 {
		pci_num = 1
	}

	paxb.pcie_ep_axi_config.set(0)
	paxb.oarr_2[0].upper.set(1 << pci_num) // hi bits 0x1 for pci0; 0x2 for pci1
	paxb.oarr_2[0].lower.set(1 << 0)       // set valid bit

	v := paxb.oarr_msi_page[0].get()
	paxb.oarr_msi_page[0].set(v | (1 << 0)) // set valid bit to enable msi interrupt page

	return uint64(1) << (32 + pci_num)
}
