// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"math"
	"sort"
	"unsafe"
)

// Next node index stored in ip adjacencies.
type LookupNext uint16

const (
	// Packet does not match any route in table.
	LookupNextMiss LookupNext = iota

	// Adjacency says to drop or punt this packet.
	LookupNextDrop
	LookupNextPunt

	// This packet matches an IP address of one of our interfaces.
	LookupNextLocal

	// Glean node.
	// This packet matches an "interface route" and packets need to be passed to ip4 ARP (or ip6 neighbor discovery)
	// to find rewrite string for this destination.
	LookupNextGlean

	// This packet is to be rewritten and forwarded to the next
	// processing node.  This is typically the output interface but
	// might be another node for further output processing.
	LookupNextRewrite

	LookupNNext
)

var lookupNextNames = [...]string{
	LookupNextMiss:    "miss",
	LookupNextDrop:    "drop",
	LookupNextPunt:    "punt",
	LookupNextLocal:   "local",
	LookupNextGlean:   "glean",
	LookupNextRewrite: "rewrite",
}

func (n LookupNext) String() string { return elib.StringerHex(lookupNextNames[:], int(n)) }

func (n *LookupNext) Parse(in *parse.Input) {
	switch text := in.Token(); text {
	case "miss":
		*n = LookupNextMiss
	case "drop":
		*n = LookupNextDrop
	case "punt":
		*n = LookupNextPunt
	case "local":
		*n = LookupNextLocal
	case "glean":
		*n = LookupNextGlean
	case "rewrite":
		*n = LookupNextRewrite
	default:
		panic(parse.ErrInput)
	}
}

type Adjacency struct {
	// Next hop after ip4-lookup.
	LookupNextIndex LookupNext

	// Number of adjecencies in block.  Greater than 1 means multipath; otherwise equal to 1.
	NAdj uint16

	// Interface address index for local/arp adjacency.
	// Destination index for rewrite.
	Index uint32

	vnet.Rewrite
}

func (a *Adjacency) IsRewrite() bool { return a.LookupNextIndex == LookupNextRewrite }
func (a *Adjacency) IsLocal() bool   { return a.LookupNextIndex == LookupNextLocal }
func (a *Adjacency) IsGlean() bool   { return a.LookupNextIndex == LookupNextGlean }

func (a *Adjacency) IfAddr() (i IfAddr) {
	i = IfAddrNil
	switch a.LookupNextIndex {
	case LookupNextGlean, LookupNextLocal:
		i = IfAddr(a.Index)
	}
	return
}

func (a *Adjacency) String(m *Main) (lines []string) {
	lines = append(lines, a.LookupNextIndex.String())
	ni := a.LookupNextIndex
	switch ni {
	case LookupNextRewrite:
		l := a.Rewrite.String(m.v)
		lines[0] += " " + l[0]
		// If only 2 lines, fit into a single line.
		if len(l) == 2 {
			lines[0] += " " + l[1]
		} else {
			lines = append(lines, l[1:]...)
		}
	case LookupNextGlean, LookupNextLocal:
		ifa := IfAddr(a.Index)
		lines[0] += " " + ifa.String(m)
	}
	return
}

func (a *Adjacency) ParseWithArgs(in *parse.Input, args *parse.Args) {
	m := args.Get().(*Main)
	if !in.Parse("%v", &a.LookupNextIndex) {
		panic(parse.ErrInput)
	}
	a.NAdj = 1
	a.Index = ^uint32(0)
	switch a.LookupNextIndex {
	case LookupNextRewrite:
		if !in.Parse("%v", &a.Rewrite, m.v) {
			panic(in.Error())
		}
	case LookupNextLocal, LookupNextGlean:
		var si vnet.Si
		if in.Parse("%v", &si, m.v) {
			a.Index = uint32(m.IfFirstAddr(si))
		}
	}
}

// Index into adjacency table.
type Adj uint32

