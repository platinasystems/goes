// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"github.com/platinasystems/go/vnet"

	"bytes"
	"math/rand"
	"unsafe"
)

// Header for ethernet packets as they appear on the network.
type Header struct {
	Dst  Address
	Src  Address
	Type Type
}

type Vlan vnet.Uint16

// Tagged packets have VlanHeader after ethernet header.
type VlanHeader struct {
	/* 3 bit priority, 1 bit CFI and 12 bit vlan id. */
	Priority_cfi_and_id vnet.Uint16

	/* Inner ethernet type. */
	Type Type
}

const MaxVlan = 1 << 12

// Packet type from ethernet header.
type Type vnet.Uint16

func (h *Header) GetType() Type     { return Type(vnet.Uint16(h.Type).ToHost()) }
func (h *VlanHeader) GetType() Type { return Type(vnet.Uint16(h.Type).ToHost()) }
func (t Type) FromHost() Type       { return Type(vnet.Uint16(t).FromHost()) }

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

func (a *Address) Add(x uint64) {
	var i int
	i = AddressBytes - 1
	for x != 0 && i > 0 {
		ai := uint64(a[i])
		y := ai + (x & 0xff)
		a[i] = byte(ai)
		x >>= 8
		if y < ai {
			x += 1
		}
		i--
	}
}

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
func (h *Header) Len() uint                      { return HeaderBytes }
func (h *Header) Finalize(l []vnet.PacketHeader) {}
func (h *Header) Write(b *bytes.Buffer) {
	type t struct{ data [unsafe.Sizeof(*h)]byte }
	i := (*t)(unsafe.Pointer(h))
	b.Write(i.data[:])
}
func (h *Header) Read(b []byte) vnet.PacketHeader { return (*Header)(vnet.Pointer(b)) }

func (h *VlanHeader) Len() uint                      { return VlanHeaderBytes }
func (h *VlanHeader) Finalize(l []vnet.PacketHeader) {}
func (h *VlanHeader) Write(b *bytes.Buffer) {
	type t struct{ data [unsafe.Sizeof(*h)]byte }
	i := (*t)(unsafe.Pointer(h))
	b.Write(i.data[:])
}
func (h *VlanHeader) Read(b []byte) vnet.PacketHeader { return (*VlanHeader)(vnet.Pointer(b)) }
