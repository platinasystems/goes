// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package m

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/devices/ethernet/switch/fe1/internal/sbus"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"

	"fmt"
	"unsafe"
)

type MemElt byte
type Mem [MemMax]MemElt

// Tables always take 0x40000 of address space.  Max index < 0x40000 = 1<<18
const MemMax = 1 << 18

func (e *MemElt) Address() sbus.Address {
	return sbus.Address(uintptr(unsafe.Pointer(e)) - BaseAddress)
}

type M32 MemElt
type M64 MemElt

func (r *M32) Offset() uint { return uint(uintptr(unsafe.Pointer(r)) - BaseAddress) }
func (r *M64) Offset() uint { return (*M32)(r).Offset() }

func (r *M32) Address() sbus.Address { return sbus.Address(r.Offset()) }
func (r *M64) Address() sbus.Address { return sbus.Address(r.Offset()) }

func (r *M32) Geta(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, a sbus.Address, v *uint32) {
	q.GetM32(v, b, r.Address()|a, c)
}
func (r *M32) Get(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, v *uint32) {
	r.Geta(q, b, c, 0, v)
}

func (r *M32) Seta(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, a sbus.Address, v uint32) {
	q.SetM32(v, b, r.Address()|a, c)
}
func (r *M32) Set(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, v uint32) {
	r.Seta(q, b, c, 0, v)
}

func (r *M64) Geta(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, a sbus.Address, v *uint64) {
	q.GetM64(v, b, r.Address()|a, c)
}
func (r *M64) Get(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, v *uint64) {
	r.Geta(q, b, c, 0, v)
}
func (r *M64) Seta(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, a sbus.Address, v uint64) {
	q.SetM64(v, b, r.Address()|a, c)
}
func (r *M64) Set(q *sbus.DmaRequest, b sbus.Block, c sbus.AccessType, v uint64) {
	r.Seta(q, b, c, 0, v)
}

type MemGetSetter interface {
	// Number of bits in memory.
	MemBits() int
	// Method to encode/decode memory as slice of uint32.
	MemGetSet(b []uint32, isSet bool)
}

type memDmaRw struct {
	sbus.DmaCmd
	m MemGetSetter
	// Enough space for largest schan command.
	buf [22]uint32
}

func (r *memDmaRw) Pre() {
	n := r.m.MemBits()
	// Round to next 32 bit word.
	b := r.buf[:(n+31)/32]
	if r.IsWrite() {
		r.m.MemGetSet(b, true)
		r.Tx = b
	} else {
		r.Rx = b
	}
}
func (r *memDmaRw) Post() {
	if r.IsRead() {
		r.m.MemGetSet(r.Rx, false)
	}
}

func (r *MemElt) memDmaGetSet(q *sbus.DmaRequest, m MemGetSetter, o sbus.Opcode, b sbus.Block, t sbus.AccessType, a sbus.Address) {
	rw := memDmaRw{
		DmaCmd: sbus.DmaCmd{
			Command: sbus.Command{Opcode: o, Block: b, AccessType: t},
			Address: r.Address() | a,
		},
		m: m,
	}
	q.Add(&rw)
}

func (r *MemElt) MemDmaGet(q *sbus.DmaRequest, m MemGetSetter, b sbus.Block, t sbus.AccessType) {
	r.memDmaGetSet(q, m, sbus.ReadMemory, b, t, 0)
}
func (r *MemElt) MemDmaSet(q *sbus.DmaRequest, m MemGetSetter, b sbus.Block, t sbus.AccessType) {
	r.memDmaGetSet(q, m, sbus.WriteMemory, b, t, 0)
}
func (r *MemElt) MemDmaGeta(q *sbus.DmaRequest, m MemGetSetter, b sbus.Block, t sbus.AccessType, a sbus.Address) {
	r.memDmaGetSet(q, m, sbus.ReadMemory, b, t, a)
}
func (r *MemElt) MemDmaSeta(q *sbus.DmaRequest, m MemGetSetter, b sbus.Block, t sbus.AccessType, a sbus.Address) {
	r.memDmaGetSet(q, m, sbus.WriteMemory, b, t, a)
}

func MemGet1(x []uint32, lo int) bool {
	l0, l1 := uint(lo/32), uint(lo%32)
	return x[l0]&(1<<l1) != 0
}