// Miss adjacency is always first in adjacency table.
const (
	AdjMiss Adj = 0
	AdjDrop Adj = 1
	AdjPunt Adj = 2
	AdjNil  Adj = (^Adj(0) - 1) // so AdjNil + 1 is non-zero for remaps.
)

//go:generate gentemplate -d Package=ip -id adjacencyHeap -d HeapType=adjacencyHeap -d Data=elts -d Type=Adjacency github.com/platinasystems/go/elib/heap.tmpl

type adjacencyThread struct {
	// Packet/byte counters for each adjacency.
	counters vnet.CombinedCounters
}

type adjacencyMain struct {
	adjacencyHeap

	multipathMain multipathMain

	threads []*adjacencyThread

	adjAddDelHookVec
	adjSyncCounterHookVec
	adjGetCounterHookVec

	specialAdj [3]Adj
}

type adjAddDelHook func(m *Main, adj Adj, isDel bool)
type adjSyncCounterHook func(m *Main)
type AdjGetCounterHandler func(tag string, v vnet.CombinedCounter)
type adjGetCounterHook func(m *Main, adj Adj, f AdjGetCounterHandler)

//go:generate gentemplate -id adjAddDelHook -d Package=ip -d DepsType=adjAddDelHookVec -d Type=adjAddDelHook -d Data=adjAddDelHooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id adjSyncHook -d Package=ip -d DepsType=adjSyncCounterHookVec -d Type=adjSyncCounterHook -d Data=adjSyncCounterHooks github.com/platinasystems/go/elib/dep/dep.tmpl
//go:generate gentemplate -id adjGetCounterHook -d Package=ip -d DepsType=adjGetCounterHookVec -d Type=adjGetCounterHook -d Data=adjGetCounterHooks github.com/platinasystems/go/elib/dep/dep.tmpl

type NextHopWeight uint32

// A next hop in a multipath.
type nextHop struct {
	// Adjacency index for next hop's rewrite.
	Adj Adj

	// Relative weight for this next hop.
	Weight NextHopWeight
}

//go:generate gentemplate -d Package=ip -id nextHopHeap -d HeapType=nextHopHeap -d Data=elts -d Type=nextHop github.com/platinasystems/go/elib/heap.tmpl
//go:generate gentemplate -d Package=ip -id nextHopVec -d VecType=nextHopVec -d Type=nextHop github.com/platinasystems/go/elib/vec.tmpl
//go:generate gentemplate -d Package=ip -id multipathAdjacencyPool -d PoolType=multipathAdjacencyPool -d Type=multipathAdjacency -d Data=elts github.com/platinasystems/go/elib/pool.tmpl

type nextHopHashValue struct {
	heapOffset uint32
	adj        Adj
}

type multipathMain struct {
	// Expressed as max allowable fraction of traffic sent the wrong way assuming perfectly random distribution.
	multipathErrorTolerance float64
	nextHopHash             elib.Hash
	nextHopHashValues       []nextHopHashValue

	cachedNextHopVec [3]nextHopVec

	nextHopHeap

	// Indexed by heap id.  So, one element per heap block.
	mpAdjPool multipathAdjacencyPool
}

func (m *multipathMain) GetNextHops(i uint) []nextHop {
	return m.nextHopHeap.Slice(uint(m.nextHopHashValues[i].heapOffset))
}
func (m *multipathMain) HashIndex(s *elib.HashState, i uint) {
	nextHopVec(m.GetNextHops(i)).HashKey(s)
}
func (m *multipathMain) HashResize(newCap uint, rs []elib.HashResizeCopy) {
	src, dst := m.nextHopHashValues, make([]nextHopHashValue, newCap)
	for i := range rs {
		dst[rs[i].Dst] = src[rs[i].Src]
	}
	m.nextHopHashValues = dst
}

