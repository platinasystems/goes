// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"

	"math/rand"
	"unsafe"
)

// Header for ethernet packets as they appear on the network.
type Header struct {
	Dst  Address
	Src  Address
	Type Type
}

// A 12 bit vlan id.
type Vlan vnet.Uint16

// A 16 bit vlan tag in network byte order.
type VlanTag vnet.Uint16

func (t VlanTag) ToHost() VlanTag   { return VlanTag(vnet.Uint16(t).ToHost()) }
func (t VlanTag) FromHost() VlanTag { return VlanTag(vnet.Uint16(t).FromHost()) }
func (v VlanTag) Id() uint16        { return uint16(v.ToHost()) & 0xfff }
func (v VlanTag) Priority() uint8   { return uint8(v.ToHost() >> 13) }
func (v VlanTag) CFI() bool         { return v.ToHost()&(1<<12) != 0 }

// Tagged packets have VlanHeader after ethernet header.
type VlanHeader struct {
	/* 3 bit priority, 1 bit CFI and 12 bit vlan id. */
	Tag VlanTag

	/* Inner ethernet type. */
	Type Type
}

const MaxVlan = 1 << 12

// Packet type from ethernet header.
type Type vnet.Uint16

func (t Type) ToHost() Type   { return Type(vnet.Uint16(t).ToHost()) }
func (t Type) FromHost() Type { return Type(vnet.Uint16(t).FromHost()) }

func (h *Header) GetType() Type      { return h.Type.ToHost() }
func (h *VlanHeader) GetType() Type  { return h.Type.ToHost() }
func (h *Header) TypeIs(t Type) bool { return t.ToHost() == h.Type }

func (h *Header) GetPayload() unsafe.Pointer {
	return unsafe.Pointer(uintptr(unsafe.Pointer(h)) + unsafe.Sizeof(*h))
}

const (
	AddressBytes    = 6
	HeaderBytes     = 14
	VlanHeaderBytes = 4
)

type Address [AddressBytes]byte

var BroadcastAddr = Address{0xff, 0xff, 0xff, 0xff, 0xff, 0xff}

const (
	isBroadcast           = 1 << 0
	isLocallyAdministered = 1 << 1
)

func (a *Address) IsBroadcast() bool {
	return a[0]&isBroadcast != 0
}
func (a *Address) IsLocallyAdministered() bool {
	return a[0]&isLocallyAdministered != 0
}
func (a *Address) IsUnicast() bool {
	return !a.IsBroadcast()
}

func (h *Header) IsBroadcast() bool {
	return h.Dst.IsBroadcast()
}
func (h *Header) IsUnicast() bool {
	return !h.Dst.IsBroadcast()
}

func (a *Address) Add(x uint64) { vnet.ByteAdd(a[:], x) }

func (a *Address) FromUint64(x vnet.Uint64) {
	for i := 0; i < AddressBytes; i++ {
		a[i] = byte((x >> uint(40-8*i)) & 0xff)
	}
}

func (a *Address) ToUint64() (x vnet.Uint64) {
	for i := 0; i < AddressBytes; i++ {
		x |= vnet.Uint64(a[i]) << uint(40-8*i)
	}
	return
}

func (a *Address) Equal(b Address) bool {
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func (a *Address) IsZero() bool {
	for i := range a {
		if a[i] != 0 {
			return false
		}
	}
	return true
}

func RandomAddress() (a Address) {
	for i := range a {
		a[i] = uint8(rand.Int())
	}
	// Make address unicast and locally administered.
	a[0] &^= isBroadcast
	a[0] |= isLocallyAdministered
	return
}

// Implement vnet.Header interface.
func (h *Header) Len() uint { return HeaderBytes }
func (h *Header) Write(b []byte) {
	type t struct{ data [HeaderBytes]byte }
	i := (*t)(unsafe.Pointer(h))
	copy(b[:], i.data[:])
}
func (h *Header) Read(b []byte) vnet.PacketHeader { return (*Header)(vnet.Pointer(b)) }

func (h *VlanHeader) Len() uint { return VlanHeaderBytes }
func (h *VlanHeader) Write(b []byte) {
	type t struct{ data [VlanHeaderBytes]byte }
	i := (*t)(unsafe.Pointer(h))
	copy(b[:], i.data[:])
}
func (h *VlanHeader) Read(b []byte) vnet.PacketHeader { return (*VlanHeader)(vnet.Pointer(b)) }