func MemSet1(x []uint32, lo int, v bool) {
	l0, l1 := uint(lo/32), uint(lo%32)
	m := uint32(1) << l1
	if v {
		x[l0] |= m
	} else {
		x[l0] &^= m
	}
}
func MemGetSet1(v *bool, x []uint32, lo int, isSet bool) int {
	if isSet {
		MemSet1(x, lo, *v)
	} else {
		*v = MemGet1(x, lo)
	}
	return lo + 1
}

// Get or Set bits lo <= i <= hi, so hi - lo + 1 bits total.
func MemGetSet(v *uint64, x []uint32, hi, lo int, isSet bool) int {
	nBits := 1 + uint(hi-lo)
	if nBits > 64 {
		panic(fmt.Errorf("more than 64 bits"))
	}
	nLeft := nBits
	r := uint64(0)
	if isSet {
		r = *v
	}
	nDone := uint(0)
	i := uint(lo)
	for nLeft > 0 {
		i0, i1 := uint(i/32), uint(i%32)
		m := 32 - i1
		if m > nLeft {
			m = nLeft
		}
		mask := uint64(1)<<m - 1
		if isSet {
			x[i0] |= uint32(((r >> nDone) & mask) << i1)
		} else {
			r |= ((uint64(x[i0]) >> i1) & mask) << nDone
		}
		nDone += m
		nLeft -= m
		i += m
	}
	if !isSet {
		*v = r
	}
	return hi + 1
}

func MemGet(x []uint32, hi, lo int) (v uint64) { MemGetSet(&v, x, hi, lo, false); return }
func MemSet(x []uint32, hi, lo int, v uint64)  { MemGetSet(&v, x, hi, lo, true) }

func MemGetSetUint8(v *uint8, x []uint32, hi, lo int, isSet bool) int {
	w := uint64(*v)
	MemGetSet(&w, x, hi, lo, isSet)
	*v = uint8(w)
	return hi + 1
}

func MemGetSetUint16(v *uint16, x []uint32, hi, lo int, isSet bool) int {
	w := uint64(*v)
	MemGetSet(&w, x, hi, lo, isSet)
	*v = uint16(w)
	return hi + 1
}

func MemGetSetUint32(v *uint32, x []uint32, hi, lo int, isSet bool) int {
	w := uint64(*v)
	MemGetSet(&w, x, hi, lo, isSet)
	*v = uint32(w)
	return hi + 1
}

func MemGetSetUint64(v *uint64, x []uint32, hi, lo int, isSet bool) int {
	w := *v
	MemGetSet(&w, x, hi, lo, isSet)
	*v = uint64(w)
	return hi + 1
}

// TCAM x/y cell encoding: x = mask & key; y = mask &^ key
// x/y are the bits that are masked with value of 1/0 respectively.
// Decoding: key = x, mask = x | y
type TcamUint8 uint8
type TcamUint16 uint16
type TcamUint32 uint32
type TcamUint64 uint64

func tcamEncodeBool(a, b bool, isSet bool) (c, d bool) {
	if isSet {
		c, d = b && a, b && !a
	} else {
		c, d = a, a || b
	}
	return
}

func (a TcamUint8) TcamEncode(b TcamUint8, isSet bool) (c, d TcamUint8) {
	if isSet {
		c, d = b&a, b&^a
	} else {
		c, d = a, a|b
	}
	return
}
func (a TcamUint16) TcamEncode(b TcamUint16, isSet bool) (c, d TcamUint16) {
	if isSet {
		c, d = b&a, b&^a
	} else {
		c, d = a, a|b
	}
	return
}
func (a TcamUint32) TcamEncode(b TcamUint32, isSet bool) (c, d TcamUint32) {
	if isSet {
		c, d = b&a, b&^a
	} else {
		c, d = a, a|b
	}
	return
}
func (a TcamUint64) TcamEncode(b TcamUint64, isSet bool) (c, d TcamUint64) {
	if isSet {
		c, d = b&a, b&^a
	} else {
		c, d = a, a|b
	}
	return
}

type LogicalPort struct {
	isTrunk bool

	// Module number when not a trunk; reserved otherwise.
	module uint8

	// Port number when not a trunk; trunk id when port is a trunk.
	number uint16
}

