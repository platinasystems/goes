// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Network counters + packet/byte counters optimized for large tables and cache locality.
package vnet

import (
	"github.com/platinasystems/go/elib"
	"math"
)

// Array of single counters
type Counters struct {
	// Compact cache-friendly counters that (rarely) can overflow.
	mini elib.Uint16Vec
	// 64 bit counters to hold overflow from mini counters.
	maxi elib.Uint64Vec
	// For testing; nil for normal operation.
	debug elib.Uint64Vec
	// Counter values when last cleared or zero if counters have never been cleared.
	valuesLastClear elib.Uint64Vec
}

//go:generate gentemplate -d Package=vnet -id countersVec -d VecType=CountersVec -d Type=Counters github.com/platinasystems/go/elib/vec.tmpl

func (c *Counters) Validate(i uint) {
	c.mini.Validate(i)
	c.maxi.Validate(i)
	c.valuesLastClear.Validate(i)
}

// Add x to counter with index i.
func (c *Counters) Add(i, x uint) {
	mini := &c.mini[i]
	old := uint(*mini)
	new := old + x
	*mini = uint16(new)
	if uint(*mini) != new {
		c.maxi[i] += uint64(old + x)
		*mini = 0
	}
}

func (c *Counters) Add64(i uint, x uint64) { c.maxi[i] += x }

func (c *Counters) Get(i uint, v *uint64) {
	*v = c.maxi[i] + uint64(c.mini[i]) - c.valuesLastClear[i]
	return
}

func (c *Counters) Value(i uint) (v uint64) {
	c.Get(i, &v)
	return
}

func (c *Counters) Clear(i uint) {
	v := c.maxi[i] + uint64(c.mini[i])
	c.valuesLastClear[i] = v
	c.maxi[i] = v
	c.mini[i] = 0
}

func (c *Counters) ClearAll() {
	for i := range c.valuesLastClear {
		c.Clear(uint(i))
	}
}

func (c CountersVec) ClearAll() {
	for i := range c {
		c[i].ClearAll()
	}
}

type CombinedCounter struct{ Packets, Bytes uint64 }

func (c *CombinedCounter) Zero() {
	c.Packets = 0
	c.Bytes = 0
}

func (c *CombinedCounter) Add(d *CombinedCounter) {
	c.Packets += d.Packets
	c.Bytes += d.Bytes
}

func (c *CombinedCounter) subNoValidate(d *CombinedCounter) {
	c.Packets -= d.Packets
	c.Bytes -= d.Bytes
}

func (c *CombinedCounter) Sub(d *CombinedCounter) {
	if d.Packets > c.Packets || d.Bytes > c.Bytes {
		panic("underflow")
	}
	c.subNoValidate(d)
}

type miniCombinedCounter struct {
	// Packet count.
	packets uint16

	// Byte count = packets * ave-packet-size + byte_difference
	byteDiff int16
}

func (c *miniCombinedCounter) Zero() {
	c.packets = 0
	c.byteDiff = 0
}

//go:generate gentemplate -d Package=vnet -id miniCombinedCounter -d VecType=miniCombinedCounterVec -d Type=miniCombinedCounter github.com/platinasystems/go/elib/vec.tmpl
//go:generate gentemplate -d Package=vnet -id CombinedCounter -d VecType=CombinedCounterVec -d Type=CombinedCounter github.com/platinasystems/go/elib/vec.tmpl

type CombinedCounters struct {
	// Compact cache-friendly counters that may (rarely) overflow.
	mini miniCombinedCounterVec
	// 64 bit counters to hold overflow from mini counters.
	maxi CombinedCounterVec

	valuesLastClear CombinedCounterVec

	// Average packet size in bytes used to compute mini-counter byte differences.
	avePacketSize uint

	// Current summed packets and bytes for average computation.
	sumPackets, sumBytes uint

	avePacketSizeRecomputeInterval uint
}

//go:generate gentemplate -d Package=vnet -id combinedCountersVec -d VecType=CombinedCountersVec -d Type=CombinedCounters github.com/platinasystems/go/elib/vec.tmpl

func (c *CombinedCounters) Validate(i uint) {
	c.mini.Validate(i)
	c.maxi.Validate(i)
	c.valuesLastClear.Validate(i)
}

// Add packet and byte increment to counter.
func (c *CombinedCounters) Add(i uint, p, b uint) {
	mini := &c.mini[i]

	op := uint(mini.packets)
	od := int(mini.byteDiff)

	np := op + p
	nd := od + int(b-p*uint(c.avePacketSize))

	mini.packets = uint16(np)
	mini.byteDiff = int16(nd)

	if uint(mini.packets) != np || int(mini.byteDiff) != nd {
		maxi := &c.maxi[i]

		nb := uint(int(op*uint(c.avePacketSize))+od) + b

		maxi.Packets += uint64(np)
		maxi.Bytes += uint64(nb)
		mini.Zero()

		// Update average packet size.
		c.sumPackets += np
		c.sumBytes += nb
		if c.sumPackets >= c.avePacketSizeRecomputeInterval {
			c.recomputeAvePacketSize()
		}
	}
}

func (c *CombinedCounters) Add64(i uint, p, b uint64) {
	c.maxi[i].Packets += p
	c.maxi[i].Bytes += b
}

// Get counter value: 2 flavors.
func (c *CombinedCounters) Get(i uint, r *CombinedCounter) {
	*r = c.maxi[i]
	c.addMini(&c.mini[i], r)
	r.subNoValidate(&c.valuesLastClear[i])
}

func (c *CombinedCounters) Value(i uint) (v CombinedCounter) {
	c.Get(i, &v)
	return
}

func (c *CombinedCounters) Clear(i uint) {
	v := c.maxi[i]
	c.addMini(&c.mini[i], &v)
	c.valuesLastClear[i] = v
	c.maxi[i] = v
	c.mini[i].Zero()
}

func (c *CombinedCounters) ClearAll() {
	for i := range c.valuesLastClear {
		c.Clear(uint(i))
	}
}

func (c CombinedCountersVec) ClearAll() {
	for i := range c {
		c[i].ClearAll()
	}
}

func (c *CombinedCounters) addMini(mini *miniCombinedCounter, maxi *CombinedCounter) {
	maxi.Packets += uint64(mini.packets)
	maxi.Bytes += uint64(int(uint(mini.packets)*uint(c.avePacketSize)) + int(mini.byteDiff))
}

func (c *CombinedCounters) flushMini(mini *miniCombinedCounter, maxi *CombinedCounter) {
	c.addMini(mini, maxi)
	mini.Zero()
}

func (c *CombinedCounters) recomputeAvePacketSize() {
	if c.avePacketSizeRecomputeInterval == 0 {
		c.avePacketSizeRecomputeInterval = 4 << 10 // sane default
		return
	}

	newAve := uint(math.Floor(.5 + float64(c.sumBytes)/float64(c.sumPackets)))
	if newAve != uint(c.avePacketSize) {
		// Flush counters since ave packet size is about to change.
		i, n_left := 0, len(c.mini)
		for n_left >= 4 {
			c.flushMini(&c.mini[i+0], &c.maxi[i+0])
			c.flushMini(&c.mini[i+1], &c.maxi[i+1])
			c.flushMini(&c.mini[i+2], &c.maxi[i+2])
			c.flushMini(&c.mini[i+3], &c.maxi[i+3])
			i += 4
			n_left -= 4
		}

		for n_left > 0 {
			c.flushMini(&c.mini[i+0], &c.maxi[i+0])
			i++
			n_left--
		}

		c.avePacketSize = newAve
		c.sumBytes, c.sumPackets = 0, 0
	}
}
