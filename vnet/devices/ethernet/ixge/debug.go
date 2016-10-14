// +build debug

package ixge

import (
	"github.com/platinasystems/go/elib/hw"

	"unsafe"
)

func check(tag string, p unsafe.Pointer, expect uint) {
	hw.CheckRegAddr(tag, uint(uintptr(p)-hw.RegsBaseAddress), expect)
}

// Validate memory map.
func init() {
	r := (*regs)(hw.RegsBasePointer)
	check("pf_0", unsafe.Pointer(&r.pf_0), 0x700)
	check("interrupt", unsafe.Pointer(&r.interrupt), 0x800)
	check("rx_dma0", unsafe.Pointer(&r.rx_dma0[0]), 0x1000)
	check("rx_dma_control", unsafe.Pointer(&r.rx_dma_control), 0x2f00)
	check("rx_enable", unsafe.Pointer(&r.rx_enable), 0x3000)
	check("ge_mac", unsafe.Pointer(&r.ge_mac), 0x4200)
	check("xge_mac", unsafe.Pointer(&r.xge_mac), 0x4240)
	check("pf_mailbox", unsafe.Pointer(&r.pf_mailbox), 0x4b00)
	check("checksum_control", unsafe.Pointer(&r.checksum_control), 0x5000)
	check("pf_virtual_control", unsafe.Pointer(&r.pf_virtual_control), 0x51b0)
	check("tx_dma", unsafe.Pointer(&r.tx_dma[0]), 0x6000)
	check("flexible_filters", unsafe.Pointer(&r.flexible_filters), 0x9000)
	check("vlan_filter", unsafe.Pointer(&r.vlan_filter), 0xa000)
	check("rx_dma1", unsafe.Pointer(&r.rx_dma1[0]), 0xd000)
	check("ethernet_type_queue_select", unsafe.Pointer(&r.ethernet_type_queue_select[0]), 0xec00)
	check("fcoe_redirection", unsafe.Pointer(&r.fcoe_redirection), 0xed00)
	check("flow_director", unsafe.Pointer(&r.flow_director), 0xee00)
	check("pf_1", unsafe.Pointer(&r.pf_1), 0xf000)
	check("eeprom_flash_control", unsafe.Pointer(&r.eeprom_flash_control), 0x10010)
	check("pcie", unsafe.Pointer(&r.pcie), 0x11000)
	check("sfp_i2c", unsafe.Pointer(&r.sfp_i2c), 0x15f58)
}
