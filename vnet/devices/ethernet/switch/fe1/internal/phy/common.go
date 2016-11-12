// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package phy

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/port"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"unsafe"
)

type Common struct {
	m.Switch
	m.PhyCommon
	m.PhyConfig
	PortBlock *port.PortBlock
	req       DmaRequest
}

func (p *Common) GetPhyCommon() *m.PhyCommon { return &p.PhyCommon }
func (p *Common) dmaReq() *DmaRequest        { p.req.Common = p; return &p.req }

type DmaRequest struct {
	sbus.DmaRequest
	*Common
}

func (q *DmaRequest) Do() {
	s := q.Switch.GetSwitchCommon()
	s.CpuMain.Dma.Do(&q.DmaRequest)
}

// Number of serdes lanes 4 x 25G = 100G
const N_lane = 4

type u16 byte
type u32 [2]u16

func (r *u16) Get(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v *uint16) {
	q.PortBlock.GetPhyU16(&q.DmaRequest, isPmd, lane_mask, r.offset(), v)
}
func (r *u16) Modify(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, write_value, write_mask uint16) {
	q.PortBlock.SetPhyU16(&q.DmaRequest, isPmd, lane_mask, r.offset(), write_value, write_mask)
}
func (r *u16) Set(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v uint16) {
	r.Modify(q, isPmd, lane_mask, v, 0xffff)
}

func (r *u32) Set(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v uint32) {
	q.PortBlock.SetPhyU32(&q.DmaRequest, isPmd, lane_mask, r.offset(), v)
}

func (r *u32) Get(q *DmaRequest, isPmd bool, lane_mask m.LaneMask, v *uint32) {
	q.PortBlock.GetPhyU32(&q.DmaRequest, isPmd, lane_mask, r.offset(), v)
}

func (r *u16) GetSync(b *port.PortBlock, isPmd bool, lane_mask m.LaneMask) uint16 {
	return b.GetPhyU16Sync(isPmd, lane_mask, r.offset())
}
func (r *u16) ModifySync(b *port.PortBlock, isPmd bool, lane_mask m.LaneMask, write_value, write_mask uint16) {
	b.SetPhyU16Sync(isPmd, lane_mask, r.offset(), write_value, write_mask)
}
func (r *u16) SetSync(b *port.PortBlock, isPmd bool, lane_mask m.LaneMask, v uint16) {
	r.ModifySync(b, isPmd, lane_mask, v, 0xffff)
}

// Differentiate PCS/PMD so that we can derive DEVAD bit.
type pcs_u16 u16
type pcs_lane_u16 u16
type pcs_u32 [2]u16
type pmd_lane_u16 u16
type pmd_u32 [2]u16
type pmd_lane_u32 [2]u16
type pad_u16 u16 // for padding

func (r *u16) offset() uint16 { return uint16(uintptr(unsafe.Pointer(r)) - m.BaseAddress) }
func (r *u32) offset() uint16 { return r[0].offset() }

const default_lane_mask = 0x1

func (r *pcs_u16) Get(q *DmaRequest, v *uint16) {
	(*u16)(r).Get(q, false, default_lane_mask, v)
}
func (r *pcs_u16) GetDo(q *DmaRequest) (v uint16) {
	(*u16)(r).Get(q, false, default_lane_mask, &v)
	q.Do()
	return
}
func (r *pcs_u16) Set(q *DmaRequest, v uint16) {
	(*u16)(r).Set(q, false, default_lane_mask, v)
}
func (r *pcs_u16) Modify(q *DmaRequest, v, m uint16) {
	(*u16)(r).Modify(q, false, default_lane_mask, v, m)
}

func (r *pcs_u16) GetSync(b *port.PortBlock) uint16 {
	return (*u16)(r).GetSync(b, false, default_lane_mask)
}
func (r *pcs_u16) SetSync(b *port.PortBlock, v uint16) {
	(*u16)(r).SetSync(b, false, default_lane_mask, v)
}
func (r *pcs_u16) ModifySync(b *port.PortBlock, v, m uint16) {
	(*u16)(r).ModifySync(b, false, default_lane_mask, v, m)
}

