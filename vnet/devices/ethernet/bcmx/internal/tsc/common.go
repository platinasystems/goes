// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tsc

import (
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/port"
	"github.com/platinasystems/vnetdevices/ethernet/switch/bcm/internal/sbus"

	"unsafe"
)

type Tsc struct {
	m.Switch
	m.PhyCommon
	m.PhyConfig
	PortBlock *port.PortBlock
	req       DmaRequest
}

func (phy *Tsc) GetPhyCommon() *m.PhyCommon { return &phy.PhyCommon }

func (phy *Tsc) dmaReq() *DmaRequest  { phy.req.Tsc = phy; return &phy.req }
func (phy *Tscf) dmaReq() *DmaRequest { return phy.Tsc.dmaReq() }
func (phy *Tsce) dmaReq() *DmaRequest { return phy.Tsc.dmaReq() }

type DmaRequest struct {
	sbus.DmaRequest
	*Tsc
}

func (q *DmaRequest) Do() {
	s := q.Tsc.Switch.GetSwitchCommon()
	s.Dma.Do(&q.DmaRequest)
}

const (
	// Number of serdes lanes 4 x 25G = 100G
	N_lane = 4
)

type reg byte
type reg32 [2]reg

func (r *reg) Get(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v *uint16) {
	q.PortBlock.GetPhyReg(&q.DmaRequest, isPmd, lane_mask, r.offset(), v)
}
func (r *reg) Modify(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, write_value, write_mask uint16) {
	q.PortBlock.SetPhyReg(&q.DmaRequest, isPmd, lane_mask, r.offset(), write_value, write_mask)
}
func (r *reg) Set(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v uint16) {
	r.Modify(q, isPmd, lane_mask, v, 0xffff)
}

func (r *reg32) Set(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v uint32) {
	q.PortBlock.SetPhyReg32(&q.DmaRequest, isPmd, lane_mask, r.offset(), v)
}

func (r *reg32) Get(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v *uint32) {
	q.PortBlock.GetPhyReg32(&q.DmaRequest, isPmd, lane_mask, r.offset(), v)
}

func (r *reg) GetSync(b *port.PortBlock, isPmd bool, lane_mask m.LaneMask) uint16 {
	return b.GetPhyRegSync(isPmd, lane_mask, r.offset())
}
func (r *reg) ModifySync(b *port.PortBlock, isPmd bool, lane_mask m.LaneMask, write_value, write_mask uint16) {
	b.SetPhyRegSync(isPmd, lane_mask, r.offset(), write_value, write_mask)
}
func (r *reg) SetSync(b *port.PortBlock, isPmd bool, lane_mask m.LaneMask, v uint16) {
	r.ModifySync(b, isPmd, lane_mask, v, 0xffff)
}

// Differentiate PCS(reg0)/PMD(reg1) registers so that we can derive DEVAD.
type pcs_reg reg
type pcs_lane_reg reg
type pcs_reg_32 [2]reg
type pmd_lane_reg reg
type pmd_reg_32 [2]reg
type pmd_lane_reg_32 [2]reg
type pad_reg reg // for padding

func (r *reg) offset() uint16   { return uint16(uintptr(unsafe.Pointer(r)) - m.RegsBaseAddress) }
func (r *reg32) offset() uint16 { return r[0].offset() }

const default_lane_mask = 0x1

func (r *pcs_reg) Get(q *DmaRequest, v *uint16) {
	(*reg)(r).Get(q, false, default_lane_mask, v)
}
func (r *pcs_reg) GetDo(q *DmaRequest) (v uint16) {
	(*reg)(r).Get(q, false, default_lane_mask, &v)
	q.Do()
	return
}
func (r *pcs_reg) Set(q *DmaRequest, v uint16) {
	(*reg)(r).Set(q, false, default_lane_mask, v)
}
func (r *pcs_reg) Modify(q *DmaRequest, v, m uint16) {
	(*reg)(r).Modify(q, false, default_lane_mask, v, m)
}

