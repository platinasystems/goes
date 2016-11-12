// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package mdio

import (
	"github.com/platinasystems/go/elib/hw"

	"fmt"
)

type Controller struct {
	param hw.U32

	read_data hw.U32

	address hw.U32

	control hw.U32

	status hw.U32
}

type request struct {
	ExternalPhy bool
	BusID       uint8
	PhyID       uint8
	Address     uint16

	ReadData *uint16

	WriteData uint16

	Done chan *request
}

type Mdio struct {
	Controller  *Controller
	requestFifo chan *request
}

func (a *request) start(s *Mdio) {
	r := s.Controller

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
	r := s.Controller
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
		s.Controller.control.Set(0)
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