func (a nextHopVec) HashKey(s *elib.HashState) {
	s.HashPointer(unsafe.Pointer(&a[0]), uintptr(a.Len())*unsafe.Sizeof(a[0]))
}
func (a nextHopVec) Equal(b nextHopVec) bool {
	if la, lb := a.Len(), b.Len(); la != lb {
		return false
	} else {
		for i := uint(0); i < la; i++ {
			if a[i].compare(&b[i]) != 0 {
				return false
			}
		}
	}
	return true
}
func (a nextHopVec) HashKeyEqual(h elib.Hasher, i uint) bool {
	b := nextHopVec(h.(*multipathMain).GetNextHops(i))
	return a.Equal(b)
}

func (a *nextHop) compare(b *nextHop) (cmp int) {
	cmp = int(b.Weight) - int(a.Weight)
	if cmp == 0 {
		cmp = int(a.Adj) - int(b.Adj)
	}
	return
}

func (a nextHopVec) sort() {
	sort.Slice(a, func(i, j int) bool {
		return a[i].compare(&a[j]) < 0
	})
}

// Normalize next hops: find a power of 2 sized block of next hops within error tolerance of given raw next hops.
func (given nextHopVec) normalizePow2(m *multipathMain, result *nextHopVec) (nAdj uint, norm nextHopVec) {
	nGiven := given.Len()

	if nGiven == 0 {
		return
	}

	// Allocate enough space for 2 copies; we'll use second copy to save original weights.
	t := *result
	t.Validate(2*nGiven - 1)
	// Save allocation for next caller.
	*result = t

	n := nGiven
	nAdj = n
	switch n {
	case 1:
		t[0] = given[0]
		t[0].Weight = 1
		norm = t[:1]
		return
	case 2:
		cmp := 0
		if given[0].compare(&given[1]) > 0 {
			cmp = 1
		}
		t[0], t[1] = given[cmp], given[cmp^1]
		if t[0].Weight == t[1].Weight {
			t[0].Weight = 1
			t[1].Weight = 1
			norm = t[:2]
			return
		}

	default:
		copy(t, given)
		// Order by decreasing weight and increasing adj index for equal weights.
		sort.Slice(t[0:n], func(i, j int) bool {
			return t[i].compare(&t[j]) < 0
		})
	}

	// Find total weight to normalize weights.
	sumWeight := float64(0)
	for i := uint(0); i < n; i++ {
		sumWeight += float64(t[i].Weight)
	}

	// In the unlikely case that all weights are given as 0, set them all to 1.
	if sumWeight == 0 {
		for i := uint(0); i < n; i++ {
			t[i].Weight = 1
		}
		sumWeight = float64(n)
	}

	// Save copies of all next hop weights to avoid being overwritten in loop below.
	copy(t[n:], t[:n])
	t_save := t[n:]

	if m.multipathErrorTolerance == 0 {
		m.multipathErrorTolerance = .01
	}

	// Try larger and larger power of 2 sized adjacency blocks until we
	// find one where traffic flows to within 1% of specified weights.
	nAdj = uint(elib.Word(n).MaxPow2())
	for {
		w := float64(nAdj) / sumWeight
		nLeft := nAdj

		for i := uint(0); i < n; i++ {
			nf := w * float64(t_save[i].Weight)
			n := uint(nf + .5) // round to nearest integer
			if n > nLeft {
				n = nLeft
			}
			nLeft -= n
			t[i].Weight = NextHopWeight(n)
		}
		// Add left over weight to largest weight next hop.
		t[0].Weight += NextHopWeight(nLeft)
		error := float64(0)
		i_zero := n
		for i := uint(0); i < n; i++ {
			if t[i].Weight == 0 && i_zero == n {
				i_zero = i
			}
			// perfect distribution of weight:
			want := float64(t_save[i].Weight) / sumWeight
			// approximate distribution with nAdj
			have := float64(t[i].Weight) / float64(nAdj)
			error += math.Abs(want - have)
		}
		if error < m.multipathErrorTolerance {
			// Truncate any next hops with zero weight.
			norm = t[:i_zero]
			break
		}

		// Try next power of 2 size.
		nAdj *= 2
	}
	return
}

