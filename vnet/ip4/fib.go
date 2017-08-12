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

type mapFibResult struct {
	adj ip.Adj
	nh  mapFibResultNextHop
}

// Maps for prefixes for /0 through /32; key in network byte order.
type MapFib [1 + 32]map[vnet.Uint32]mapFibResult

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
func (a *Address) Mask(l uint) (v Address) {
	v.FromUint32(a.AsUint32() & netMasks[l])
	return
}

func (m *MapFib) validateLen(l uint32) {
	if m[l] == nil {
		m[l] = make(map[vnet.Uint32]mapFibResult)
	}
}
func (m *MapFib) Set(p *Prefix, newAdj ip.Adj) (oldAdj ip.Adj, ok bool) {
	l := p.Len
	m.validateLen(l)
	k := p.mapFibKey()
	var r mapFibResult
	if r, ok = m[l][k]; !ok {
		oldAdj = ip.AdjNil
	} else {
		oldAdj = r.adj
	}
	ok = true // set never fails
	r.adj = newAdj
	m[l][k] = r
	return
}

func (m *MapFib) Unset(p *Prefix) (oldAdj ip.Adj, ok bool) {
	k := p.mapFibKey()
	var r mapFibResult
	if r, ok = m[p.Len][k]; ok {
		oldAdj = r.adj
		delete(m[p.Len], k)
	} else {
		oldAdj = ip.AdjNil
	}
	return
}

func (m *MapFib) Get(p *Prefix) (r mapFibResult, ok bool) {
	r, ok = m[p.Len][p.mapFibKey()]
	return
}

func (m *MapFib) Lookup(a Address) (r mapFibResult, p Prefix, ok bool) {
	p = a.toPrefix()
	for l := 32; l >= 0; l-- {
		if m[l] == nil {
			continue
		}
		p.SetLen(uint(l))
		k := p.mapFibKey()
		if r, ok = m[l][k]; ok {
			p.Address.FromUint32(k)
			return
		}
	}
	r = mapFibResult{adj: ip.AdjMiss}
	p = Prefix{}
	return
}

func (f *MapFib) lookupReachable(m *Main, a Address) (r mapFibResult, p Prefix, ok bool) {
	if r, p, ok = f.Lookup(a); ok {
		as := m.GetAdj(r.adj)
		// Anything that is not a Glean (e.g. matching an interface's address) is "reachable".
		for i := range as {
			if ok = !as[i].IsGlean(); !ok {
				break
			}
		}
	}
	return
}

// Calls function for each more specific prefix matching given key.
func (m *MapFib) foreachMatchingPrefix(key *Prefix, fn func(p *Prefix, r mapFibResult)) {
	p := Prefix{Address: key.Address}
	for l := key.Len + 1; l <= 32; l++ {
		p.Len = l
		if r, ok := m[l][p.mapFibKey()]; ok {
			fn(&p, r)
		}
	}
}

func (m *MapFib) foreach(fn func(p *Prefix, r mapFibResult)) {
	var p Prefix
	for l := 32; l >= 0; l-- {
		p.Len = uint32(l)
		for k, r := range m[l] {
			p.Address.FromUint32(k)
			fn(&p, r)
		}
	}
}

func (m *MapFib) reset() {
	for i := range m {
		m[i] = nil
	}
}

func (m *MapFib) clean(fi ip.FibIndex) {
	for i := range m {
		for _, r := range m[i] {
			for dst, dstMap := range r.nh {
				for dp := range dstMap {
					if dp.i == fi {
						delete(dstMap, dp)
					}
				}
				if len(dstMap) == 0 {
					delete(r.nh, dst)
				}
			}
		}
	}
}

type Fib struct {
	index ip.FibIndex

	// Map-based fib for general accounting and to maintain mtrie (e.g. setLessSpecific).
	reachable, unreachable MapFib

	// Mtrie for fast lookups.
	mtrie
}