var LogicalPortMaskAll = LogicalPort{
	isTrunk: true,
	module:  ^uint8(0),
	number:  ^uint16(0),
}

func (x *LogicalPort) Set(n uint) { x.number = uint16(n) }

func (x *LogicalPort) Uint32() (v uint32) {
	v = uint32(x.number)
	v |= uint32(x.module) << 8
	if x.isTrunk {
		v |= 1 << 17
	}
	return
}

func (x *LogicalPort) FromUint32(v uint32) {
	x.isTrunk = v&(1<<17) != 0
	x.module = uint8((v >> 8) & 0xff)
	x.number = uint16(v & 0xff)
}

func (a *LogicalPort) TcamEncode(b *LogicalPort, isSet bool) (c, d LogicalPort) {
	c.isTrunk, d.isTrunk = tcamEncodeBool(a.isTrunk, b.isTrunk, isSet)

	{
		x, y := TcamUint8(a.module).TcamEncode(TcamUint8(b.module), isSet)
		c.module, d.module = uint8(x), uint8(y)
	}

	{
		x, y := TcamUint16(a.number).TcamEncode(TcamUint16(b.number), isSet)
		c.number, d.number = uint16(x), uint16(y)
	}

	return
}

func (x *LogicalPort) MemGetSet(b []uint32, lo int, isSet bool) int {
	p := uint64(0)
	if isSet {
		p = uint64(x.number)
		if x.isTrunk {
			p |= 1 << 17
		} else {
			p |= uint64(x.module) << 8
		}
	}
	MemGetSet(&p, b, lo+16, lo+0, isSet)
	if !isSet {
		var hp LogicalPort
		if p&(1<<17) != 0 {
			hp.isTrunk = true
			hp.number = uint16(p & 0xffff)
		} else {
			hp.number = uint16(p & 0xff)
			hp.module = uint8((p >> 8) & 0xff)
		}
		*x = hp
	}
	return lo + 17
}

type NextHop struct {
	// Whether index is next hop index of ECMP index.
	IsECMP bool

	Index uint16
}

func (n *NextHop) MemGetSet(b []uint32, lo int, isSet, hasReservedBit bool) int {
	i := MemGetSet1(&n.IsECMP, b, lo, isSet)
	nBits := 15
	if hasReservedBit {
		nBits = 16
	}
	i = MemGetSetUint16(&n.Index, b, i+nBits-1, i, isSet)
	return i
}

type NatEditIndex struct {
	// Selects entry 1 else entry 0
	use_entry_1 bool
	index       uint16
}

func (x *NatEditIndex) MemGetSet(b []uint32, i int, isSet bool) int {
	i = MemGetSetUint16(&x.index, b, i+9, i, isSet)
	i = MemGetSet1(&x.use_entry_1, b, i, isSet)
	return i
}

type PriorityChange struct {
	Enable bool

	// 4 bit priority
	Priority uint8
}

func (x *PriorityChange) MemGetSet(b []uint32, lo int, isSet bool) int {
	i := MemGetSetUint8(&x.Priority, b, lo+3, lo, isSet)
	i = MemGetSet1(&x.Enable, b, i, isSet)
	return i
}

// 6 bit class id for *FP lookups.
type FpClassId uint8

func (x *FpClassId) MemGetSet(b []uint32, lo int, isSet bool) int {
	return MemGetSetUint8((*uint8)(x), b, lo+5, lo, isSet)
}

type EthernetAddress ethernet.Address