func (m *multipathMain) allocNextHopBlock(b *nextHopBlock, key nextHopVec) {
	n := uint(len(key))
	o := m.nextHopHeap.Get(n)
	copy(m.nextHopHeap.Slice(o), key)
	b.size = uint32(n)
	b.offset = uint32(o)
}

func (m *multipathMain) freeNextHopBlock(b *nextHopBlock) {
	m.nextHopHeap.Put(uint(b.offset))
	b.offset = ^uint32(0)
	b.size = ^uint32(0)
}

func (m *multipathMain) getNextHopBlock(b *nextHopBlock) nextHopVec {
	return nextHopVec(m.nextHopHeap.Slice(uint(b.offset)))
}

type nextHopBlock struct {
	// Heap offset of first next hop.
	offset uint32

	// Block size.
	size uint32
}

type multipathAdjacency struct {
	// Index of first adjacency in block.
	adj Adj

	// Power of 2 size of block.
	nAdj uint32

	// Number of prefixes that point to this adjacency.
	referenceCount uint32

	// Pool index.
	index uint

	// Given next hops are saved so that control plane has a record of exactly
	// what the RIB told it.
	givenNextHops nextHopBlock

	// Resolved next hops are used as hash keys: they are sorted by weight
	// and weights are chosen so they add up to nAdj (with zero-weighted next hops being deleted).
	resolvedNextHops nextHopBlock
}

func (p *nextHopVec) sum_duplicates() {
	var i, j, l int
	r := *p
	l = len(r)
	lastAdj := AdjNil
	for i = 0; i < l; i++ {
		a := &r[i]
		if i > 0 && a.Adj == lastAdj {
			r[j].Weight += a.Weight
		} else {
			r[j] = *a
			j++
		}
		lastAdj = a.Adj
	}
	*p = r[:j]
}

func (r *nextHopVec) resolve(m *Main, given nextHopVec, level uint) {
	mp := &m.multipathMain
	if level == 0 {
		r.ResetLen()
	}
	for i := range given {
		g := &given[i]
		i0 := r.Len()
		ma := m.mpAdjForAdj(g.Adj, false)
		if ma != nil {
			r.resolve(m, mp.getNextHopBlock(&ma.givenNextHops), level+1)
		} else {
			*r = append(*r, *g)
		}
		i1 := r.Len()
		for i0 < i1 {
			(*r)[i0].Weight *= g.Weight
			i0++
		}
	}
	if level == 0 {
		r.sort()
		r.sum_duplicates()
	}
}

type AdjacencyFinalizer interface {
	FinalizeAdjacency(a *Adjacency)
}

func (m *Main) createMpAdj(given nextHopVec, af AdjacencyFinalizer) (madj *multipathAdjacency) {
	mp := &m.multipathMain

	mp.cachedNextHopVec[1].resolve(m, given, 0)
	resolved := mp.cachedNextHopVec[1]

	nAdj, norm := resolved.normalizePow2(mp, &mp.cachedNextHopVec[2])

	// Use given next hops to see if we've seen a block equivalent to this one before.
	i, ok := mp.nextHopHash.Get(norm)
	if ok {
		ai := mp.nextHopHashValues[i].adj
		madj = m.mpAdjForAdj(ai, false)
		return
	}

	// Copy next hops into power of 2 adjacency block one for each weight.
	ai, as := m.NewAdj(nAdj)
	for nhi := range norm {
		nh := &norm[nhi]
		nextHopAdjacency := &m.adjacencyHeap.elts[nh.Adj]
		for w := NextHopWeight(0); w < nh.Weight; w++ {
			as[i] = *nextHopAdjacency
			if af != nil {
				af.FinalizeAdjacency(&as[i])
			}
			as[i].NAdj = uint16(nAdj)
			i++
		}
	}

	madj = m.mpAdjForAdj(ai, true)

	madj.adj = ai
	madj.nAdj = uint32(nAdj)
	madj.referenceCount = 0 // caller will set to 1

	mp.allocNextHopBlock(&madj.resolvedNextHops, norm)
	mp.allocNextHopBlock(&madj.givenNextHops, given)

	i, _ = mp.nextHopHash.Set(norm)
	mp.nextHopHashValues[i] = nextHopHashValue{
		heapOffset: madj.resolvedNextHops.offset,
		adj:        ai,
	}

	m.CallAdjAddHooks(ai)
	return
}

