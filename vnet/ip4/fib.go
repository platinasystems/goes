// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"errors"
	"fmt"
)

type Prefix struct {
	Address
	Len uint32
}

var masks = compute_masks()

func compute_masks() (m [33]Address) {
	for l := uint(0); l < uint(len(m)); l++ {
		mask := vnet.Uint32(0)
		if l > 0 {
			mask = (vnet.Uint32(1)<<l - 1) << (32 - l)
		}
		m[l].FromUint32(mask.FromHost())
	}
	return
}

func (a *Address) MaskLen() (l uint, ok bool) {
	m := ^a.AsUint32().ToHost()
	l = ^uint(0)
	if ok = (m+1)&m == 0; ok {
		l = 32
		if m != 0 {
			l -= 1 + elib.Word(m).MinLog2()
		}
	}
	return
}

func (v *Address) MaskedString(r vnet.MaskedStringer) (s string) {
	m := r.(*Address)
	s = v.String() + "/"
	if l, ok := m.MaskLen(); ok {
		s += fmt.Sprintf("%d", l)
	} else {
		s += fmt.Sprintf("%s", m.HexString())
	}
	return
}

func AddressMaskForLen(l uint) Address { return masks[l] }

func (p *Prefix) SetLen(l uint) { p.Len = uint32(l) }
func (a *Address) toPrefix() (p Prefix) {
	p.Address = *a
	return
}

func (p *Prefix) IsEqual(q *Prefix) bool { return p.Len == q.Len && p.Address.IsEqual(&q.Address) }

func (p *Prefix) LessThan(q *Prefix) bool {
	if cmp := p.Address.Diff(&q.Address); cmp != 0 {
		return cmp < 0
	}
	return p.Len < q.Len
}

// Add adds offset to prefix.  For example, 1.2.3.0/24 + 1 = 1.2.4.0/24.
func (p *Prefix) Add(offset uint) (q Prefix) {
	a := p.Address.AsUint32().ToHost()
	a += uint32(offset << (32 - p.Len))
	q = *p
	q.Address.FromUint32(vnet.Uint32(a).FromHost())
	return
}

// True if given destination matches prefix.
func (dst *Address) MatchesPrefix(p *Prefix) bool {
	return 0 == (dst.AsUint32()^p.Address.AsUint32())&p.Mask()
}

func FromIp4Prefix(i *ip.Prefix) (p Prefix) {
	copy(p.Address[:], i.Address[:AddressBytes])
	p.Len = i.Len
	return
}
func (p *Prefix) ToIpPrefix() (i ip.Prefix) {
	copy(i.Address[:], p.Address[:])
	i.Len = p.Len
	return
}

// Maps for prefixes for /0 through /32; key in network byte order.
type mapFib [1 + 32]map[vnet.Uint32]ip.Adj

// Cache of prefix length network masks: entry LEN has high LEN bits set.
// So, 10/8 has top 8 bits set.
var netMasks = computeNetMasks()

func computeNetMasks() (r [33]vnet.Uint32) {
	for i := range netMasks {
		m := ^vnet.Uint32(0)
		if i < 32 {
			m = vnet.Uint32(1<<uint(i)-1) << uint(32-i)
		}
		r[i] = vnet.Uint32(m).FromHost()
	}
	return
}

func (p *Prefix) Mask() vnet.Uint32          { return netMasks[p.Len] }
func (p *Prefix) MaskAsAddress() (a Address) { a.FromUint32(p.Mask()); return }
func (p *Prefix) mapFibKey() vnet.Uint32     { return p.Address.AsUint32() & p.Mask() }

func (m *mapFib) set(p *Prefix, r ip.Adj) (oldAdj ip.Adj, ok bool) {
	l := p.Len
	if m[l] == nil {
		m[l] = make(map[vnet.Uint32]ip.Adj)
	}
	k := p.mapFibKey()
	if oldAdj, ok = m[l][k]; !ok {
		oldAdj = ip.AdjNil
	}
	ok = true // set never fails
	m[l][k] = r
	return
}

func (m *mapFib) unset(p *Prefix) (oldAdj ip.Adj, ok bool) {
	k := p.mapFibKey()
	if oldAdj, ok = m[p.Len][k]; ok {
		delete(m[p.Len], k)
	} else {
		oldAdj = ip.AdjNil
	}
	return
}