var EthernetAddressMaskAll = EthernetAddress{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

func (x *EthernetAddress) MemGetSet(b []uint32, lo int, isSet bool) int {
	var mac [2]uint32
	y := (*ethernet.Address)(x)
	if isSet {
		mac[0] = uint32(y[2])<<24 | uint32(y[3])<<16 | uint32(y[4])<<8 | uint32(y[5])
		mac[1] = uint32(y[0])<<8 | uint32(y[1])
	}
	i := MemGetSetUint32(&mac[0], b, lo+31, lo, isSet)
	i = MemGetSetUint32(&mac[1], b, i+15, i, isSet)
	if !isSet {
		y[0] = byte(mac[1] >> 8)
		y[1] = byte(mac[1] >> 0)
		y[2] = byte(mac[0] >> 24)
		y[3] = byte(mac[0] >> 16)
		y[4] = byte(mac[0] >> 8)
		y[5] = byte(mac[0] >> 0)
	}
	return lo + 48
}

func (a *EthernetAddress) TcamEncode(b *EthernetAddress, isSet bool) (c, d EthernetAddress) {
	aʹ, bʹ := TcamUint64((*ethernet.Address)(a).ToUint64()), TcamUint64((*ethernet.Address)(b).ToUint64())
	cʹ, dʹ := aʹ.TcamEncode(bʹ, isSet)
	var ec, ed ethernet.Address
	ec.FromUint64(vnet.Uint64(cʹ))
	ed.FromUint64(vnet.Uint64(dʹ))
	c, d = EthernetAddress(ec), EthernetAddress(ed)
	return
}

type Vlan ethernet.Vlan

func (a Vlan) TcamEncode(b Vlan, isSet bool) (c, d Vlan) {
	x, y := TcamUint16(a).TcamEncode(TcamUint16(b), isSet)
	c, d = Vlan(x), Vlan(y)
	return
}

func (a *Vlan) MemGetSet(b []uint32, lo int, isSet bool) int {
	v := uint64(*a)
	MemGetSet(&v, b, lo+11, lo, isSet)
	if !isSet {
		*a = Vlan(v)
	}
	return lo + 12
}

type Ip4Address ip4.Address

func (a Ip4Address) TcamEncode(b Ip4Address, isSet bool) (c, d Ip4Address) {
	var s, t ip4.Address
	s, t = ip4.Address(a), ip4.Address(b)
	q, r := TcamUint32(s.AsUint32()), TcamUint32(t.AsUint32())
	qʹ, rʹ := q.TcamEncode(r, isSet)
	s.FromUint32(vnet.Uint32(qʹ))
	t.FromUint32(vnet.Uint32(rʹ))
	c, d = Ip4Address(s), Ip4Address(t)
	return
}

func (x *Ip4Address) MemGetSet(b []uint32, lo int, isSet bool) int {
	ip := vnet.Uint32(0)
	y := (*ip4.Address)(x)
	if isSet {
		ip = y.AsUint32().FromHost()
	}
	MemGetSetUint32((*uint32)(&ip), b, lo+31, lo, isSet)
	if !isSet {
		y.FromUint32(vnet.Uint32(ip.ToHost()))
	}
	return lo + 32
}

type Ip6Address ip6.Address

func (x *Ip6Address) MemGetSet(b []uint32, lo int, isSet bool) int {
	var v uint32
	y := (*ip6.Address)(x)
	if isSet {
		for i := 0; i < 4; i++ {
			v = y.Uint32(i)
			lo = MemGetSetUint32(&v, b, lo+31, lo, isSet)
		}
	} else {
		for i := 0; i < 4; i++ {
			lo = MemGetSetUint32(&v, b, lo+31, lo, isSet)
			y.FromUint32(i, v)
		}
	}
	return lo + 128
}

type IpPort uint16

func (x *IpPort) MemGetSet(b []uint32, lo int, isSet bool) int {
	return MemGetSetUint16((*uint16)(x), b, lo+15, lo, isSet)
}

type Vrf uint16

func (a Vrf) TcamEncode(b Vrf, isSet bool) (c, d Vrf) {
	x, y := TcamUint16(a).TcamEncode(TcamUint16(b), isSet)
	c, d = Vrf(x), Vrf(y)
	return
}

func (x *Vrf) MemGetSet(b []uint32, lo int, isSet bool) int {
	return MemGetSetUint16((*uint16)(x), b, lo+10, lo, isSet)
}

type MplsLabel uint32

func (x *MplsLabel) MemGetSet(b []uint32, lo int, isSet bool) int {
	return MemGetSetUint32((*uint32)(x), b, lo+19, lo, isSet)
}

type SpanningTreeState uint8

const (
	SpanningTreeDisabled SpanningTreeState = iota
	SpanningTreeBlocking
	SpanningTreeLearning
	SpanningTreeForwarding
)

func (x *SpanningTreeState) MemGetSet(b []uint32, lo int, isSet bool) int {
	return MemGetSetUint8((*uint8)(x), b, lo+1, lo, isSet)
}
