// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package tomahawk

import (
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/m"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/bcm/internal/sbus"
)

type ecmp_load_balencing_mode uint8

const (
	ecmp_mode_regular_hash ecmp_load_balencing_mode = iota
	ecmp_mode_resilient_hash
	ecmp_mode_random
)

type ecmp_group_entry struct {
	mode  ecmp_load_balencing_mode
	count uint32
	base  uint32
}

func (r *ecmp_group_entry) MemBits() int { return 30 }
func (r *ecmp_group_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint32(&r.count, b, i+10, i, isSet)
	i = m.MemGetSetUint32(&r.base, b, i+15, i, isSet)
	i = m.MemGetSetUint8((*uint8)(&r.mode), b, i+1, i, isSet)
	if i != 29 {
		panic("29")
	}
}

type ecmp_group_mem m.MemElt

func (r *ecmp_group_mem) geta(q *DmaRequest, v *ecmp_group_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ecmp_group_mem) seta(q *DmaRequest, v *ecmp_group_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ecmp_group_mem) get(q *DmaRequest, v *ecmp_group_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *ecmp_group_mem) set(q *DmaRequest, v *ecmp_group_entry) {
	r.seta(q, v, sbus.Duplicate)
}

type ecmp_entry struct {
	is_ecmp bool
	index   uint32
}

func (r *ecmp_entry) MemBits() int { return 18 }
func (r *ecmp_entry) MemGetSet(b []uint32, isSet bool) {
	i := 0
	i = m.MemGetSetUint32(&r.index, b, i+15, i, isSet)
	i = m.MemGetSet1(&r.is_ecmp, b, i, isSet)
}

type ecmp_mem m.MemElt

func (r *ecmp_mem) geta(q *DmaRequest, v *ecmp_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaGet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ecmp_mem) seta(q *DmaRequest, v *ecmp_entry, t sbus.AccessType) {
	(*m.MemElt)(r).MemDmaSet(&q.DmaRequest, v, BlockRxPipe, t)
}
func (r *ecmp_mem) get(q *DmaRequest, v *ecmp_entry) {
	r.geta(q, v, sbus.Duplicate)
}
func (r *ecmp_mem) set(q *DmaRequest, v *ecmp_entry) {
	r.seta(q, v, sbus.Duplicate)
}