func (m *mapFib) get(p *Prefix) (r ip.Adj, ok bool) {
	r, ok = m[p.Len][p.mapFibKey()]
	return
}

func (m *mapFib) lookup(a *Address) ip.Adj {
	p := a.toPrefix()
	for l := 32; l >= 0; l-- {
		if m[l] == nil {
			continue
		}
		p.SetLen(uint(l))
		if r, ok := m[l][p.mapFibKey()]; ok {
			return r
		}
	}
	return ip.AdjMiss
}

// Calls function for each more specific prefix matching given key.
func (m *mapFib) foreachMatchingPrefix(key *Prefix, fn func(p *Prefix, a ip.Adj)) {
	p := Prefix{Address: key.Address}
	for l := key.Len + 1; l <= 32; l++ {
		p.Len = l
		if a, ok := m[l][p.mapFibKey()]; ok {
			fn(&p, a)
		}
	}
}

func (m *mapFib) foreach(fn func(p *Prefix, a ip.Adj)) {
	var p Prefix
	for l := 32; l >= 0; l-- {
		p.Len = uint32(l)
		for k, a := range m[l] {
			p.Address.FromUint32(k)
			fn(&p, a)
		}
	}
}

type Fib struct {
	index ip.FibIndex

	// Map-based fib for general accounting and to maintain mtrie (e.g. setLessSpecific).
	reachable, unreachable mapFib

	// Mtrie for fast lookups.
	mtrie

	remapUnreachablePrefixes []Prefix
}

// Total number of routes in FIB.
func (f *Fib) Len() (n uint) {
	for i := range f.reachable {
		n += uint(len(f.reachable[i]))
		n += uint(len(f.unreachable[i]))
	}
	return
}

type FibAddDelHook func(i ip.FibIndex, p *Prefix, r ip.Adj, isDel bool, isRemap bool)
type IfAddrAddDelHook func(ia ip.IfAddr, isDel bool)

//go:generate gentemplate -id FibAddDelHook -d Package=ip4 -d DepsType=FibAddDelHookVec -d Type=FibAddDelHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id IfAddrAddDelHook -d Package=ip4 -d DepsType=IfAddrAddDelHookVec -d Type=IfAddrAddDelHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl

func (f *Fib) addDel(main *Main, p *Prefix, r ip.Adj, isDel bool) (oldAdj ip.Adj, ok bool) {
	// Call hooks before unset.
	if isDel {
		for i := range main.fibAddDelHooks.hooks {
			main.fibAddDelHooks.Get(i)(f.index, p, r, isDel, false)
		}
	}

	// Add/delete in map fib.
	if isDel {
		oldAdj, ok = f.reachable.unset(p)
	} else {
		oldAdj, ok = f.reachable.set(p, r)
	}

	// Add/delete in mtrie fib.
	m := &f.mtrie

	if len(m.plys) == 0 {
		m.init()
	}

	s := addDelLeaf{
		key:    p.Address,
		keyLen: uint8(p.Len),
		result: r,
	}
	if isDel {
		if p.Len == 0 {
			m.defaultLeaf = emptyLeaf
		} else {
			s.unset(m)
			f.setLessSpecific(p)
		}
	} else {
		if p.Len == 0 {
			m.defaultLeaf = setResult(s.result)
		} else {
			s.set(m)
		}
	}

	// Call hooks after add.
	if !isDel {
		for i := range main.fibAddDelHooks.hooks {
			main.fibAddDelHooks.Get(i)(f.index, p, r, isDel, false)
		}
	}

	return
}

// Find first less specific route matching address and insert into mtrie.
func (f *Fib) setLessSpecific(pʹ *Prefix) (adj ip.Adj, ok bool) {
	// Copy to avoid modifying pʹ.
	p := pʹ.Address.toPrefix()

	// No need to consider length 0 since that's not in mtrie.
	for l := pʹ.Len - 1; l >= 1; l-- {
		if f.reachable[l] == nil {
			continue
		}
		p.Len = l
		k := p.mapFibKey()
		if adj, ok = f.reachable[l][k]; ok {
			s := addDelLeaf{
				result: adj,
				keyLen: uint8(l),
			}
			s.key.FromUint32(k)
			s.set(&f.mtrie)
			return
		}
	}
	return
}

