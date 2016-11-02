// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"time"
)

type swIfCounterKind uint16
type swIfCombinedCounterKind uint16
type HwIfCounterKind uint16
type HwIfCombinedCounterKind uint16

const (
	IfDrops swIfCounterKind = iota
	IfPunts
	nBuiltinSingleIfCounters
)
const (
	IfRxCounter swIfCombinedCounterKind = iota
	IfTxCounter
	nBuiltinCombinedIfCounters
)

var builtinSingleIfCounterNames = [...]string{
	IfDrops: "drops", IfPunts: "punts",
}
var builtinCombinedIfCounterNames = [...]string{
	IfRxCounter: "rx", IfTxCounter: "tx",
}

type InterfaceCounterNames struct {
	Single, Combined []string
}

func (n *InterfaceCounterNames) add(name string, is_combined bool) (i uint) {
	p := &n.Single
	if is_combined {
		p = &n.Combined
	}
	i = uint(len(*p))
	*p = append(*p, name)
	return
}

func (v *interfaceMain) addSwCounter(name string) (i uint) {
	return v.swIfCounterNames.add(name, false)
}
func (v *interfaceMain) addSwCombinedCounter(name string) (i uint) {
	return v.swIfCounterNames.add(name, true)
}

func (cn *InterfaceCounterNames) newCounters(v *interfaceMain, names []string, is_hw, is_combined bool) (kind uint) {
	n := uint(len(names))
	m := v.swInterfaces.Len()
	if is_hw {
		m = v.hwIferPool.Len()
	}
	for _, t := range v.ifThreads {
		c := &t.sw
		if is_hw {
			c = &t.hw
		}
		if is_combined {
			kind = c.resizeCombined(n, m)
		} else {
			kind = c.resizeSingle(n, m)
		}
	}
	for _, name := range names {
		cn.add(name, is_combined)
	}
	return
}

func (v *interfaceMain) NewSwCounters(names []string) swIfCounterKind {
	i := v.swIfCounterNames.newCounters(v, names, false, false)
	return swIfCounterKind(i)
}

func (v *interfaceMain) NewSwCombinedCounters(names []string) swIfCombinedCounterKind {
	i := v.swIfCounterNames.newCounters(v, names, false, true)
	return swIfCombinedCounterKind(i)
}

// Add to given interface counters value.
func (c swIfCounterKind) Add(t *InterfaceThread, swIfIndex Si, value uint) {
	t.sw.single[c].Add(uint(swIfIndex), value)
}

// Add to given interface counters packets and bytes values.
func (c swIfCombinedCounterKind) Add(t *InterfaceThread, si Si, packets, bytes uint) {
	t.sw.combined[c].Add(uint(si), packets, bytes)
}

func (c swIfCombinedCounterKind) Add64(t *InterfaceThread, si Si, packets, bytes uint64) {
	t.sw.combined[c].Add64(uint(si), packets, bytes)
}

func (c HwIfCounterKind) Add(t *InterfaceThread, hi Hi, value uint) {
	t.hw.single[c].Add(uint(hi), value)
}

func (c HwIfCounterKind) Add64(t *InterfaceThread, hi Hi, value uint64) {
	t.hw.single[c].Add64(uint(hi), value)
}

func (c HwIfCombinedCounterKind) Add(t *InterfaceThread, hi Hi, packets, bytes uint) {
	t.hw.combined[c].Add(uint(hi), packets, bytes)
}

func (c HwIfCombinedCounterKind) Add64(t *InterfaceThread, hi Hi, packets, bytes uint64) {
	t.hw.combined[c].Add64(uint(hi), packets, bytes)
}

type foreachFn func(name string, value uint64)

func (m *interfaceMain) doSwCombined(f foreachFn, nm *InterfaceCounterNames, zero bool, nk, k, i uint) {
	var v, w CombinedCounter
	for _, t := range m.ifThreads {
		if t != nil {
			t.sw.combined[k].Get(i, &w)
			v.Add(&w)
		}
	}
	if v.Packets != 0 || zero {
		f(nm.Combined[nk]+" packets", v.Packets)
		f(nm.Combined[nk]+" bytes", v.Bytes)
	}
	return
}

func (m *interfaceMain) doSwSingle(f foreachFn, nm *InterfaceCounterNames, zero bool, nk, k, i uint) {
	var v, w uint64
	for _, t := range m.ifThreads {
		if t != nil {
			t.sw.single[k].Get(i, &w)
			v += w
		}
	}
	if v != 0 || zero {
		f(nm.Single[nk], v)
	}
	return
}

