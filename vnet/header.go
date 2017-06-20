// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"fmt"
	"unsafe"
)

type PacketHeader interface {
	// Number of packet bytes in this layer's payload.
	Len() uint

	// Write this layer's packet data to given slice.
	Write([]byte)

	// Return given slice as packet header type.
	Read([]byte) PacketHeader

	String() string
}

func Pointer(b []byte) unsafe.Pointer { return unsafe.Pointer(&b[0]) }

func MakePacket(args ...PacketHeader) (b []byte) {
	na := len(args)

	// Find total length of packet.
	l := uint(0)
	for _, a := range args {
		l += a.Len()
	}

	// Allocate buffer.
	b = make([]byte, l)

	// Write packet from innermost header to outermost.
	k := uint(0)
	for i := 0; i < na; i++ {
		a := args[na-1-i]
		k += a.Len()
		a.Write(b[l-k:])
	}
	return
}

func ReadPacket(b []byte, args ...PacketHeader) (hs []PacketHeader) {
	i := uint(0)
	for a := range args {
		l := args[a]
		hs = append(hs, l.Read(b[i:]))
		i += l.Len()
	}
	return hs
}

// Packet layer with incrementing data of given byte count.
type IncrementingPayload struct{ Count uint }

func (i *IncrementingPayload) Len() uint                 { return i.Count }
func (i *IncrementingPayload) Finalize(l []PacketHeader) {}
func (i *IncrementingPayload) String() string            { return fmt.Sprintf("incrementing %d", i.Count) }

func (i *IncrementingPayload) Write(b []byte) {
	for j := uint(0); j < i.Count; j++ {
		b[j] = byte(j % 256)
	}
}
func (i *IncrementingPayload) Read(b []byte) PacketHeader { return i }

type GivenPayload struct{ Payload []byte }

func (i *GivenPayload) Len() uint                 { return uint(len(i.Payload)) }
func (i *GivenPayload) Finalize(l []PacketHeader) {}
func (i *GivenPayload) String() string            { return fmt.Sprintf("payload %x", i.Payload) }

func (i *GivenPayload) Write(b []byte) {
	copy(b[:], i.Payload)
}
func (i *GivenPayload) Read(b []byte) PacketHeader { return i }