func (m *Main) GetAdjacencyUsage() elib.HeapUsage { return m.adjacencyHeap.GetUsage() }

func (m *adjacencyMain) mpAdjForAdj(a Adj, create bool) (ma *multipathAdjacency) {
	adj := &m.adjacencyHeap.elts[a]
	if !adj.IsRewrite() {
		return
	}
	i := uint(adj.Index)
	mm := &m.multipathMain
	if create {
		i = mm.mpAdjPool.GetIndex()
		adj.Index = uint32(i)
		mm.mpAdjPool.elts[i].index = i
	}
	if i < mm.mpAdjPool.Len() {
		ma = &mm.mpAdjPool.elts[i]
	}
	return
}

func (m *adjacencyMain) NextHopsForAdj(a Adj) (nhs nextHopVec) {
	mm := &m.multipathMain
	ma := m.mpAdjForAdj(a, false)
	if ma != nil && ma.nAdj > 0 {
		nhs = mm.getNextHopBlock(&ma.resolvedNextHops)
	} else {
		nhs = []nextHop{{Adj: a, Weight: 1}}
	}
	return
}

func (m *Main) AddDelNextHop(oldAdj Adj, nextHopAdj Adj, nextHopWeight NextHopWeight, af AdjacencyFinalizer, isDel bool) (newAdj Adj, ok bool) {
	mm := &m.multipathMain
	var (
		old, new *multipathAdjacency
		nhs      nextHopVec
		nnh, nhi uint
	)

	if oldAdj != AdjNil {
		if old = m.mpAdjForAdj(oldAdj, false); old != nil {
			nhs = mm.getNextHopBlock(&old.givenNextHops)
			nnh = nhs.Len()
			nhi, ok = nhs.find(nextHopAdj)

			// For delete next hop must be found.
			if nhi >= nnh && isDel {
				return
			}
		}
	}

	// Re-use vector for each call.
	mm.cachedNextHopVec[0].Validate(nnh)
	newNhs := mm.cachedNextHopVec[0]
	newAdj = AdjNil
	if isDel {
		// Delete next hop at previously found index.
		if nhi > 0 {
			copy(newNhs[:nhi], nhs[:nhi])
		}
		if nhi+1 < nnh {
			copy(newNhs[nhi:], nhs[nhi+1:])
		}
		newNhs = newNhs[:nnh-1]
	} else {
		// If next hop is already there with the same weight, we have nothing to do.
		if nhi < nnh && nhs[nhi].Weight == nextHopWeight {
			newAdj = oldAdj
			ok = true
			return
		}

		newNhs = newNhs[:nnh+1]

		// Copy old next hops to lookup key.
		copy(newNhs, nhs)

		var nh *nextHop
		if nhi < nnh {
			// Change weight of existing next hop.
			nh = &newNhs[nhi]
		} else {
			// Add a new next hop.
			nh = &newNhs[nnh]
			nh.Adj = nextHopAdj
		}
		// In either case set next hop weight.
		nh.Weight = nextHopWeight
	}

	new = m.addDelHelper(newNhs, old, af)
	if new != nil {
		ok = true
		newAdj = new.adj
	}
	return
}