func (f *Fib) mapFibRemapAdjacency(m *Main, from, to ip.Adj) {
	for l := 0; l <= 32; l++ {
		for dst, adj := range f.reachable[l] {
			if adj == from {
				isDel := to == ip.AdjNil
				if isDel {
					delete(f.reachable[l], dst)
				} else {
					f.reachable[l][dst] = to
				}
				p := Prefix{Len: uint32(l)}
				p.Address.FromUint32(dst)
				if isDel {
					f.remapUnreachablePrefixes = append(f.remapUnreachablePrefixes, p)
				}
				for i := range m.fibAddDelHooks.hooks {
					m.fibAddDelHooks.Get(i)(f.index, &p, to, isDel, true)
				}
			}
		}
	}
}

func (m *Main) remapAdjacency(from, to ip.Adj) {
	for _, f := range m.fibs {
		if f == nil {
			continue
		}
		f.mapFibRemapAdjacency(m, from, to)
		f.mtrie.remapAdjacency(from, to)

		// Insert less-specific routes into mtrie for any deleted prefixes.
		if f.remapUnreachablePrefixes != nil {
			for i := range f.remapUnreachablePrefixes {
				p := &f.remapUnreachablePrefixes[i]
				if _, found := f.setLessSpecific(p); !found {
					f.unreachable.set(p, ip.AdjNil)
				}
			}
			// Save prefix slice for next call.
			f.remapUnreachablePrefixes = f.remapUnreachablePrefixes[:0]
		}
	}
}

func (f *Fib) Get(p *Prefix) (a ip.Adj, ok bool) {
	if a, ok = f.reachable[p.Len][p.mapFibKey()]; !ok {
		a = ip.AdjNil
	}
	return
}

func (f *Fib) Add(m *Main, p *Prefix, r ip.Adj) (ip.Adj, bool) { return f.addDel(m, p, r, true) }
func (f *Fib) Del(m *Main, p *Prefix) (ip.Adj, bool)           { return f.addDel(m, p, ip.AdjMiss, false) }
func (f *Fib) Lookup(a *Address) (r ip.Adj) {
	r = f.mtrie.lookup(a)
	return
}

func (m *Main) setInterfaceAdjacency(a *ip.Adjacency, si vnet.Si, ia ip.IfAddr) {
	sw := m.Vnet.SwIf(si)
	hw := m.Vnet.SupHwIf(sw)
	h := m.Vnet.HwIfer(hw.Hi())

	next := ip.LookupNextRewrite
	noder := &m.rewriteNode
	packetType := vnet.IP4

	if _, ok := h.(vnet.Arper); ok {
		next = ip.LookupNextGlean
		noder = &m.arpNode
		packetType = vnet.ARP
		a.IfAddr = ia
	}

	a.LookupNextIndex = next
	m.Vnet.SetRewrite(&a.Rewrite, si, noder, packetType, nil /* dstAdr meaning broadcast */)
}

type fibMain struct {
	fibs FibVec
	// Hooks to call on set/unset.
	fibAddDelHooks      FibAddDelHookVec
	ifRouteAdjIndexBySi map[vnet.Si]ip.Adj
}

//go:generate gentemplate -d Package=ip4 -id Fib -d VecType=FibVec -d Type=*Fib github.com/platinasystems/go/elib/vec.tmpl

func (m *fibMain) RegisterFibAddDelHook(f FibAddDelHook, dep ...*dep.Dep) {
	m.fibAddDelHooks.Add(f, dep...)
}

func (m *Main) fibByIndex(i ip.FibIndex, create bool) (f *Fib) {
	m.fibs.Validate(uint(i))
	if create && m.fibs[i] == nil {
		m.fibs[i] = &Fib{index: i}
	}
	f = m.fibs[i]
	return
}

func (m *Main) fibById(id ip.FibId, create bool) *Fib {
	var (
		i  ip.FibIndex
		ok bool
	)
	if i, ok = m.FibIndexForId(id); !ok {
		i = ip.FibIndex(m.fibs.Len())
	}
	return m.fibByIndex(i, create)
}