func (r *pcs_lane_u16) Get(q *DmaRequest, laneMask m.LaneMask, v *uint16) {
	(*u16)(r).Get(q, false, laneMask, v)
}
func (r *pcs_lane_u16) GetDo(q *DmaRequest, laneMask m.LaneMask) (v uint16) {
	(*u16)(r).Get(q, false, laneMask, &v)
	q.Do()
	return
}
func (r *pcs_lane_u16) Set(q *DmaRequest, laneMask m.LaneMask, v uint16) {
	(*u16)(r).Set(q, false, laneMask, v)
}
func (r *pcs_lane_u16) Modify(q *DmaRequest, laneMask m.LaneMask, v, m uint16) {
	(*u16)(r).Modify(q, false, laneMask, v, m)
}

func (r *pcs_lane_u16) GetSync(b *port.PortBlock, laneMask m.LaneMask) uint16 {
	return (*u16)(r).GetSync(b, false, laneMask)
}
func (r *pcs_lane_u16) SetSync(b *port.PortBlock, laneMask m.LaneMask, v uint16) {
	(*u16)(r).SetSync(b, false, laneMask, v)
}
func (r *pcs_lane_u16) ModifySync(b *port.PortBlock, laneMask m.LaneMask, v, m uint16) {
	(*u16)(r).ModifySync(b, false, laneMask, v, m)
}

func (r *pmd_lane_u16) Get(q *DmaRequest, laneMask m.LaneMask, v *uint16) {
	(*u16)(r).Get(q, true, laneMask, v)
}
func (r *pmd_lane_u16) GetDo(q *DmaRequest, laneMask m.LaneMask) (v uint16) {
	(*u16)(r).Get(q, true, laneMask, &v)
	q.Do()
	return
}
func (r *pmd_lane_u16) Set(q *DmaRequest, laneMask m.LaneMask, v uint16) {
	(*u16)(r).Set(q, true, laneMask, v)
}
func (r *pmd_lane_u16) Modify(q *DmaRequest, laneMask m.LaneMask, v, m uint16) {
	(*u16)(r).Modify(q, true, laneMask, v, m)
}

func (r *pmd_lane_u16) GetSync(b *port.PortBlock, laneMask m.LaneMask) uint16 {
	return (*u16)(r).GetSync(b, true, laneMask)
}
func (r *pmd_lane_u16) SetSync(b *port.PortBlock, laneMask m.LaneMask, v uint16) {
	(*u16)(r).SetSync(b, true, laneMask, v)
}
func (r *pmd_lane_u16) ModifySync(b *port.PortBlock, laneMask m.LaneMask, v, m uint16) {
	(*u16)(r).ModifySync(b, true, laneMask, v, m)
}

func (r *pcs_u32) Get(q *DmaRequest, v *uint32) {
	(*u32)(r).Get(q, false, default_lane_mask, v)
}
func (r *pcs_u32) Set(q *DmaRequest, v uint32) {
	(*u32)(r).Set(q, false, default_lane_mask, v)
}

func (r *pmd_u32) Get(q *DmaRequest, v *uint32) {
	(*u32)(r).Get(q, true, default_lane_mask, v)
}
func (r *pmd_u32) Set(q *DmaRequest, v uint32) {
	(*u32)(r).Set(q, true, default_lane_mask, v)
}

func (r *pmd_lane_u32) Get(q *DmaRequest, laneMask m.LaneMask, v *uint32) {
	(*u32)(r).Get(q, true, laneMask, v)
}
func (r *pmd_lane_u32) Set(q *DmaRequest, laneMask m.LaneMask, v uint32) {
	(*u32)(r).Set(q, true, laneMask, v)
}
