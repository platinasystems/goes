// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Ethernet hardware interfaces.
package ethernet

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"

	"fmt"
	"unsafe"
)

// Spanning Tree State IEEE 802.1d
type IfSpanningTreeState int

// Possible spanning tree states.
const (
	Disable IfSpanningTreeState = iota + 1
	Block
	Listen
	Learn
	Forward
)

var spanningTreeStateNames = [...]string{
	Disable: "disable",
	Block:   "block",
	Listen:  "listen",
	Learn:   "learn",
	Forward: "forward",
}

func (x IfSpanningTreeState) String() string {
	return elib.StringerHex(spanningTreeStateNames[:], int(x))
}

// Full or half duplex.
type IfDuplex int

const (
	Full IfDuplex = iota + 1
	Half
)

var ifDuplexNames = [...]string{
	Full: "full",
	Half: "half",
}

func (x IfDuplex) String() string { return elib.StringerHex(ifDuplexNames[:], int(x)) }

// Physical interface between ethernet MAC and PHY.
type PhyInterface int

// Mac to PHY physical interface types.  Sorted alphabetically.
const (
	CAUI PhyInterface = iota + 1
	CR
	CR2
	CR4
	GMII
	INTERLAKEN
	KR
	KR2
	KR4
	KX
	LR
	LR4
	MII
	QSGMII
	RGMII
	RXAUI
	SFI
	SGMII
	SPAUI
	SR
	SR10
	SR2
	SR4
	XAUI
	XFI
	XGMII
	XLAUI
	XLAUI2
	ZR
)

var phyInterfaceNames = [...]string{
	CAUI:       "caui",
	CR:         "cr",
	CR2:        "cr2",
	CR4:        "cr4",
	GMII:       "gmii",
	INTERLAKEN: "interlaken",
	KR:         "kr",
	KR2:        "kr2",
	KR4:        "kr4",
	KX:         "kx",
	LR:         "lr",
	LR4:        "lr4",
	MII:        "mii",
	QSGMII:     "qsgmii",
	RGMII:      "rgmii",
	RXAUI:      "rxaui",
	SFI:        "sfi",
	SGMII:      "sgmii",
	SPAUI:      "spaui",
	SR:         "sr",
	SR10:       "sr10",
	SR2:        "sr2",
	SR4:        "sr4",
	XAUI:       "xaui",
	XFI:        "xfi",
	XGMII:      "xgmii",
	XLAUI:      "xlaui",
	XLAUI2:     "xlaui2",
	ZR:         "zr",
}

func (x PhyInterface) String() string { return elib.StringerHex(phyInterfaceNames[:], int(x)) }

type InterfaceConfig struct {
	Address       Address
	PhyInterface  PhyInterface
	NativeVlan    Vlan
	Unprovisioned bool
}

type Interface struct {
	vnet.HwIf

	InterfaceConfig

	duplex IfDuplex

	autoNegotiation bool

	spanningTreeState IfSpanningTreeState
	loopback          vnet.IfLoopbackType
}

func (i *Interface) GetInterface() *Interface { return i }

type HwInterfacer interface {
	GetInterface() *Interface
	vnet.HwInterfacer
}

type IfId vnet.IfId

// 32 bit Id: 16 bit outer/inner id: 12 bit id + valid bit
func (i IfId) inner() IfId { return i >> 16 }
func (i IfId) outer() IfId { return i & 0xffff }
func (i IfId) valid() bool { return i&(1<<15) != 0 }
func (i IfId) id() (id vnet.Uint16, valid bool) {
	valid = i.valid()
	if valid {
		id = vnet.Uint16(i & 0xfff)
	}
	return
}
func (i IfId) OuterVlan() (id vnet.Uint16, valid bool) { return i.outer().id() }
func (i IfId) InnerVlan() (id vnet.Uint16, valid bool) { return i.inner().id() }
func (i *IfId) Set(outer vnet.Uint16)                  { *i = IfId(outer) | 1<<15 }
func (i *IfId) Set2(outer, inner vnet.Uint16)          { *i = IfId(outer) | 1<<15 | IfId(inner)<<16 | 1<<31 }