func (m *Main) fibBySi(si vnet.Si) *Fib {
	i := m.FibIndexForSi(si)
	return m.fibByIndex(i, true)
}

func (m *Main) validateDefaultFibForSi(si vnet.Si) {
	i := m.ValidateFibIndexForSi(si)
	m.fibByIndex(i, true)
}

func (m *Main) getRoute(p *ip.Prefix, si vnet.Si) (ai ip.Adj, as []ip.Adjacency, ok bool) {
	f := m.fibBySi(si)
	q := FromIp4Prefix(p)
	ai, ok = f.Get(&q)
	if ok {
		as = m.GetAdj(ai)
	}
	return
}

func (m *Main) GetRoute(p *Prefix, si vnet.Si) (ai ip.Adj, ok bool) {
	f := m.fibBySi(si)
	ai, ok = f.Get(p)
	return
}

func (m *Main) getRouteFibIndex(p *ip.Prefix, fi ip.FibIndex) (ai ip.Adj, ok bool) {
	f := m.fibByIndex(fi, false)
	q := FromIp4Prefix(p)
	ai, ok = f.Get(&q)
	return
}

func (m *Main) addDelRoute(p *ip.Prefix, fi ip.FibIndex, newAdj ip.Adj, isDel bool) (oldAdj ip.Adj, err error) {
	createFib := !isDel
	f := m.fibByIndex(fi, createFib)
	q := FromIp4Prefix(p)
	var ok bool
	oldAdj, ok = f.addDel(m, &q, newAdj, isDel)
	if !ok {
		err = fmt.Errorf("prefix %s not found", &q)
	}
	return
}

type NextHop struct {
	Address Address
	Si      vnet.Si
	Weight  ip.NextHopWeight
}

func (x *NextHop) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*vnet.Vnet)
	if !in.Parse("%v %v", &x.Si, v, &x.Address) {
		panic(parse.ErrInput)
	}
	x.Weight = 1
	in.Parse("weight %d", &x.Weight)
}

var ErrNextHopNotFound = errors.New("next hop not found")

func (m *Main) AddDelRouteNextHop(p *Prefix, nh *NextHop, isDel bool) (err error) {
	f := m.fibBySi(nh.Si)

	var (
		nhAdj, oldAdj, newAdj ip.Adj
		adjs                  []ip.Adjacency
		ok                    bool
	)

	if !isDel && p.Len == 32 && p.Address.IsEqual(&nh.Address) {
		err = fmt.Errorf("prefix %s matches next-hop %s", p, &nh.Address)
		return
	}

	// Zero address means interface next hop.
	if nh.Address.IsZero() && !isDel {
		if nhAdj, ok = m.ifRouteAdjIndexBySi[nh.Si]; !ok {
			nhAdj, adjs = m.NewAdj(1)
			m.setInterfaceAdjacency(&adjs[0], nh.Si, ip.IfAddrNil)
			m.CallAdjAddHooks(nhAdj)
			if m.ifRouteAdjIndexBySi == nil {
				m.ifRouteAdjIndexBySi = make(map[vnet.Si]ip.Adj)
			}
			m.ifRouteAdjIndexBySi[nh.Si] = nhAdj
		}
	} else {
		if nhAdj, ok = f.Get(&Prefix{Address: nh.Address, Len: 32}); !ok {
			err = ErrNextHopNotFound
			return
		}
	}

	oldAdj, ok = f.Get(p)
	if isDel && !ok {
		err = fmt.Errorf("unknown destination %s", p)
		return
	}

	if newAdj, ok = m.AddDelNextHop(oldAdj, isDel, nhAdj, nh.Weight); !ok {
		err = fmt.Errorf("requested next-hop %s not found in multipath", &nh.Address)
		return
	}

	if oldAdj != newAdj {
		// Only remove from fib on delete of final adjacency.
		isFibDel := isDel
		if isFibDel && newAdj != ip.AdjNil {
			isFibDel = false
		}
		f.addDel(m, p, newAdj, isFibDel)
	}

	return
}

