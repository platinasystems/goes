// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package cpu

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet"

	"fmt"
)

type miim_controller struct {
	rate_adjust_external_mdio hw.U32
	rate_adjust_internal_mdio hw.U32

	control hw.U32

	status hw.U32

	auto_scan_address hw.U32

	pause_address hw.U32

	port_link_is_down_0 [3]hw.U32
	rx_pause_status_0   [3]hw.U32
	tx_pause_status_0   [3]hw.U32

	port_enable_link_scan_0  [3]hw.U32
	port_enable_pause_scan_0 [3]hw.U32

	port_scan_is_clause45_0 [3]hw.U32

	port_phy_is_internal_0 [3]hw.U32

	port_bus_index_0 [10]hw.U32

	port_phy_address_0 [24]hw.U32

	rx_pause_capability_0       [3]hw.U32
	rx_pause_override_control_0 [3]hw.U32
	tx_pause_capability_0       [3]hw.U32
	tx_pause_override_control_0 [3]hw.U32

	clear_scan_status hw.U32

	port_link_is_down_1         [1]hw.U32
	rx_pause_status_1           [1]hw.U32
	tx_pause_status_1           [1]hw.U32
	port_enable_link_scan_1     [1]hw.U32
	port_enable_pause_scan_1    [1]hw.U32
	port_scan_is_clause45_1     [1]hw.U32
	port_phy_is_internal_1      [1]hw.U32
	port_bus_index_1            [4]hw.U32
	port_phy_address_1          [8]hw.U32
	rx_pause_capability_1       [1]hw.U32
	rx_pause_override_control_1 [1]hw.U32
	tx_pause_capability_1       [1]hw.U32
	tx_pause_override_control_1 [1]hw.U32

	config hw.U32

	port_link_is_down_2         [2]hw.U32
	rx_pause_status_2           [2]hw.U32
	tx_pause_status_2           [2]hw.U32
	port_enable_link_scan_2     [2]hw.U32
	port_enable_pause_scan_2    [2]hw.U32
	port_scan_is_clause45_2     [2]hw.U32
	port_phy_is_internal_2      [2]hw.U32
	port_bus_index_2            [7]hw.U32
	port_phy_address_2          [16]hw.U32
	rx_pause_capability_2       [2]hw.U32
	rx_pause_override_control_2 [2]hw.U32
	tx_pause_capability_2       [2]hw.U32
	tx_pause_override_control_2 [2]hw.U32
}

type linkStatusChanger interface {
	LinkStatusChange(v *LinkStatus)
}

type linkScanMain struct {
	validMask LinkStatus

	changer linkStatusChanger
}

func i32(i uint16) (uint32, uint32) {
	return uint32(i / 32), uint32(1) << (i % 32)
}

func findAddr(i0 uint32, r0 *[3]hw.U32, r1 *[1]hw.U32, r2 *[2]hw.U32) (r *hw.U32) {
	switch {
	case i0 < 3:
		r = &r0[i0]
	case i0 < 4:
		r = &r1[i0-3]
	case i0 < 6:
		r = &r2[i0-4]
	default:
		panic("port index")
	}
	return
}

func (c *miim_controller) set_internal(i uint16, isExternal bool) {
	i0, i1 := i32(i)
	r := findAddr(i0, &c.port_phy_is_internal_0, &c.port_phy_is_internal_1, &c.port_phy_is_internal_2)
	v := r.Get()
	if isExternal {
		v &^= i1
	} else {
		v |= i1
	}
	r.Set(v)
}

func (c *miim_controller) set_enable(i uint16, enable bool) {
	i0, i1 := i32(i)
	r := findAddr(i0, &c.port_enable_link_scan_0, &c.port_enable_link_scan_1, &c.port_enable_link_scan_2)
	v := r.Get()
	if enable {
		v |= i1
	} else {
		v &^= i1
	}
	r.Set(v)
}

func (c *miim_controller) set_clause45_enable(i uint16, enable bool) {
	i0, i1 := i32(i)
	r := findAddr(i0, &c.port_scan_is_clause45_0, &c.port_scan_is_clause45_1, &c.port_scan_is_clause45_2)
	v := r.Get()
	if enable {
		v |= i1
	} else {
		v &^= i1
	}
	r.Set(v)
}

func (c *miim_controller) set_phy_id(i uint16, id uint8) {
	i0, i1 := uint32(i/4), uint32(5*(i%4))
	var r *hw.U32
	switch {
	case i0 < 24:
		r = &c.port_phy_address_0[i0]
	case i0 < 32:
		r = &c.port_phy_address_1[i0-24]
	case i0 < 48:
		r = &c.port_phy_address_2[i0-32]
	default:
		panic("port index")
	}
	v := r.Get()
	m := uint32(0x1f) << i1
	v = (v &^ m) | uint32(id)<<i1
	r.Set(v)
}

