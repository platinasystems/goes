// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip6

import (
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ip"

	"unsafe"
)

const (
	AddressBytes = 16
	SizeofHeader = 40
)

type Address [AddressBytes]byte

type Header struct {
	/* 4 bit version, 8 bit traffic class and 20 bit flow label. */
	Ip_version_traffic_class_and_flow_label uint32

	/* Total packet length not including this header (but including
	   any extension headers if present). */
	Payload_length uint16

	/* Protocol for next header. */
	Protocol ip.Protocol

	/* Hop limit decremented by router at each hop. */
	Ttl uint8

	/* Source and destination address. */
	Src, Dst Address
}

func (a *Address) Uint32(i int) uint32 {
	return uint32(a[4*i+3]) | uint32(a[4*i+2])<<8 | uint32(a[4*i+1])<<16 | uint32(a[4*i+0])<<24
}

func (a *Address) IsZero() bool {
	for i := range a {
		if a[i] != 0 {
			return false
		}
	}
	return true
}

func (a *Address) FromUint32(i int, x uint32) {
	a[4*i+0] = byte(x >> 24)
	a[4*i+1] = byte(x >> 16)
	a[4*i+2] = byte(x >> 8)
	a[4*i+3] = byte(x)
}

func IpAddress(a *ip.Address) *Address { return (*Address)(unsafe.Pointer(&a[0])) }

func (h *Header) Len() int { return SizeofHeader }
func (h *Header) Write(b []byte) {
	type t struct{ data [SizeofHeader]byte }
	i := (*t)(unsafe.Pointer(h))
	copy(b[:], i.data[:])
}
func (h *Header) Finalize(hs []vnet.PacketHeader) {
	sum := uint(0)
	for _, h := range hs {
		sum += h.Len()
	}
	h.Payload_length = uint16(sum)
}