func (m *Main) addDelHelper(newNhs nextHopVec, old *multipathAdjacency, af AdjacencyFinalizer) (new *multipathAdjacency) {
	mm := &m.multipathMain

	if len(newNhs) > 0 {
		new = m.createMpAdj(newNhs, af)
		// Fetch again since create may have moved multipath adjacency vector.
		if old != nil {
			old = &mm.mpAdjPool.elts[old.index]
		}
	}

	if new != old {
		if old != nil {
			old.referenceCount--
		}
		if new != nil {
			new.referenceCount++
		}
	}
	if old != nil && old.referenceCount == 0 {
		old.free(m)
	}
	return
}

func (m *adjacencyMain) PoisonAdj(a Adj) {
	as := m.adjacencyHeap.Slice(uint(a))
	elib.PointerPoison(unsafe.Pointer(&as[0]), uintptr(len(as))*unsafe.Sizeof(as[0]))
}

func (m *Main) FreeAdj(a Adj) { m.adjacencyHeap.Put(uint(a)) }
func (m *Main) DelAdj(a Adj) {
	m.CallAdjDelHooks(a)
	m.FreeAdj(a)
}

func (nhs nextHopVec) find(target Adj) (i uint, ok bool) {
	for i = 0; i < uint(len(nhs)); i++ {
		if ok = nhs[i].Adj == target; ok {
			break
		}
	}
	return
}

func (m *Main) ReplaceNextHop(ai, fromNextHopAdj, toNextHopAdj Adj, af AdjacencyFinalizer) (ok bool) {
	if fromNextHopAdj == toNextHopAdj {
		return
	}

	mm := &m.multipathMain
	ma := m.mpAdjForAdj(ai, false)
	given := mm.getNextHopBlock(&ma.givenNextHops)
	for i := range given {
		nh := &given[i]
		if ok = nh.Adj == fromNextHopAdj; ok {
			nh.Adj = toNextHopAdj
			break
		}
	}
	if !ok {
		return
	}

	m.addDelHelper(given, ma, af)
	return
}

func (a *multipathAdjacency) invalidate()   { a.nAdj = 0 }
func (a *multipathAdjacency) isValid() bool { return a.nAdj != 0 }

func (m *Main) callAdjAddDelHooks(a Adj, isDel bool) {
	for i := range m.adjAddDelHooks {
		m.adjAddDelHookVec.Get(i)(m, a, isDel)
	}
}
func (m *Main) CallAdjAddHooks(a Adj) { m.callAdjAddDelHooks(a, false) }
func (m *Main) CallAdjDelHooks(a Adj) { m.callAdjAddDelHooks(a, true) }

func (m *Main) RegisterAdjAddDelHook(f adjAddDelHook, dep ...*dep.Dep) {
	m.adjAddDelHookVec.Add(f, dep...)
}

func (m *Main) CallAdjSyncCounterHooks() {
	for i := range m.adjSyncCounterHooks {
		m.adjSyncCounterHookVec.Get(i)(m)
	}
}

func (m *Main) RegisterAdjSyncCounterHook(f adjSyncCounterHook, dep ...*dep.Dep) {
	m.adjSyncCounterHookVec.Add(f, dep...)
}

func (m *Main) CallAdjGetCounterHooks(adj Adj, f AdjGetCounterHandler) {
	for i := range m.adjGetCounterHooks {
		m.adjGetCounterHookVec.Get(i)(m, adj, f)
	}
}

func (m *Main) RegisterAdjGetCounterHook(f adjGetCounterHook, dep ...*dep.Dep) {
	m.adjGetCounterHookVec.Add(f, dep...)
}