func (m *interfaceMain) HwSwSingleIfCounter(i uint) swIfCounterKind {
	return swIfCounterKind(uint(len(m.swIfCounterNames.Single)) + i)
}

func (m *interfaceMain) HwSwCombinedIfCounter(i uint) swIfCombinedCounterKind {
	return swIfCombinedCounterKind(uint(len(m.swIfCounterNames.Combined)) + i)
}

func (m *interfaceMain) foreachSwIfCounter(zero bool, si Si, f func(name string, value uint64)) {
	i := uint(si)

	// Make sure at least one interface thread exists.
	m.GetIfThread(0)

	var k0, k1 uint

	// First builtin counters.
	for k0 = 0; k0 < uint(len(builtinCombinedIfCounterNames)); k0++ {
		m.doSwCombined(f, &m.swIfCounterNames, zero, k0, k0, i)
	}
	for k1 = 0; k1 < uint(len(builtinSingleIfCounterNames)); k1++ {
		m.doSwSingle(f, &m.swIfCounterNames, zero, k1, k1, i)
	}

	// Next user-defined counters.
	for ; k0 < uint(len(m.swIfCounterNames.Combined)); k0++ {
		m.doSwCombined(f, &m.swIfCounterNames, zero, k0, k0, i)
	}
	for ; k1 < uint(len(m.swIfCounterNames.Single)); k1++ {
		m.doSwSingle(f, &m.swIfCounterNames, zero, k1, k1, i)
	}

	// Next hardware software interface counters.
	h := m.HwIfer(m.SupHi(si))
	nm := h.GetSwInterfaceCounterNames()
	for k := uint(0); k < uint(len(nm.Combined)); k++ {
		m.doSwCombined(f, &nm, zero, k, k0+k, i)
	}
	for k := uint(0); k < uint(len(nm.Single)); k++ {
		m.doSwSingle(f, &nm, zero, k, k1+k, i)
	}
}

func (v *Vnet) ForeachSwIfCounter(zero bool, f func(si Si, name string, value uint64)) {
	v.swInterfaces.Foreach(func(x swIf) {
		v.foreachSwIfCounter(zero, x.si, func(name string, value uint64) {
			f(x.si, name, value)
		})
	})
}

func (m *interfaceMain) doHwCombined(f foreachFn, nm *InterfaceCounterNames, zero bool, k, i uint) {
	var v, w CombinedCounter
	for _, t := range m.ifThreads {
		if t != nil {
			t.hw.combined[k].Get(i, &w)
			v.Add(&w)
		}
	}
	if v.Packets != 0 || (zero && k < uint(len(nm.Combined))) {
		f(nm.Combined[k]+" packets", v.Packets)
		f(nm.Combined[k]+" bytes", v.Bytes)
	}
	return
}

func (m *interfaceMain) doHwSingle(f foreachFn, nm *InterfaceCounterNames, zero bool, k, i uint) {
	var v, w uint64
	for _, t := range m.ifThreads {
		if t != nil {
			t.hw.single[k].Get(i, &w)
			v += w
		}
	}
	if v != 0 || (zero && k < uint(len(nm.Single))) {
		f(nm.Single[k], v)
	}
	return
}

func (m *interfaceMain) foreachHwIfCounter(zero bool, hi Hi, f func(name string, value uint64)) {
	h := m.HwIfer(hi)
	t := m.GetIfThread(0)
	nm := h.GetHwInterfaceCounterNames()
	h.GetHwInterfaceCounterValues(t)
	i := uint(hi)
	for k := range t.hw.combined {
		m.doHwCombined(f, &nm, zero, uint(k), i)
	}
	for k := range t.hw.single {
		m.doHwSingle(f, &nm, zero, uint(k), i)
	}
}

func (v *Vnet) ForeachHwIfCounter(zero bool, unixOnly bool, f func(hi Hi, name string, value uint64)) {
	for i := range v.hwIferPool.elts {
		if v.hwIferPool.IsFree(uint(i)) {
			continue
		}
		hwifer := v.hwIferPool.elts[i]
		if unixOnly && !hwifer.IsUnix() {
			continue
		}
		h := hwifer.GetHwIf()
		if h.unprovisioned {
			continue
		}
		v.foreachHwIfCounter(zero, h.hi, func(name string, value uint64) {
			f(h.hi, name, value)
		})
	}
}