//go:generate gentemplate -d Package=ip4 -id Fib -d VecType=FibVec -d Type=*Fib github.com/platinasystems/go/elib/vec.tmpl

// Total number of routes in FIB.
func (f *Fib) Len() (n uint) {
	for i := range f.reachable {
		n += uint(len(f.reachable[i]))
	}
	return
}

type IfAddrAddDelHook func(ia ip.IfAddr, isDel bool)

//go:generate gentemplate -id FibAddDelHook -d Package=ip4 -d DepsType=FibAddDelHookVec -d Type=FibAddDelHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id IfAddrAddDelHook -d Package=ip4 -d DepsType=IfAddrAddDelHookVec -d Type=IfAddrAddDelHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl

func (f *Fib) addDel(main *Main, p *Prefix, r ip.Adj, isDel bool) (oldAdj ip.Adj, ok bool) {
	if isDel {
		// Call hooks before delete.
		main.callFibAddDelHooks(f.index, p, r, isDel)
		f.addDelReachable(main, p, r, isDel)
	}

	// Add/delete in map fib.
	if isDel {
		oldAdj, ok = f.reachable.Unset(p)
	} else {
		oldAdj, ok = f.reachable.Set(p, r)
	}

	// Add/delete in mtrie fib.
	m := &f.mtrie

	if len(m.plys) == 0 {
		m.init()
	}

	s := addDelLeaf{
		key:    p.Address.Mask(uint(p.Len)),
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
		main.callFibAddDelHooks(f.index, p, r, isDel)
		f.addDelReachable(main, p, r, isDel)
	}

	return
}

type NextHopper interface {
	ip.AdjacencyFinalizer
	NextHopFibIndex(m *Main) ip.FibIndex
	NextHopWeight() ip.NextHopWeight
}

type idst struct {
	a Address
	i ip.FibIndex
}

type ipre struct {
	p Prefix
	i ip.FibIndex
}

type mapFibResultNextHop map[idst]map[ipre]NextHopper

func (x *mapFibResult) addDelNextHop(m *Main, pf *Fib, p Prefix, a Address, r NextHopper, isDel bool) {
	id := idst{a: a, i: r.NextHopFibIndex(m)}
	ip := ipre{p: p, i: pf.index}
	if isDel {
		delete(x.nh[id], ip)
		if len(x.nh[id]) == 0 {
			delete(x.nh, id)
		}
	} else {
		if x.nh == nil {
			x.nh = make(map[idst]map[ipre]NextHopper)
		}
		if x.nh[id] == nil {
			x.nh[id] = make(map[ipre]NextHopper)
		}
		x.nh[id][ip] = r
	}
}

func (x *mapFibResultNextHop) String() (s string) {
	for a, m := range *x {
		for p, w := range m {
			s += fmt.Sprintf("  %v %v x %d\n", &p, &a.a, w)
		}
	}
	return
}

func (f *Fib) setReachable(m *Main, p *Prefix, pf *Fib, via *Prefix, a Address, r NextHopper, isDel bool) {
	va, vl := via.Address.AsUint32(), via.Len
	x := f.reachable[vl][va]
	x.addDelNextHop(m, pf, *p, a, r, isDel)
	f.reachable[vl][va] = x
}

// Delete more specific reachable with less specific reachable.
func (less *mapFibResult) replaceWithLessSpecific(m *Main, f *Fib, more *mapFibResult) {
	for dst, dstMap := range more.nh {
		// Move all destinations from more -> less.
		delete(more.nh, dst)
		if less.nh == nil {
			less.nh = make(map[idst]map[ipre]NextHopper)
		}
		less.nh[dst] = dstMap
		// Replace adjacencies: more -> less.
		for dp, r := range dstMap {
			g := m.fibByIndex(dp.i, false)
			g.replaceNextHop(m, &dp.p, f, more.adj, less.adj, dst.a, r)
		}
	}
}

