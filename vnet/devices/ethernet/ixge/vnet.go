package ixge

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
)

type vnet_dev struct {
	vnet.InterfaceNode
	ethernet.Interface
	ethIfConfig ethernet.InterfaceConfig
}

func (d *dev) IsUnix() bool { return true }

func (d *dev) vnetInit() {
	v := d.m.Vnet

	d.Next = []string{
		rx_next_error:                    "error",
		rx_next_punt:                     "punt",
		rx_next_ethernet_input:           "ethernet-input",
		rx_next_ip4_input_valid_checksum: "ip4-input-valid-checksum",
		rx_next_ip6_input:                "ip6-input",
	}
	d.Errors = []string{
		rx_error_none:                 "no error",
		rx_error_ip4_invalid_checksum: "invalid ip4 checksum",
	}

	a := d.pciDev.Addr
	ethernet.RegisterInterface(v, d, &d.ethIfConfig, "ixge%d-%d-%d", a.Bus, a.Slot, a.Fn)
	v.RegisterInterfaceNode(d, d.Hi(), d.Name())
}

func (d *dev) SetLoopback(x vnet.IfLoopbackType) (err error) {
	const (
		force_link_up = 1 << 0
		mac_loopback  = 1 << 15
	)
	switch x {
	case vnet.IfLoopbackMac:
		d.regs.xge_mac.mac_control.or(d, force_link_up)
		d.regs.xge_mac.control.or(d, mac_loopback)
	case vnet.IfLoopbackNone:
		d.regs.xge_mac.control.andnot(d, mac_loopback)
		d.regs.xge_mac.mac_control.andnot(d, force_link_up)
	default:
		return vnet.ErrNotSupported
	}

	return
}

func (d *dev) ValidateSpeed(speed vnet.Bandwidth) (err error) {
	return
}
