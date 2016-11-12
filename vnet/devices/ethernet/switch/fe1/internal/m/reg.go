// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package m

import (
	"github.com/platinasystems/go/elib/hw"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"

	"unsafe"
)

var (
	BasePointer = hw.BasePointer
	BaseAddress = hw.BaseAddress
)

type Gu32 byte
type Gu64 byte
type Pu32 byte
type Pu64 byte

const Log2NPorts = 8

type U32 [1 << Log2NPorts]Gu32
type U64 [1 << Log2NPorts]Gu64
type PortU32 [1 << Log2NPorts]Pu32
type PortU64 [1 << Log2NPorts]Pu64

func (r *Gu32) Offset() uint { return uint(uintptr(unsafe.Pointer(r)) - BaseAddress) }
func (r *Gu64) Offset() uint { return (*Gu32)(r).Offset() }
func (r *Pu32) Offset() uint { return (*Gu32)(r).Offset() }
func (r *Pu64) Offset() uint { return (*Gu32)(r).Offset() }

func (r *U32) Offset() uint { return r[0].Offset() }
func (r *U64) Offset() uint { return r[0].Offset() }

func (r *Pu32) Address() sbus.Address { return sbus.Address(r.Offset()) | sbus.PortReg }
func (r *U32) Address() sbus.Address  { return sbus.Address(r.Offset()) | sbus.GenReg }
func (r *Pu64) Address() sbus.Address { return sbus.Address(r.Offset()) | sbus.PortReg }
func (r *U64) Address() sbus.Address  { return sbus.Address(r.Offset()) | sbus.GenReg }

func (r *U32) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint32) {
	q.GetU32(v, b, a|r.Address(), c)
}
func (r *U32) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint32) {
	q.SetU32(v, b, a|r.Address(), c)
}

func (r *U64) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint64) {
	q.GetU64(v, b, a|r.Address(), c)
}
func (r *U64) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint64) {
	q.SetU64(v, b, a|r.Address(), c)
}

func (r *Pu32) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint32) {
	q.GetU32(v, b, a|r.Address(), c)
}
func (r *Pu32) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint32) {
	q.SetU32(v, b, a|r.Address(), c)
}

func (r *Pu64) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint64) {
	q.GetU64(v, b, a|r.Address(), c)
}
func (r *Pu64) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint64) {
	q.SetU64(v, b, a|r.Address(), c)
}

func (r *Gu32) address() sbus.Address { return sbus.Address(r.Offset()) | sbus.GenReg }
func (r *Gu64) address() sbus.Address { return sbus.Address(r.Offset()) | sbus.GenReg }

func (r *Gu32) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint32) {
	q.GetU32(v, b, a|r.address(), c)
}
func (r *Gu32) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint32) {
	q.SetU32(v, b, a|r.address(), c)
}

func (r *Gu64) Get(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v *uint64) {
	q.GetU64(v, b, a|r.address(), c)
}
func (r *Gu64) Set(q *sbus.DmaRequest, a sbus.Address, b sbus.Block, c sbus.AccessType, v uint64) {
	q.SetU64(v, b, a|r.address(), c)
}