func (x *mapFibResult) delReachableVia(m *Main, f *Fib) {
	for dst, dstMap := range x.nh {
		delete(x.nh, dst)
		for dp, r := range dstMap {
			g := m.fibByIndex(dp.i, false)
			const isDel = true
			g.addDelRouteNextHop(m, &dp.p, dst.a, r, isDel)
			// Prefix is now unreachable.
			f.addDelUnreachable(m, &dp.p, g, dst.a, r, !isDel, false)
		}
	}
}

func (less *mapFibResult) replaceWithMoreSpecific(m *Main, f *Fib, p *Prefix, adj ip.Adj, more *mapFibResult) {
	for dst, dstMap := range less.nh {
		if dst.a.MatchesPrefix(p) {
			delete(less.nh, dst)
			for dp, r := range dstMap {
				const isDel = false
				g := m.fibByIndex(dp.i, false)
				more.addDelNextHop(m, g, dp.p, dst.a, r, isDel)
				g.replaceNextHop(m, &dp.p, f, less.adj, adj, dst.a, r)
			}
		}
	}
	f.reachable[p.Len][p.mapFibKey()] = *more
}

func (r *mapFibResult) makeReachable(m *Main, f *Fib, p *Prefix, adj ip.Adj) {
	for dst, dstMap := range r.nh {
		if dst.a.MatchesPrefix(p) {
			delete(r.nh, dst)
			for dp, r := range dstMap {
				g := m.fibByIndex(dp.i, false)
				const isDel = false
				g.addDelRouteNextHop(m, &dp.p, dst.a, r, isDel)
			}
		}
	}
}

func (x *mapFibResult) addUnreachableVia(m *Main, f *Fib, p *Prefix) {
	for dst, dstMap := range x.nh {
		if dst.a.MatchesPrefix(p) {
			delete(x.nh, dst)
			for dp, r := range dstMap {
				g := m.fibByIndex(dp.i, false)
				const isDel = false
				f.addDelUnreachable(m, &dp.p, g, dst.a, r, isDel, false)
			}
		}
	}
}

func (f *Fib) addDelReachable(m *Main, p *Prefix, a ip.Adj, isDel bool) {
	r, _ := f.reachable.Get(p)
	// Look up less specific reachable route for prefix.
	lr, _, lok := f.reachable.getLessSpecific(p)
	if isDel {
		if lok {
			lr.replaceWithLessSpecific(m, f, &r)
		} else {
			r.delReachableVia(m, f)
		}
	} else {
		if lok {
			lr.replaceWithMoreSpecific(m, f, p, a, &r)
		}
		if r, _, ok := f.unreachable.Lookup(p.Address); ok {
			r.makeReachable(m, f, p, a)
		}
	}
}

func (f *Fib) addDelUnreachable(m *Main, p *Prefix, pf *Fib, a Address, r NextHopper, isDel bool, recurse bool) (err error) {
	nr, np, _ := f.unreachable.Lookup(a)
	if isDel && recurse {
		nr.delReachableVia(m, f)
	}
	if !isDel && recurse {
		nr.addUnreachableVia(m, f, p)
	}
	nr.addDelNextHop(m, pf, *p, a, r, isDel)
	f.unreachable.validateLen(np.Len)
	nr.adj = ip.AdjNil
	f.unreachable[np.Len][np.mapFibKey()] = nr
	return
}

// Find first less specific route matching address and insert into mtrie.
func (f *MapFib) getLessSpecific(pʹ *Prefix) (r mapFibResult, p Prefix, ok bool) {
	p = pʹ.Address.toPrefix()

	// No need to consider length 0 since that's not in mtrie.
	for l := pʹ.Len - 1; l >= 1; l-- {
		if f[l] == nil {
			continue
		}
		p.Len = l
		k := p.mapFibKey()
		if r, ok = f[l][k]; ok {
			return
		}
	}
	return
}

