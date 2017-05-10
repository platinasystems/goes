// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip4

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"bytes"
	"unsafe"
)

const (
	AddressBytes              = 4
	AddressBits               = 8 * AddressBytes
	HeaderBytes               = 20
	MoreFragments HeaderFlags = 1 << 13
	DontFragment  HeaderFlags = 1 << 14
	Congestion    HeaderFlags = 1 << 15
)

type Address [AddressBytes]uint8

type HeaderFlags vnet.Uint16

func (h *Header) GetHeaderFlags() HeaderFlags {
	return HeaderFlags(h.Flags_and_fragment_offset.ToHost())
}
func (t HeaderFlags) FromHost() vnet.Uint16 { return vnet.Uint16(t).FromHost() }

type Header struct {
	// 4 bit header length (in 32bit units) and version VVVVLLLL.
	// e.g. for packets w/ no options ip_version_and_header_length == 0x45.
	Ip_version_and_header_length uint8

	// Type of service.
	Tos uint8

	// Total layer 3 packet length including this header.
	Length vnet.Uint16

	// 16-bit number such that Src, Dst, Protocol and Fragment ID together uniquely
	// identify packet for fragment re-assembly.
	Fragment_id vnet.Uint16

	// 3 bits of flags and 13 bits of fragment offset (in units of 8 bytes).
	Flags_and_fragment_offset vnet.Uint16

	// Time to live decremented by router at each hop.
	Ttl uint8

	// Next layer protocol.
	Protocol ip.Protocol

	Checksum vnet.Uint16

	// Source and destination address.
	Src, Dst Address
}

func (a Address) AsUint32() vnet.Uint32     { return *(*vnet.Uint32)(unsafe.Pointer(&a[0])) }
func (a *Address) FromUint32(x vnet.Uint32) { *(*vnet.Uint32)(unsafe.Pointer(&a[0])) = x }
func (a *Address) IsEqual(b *Address) bool  { return a.AsUint32() == b.AsUint32() }
func (a *Address) IsZero() bool             { return a.AsUint32() == 0 }
func (a *Address) Add(x uint64)             { vnet.ByteAdd(a[:], x) }

// Compare 2 addresses for sorting.
func (a *Address) Diff(b *Address) (v int) {
	cmp := int(a.AsUint32().ToHost()) - int(b.AsUint32().ToHost())
	v = 0
	if cmp != 0 {
		v = 1
		if cmp < 0 {
			v = -1
		}
	}
	return
}

func IpAddress(a *ip.Address) *Address { return (*Address)(unsafe.Pointer(&a[0])) }
func (a *Address) ToIp() (v ip.Address) {
	for i := range a {
		v[i] = a[i]
	}
	return
}

// 20 byte ip4 header wide access for efficient checksum.
type header64 struct {
	d64 [2]uint64
	d32 [1]uint32
}

func (h *Header) checksum() vnet.Uint16 {
	i := (*header64)(unsafe.Pointer(h))
	c := ip.Checksum(i.d64[0])
	c = c.AddWithCarry(ip.Checksum(i.d64[1]))
	c = c.AddWithCarry(ip.Checksum(i.d32[0]))
	return ^c.Fold()
}

func (h *Header) ComputeChecksum() vnet.Uint16 {
	var tmp Header = *h
	tmp.Checksum = 0
	return tmp.checksum()
}

func (h *Header) Len() uint { return HeaderBytes }
func (h *Header) Finalize(payload []vnet.PacketHeader) {
	var sum uint
	for _, l := range payload {
		sum += l.Len()
	}
	h.Length.Set(HeaderBytes + sum)
	h.Checksum = 0
	h.Checksum = h.checksum()
}

func (h *Header) Write(b *bytes.Buffer) {
	type t struct{ data [20]byte }
	i := (*t)(unsafe.Pointer(h))
	b.Write(i.data[:])
}
func (h *Header) Read(b []byte) vnet.PacketHeader { return (*Header)(vnet.Pointer(b)) }