func (c *miim_controller) set_phy_bus_id(i uint16, id uint8) {
	i0, i1 := uint32(i/10), uint32(3*(i%10))
	var r *hw.U32
	switch {
	case i0 < 10:
		r = &c.port_bus_index_0[i0]
	case i0 < 14:
		r = &c.port_bus_index_1[i0-10]
	case i0 < 21:
		r = &c.port_bus_index_2[i0-14]
	default:
		panic("port index")
	}
	v := r.Get()
	m := uint32(0x7) << i1
	v = (v &^ m) | uint32(id)<<i1
	r.Set(v)
}

func (c *Main) MdioInit(coreFreqInHz float64, ch linkStatusChanger) {
	c.changer = ch
	c.setMdioFreq(coreFreqInHz)
}

func (c *Main) LinkScanEnable(vn *vnet.Vnet, enable bool) {
	if defaultLinkStatusNode.Vnet != vn {
		vn.RegisterNode(defaultLinkStatusNode, "fe1-link-status")
	}

	r := &c.controller.miim

	const link_scan_enable = 1 << 0

	v := r.control.Get()
	if enable {
		v |= link_scan_enable
	} else {
		v &^= link_scan_enable
	}
	r.control.Set(v)
}

type LinkStatus [6]uint32

const LinkStatusWordBits = 32

func (c *Main) getLinkStatus(v *LinkStatus) {
	r := &c.controller.miim
	lm := &c.linkScanMain
	for i := range r.port_link_is_down_0 {
		v[i+0] = r.port_link_is_down_0[i].Get()
	}
	for i := range r.port_link_is_down_1 {
		v[i+3] = r.port_link_is_down_1[i].Get()
	}
	for i := range r.port_link_is_down_2 {
		v[i+4] = r.port_link_is_down_2[i].Get()
	}
	// Hardware puts spurious 1s in unused bits; clear them.
	for i := range v {
		v[i] &= lm.validMask[i]
	}
	return
}

type linkStatusNode struct{ vnet.Node }

func (n *linkStatusNode) EventHandler() {}

var defaultLinkStatusNode = &linkStatusNode{}

type linkStatusEvent struct {
	c *Main
	v LinkStatus
}

func (e *linkStatusEvent) EventAction()   { e.c.changer.LinkStatusChange(&e.v) }
func (e *linkStatusEvent) String() string { return fmt.Sprintf("fe1 link status change %x", e.v) }

func (c *Main) LinkStatusChangeInterrupt() {
	r := &c.controller.miim
	r.clear_scan_status = 1 << 4
	e := &linkStatusEvent{c: c}
	c.getLinkStatus(&e.v)
	defaultLinkStatusNode.AddEvent(e, defaultLinkStatusNode)
	if elog.Enabled() {
		e := linkStatusElogEvent{LinkStatus: e.v}
		e.Log()
	}
}

func (c *Main) PauseStatusChangeInterrupt() {
	panic("not yet")
}

// Event logging.
type linkStatusElogEvent struct {
	LinkStatus
}

func (e *linkStatusElogEvent) String() string { return fmt.Sprintf("link-change %x", e.LinkStatus) }
func (e *linkStatusElogEvent) Encode(b []byte) int {
	i := 0
	for j := range e.LinkStatus {
		i += elog.EncodeUint32(b[i:], e.LinkStatus[j])
	}
	return i
}
func (e *linkStatusElogEvent) Decode(b []byte) (i int) {
	for j := range e.LinkStatus {
		e.LinkStatus[j], i = elog.DecodeUint32(b, i)
	}
	return
}

//go:generate gentemplate -d Package=cpu.-id linkStatusElogEvent -d Type=linkStatusElogEvent github.com/platinasystems/go/elib/elog/event.tmpl

type LinkScanPort struct {
	IsExternal bool
	Enable     bool
	// true => clause 45; false clause 22
	IsClause45 bool
	PhyId      uint8
	PhyBusId   uint8
	Index      uint16
}

func (c *Main) LinkScanAdd(p *LinkScanPort) {
	r := &c.controller.miim
	r.set_internal(p.Index, p.IsExternal)
	r.set_phy_id(p.Index, p.PhyId)
	r.set_phy_bus_id(p.Index, p.PhyBusId)
	r.set_clause45_enable(p.Index, p.IsClause45)
	r.set_enable(p.Index, p.Enable)
	lm := &c.linkScanMain
	i0, i1 := p.Index/32, uint32(1)<<(p.Index%32)
	if p.Enable {
		lm.validMask[i0] |= i1
	} else {
		lm.validMask[i0] &^= i1
	}
}

func (c *Main) setMdioFreq(coreFreqInHz float64) {
	r := &c.controller.miim

	// Set external mdio frequency to ~2Mhz
	{
		target := 2e6
		divisor := uint32((coreFreqInHz + (2*target - 1)) / (2 * target))
		dividend := uint32(1)
		r.rate_adjust_external_mdio.Set(dividend<<16 | divisor)
	}

	// Set external mdio frequency to ~12Mhz
	{
		target := 12e6
		divisor := uint32((coreFreqInHz + (2*target - 1)) / (2 * target))
		dividend := uint32(1)
		r.rate_adjust_internal_mdio.Set(dividend<<16 | divisor)
	}

	// Set MDIO output delay
	{
		v := uint32(10)
		r.config.Set(v)
	}
}