// Find first less specific route matching address and insert into mtrie.
func (f *Fib) setLessSpecific(pʹ *Prefix) (r mapFibResult, p Prefix, ok bool) {
	r, p, ok = f.reachable.getLessSpecific(pʹ)
	if ok {
		s := addDelLeaf{
			result: r.adj,
			keyLen: uint8(p.Len),
		}
		s.key = p.Address
		s.set(&f.mtrie)
	}
	return
}

func (f *Fib) Get(p *Prefix) (a ip.Adj, ok bool) {
	var r mapFibResult
	a = ip.AdjNil
	if r, ok = f.reachable[p.Len][p.mapFibKey()]; ok {
		a = r.adj
	}
	return
}

func (f *Fib) Add(m *Main, p *Prefix, r ip.Adj) (ip.Adj, bool) { return f.addDel(m, p, r, true) }
func (f *Fib) Del(m *Main, p *Prefix) (ip.Adj, bool)           { return f.addDel(m, p, ip.AdjMiss, false) }
func (f *Fib) Lookup(a *Address) (r ip.Adj) {
	r = f.mtrie.lookup(a)
	return
}
func (m *Main) Lookup(a *Address, i ip.FibIndex) (r ip.Adj) {
	f := m.fibByIndex(i, true)
	return f.Lookup(a)
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
		a.Index = uint32(ia)
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

type FibAddDelHook func(i ip.FibIndex, p *Prefix, r ip.Adj, isDel bool)

func (m *fibMain) RegisterFibAddDelHook(f FibAddDelHook, dep ...*dep.Dep) {
	m.fibAddDelHooks.Add(f, dep...)
}

func (m *fibMain) callFibAddDelHooks(fi ip.FibIndex, p *Prefix, r ip.Adj, isDel bool) {
	for i := range m.fibAddDelHooks.hooks {
		m.fibAddDelHooks.Get(i)(fi, p, r, isDel)
	}
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
		err = fmt.Errorf("prefix %v not found", q)
	}
	return
}

type NextHop struct {
	Address Address
	Si      vnet.Si
	Weight  ip.NextHopWeight
}

func (n *NextHop) NextHopWeight() ip.NextHopWeight     { return n.Weight }
func (n *NextHop) NextHopFibIndex(m *Main) ip.FibIndex { return m.FibIndexForSi(n.Si) }
func (n *NextHop) FinalizeAdjacency(a *ip.Adjacency)   {}

func (x *NextHop) ParseWithArgs(in *parse.Input, args *parse.Args) {
	v := args.Get().(*vnet.Vnet)
	switch {
	case in.Parse("%v %v", &x.Si, v, &x.Address):
	default:
		panic(fmt.Errorf("expecting INTERFACE ADDRESS; got %s", in))
	}
	x.Weight = 1
	in.Parse("weight %d", &x.Weight)
}

type prefixError struct {
	s string
	p Prefix
}

func (e *prefixError) Error() string { return e.s + ": " + e.p.String() }

func (m *Main) AddDelRouteNextHop(p *Prefix, nh *NextHop, isDel bool) (err error) {
	f := m.fibBySi(nh.Si)
	return f.addDelRouteNextHop(m, p, nh.Address, nh, isDel)
}