func (ma *multipathAdjacency) free(m *Main) {
	m.CallAdjDelHooks(ma.adj)

	mm := &m.multipathMain
	nhs := mm.getNextHopBlock(&ma.resolvedNextHops)
	i, ok := mm.nextHopHash.Unset(nhs)
	if !ok {
		panic("unknown multipath adjacency")
	}
	mm.nextHopHashValues[i] = nextHopHashValue{
		heapOffset: ^uint32(0),
		adj:        AdjNil,
	}

	m.PoisonAdj(ma.adj)
	m.FreeAdj(ma.adj)
	mm.freeNextHopBlock(&ma.givenNextHops)
	mm.freeNextHopBlock(&ma.resolvedNextHops)

	// Nothing to free since multipath adjacencies are indexed by adjacency index.

	ma.invalidate()
}

func (m *adjacencyMain) validateCounter(a Adj) {
	for _, t := range m.threads {
		t.counters.Validate(uint(a))
	}
}
func (m *adjacencyMain) clearCounter(a Adj) {
	for _, t := range m.threads {
		t.counters.Clear(uint(a))
	}
}

func (m *Main) ForeachAdjCounter(a Adj, f func(tag string, v vnet.CombinedCounter)) {
	var v vnet.CombinedCounter
	for _, t := range m.threads {
		var u vnet.CombinedCounter
		t.counters.Get(uint(a), &u)
		v.Add(&u)
	}
	f("", v)
	m.CallAdjGetCounterHooks(a, f)
	return
}

func (m *adjacencyMain) EqualAdj(ai0, ai1 Adj) (same bool) {
	a0, a1 := &m.adjacencyHeap.elts[ai0], &m.adjacencyHeap.elts[ai1]
	ni0, ni1 := a0.LookupNextIndex, a1.LookupNextIndex
	if ni0 != ni1 {
		return
	}
	switch ni0 {
	case LookupNextGlean:
		if a0.Index != a1.Index {
			return
		}
	case LookupNextRewrite:
		if string(a0.Rewrite.Slice()) != string(a1.Rewrite.Slice()) {
			return
		}
	}
	same = true
	return
}

func (m *adjacencyMain) GetAdj(a Adj) (as []Adjacency) { return m.adjacencyHeap.Slice(uint(a)) }
func (m *adjacencyMain) GetAdjRewriteSi(a Adj) (si vnet.Si) {
	as := m.GetAdj(a)
	if as[0].LookupNextIndex == LookupNextRewrite {
		si = as[0].Rewrite.Si
	}
	return
}

func (m *adjacencyMain) NewAdjWithTemplate(n uint, template *Adjacency) (ai Adj, as []Adjacency) {
	ai = Adj(m.adjacencyHeap.Get(n))
	m.validateCounter(ai)
	as = m.GetAdj(ai)
	for i := range as {
		if template != nil {
			as[i] = *template
		}
		as[i].Si = vnet.SiNil
		as[i].NAdj = uint16(n)
		as[i].Index = ^uint32(0)
		m.clearCounter(ai + Adj(i))
	}
	return
}
func (m *adjacencyMain) NewAdj(n uint) (Adj, []Adjacency) { return m.NewAdjWithTemplate(n, nil) }

func (m *multipathMain) init() {
	m.nextHopHash.Init(m, 32)
}

func (m *Main) adjacencyInit() {
	m.multipathMain.init()

	// Build special adjacencies.
	for i, v := range []struct {
		Adj
		LookupNext
	}{
		{Adj: AdjMiss, LookupNext: LookupNextMiss}, // must be in order 0 1 2 else panic below triggers.
		{Adj: AdjDrop, LookupNext: LookupNextDrop},
		{Adj: AdjPunt, LookupNext: LookupNextPunt},
	} {
		var as []Adjacency
		a, as := m.NewAdj(1)
		m.specialAdj[i] = a
		as[0].LookupNextIndex = v.LookupNext
		if got, want := a, v.Adj; got != want {
			panic(fmt.Errorf("special adjacency index mismatch got %d != want %d", got, want))
		}
		m.CallAdjAddHooks(a)
	}
}
