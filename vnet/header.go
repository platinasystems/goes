// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"bytes"
	"fmt"
	"unsafe"
)

type PacketHeader interface {
	// Number of packet bytes in this layer's payload.
	Len() uint

	// Finalize this layer given finalized inner layers.
	// This allows, for example, the IP4 layer to compute length and checksum based on payload.
	Finalize([]PacketHeader)

	String() string

	// Append this layer's packet data to given buffer.
	Write(*bytes.Buffer)

	Read([]byte) PacketHeader
}

func Pointer(b []byte) unsafe.Pointer { return unsafe.Pointer(&b[0]) }

func MakePacket(args ...PacketHeader) []byte {
	n := len(args)
	if n == 0 {
		return nil
	}
	for i := 0; i < n; i++ {
		args[n-1-i].Finalize(args[n-i:])
	}
	b := new(bytes.Buffer)
	for i := 0; i < n; i++ {
		b.Grow(int(args[i].Len()))
		args[i].Write(b)
	}
	return b.Bytes()
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

func (i *IncrementingPayload) Write(b *bytes.Buffer) {
	for j := uint(0); j < i.Count; j++ {
		b.WriteByte(byte(j % 256))
	}
}
func (i *IncrementingPayload) Read(b []byte) PacketHeader { return i }