func (i *Interface) LessThanId(aʹ, bʹ vnet.IfId) bool {
	a, b := IfId(aʹ), IfId(bʹ)

	// Compare outer then inner vlan.
	{
		ai, av := a.OuterVlan()
		bi, bv := b.OuterVlan()
		if av && bv && ai != bi {
			return ai < bi
		}
	}
	{
		ai, av := a.InnerVlan()
		bi, bv := b.InnerVlan()
		if av && bv && ai != bi {
			return ai < bi
		}
	}
	// Vlans not valid.
	return a < b
}

func (intf *Interface) ParseId(a *vnet.IfId, in *parse.Input) bool {
	var (
		i int
		v []int
	)
	for !in.End() {
		switch {
		case in.Parse(".%d", &i) && i <= 0xfff:
			v = append(v, i)
		default:
			return false
		}
		if len(v) > 2 {
			break
		}
	}
	switch {
	case len(v) == 1:
		*a = vnet.IfId(1<<15 | v[0])
	case len(v) == 2:
		*a = vnet.IfId(1<<15 | v[0] | 1<<31 | v[1]<<16)
	default:
		return false
	}
	return true
}

func (i *Interface) FormatId(aʹ vnet.IfId) (v string) {
	a := IfId(aʹ)
	oi, ov := a.OuterVlan()
	ii, iv := a.InnerVlan()
	if ov {
		v += fmt.Sprintf(".%d", oi)
	}
	if iv {
		v += fmt.Sprintf(".%d", ii)
	}
	if !iv && !ov {
		v = fmt.Sprintf("invalid 0x%x", a)
	}
	return
}

// See vnet.Arper interface.
// Dummy function to mark ethernet interfaces as supporting ARP.
func (i *Interface) SupportsArp() {}

func RegisterInterface(v *vnet.Vnet, hi HwInterfacer, config *InterfaceConfig, format string, args ...interface{}) {
	i := hi.GetInterface()
	i.InterfaceConfig = *config
	v.RegisterAndProvisionHwInterface(hi, !config.Unprovisioned, format, args...)
}

func (hi *Interface) FormatAddress() string    { return hi.Address.String() }
func (hi *Interface) GetAddress() []byte       { return hi.Address[:] }
func (hi *Interface) SetAddress(a []byte)      { copy(hi.Address[:], a) }
func (hi *Interface) EthernetAddress() Address { return hi.Address }

var rewriteTypeMap = [...]Type{
	vnet.IP4:            TYPE_IP4,
	vnet.IP6:            TYPE_IP6,
	vnet.MPLS_UNICAST:   TYPE_MPLS_UNICAST,
	vnet.MPLS_MULTICAST: TYPE_MPLS_MULTICAST,
	vnet.ARP:            TYPE_ARP,
}

func (et *Type) SetPacketType(pt vnet.PacketType) { *et = rewriteTypeMap[pt].FromHost() }

type rwHeader struct {
	Header
	vlan [2]VlanHeader
}

func (hi *Interface) SetRewrite(v *vnet.Vnet, rw *vnet.Rewrite, packetType vnet.PacketType, da []byte) {
	var h rwHeader
	sw := v.SwIf(rw.Si)
	t := rewriteTypeMap[packetType].FromHost()
	size := uintptr(SizeofHeader)
	id := IfId(sw.Id(v))
	outer_id, outer_valid := id.OuterVlan()
	inner_id, inner_valid := id.InnerVlan()
	if outer_valid {
		if inner_valid {
			h.Type = TYPE_VLAN.FromHost()
			h.vlan[0].Tag = VlanTag(outer_id).FromHost()
			h.vlan[0].Type = h.Type
			h.vlan[1].Tag = VlanTag(inner_id).FromHost()
			h.vlan[1].Type = t
			size += 2 * SizeofVlanHeader
		} else {
			h.Type = TYPE_VLAN.FromHost()
			h.vlan[0].Tag = VlanTag(outer_id).FromHost()
			h.vlan[0].Type = t
			size += SizeofVlanHeader
		}
	} else {
		h.Type = t
	}
	if len(da) > 0 {
		copy(h.Dst[:], da)
	} else {
		h.Dst = BroadcastAddr
	}
	copy(h.Src[:], hi.Address[:])
	rw.ResetData()
	rw.AddData(unsafe.Pointer(&h), size)
}