func (v *Vnet) syncSwIfCounters() {
	for i := range v.swIfCounterSyncHooks.hooks {
		v.swIfCounterSyncHooks.Get(i)(v)
	}
}

func (m *interfaceMain) syncHwIfCounters() {
	t := m.GetIfThread(0)
	m.hwIferPool.Foreach(func(h HwInterfacer) {
		h.GetHwInterfaceCounterValues(t)
	})
}

func (v *Vnet) clearIfCounters() {
	// Sync before clear so counters have accurate values.
	v.syncSwIfCounters()
	v.syncHwIfCounters()

	m := &v.interfaceMain
	m.timeLastClear = time.Now()
	for _, t := range m.ifThreads {
		if t == nil {
			continue
		}
		t.sw.combined.ClearAll()
		t.sw.single.ClearAll()
		t.hw.combined.ClearAll()
		t.hw.single.ClearAll()
	}
}

func (m *interfaceMain) counterValidate(is_hw bool, i uint) {
	for _, t := range m.ifThreads {
		c := &t.sw
		if is_hw {
			c = &t.hw
		}
		for k := range c.combined {
			c.combined[k].Validate(i)
		}
		for k := range c.single {
			c.single[k].Validate(i)
		}
	}
}

func (m *interfaceMain) counterValidateSw(si Si) { m.counterValidate(false, uint(si)) }
func (m *interfaceMain) counterValidateHw(hi Hi) { m.counterValidate(true, uint(hi)) }

func (m *interfaceMain) counterInit(t *InterfaceThread) {
	t.sw.single.Validate(uint(nBuiltinSingleIfCounters))
	t.sw.combined.Validate(uint(nBuiltinCombinedIfCounters))

	if len(m.swIfCounterNames.Single) < len(builtinSingleIfCounterNames) {
		for i := range builtinSingleIfCounterNames {
			m.addSwCounter(builtinSingleIfCounterNames[i])
		}
	}
	if len(m.swIfCounterNames.Combined) < len(builtinCombinedIfCounterNames) {
		for i := range builtinCombinedIfCounterNames {
			m.addSwCombinedCounter(builtinCombinedIfCounterNames[i])
		}
	}

	m.swInterfaces.Foreach(func(x swIf) {
		h := m.HwIfer(m.SupHi(x.si))
		nm := h.GetSwInterfaceCounterNames()
		if len(nm.Single) > 0 {
			t.sw.single.Validate(uint(len(m.swIfCounterNames.Single)) + uint(len(nm.Single)) - 1)
		}
		if len(nm.Combined) > 0 {
			t.sw.combined.Validate(uint(len(m.swIfCounterNames.Combined)) + uint(len(nm.Combined)) - 1)
		}
	})

	if nSwIfs := m.swInterfaces.Len(); nSwIfs > 0 {
		for i := range t.sw.single {
			t.sw.single[i].Validate(nSwIfs - 1)
		}
		for i := range t.sw.combined {
			t.sw.combined[i].Validate(nSwIfs - 1)
		}
	}

	// Allocate hardware counters based on largest number of names.
	m.hwIferPool.Foreach(func(h HwInterfacer) {
		nm := h.GetHwInterfaceCounterNames()
		if len(nm.Single) > 0 {
			t.hw.single.Validate(uint(len(nm.Single)) - 1)
		}
		if len(nm.Combined) > 0 {
			t.hw.combined.Validate(uint(len(nm.Combined)) - 1)
		}
	})

	if nHwIfs := m.hwIferPool.Len(); nHwIfs > 0 {
		for i := range t.hw.single {
			t.hw.single[i].Validate(nHwIfs - 1)
		}
		for i := range t.hw.combined {
			t.hw.combined[i].Validate(nHwIfs - 1)
		}
	}
}

type interfaceThreadCounters struct {
	single   CountersVec
	combined CombinedCountersVec
}

type InterfaceThread struct {
	// This threads' sw/hw interface counters indexed by counter kind.
	sw, hw interfaceThreadCounters
}

func (c *interfaceThreadCounters) resizeSingle(n, m uint) (i uint) {
	i = c.single.Len()
	c.single.Resize(n)
	for j := uint(0); j < n; j++ {
		c.single[i+j].Validate(m)
	}
	return
}

func (c *interfaceThreadCounters) resizeCombined(n, m uint) (i uint) {
	i = c.combined.Len()
	c.combined.Resize(n)
	for j := uint(0); j < n; j++ {
		c.combined[i+j].Validate(m)
	}
	return
}