func (r *pcs_reg) GetSync(b *port.PortBlock) uint16 {
	return (*reg)(r).GetSync(b, false, default_lane_mask)
}
func (r *pcs_reg) SetSync(b *port.PortBlock, v uint16) {
	(*reg)(r).SetSync(b, false, default_lane_mask, v)
}
func (r *pcs_reg) ModifySync(b *port.PortBlock, v, m uint16) {
	(*reg)(r).ModifySync(b, false, default_lane_mask, v, m)
}

func (r *pcs_lane_reg) Get(q *DmaRequest, laneMask m.LaneMask, v *uint16) {
	(*reg)(r).Get(q, false, laneMask, v)
}
func (r *pcs_lane_reg) GetDo(q *DmaRequest, laneMask m.LaneMask) (v uint16) {
	(*reg)(r).Get(q, false, laneMask, &v)
	q.Do()
	return
}
func (r *pcs_lane_reg) Set(q *DmaRequest, laneMask m.LaneMask, v uint16) {
	(*reg)(r).Set(q, false, laneMask, v)
}
func (r *pcs_lane_reg) Modify(q *DmaRequest, laneMask m.LaneMask, v, m uint16) {
	(*reg)(r).Modify(q, false, laneMask, v, m)
}

func (r *pcs_lane_reg) GetSync(b *port.PortBlock, laneMask m.LaneMask) uint16 {
	return (*reg)(r).GetSync(b, false, laneMask)
}
func (r *pcs_lane_reg) SetSync(b *port.PortBlock, laneMask m.LaneMask, v uint16) {
	(*reg)(r).SetSync(b, false, laneMask, v)
}
func (r *pcs_lane_reg) ModifySync(b *port.PortBlock, laneMask m.LaneMask, v, m uint16) {
	(*reg)(r).ModifySync(b, false, laneMask, v, m)
}

func (r *pmd_lane_reg) Get(q *DmaRequest, laneMask m.LaneMask, v *uint16) {
	(*reg)(r).Get(q, true, laneMask, v)
}
func (r *pmd_lane_reg) GetDo(q *DmaRequest, laneMask m.LaneMask) (v uint16) {
	(*reg)(r).Get(q, true, laneMask, &v)
	q.Do()
	return
}
func (r *pmd_lane_reg) Set(q *DmaRequest, laneMask m.LaneMask, v uint16) {
	(*reg)(r).Set(q, true, laneMask, v)
}
func (r *pmd_lane_reg) Modify(q *DmaRequest, laneMask m.LaneMask, v, m uint16) {
	(*reg)(r).Modify(q, true, laneMask, v, m)
}

func (r *pmd_lane_reg) GetSync(b *port.PortBlock, laneMask m.LaneMask) uint16 {
	return (*reg)(r).GetSync(b, true, laneMask)
}
func (r *pmd_lane_reg) SetSync(b *port.PortBlock, laneMask m.LaneMask, v uint16) {
	(*reg)(r).SetSync(b, true, laneMask, v)
}
func (r *pmd_lane_reg) ModifySync(b *port.PortBlock, laneMask m.LaneMask, v, m uint16) {
	(*reg)(r).ModifySync(b, true, laneMask, v, m)
}

func (r *pcs_reg_32) Get(q *DmaRequest, v *uint32) {
	(*reg32)(r).Get(q, false, default_lane_mask, v)
}
func (r *pcs_reg_32) Set(q *DmaRequest, v uint32) {
	(*reg32)(r).Set(q, false, default_lane_mask, v)
}

func (r *pmd_reg_32) Get(q *DmaRequest, v *uint32) {
	(*reg32)(r).Get(q, true, default_lane_mask, v)
}
func (r *pmd_reg_32) Set(q *DmaRequest, v uint32) {
	(*reg32)(r).Set(q, true, default_lane_mask, v)
}

func (r *pmd_lane_reg_32) Get(q *DmaRequest, laneMask m.LaneMask, v *uint32) {
	(*reg32)(r).Get(q, true, laneMask, v)
}
func (r *pmd_lane_reg_32) Set(q *DmaRequest, laneMask m.LaneMask, v uint32) {
	(*reg32)(r).Set(q, true, laneMask, v)
}