func (f *Fib) addDelRouteNextHop(m *Main, p *Prefix, nha Address, nhr NextHopper, isDel bool) (err error) {
	var (
		nhAdj, oldAdj, newAdj ip.Adj
		ok                    bool
	)

	if !isDel && nha.MatchesPrefix(p) {
		err = fmt.Errorf("prefix %s matches next-hop %s", p, &nha)
		return
	}

	nhf := m.fibByIndex(nhr.NextHopFibIndex(m), true)

	var reachable_via_prefix Prefix
	if r, np, found := nhf.reachable.lookupReachable(m, nha); found {
		nhAdj = r.adj
		reachable_via_prefix = np
	} else {
		const recurse = true
		err = nhf.addDelUnreachable(m, p, f, nha, nhr, isDel, recurse)
		return
	}

	oldAdj, ok = f.Get(p)
	if isDel && !ok {
		err = &prefixError{s: "unknown destination", p: *p}
		return
	}

	if oldAdj == nhAdj && isDel {
		newAdj = ip.AdjNil
	} else if newAdj, ok = m.AddDelNextHop(oldAdj, nhAdj, nhr.NextHopWeight(), nhr, isDel); !ok {
		err = fmt.Errorf("requested next-hop %s not found in multipath", &nha)
		return
	}

	if oldAdj != newAdj {
		// Only remove from fib on delete of final adjacency.
		isFibDel := isDel
		if isFibDel && newAdj != ip.AdjNil {
			isFibDel = false
		}
		f.addDel(m, p, newAdj, isFibDel)
		nhf.setReachable(m, p, f, &reachable_via_prefix, nha, nhr, isDel)
	}

	return
}

func (f *Fib) replaceNextHop(m *Main, p *Prefix, pf *Fib, fromNextHopAdj, toNextHopAdj ip.Adj, nha Address, r NextHopper) (err error) {
	if adj, ok := f.Get(p); !ok {
		err = &prefixError{s: "unknown destination", p: *p}
	} else {
		as := m.GetAdj(toNextHopAdj)
		// If replacement is glean (interface route) then next hop becomes unreachable.
		isDel := len(as) == 1 && as[0].IsGlean()
		if isDel {
			err = pf.addDelRouteNextHop(m, p, nha, r, isDel)
			if err == nil {
				err = f.addDelUnreachable(m, p, pf, nha, r, !isDel, false)
			}
		} else {
			if err = m.ReplaceNextHop(adj, fromNextHopAdj, toNextHopAdj, r); err != nil {
				err = fmt.Errorf("replace next hop: %v", err)
			} else {
				m.callFibAddDelHooks(pf.index, p, adj, isDel)
			}
		}
	}
	if err != nil {
		panic(err)
	}
	return
}

func (f *Fib) deleteMatchingRoutes(m *Main, key *Prefix) {
	f.reachable.foreachMatchingPrefix(key, func(p *Prefix, r mapFibResult) {
		f.Del(m, p)
	})
}

func (f *Fib) addDelReplace(m *Main, p *Prefix, r ip.Adj, isDel bool) {
	if oldAdj, ok := f.addDel(m, p, r, isDel); ok && oldAdj != ip.AdjNil {
		m.DelAdj(oldAdj)
	}
}

func (m *Main) addDelInterfaceAddressRoutes(ia ip.IfAddr, isDel bool) {
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
			as[0].Index = uint32(ia)
			as[0].Si = si
			if hw != nil {
				as[0].SetMaxPacketSize(hw)
			}
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
			m.addDelInterfaceAddressRoutes(ia, isDel)
		}
	}

	// Delete interface address.  Return error if deleting non-existent address.
	if ia, exists, err = m.Main.AddDelInterfaceAddress(si, &pa, isDel); err != nil {
		return
	}

	// If interface is up add interface routes.
	if isUp && !isDel && !exists {
		m.addDelInterfaceAddressRoutes(ia, isDel)
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
		m.addDelInterfaceAddressRoutes(ia, isDel)
		return
	})
	return
}

func (f *Fib) Reset() {
	f.reachable.reset()
	f.unreachable.reset()
	f.mtrie.reset()
}

func (m *Main) FibReset(fi ip.FibIndex) {
	for i := range m.fibs {
		if i != int(fi) && m.fibs[i] != nil {
			m.fibs[i].reachable.clean(fi)
			m.fibs[i].unreachable.clean(fi)
		}
	}

	f := m.fibByIndex(fi, true)
	f.Reset()
}