func (f *Fib) deleteMatchingRoutes(m *Main, key *Prefix) {
	f.reachable.foreachMatchingPrefix(key, func(p *Prefix, a ip.Adj) {
		f.Del(m, p)
	})
}

func (f *Fib) addDelReplace(m *Main, p *Prefix, r ip.Adj, isDel bool) {
	if oldAdj, ok := f.addDel(m, p, r, isDel); ok && oldAdj != ip.AdjNil {
		m.DelAdj(oldAdj)
	}
}

func (m *Main) addDelInterfaceRoutes(ia ip.IfAddr, isDel bool) {
	ifa := m.GetIfAddr(ia)
	si := ifa.Si
	sw := m.Vnet.SwIf(si)
	hw := m.Vnet.SupHwIf(sw)
	fib := m.fibBySi(si)
	p := FromIp4Prefix(&ifa.Prefix)

	// Add interface's prefix as route tied to glean adjacency (arp for Ethernet).
	// Suppose interface has address 1.1.1.1/8; here we add 1.0.0.0/8 tied to glean adjacency.
	if p.Len < 32 {
		addDelAdj := ip.AdjNil
		if !isDel {
			ai, as := m.NewAdj(1)
			m.setInterfaceAdjacency(&as[0], si, ia)
			m.CallAdjAddHooks(ai)
			addDelAdj = ai
		}
		fib.addDelReplace(m, &p, addDelAdj, isDel)
		ifa.NeighborProbeAdj = addDelAdj
	}

	// Add 1.1.1.1/32 as a local address.
	{
		addDelAdj := ip.AdjNil
		if !isDel {
			ai, as := m.NewAdj(1)
			as[0].LookupNextIndex = ip.LookupNextLocal
			as[0].IfAddr = ia
			as[0].Si = si
			as[0].SetMaxPacketSize(hw)
			m.CallAdjAddHooks(ai)
			addDelAdj = ai
		}
		p.Len = 32
		fib.addDelReplace(m, &p, addDelAdj, isDel)
	}

	if isDel {
		fib.deleteMatchingRoutes(m, &p)
	}
}

func (m *Main) AddDelInterfaceAddress(si vnet.Si, addr *Prefix, isDel bool) (err error) {
	if !isDel {
		err = m.ForeachIfAddress(si, func(ia ip.IfAddr, ifa *ip.IfAddress) (err error) {
			p := FromIp4Prefix(&ifa.Prefix)
			if !p.IsEqual(addr) && (addr.Address.MatchesPrefix(&p) || p.Address.MatchesPrefix(addr)) {
				err = fmt.Errorf("%s: add %s conflicts with existing address %s", si.Name(m.Vnet), addr, &p)
			}
			return
		})
		if err != nil {
			return
		}
	}

	var (
		ia     ip.IfAddr
		exists bool
	)

	sw := m.Vnet.SwIf(si)
	isUp := sw.IsAdminUp()
	pa := addr.ToIpPrefix()

	// If interface is admin up, delete interface routes *before* removing address.
	if isUp && isDel {
		ia, exists = m.Main.IfAddrForPrefix(&pa, si)
		// For non-existing prefixes error will be signalled by AddDelInterfaceAddress below.
		if exists {
			m.addDelInterfaceRoutes(ia, isDel)
		}
	}

	// Delete interface address.  Return error if deleting non-existent address.
	if ia, exists, err = m.Main.AddDelInterfaceAddress(si, &pa, isDel); err != nil {
		return
	}

	// If interface is up add interface routes.
	if isUp && !isDel && !exists {
		m.addDelInterfaceRoutes(ia, isDel)
	}

	// Do callbacks when new address is created or old one is deleted.
	if isDel || !exists {
		for i := range m.ifAddrAddDelHooks.hooks {
			m.ifAddrAddDelHooks.Get(i)(ia, isDel)
		}
	}

	return
}

func (m *Main) swIfAdminUpDown(v *vnet.Vnet, si vnet.Si, isUp bool) (err error) {
	m.validateDefaultFibForSi(si)
	m.ForeachIfAddress(si, func(ia ip.IfAddr, ifa *ip.IfAddress) (err error) {
		isDel := !isUp
		m.addDelInterfaceRoutes(ia, isDel)
		return
	})
	return
}
