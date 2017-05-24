// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/vnet"

	"fmt"
)

type IfAddr uint32

const IfAddrNil = ^IfAddr(0)

//go:generate gentemplate -d Package=ip -id IfAddr -d VecType=IfAddrVec -d Type=IfAddr github.com/platinasystems/go/elib/vec.tmpl

type IfAddress struct {
	// ip4/ip6 address and map key plus length.
	Prefix Prefix

	// Interface which has this address.
	Si vnet.Si

	NeighborProbeAdj Adj

	// Next and previous pointers in doubly-linked list of interface addresses for this interface.
	next, prev IfAddr
}

func (i IfAddr) String(m *Main) string {
	a := m.GetIfAddr(i)
	return a.Si.Name(m.v)
}

//go:generate gentemplate -d Package=ip -id ifaddress -d PoolType=ifAddressPool -d Type=IfAddress -d Data=ifAddrs github.com/platinasystems/go/elib/pool.tmpl

type ifAddressMain struct {
	ifAddressPool

	// Maps ip4/ip6 address to pool index.
	addrMap map[Address]IfAddr

	// Head of doubly-linked list indexed by software interface.
	headBySwIf IfAddrVec
}

func (m *ifAddressMain) swIfAddDel(v *vnet.Vnet, si vnet.Si, isDel bool) (err error) {
	m.headBySwIf.ValidateInit(uint(si), IfAddrNil)
	return
}

func (m *ifAddressMain) init(v *vnet.Vnet) {
	v.RegisterSwIfAddDelHook(m.swIfAddDel)
}

func (m *ifAddressMain) GetIfAddress(a []uint8) (ia *IfAddress) {
	var k Address
	copy(k[:], a)
	if i, ok := m.addrMap[k]; ok {
		ia = &m.ifAddrs[i]
	}
	return
}
func (m *ifAddressMain) GetIfAddr(i IfAddr) *IfAddress                 { return &m.ifAddrs[i] }
func (m *ifAddressMain) IfFirstAddr(i vnet.Si) IfAddr                  { return m.headBySwIf[i] }
func (m *ifAddressMain) IfFirstAddress(i vnet.Si) *IfAddress           { return m.GetIfAddr(m.IfFirstAddr(i)) }
func (m *ifAddressMain) IfAddressForAdjacency(a *Adjacency) *IfAddress { return m.GetIfAddr(a.IfAddr) }

func (m *ifAddressMain) ForeachIfAddress(si vnet.Si, f func(ia IfAddr, i *IfAddress) error) error {
	i := m.headBySwIf[si]
	for i != IfAddrNil {
		ia := m.GetIfAddr(i)
		if err := f(i, ia); err != nil {
			return err
		}
		i = ia.next
	}
	return nil
}

func (m *Main) IfAddrForPrefix(p *Prefix) (ai IfAddr, exists bool) {
	ai, exists = m.addrMap[p.Address]
	return
}

func (m *Main) AddDelInterfaceAddress(si vnet.Si, p *Prefix, isDel bool) (ai IfAddr, exists bool, err error) {
	var a *IfAddress
	if ai, exists = m.addrMap[p.Address]; exists {
		a = m.GetIfAddr(ai)
	}

	if isDel {
		if a == nil {
			err = fmt.Errorf("%s: address %s not found", si.Name(m.v), p.String(m))
			return
		}
		if a.prev != IfAddrNil {
			prev := m.GetIfAddr(a.prev)
			prev.next = a.next
		} else {
			// Delete list head.
			m.headBySwIf[si] = IfAddrNil
		}
		if a.next != IfAddrNil {
			next := m.GetIfAddr(a.next)
			next.prev = a.prev
		}

		delete(m.addrMap, p.Address)
		m.ifAddressPool.PutIndex(uint(ai))
		ai = IfAddrNil
	} else if a == nil {
		ai = IfAddr(m.ifAddressPool.GetIndex())
		a = m.GetIfAddr(ai)

		if m.addrMap == nil {
			m.addrMap = make(map[Address]IfAddr)
		}
		m.addrMap[p.Address] = ai
		a.Prefix = *p
		a.Si = si
		a.NeighborProbeAdj = AdjNil

		pi := m.headBySwIf[si]
		a.next = IfAddrNil
		a.prev = pi

		// Make previous head point to added element and set added element as new head.
		if pi != IfAddrNil {
			p := m.GetIfAddr(pi)
			a.next = pi
			p.prev = ai
		}
		m.headBySwIf[si] = ai
	}
	return
}
