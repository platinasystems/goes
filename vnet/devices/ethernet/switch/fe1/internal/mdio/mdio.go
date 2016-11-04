// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mdio

import (
	"github.com/platinasystems/go/elib/hw"

	"fmt"
)

type Regs struct {
	/* [31:29] cycle
	   [25] 1 => internal, 0 => external
	   [24:22] bus id
	   [21] 1 => clause45, 0 => clause 22
	   [20:16] phy id
	   [15:0] phy write data */
	param hw.Reg32

	// [15:0] phy read data
	read_data hw.Reg32

	// clause 22 [4:0] register address
	// clause 45 [20:16] DTYPE, [15:0] register address
	address hw.Reg32

	// [1] read start, [0] write start
	control hw.Reg32

	// [0] operation done; cleared by clearing control register read/write start bits.
	status hw.Reg32
}

type request struct {
	ExternalPhy bool
	BusID       uint8
	PhyID       uint8
	Address     uint16

	// Pointer to read data if operation is a read.
	// Operation is a write if ReadData is nil.
	ReadData *uint16

	// Data to write if operation is a write.
	WriteData uint16

	// Pointer sent when reply is ready.
	Done chan *request
}

type Mdio struct {
	Regs        *Regs
	requestFifo chan *request
}

func (a *request) start(s *Mdio) {
	r := s.Regs

	isWrite := a.ReadData == nil
	p := uint32(0)
	if isWrite {
		p |= uint32(a.WriteData)
	}
	if !a.ExternalPhy {
		p |= 1 << 25
	}
	if a.PhyID > 0x1f {
		panic(fmt.Errorf("PhyID > 0x1f: %x", a.PhyID))
	}
	p |= uint32(a.PhyID) << 16
	if a.BusID > 0x7 {
		panic(fmt.Errorf("BusID > 0x7: %x", a.BusID))
	}
	p |= uint32(a.BusID) << 22

	r.param.Set(p)
	r.address.Set(uint32(a.Address))
	hw.MemoryBarrier()
	if isWrite {
		r.control.Set(1 << 0)
	} else {
		r.control.Set(1 << 1)
	}
}

func (a *request) finish(s *Mdio) {
	// Fetch request status
	r := s.Regs
	if a.ReadData != nil {
		*a.ReadData = uint16(r.read_data.Get())
	}
	if a.Done != nil {
		a.Done <- a
	}

	// Either start next request or leave hardware idle.
	select {
	case b := <-s.requestFifo:
		b.start(s)
	default:
		r.control.Set(0)
	}
}

func (s *Mdio) DoneInterrupt() {
	select {
	case a := <-s.requestFifo:
		a.finish(s)
	default:
		s.Regs.control.Set(0)
	}
}

func (a *request) do(s *Mdio) {
	if s.requestFifo == nil {
		s.requestFifo = make(chan *request, 64)
	}
	if a.Done == nil {
		a.Done = make(chan *request, 1)
	}
	s.requestFifo <- a
	if len(s.requestFifo) == 1 {
		a.start(s)
	}
}

func (s *Mdio) Read(busId, phyId uint8, a uint16) (v uint16) {
	req := request{
		BusID:    busId,
		PhyID:    phyId,
		Address:  a,
		ReadData: &v,
	}
	req.do(s)
	<-req.Done
	return
}

func (s *Mdio) Write(busId, phyId uint8, a, v uint16) {
	req := request{
		BusID:     busId,
		PhyID:     phyId,
		Address:   a,
		WriteData: v,
	}
	req.do(s)
	<-req.Done
}
