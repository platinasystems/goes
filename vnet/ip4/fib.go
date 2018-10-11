// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"fmt"
	"sync"

	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"
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

func (p *Prefix) Matches(q *Prefix) bool {
	return p.Address.AsUint32()&p.Mask() == q.Address.AsUint32()&q.Mask()
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

var cached struct {
	masks struct {
		once sync.Once
		val  interface{}
	}
}

// Cache of prefix length network masks: entry LEN has high LEN bits set.
// So, 10/8 has top 8 bits set.
func netMask(i uint) vnet.Uint32 {
	const nmasks = 33
	cached.masks.once.Do(func() {
		masks := make([]vnet.Uint32, nmasks)
		for i := range masks {
			m := ^vnet.Uint32(0)
			if i < 32 {
				m = vnet.Uint32(1<<uint(i)-1) << uint(32-i)
			}
			masks[i] = vnet.Uint32(m).FromHost()
		}
		cached.masks.val = masks
	})
	if i < nmasks {
		return cached.masks.val.([]vnet.Uint32)[i]
	}
	return 0
}

func (p *Prefix) Mask() vnet.Uint32          { return netMask(uint(p.Len)) }
func (p *Prefix) MaskAsAddress() (a Address) { a.FromUint32(p.Mask()); return }
func (p *Prefix) mapFibKey() vnet.Uint32     { return p.Address.AsUint32() & p.Mask() }
func (p *Prefix) ApplyMask() (q *Prefix) {
	pm := Prefix{}
	pm.Address.FromUint32(p.Address.AsUint32() & p.Mask())
	pm.Len = p.Len
	q = &pm
	return
}
func (a *Address) Mask(l uint) (v Address) {
	v.FromUint32(a.AsUint32() & netMask(l))
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

// Reachable means that all next-hop adjacencies are rewrites.
func (f *MapFib) lookupReachable(m *Main, a Address) (r mapFibResult, p Prefix, reachable, err bool) {
	if r, p, reachable = f.Lookup(a); reachable {
		as := m.GetAdj(r.adj)
		for i := range as {
			err = as[i].IsLocal()
			reachable = as[i].IsRewrite()
			if !reachable {
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
		if vnet.AdjDebug {
			fmt.Printf("fib.go addDel %v prefix %v adj %v delete: call fe1 hooks, addDelReachable\n", f.index.Name(&main.Main), p.String(), r)
		}
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
		if vnet.AdjDebug {
			fmt.Printf("fib.go addDel %v prefix %v adj %v add: call fe1 hooks, addDelReachable\n", f.index.Name(&main.Main), p.String(), r)
		}
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

// idst is the destination or nh address and namespace
// ipre is the prefix that has idst as its nh
type mapFibResultNextHop map[idst]map[ipre]NextHopper

func (x *mapFibResult) addDelNextHop(m *Main, pf *Fib, p Prefix, a Address, r NextHopper, isDel bool) {
	id := idst{a: a, i: r.NextHopFibIndex(m)}
	ip := ipre{p: p, i: pf.index}
	if vnet.AdjDebug {
		fmt.Printf("fib.go addDelNextHop isDel %v id %v ip %v: before %v\n", isDel, id, ip, x)
	}
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
	if vnet.AdjDebug {
		fmt.Printf("fib.go addDelNextHop: after %v\n", x)
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
	if vnet.AdjDebug {
		fmt.Printf("fib.go setReachable isDel %v prefix %v via %v nha %v adj %v, new mapFibResult %v\n", isDel, p.String(), via.String(), a, x.adj, x)
	}
}

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
	// x is the mapFibResult from reachable (i.e. x is the reachable MapFib)
	// delReachableVia will traverse the map and remove x's address from all the prefixes that uses it as its nexthop address, and add them to unreachable
	// This is also called from addDelUnreachable (i.e. x is the unreachable MapFib) when doing recursive delete; not sure what the purpose is...
	if vnet.AdjDebug {
		fmt.Printf("fib.go delReachableVia adj %v IsMpAdj %v mapFibResult before: %v\n", x.adj, m.IsMpAdj(x.adj), x)
	}
	for dst, dstMap := range x.nh {
		// dstMap is map of prefixes that uses dst as its nh
		// For each of them, remove nh from prefix and add to unreachable
		for dp, r := range dstMap {
			g := m.fibByIndex(dp.i, false)
			const isDel = true
			ai, ok := g.Get(&dp.p) // Get gets from g.reachable
			if !ok || ai == ip.AdjNil || ai == ip.AdjMiss {
				return
			}
			if m.IsMpAdj(ai) {
				// ai is a mpAdj, use addDelRouteNextHop to delete
				// for mpAdj, addDelRouteNextHop will remove dst from x.nh's map as part of the cleanup and accounting
				// if len(x.nh) ends up 0 after, it will remove x.nh
				g.addDelRouteNextHop(m, &dp.p, dst.a, r, isDel)
			} else {
				// ai is either local, glean, or adjacency from neighbor
				// use addDel directly to delete adjacency
				as := m.GetAdj(ai)
				adjType := "no adjacency found at that adj index"
				if len(as) > 0 {
					adjType = as[0].LookupNextIndex.String()
					// addDel will not remove dst from x.nh's map automatically
					g.addDel(m, &dp.p, ai, true)
				} else {
					fmt.Printf("DEBUG: fib.go delReachableVia: attempt to remove nh %v from prefix %v but unexpected old adjacency %v type %v",
						dst.a, dp.p.String(), ai, adjType)
				}
			}
			// Prefix is now unreachable, add to unreachable, no recurse
			f.addDelUnreachable(m, &dp.p, g, dst.a, r, !isDel, false)
		}
		// Verify that x.nh[id] is not already deleted as part of the cleanup; and if not, delete it
		if x.nh[dst] != nil {
			// delete the etry from x's map
			delete(x.nh, dst)
		}
	}

	if vnet.AdjDebug {
		fmt.Printf("fib.go delReachableVia adj %v IsMpAdj %v mapFibResult after: %v\n", x.adj, m.IsMpAdj(x.adj), x)
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
	// r is a mapFibResult from unreachable that we will move to reachable here
	for dst, dstMap := range r.nh {
		// find the destination address from r that matches with prefix p
		if dst.a.MatchesPrefix(p) {
			// delete the entry from r's map
			delete(r.nh, dst)
			// dstMap is map of prefixes that has dst as their nh but was not acctually added to the fib table because nh was unreachable
			// For each that match prefix p, actually add nh (i.e. dst.a) to prefix via addDelRouteNextHop which makes nh reachable
			for dp, r := range dstMap {
				g := m.fibByIndex(dp.i, false)
				const isDel = false
				if vnet.AdjDebug {
					fmt.Printf("fib.go call addDelRouteNextHop prefix %v add nh %v from makeReachable\n",
						dp.p.String(), dst.a)
				}
				// Don't add nh to glean or local
				// FIXME, what to do instead? ignore and print now
				ai, _ := g.Get(&dp.p) // Get gets from g.reachable
				if ai == ip.AdjNil || ai == ip.AdjMiss || m.IsMpAdj(ai) {
					g.addDelRouteNextHop(m, &dp.p, dst.a, r, isDel)
				} else if vnet.AdjDebug {
					as := m.GetAdj(ai)
					adjType := "no adjacency found at that adj index"
					if len(as) > 0 {
						adjType = as[0].LookupNextIndex.String()
					}
					fmt.Printf("fib.go makeReachable: ignore adding nh %v to prefix %v which has has non MpAdj %v of type %v",
						dst.a, dp.p.String(), ai, adjType)
				}
			}
		}
	}
}

func (x *mapFibResult) addUnreachableVia(m *Main, f *Fib, p *Prefix) {
	// don't know how this is used in conjunction with recursive addDelUnreachable
	// seems like if there is a match, it would delete the entry, but then just add it back?
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

	if vnet.AdjDebug {
		if isDel {
			fmt.Printf("fib.go addDelReachable delete: %v %v adj %v IsMpAdj %v\n", f.index.Name(&m.Main), p.String(), a, m.IsMpAdj(a))
		} else {
			fmt.Printf("fib.go addDelReachable add: %v %v adj %v IsMpAdj %v\n", f.index.Name(&m.Main), p.String(), a, m.IsMpAdj(a))
		}
	}
	// Look up less specific reachable route for prefix.
	lr, _, lok := f.reachable.getLessSpecific(p)
	if isDel {
		if lok {
			fmt.Printf("fib.go addDelReachable delete: %v %v adj %v replaceWithLessSpecific lr %v r %v\n",
				f.index.Name(&m.Main), p.String(), a, lr, r)
			lr.replaceWithLessSpecific(m, f, &r)
		} else {
			fmt.Printf("fib.go addDelReachable delete: %v %v adj %v delReachableVia r %v\n",
				f.index.Name(&m.Main), p.String(), a, r)
			r.delReachableVia(m, f)
		}
	} else {
		if lok {
			fmt.Printf("fib.go addDelReachable add: %v %v adj %v replaceWithMoreSpecific lr %v r %v\n",
				f.index.Name(&m.Main), p.String(), a, lr, r)
			lr.replaceWithMoreSpecific(m, f, p, a, &r)
		}
		if r, _, ok := f.unreachable.Lookup(p.Address); ok {
			fmt.Printf("fib.go addDelReachable add: %v %v adj %v makeReachable\n", f.index.Name(&m.Main), p.String(), a)
			r.makeReachable(m, f, p, a)
		}
	}
	if vnet.AdjDebug {
		if isDel {
			fmt.Printf("fib.go addDelReachable delete: %v %v adj %v IsMpAdj %v finished: r %v\n", f.index.Name(&m.Main), p.String(), a, m.IsMpAdj(a), r)
		} else {
			fmt.Printf("fib.go addDelReachable add: %v %v adj %v IsMpAdj %v finished: r %v\n", f.index.Name(&m.Main), p.String(), a, m.IsMpAdj(a), r)
		}
	}
}

func (f *Fib) addDelUnreachable(m *Main, p *Prefix, pf *Fib, a Address, r NextHopper, isDel bool, recurse bool) (err error) {
	//pf is the fib that f is the nexthop of
	//a is the nexthop address from pf
	//r is the nexthopper from pf
	nr, np, _ := f.unreachable.Lookup(a)
	if isDel && recurse {
		// don't recurse on delete for now; can get into infinite loop sometimes
		/*
			nr.delReachableVia(m, f)
		*/
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
	for l := int(pʹ.Len) - 1; l >= 1; l-- {
		if f[l] == nil {
			continue
		}
		p.Len = uint32(l)
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
	// a = 0 is AdjMiss if not found in reachable
	if r, ok = f.reachable[p.Len][p.mapFibKey()]; ok {
		a = r.adj
	}
	return
}

func (f *Fib) Add(m *Main, p *Prefix, r ip.Adj) (ip.Adj, bool) { return f.addDel(m, p, r, false) }
func (f *Fib) Del(m *Main, p *Prefix) (ip.Adj, bool)           { return f.addDel(m, p, ip.AdjMiss, true) }
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
	var h vnet.HwInterfacer
	if hw != nil {
		h = m.Vnet.HwIfer(hw.Hi())
	}

	next := ip.LookupNextRewrite
	noder := &m.rewriteNode
	packetType := vnet.IP4

	if _, ok := h.(vnet.Arper); h == nil || ok {
		next = ip.LookupNextGlean
		noder = &m.arpNode
		packetType = vnet.ARP
		a.Index = uint32(ia)
	}

	a.LookupNextIndex = next
	if h != nil {
		m.Vnet.SetRewrite(&a.Rewrite, si, noder, packetType, nil /* dstAdr meaning broadcast */)
	}
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

// used by neighbor message to add/del route, e.g. from succesfull arp
func (m *Main) addDelRoute(p *ip.Prefix, fi ip.FibIndex, baseAdj ip.Adj, isDel bool) (oldAdj ip.Adj, err error) {
	createFib := !isDel
	f := m.fibByIndex(fi, createFib)
	q := FromIp4Prefix(p)
	var ok bool

	if vnet.AdjDebug {
		if isDel {
			fmt.Printf("fib.go addDelRoute delete %v adj %v\n", q.Address.String(), baseAdj.String())
		} else {
			fmt.Printf("fib.go addDelRoute add %v adj %v\n", q.Address.String(), baseAdj.String())
		}
	}

	//addDel the route to/from fib
	oldAdj, ok = f.addDel(m, &q, baseAdj, isDel)

	//don't err if deleting something that has already been deleted
	if !ok {
		if isDel {
			fmt.Printf("DEBUG: fib.go addDelRoute delete prefix %v not fount", q.String())
		} else {
			err = fmt.Errorf("fib.go addDelRoute add prefix %v error addDel", q.String())
		}
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

func (x *mapFibResult) delNhFromMatchingPrefix(m *Main, f *Fib, p *Prefix) {
	// x is the mapFibResult from reachable or unreachable
	// delReachableVia will traverse the map and remove x's address from prefix p if p uses x.nh.a as a nexthop address
	for id, dstmap := range x.nh {
		for ip, nhr := range dstmap {
			if p.Matches(&ip.p) && ip.i == f.index {
				if vnet.AdjDebug {
					fmt.Printf("fib.go call addDelRouteNextHop delete %v prefix %v nexthop %v nhAdj %v from delAllRouteNextHops\n",
						f.index.Name(&m.Main), p.String(), id.a, x.adj)
				}
				// for mpAdj, addDelRouteNextHop will take care of removing ip from dstmap as part of the cleanup and accounting?
				f.addDelRouteNextHop(m, p, id.a, nhr, true)
			}
		}
	}
}

// properly delete all nexthops in the adj
func (m *Main) delAllRouteNextHops(f *Fib, p *Prefix) {
	oldAdj, _ := f.Get(p)
	if oldAdj == ip.AdjNil || oldAdj == ip.AdjMiss || !m.IsMpAdj(oldAdj) {
		// Nothing to delete if AdjNil or AdjMiss
		// None mpAdj deletes are handled elsewhere, no concept of deleteAll there
		return
	}
	if vnet.AdjDebug {
		fmt.Printf("fib.go delAllRouteNextHops from %v %v\n", f.index.Name(&m.Main), p.String())
	}
	// find all nh from reachable that has p in its map; only need to look into length 32
	for _, r := range f.reachable[32] {
		r.delNhFromMatchingPrefix(m, f, p)
	}
	// find all nh from unreachable that has p in its map; only need to look into length 32
	for _, r := range f.unreachable[32] {
		r.delNhFromMatchingPrefix(m, f, p)
	}
}

func (m *Main) AddDelRouteNextHop(p *Prefix, nh *NextHop, isDel bool, isReplace bool) (err error) {
	f := m.fibBySi(nh.Si)
	oldAdj, _ := f.Get(p)

	if oldAdj != ip.AdjNil && oldAdj != ip.AdjMiss && !m.IsMpAdj(oldAdj) {
		// oldAdj is probably a glean or local, don't add or remove nh
		if vnet.AdjDebug {
			as := m.GetAdj(oldAdj)
			adjType := "no adjacency found at that adj index"
			if len(as) > 0 {
				adjType = as[0].LookupNextIndex.String()
			}
			if isDel {
				fmt.Printf("fib.go AddDelRouteNextHop isReplace %v: ignore deleting nh %v from prefix %v which has has non MpAdj %v of type %v\n",
					isReplace, nh.Address, p.String(), oldAdj, adjType)
			} else {
				fmt.Printf("fib.go AddDelRouteNextHop isReplace %v: ignore adding nh %v to prefix %v which has has non MpAdj %v of type %v\n",
					isReplace, nh.Address, p.String(), oldAdj, adjType)
			}
		}
		return
	}

	if isReplace { // Delete prefix and cleanup its adjacencies before add
		if vnet.AdjDebug {
			fmt.Printf("fib.go Replace: delete %v and clean up old adjacency oldAdj %v\n", p.String(), oldAdj.String())
		}
		// Do a proper cleanup and delete of old next hops
		m.delAllRouteNextHops(f, p)
	}
	if vnet.AdjDebug {
		if isDel {
			fmt.Printf("fib.go call addDelRouteNextHop %v prefix %v oldAdj %v delete %v from AddDelRouteNextHop\n",
				f.index.Name(&m.Main), p.String(), oldAdj, nh.Address)
		} else {
			fmt.Printf("fib.go call addDelRouteNextHop %v prefix %v oldAdj %v add nh %v from AddDelRouteNextHop\n",
				f.index.Name(&m.Main), p.String(), oldAdj, nh.Address)
		}
	}
	return f.addDelRouteNextHop(m, p, nh.Address, nh, isDel)
}

func (f *Fib) addDelRouteNextHop(m *Main, p *Prefix, nha Address, nhr NextHopper, isDel bool) (err error) {
	var (
		nhAdj, oldAdj, newAdj ip.Adj
		ok                    bool
	)
	if !isDel && nha.MatchesPrefix(p) && p.Address != AddressUint32(0) {
		err = fmt.Errorf("fib.go addDelRouteNextHop add: prefix %s matches next-hop %s", p, &nha)
		return
	}

	nhf := m.fibByIndex(nhr.NextHopFibIndex(m), true)

	var reachable_via_prefix Prefix
	if r, np, found, bad := nhf.reachable.lookupReachable(m, nha); found || bad {
		if bad {
			err = &prefixError{s: "unreachable next-hop", p: *p}
			return
		}
		nhAdj = r.adj
		reachable_via_prefix = np
	} else {
		// not sure what's the purpose of recurse....
		// seems like it is trying to handle if the prefix that uses nha as next hop is in a different fib table, e.g. if f != nhf?
		// If they are the same, we could end up with multipe add or multiple del to same entry?
		const recurse = true
		err = nhf.addDelUnreachable(m, p, f, nha, nhr, isDel, recurse)
		{ //debug print
			if err != nil {
				fmt.Printf("fib.go addDelUnreachable err: recurse\n")
			}
		}
		return
	}

	oldAdj, ok = f.Get(p)
	if isDel && !ok {
		//debug print, flag but don't err if deleting
		//err = &prefixError{s: "unknown destination", p: *p}
		fmt.Printf("fib.go: deleteing %v unknown destination; maybe already deleted\n", p)
		return
	}

	if oldAdj == nhAdj && isDel {
		newAdj = ip.AdjNil
	} else if newAdj, ok = m.AddDelNextHop(oldAdj, nhAdj, nhr.NextHopWeight(), nhr, isDel); !ok { //for nhr NextHopper argument here, only AdjacencyFinalizer is used
		if true { //if this is a delete, don't error (which would cause panic later); just flag
			if !isDel {
				err = fmt.Errorf("fib.go addDelRouteNextHop add: %v %v, AddDelextHop !ok, oldAdj %v, nhAdj %v\n",
					f.index.Name(&m.Main), p.String(), oldAdj, nhAdj)
			} else {
				// This is legit.  Could get a message to remove a nexthop that is no longer the next hop of this prefix.
				// Example:  2 routes to a prefix, one via a rewrite next hop, one via glean.  Not consider a multipath, but it is effectively that.
				// vnet will choose one of the 2 to populate the TCAM depending on the order the add messages come in.  If the delete rewrite nexthop come in
				// and vnet had populated the glean as its nexthop, then AddDelNextHop will return !ok as the rewrite nhAdj is not there.
				fmt.Printf("fib.go addDelRouteNextHop delete: %v %v, AddDelextHop !ok, oldAdj %v, nhAdj %v, likely because %v was not its nh which is OK\n",
					f.index.Name(&m.Main), p.String(), oldAdj, nhAdj, nhAdj)
			}
		}
		return
	}

	if vnet.AdjDebug {
		if isDel {
			fmt.Printf("fib.go addDelRouteNextHop delete: prefix %v,  oldAdj %v, newAdj %v\n", p.String(), oldAdj.String(), newAdj.String())
		} else {
			fmt.Printf("fib.go addDelRouteNextHop add: prefix %v, oldAdj %v, newAdj %v\n", p.String(), oldAdj.String(), newAdj.String())
		}
	}

	if oldAdj != newAdj {
		// oldAdj != newAdj means index changed because there is now more than 1 nexthop (multiplath)
		// or multipath members changed because of nexthop add/del

		isFibDel := isDel
		// if isDel, do not remove adjacency unless all members of the multipath adjacency have been removed
		// instead, update fib table with newAdj
		// when all member os the multipath are removed, newAdj will be ip.AdjNil
		if isFibDel && newAdj != ip.AdjNil {
			isFibDel = false
		}
		f.addDel(m, p, newAdj, isFibDel)
		nhf.setReachable(m, p, f, &reachable_via_prefix, nha, nhr, isDel)
	}
	return
}

// This is used by replaceWithLessSpecific and replaceWithMoreSpecific only
func (f *Fib) replaceNextHop(m *Main, p *Prefix, pf *Fib, fromNextHopAdj, toNextHopAdj ip.Adj, nha Address, r NextHopper) (err error) {
	if adj, ok := f.Get(p); !ok {
		//debug print instead of err; may be OK
		//err = &prefixError{s: "unknown destination", p: *p}
		fmt.Printf("fib.go: replaceNextHop, unknown destination, addr %v, nextHop %v, from-nha %v to-nha %v, namespace %s\n",
			p, nha, fromNextHopAdj, toNextHopAdj, f.index.Name(&m.Main))
	} else {
		if vnet.AdjDebug {
			fmt.Printf("fib.go replaceNextHop: prefix %v from adj %v to adj %v, nha %v\n",
				p.String(), fromNextHopAdj, toNextHopAdj, nha)
		}
		as := m.GetAdj(toNextHopAdj)
		// If replacement is glean (interface route) then next hop becomes unreachable.
		// Assume glean already exist so no need to explicity add here?
		isDel := len(as) == 1 && as[0].IsGlean()
		if isDel {
			if vnet.AdjDebug {
				fmt.Printf("fib.go call addDelRouteNextHop prefix %v delete nh %v from replaceNextHop\n",
					p.String(), nha)
			}
			err = pf.addDelRouteNextHop(m, p, nha, r, isDel)
			if err == nil {
				err = f.addDelUnreachable(m, p, pf, nha, r, !isDel, false)
			}
		} else {
			// Adjacencies in the toNextHopAj must be rewrites for ReplaceNextHop
			if err = m.ReplaceNextHop(adj, fromNextHopAdj, toNextHopAdj, r); err != nil {
				err = fmt.Errorf("replace next hop %v from-nha %v to-nha %v: %v", adj, fromNextHopAdj, toNextHopAdj, err)
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

//used for local or glean fib adjacencies
func (f *Fib) addDelReplace(m *Main, p *Prefix, r ip.Adj, isDel bool) {
	if vnet.AdjDebug {
		if isDel {
			fmt.Printf("fib.go addDelReplace delete %v %v adj %v\n", f.index.Name(&m.Main), p.String(), r.String())
		} else {
			if m.IsMpAdj(r) {
				panic(fmt.Errorf("fib.go addDelReplace %v adding a multipath adj %v to glean or local fib %v!\n",
					f.index.Name(&m.Main), r.String(), p.String()))
			} else {
				fmt.Printf("fib.go addDelReplace add %v %v adj %v\n", f.index.Name(&m.Main), p.String(), r.String())
			}
		}
	}

	oldAdj, _ := f.Get(p)
	// If oldAdj is a mpAdj, then need to do a cleanup and delete before adding the replacement adj
	if m.IsMpAdj(oldAdj) {
		// if add, clean up first by deleting oldAdj before adding
		// if delete, r argument is AdjNil or AdjMiss, so use oldAdj to delete here
		if vnet.AdjDebug {
			fmt.Printf("fib.go addDelReplace %v %v %v: first delete old adjacency %v IsMpAdj %v\n",
				f.index.Name(&m.Main), p.String(), r.String(), oldAdj, m.IsMpAdj(oldAdj))
		}
		// First delete all its next hops; this maintains the proper adjacency tracking
		// prefixes in fib, i.e. MapFib, are indexed by mapFibKey which has the mask applied
		// apply mask for the delAllRouteNextHops argument
		q := p.ApplyMask()
		m.delAllRouteNextHops(f, q)

		// This uses the original prefix
		if !isDel {
			// addDel new adjacency
			f.addDel(m, p, r, isDel)
		}
	} else {
		if oldAdj, ok := f.addDel(m, p, r, isDel); ok && oldAdj != ip.AdjNil && oldAdj != ip.AdjMiss && oldAdj != r {
			// oldAdj should not return as a mpAdj
			if m.IsMpAdj(oldAdj) {
				fmt.Printf("DEBUG: fib.go addDelReplace isDel %v %v %v adj %v:  addDel returned an oldAdj %v that is a mpAdj",
					isDel, f.index.Name(&m.Main), p.String(), r, oldAdj)
				return
			}
			if !m.IsAdjFree(oldAdj) {
				m.DelAdj(oldAdj)
			}
		}
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
			// Question why this is done independently - rtnetlink should send
			// any routes it would like deleted.
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