func (t Type) isVlan() bool {
	switch t.ToHost() {
	case TYPE_VLAN, TYPE_VLAN_IN_VLAN, TYPE_VLAN_802_1AD:
		return true
	default:
		return false
	}
}

func (h *rwHeader) nTags() (n uint) {
	if h.Type.isVlan() {
		n++
		if h.vlan[0].Type.isVlan() {
			n++
		}
	}
	return
}

func (h *rwHeader) Sizeof() uint { return SizeofHeader + h.nTags()*SizeofVlanHeader }
func (h *rwHeader) InnerType() (t Type) {
	t = h.Type
	if t.isVlan() {
		t = h.vlan[0].Type
		if t.isVlan() {
			t = h.vlan[1].Type
		}
	}
	return
}

func (h *rwHeader) String() (s string) {
	nTags := h.nTags()
	if nTags == 0 {
		return h.Header.String()
	}
	var tmp rwHeader
	tmp.Header = h.Header
	tmp.Header.Type = h.vlan[nTags-1].Type // inner type
	tmp.vlan[0] = h.vlan[0]
	tmp.vlan[0].Type = h.Type
	if nTags > 1 {
		tmp.vlan[1] = h.vlan[1]
		tmp.vlan[1].Type = h.vlan[0].Type
	}
	s = tmp.Header.String()
	for i := uint(0); i < nTags; i++ {
		s += " " + tmp.vlan[i].String()
	}
	return
}

func (hi *Interface) FormatRewrite(r *vnet.Rewrite) []string { return FormatRewrite(hi.GetVnet(), r) }

func FormatRewrite(v *vnet.Vnet, r *vnet.Rewrite) (lines []string) {
	h := (*rwHeader)(r.GetData())
	b := r.Slice()
	lines = append(lines, h.String())
	i := h.Sizeof()
	innerType := h.InnerType()
	if i < uint(len(b)) {
		m := GetMain(v)
		if l, ok := m.layerMap[innerType.ToHost()]; ok {
			lines = append(lines, l.FormatLayer(b[i:])...)
		} else {
			panic(fmt.Errorf("no formatter for type %s", innerType.FromHost()))
		}
	}
	return
}

func (hi *Interface) ParseRewrite(r *vnet.Rewrite, in *parse.Input) {
	var h HeaderParser
	innerType := h.Parse(in)
	b := r.Data()
	h.Write(b)
	i := h.Sizeof()
	if !in.End() {
		m := GetMain(hi.GetVnet())
		if l, ok := m.layerMap[innerType.ToHost()]; ok {
			i += l.ParseLayer(b[i:], in)
		} else {
			panic(fmt.Errorf("no parser for type %s: %s", innerType.FromHost(), in))
		}
	}
	r.SetData(b[:i])
}

// Block of ethernet addresses for allocation by a switch.
type AddressBlock struct {
	// Base ethernet address (stored in board's EEPROM).
	Base Address

	// Number of addresses starting at base.
	Count uint32

	nAlloc  uint32
	freeMap map[uint32]struct{}
}

func (a *Address) add(offset uint32) {
	for i, o := 0, offset; o != 0 && i < SizeofAddress; i++ {
		j := SizeofAddress - 1 - i
		x := uint8(o)
		y := a[j]
		y += x
		a[j] = y
		o >>= 8
		// Add in carry.
		if y < x {
			o++
		}
	}
}

func (b *AddressBlock) Alloc() (Address, bool) {
	a := b.Base
	ok := false
	var offset uint32
	for o, _ := range b.freeMap {
		delete(b.freeMap, o)
		offset = o
		ok = true
		break
	}
	if !ok {
		if ok = b.nAlloc < b.Count; ok {
			offset = b.nAlloc
			b.nAlloc++
		}
	}
	if ok {
		a.add(offset)
	}
	return a, ok
}

func (b *AddressBlock) Free(a *Address) {
	offset := uint64(0)
	for i := range a {
		j := SizeofAddress - 1 - i
		offset += uint64(a[j]-b.Base[j]) << uint64(8*i)
	}

	if b.freeMap == nil {
		b.freeMap = make(map[uint32]struct{})
	}
	o := uint32(offset)
	if o >= b.Count {
		panic("bad free")
	}
	if _, ok := b.freeMap[o]; ok {
		panic("duplicate free")
	}
	b.freeMap[o] = struct{}{}
}
