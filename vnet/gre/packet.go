// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package gre

import (
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"

	"unsafe"
)

type Header struct {
	VersionAndFlags
	Type ethernet.Type
}

const SizeofHeader = 4

type VersionAndFlags vnet.Uint16

const (
	ChecksumPresent VersionAndFlags = 1 << 15
	KeyPresent      VersionAndFlags = 1 << 13
	SequencePresent VersionAndFlags = 1 << 12
)

func (x VersionAndFlags) Version() uint { return uint(x) & 7 }

func (h *Header) String() string { return "GRE: " + h.Type.ToHost().String() }

// vnet.PacketHeader interface.
func (h *Header) Len() uint                       { return SizeofHeader }
func (h *Header) Read(b []byte) vnet.PacketHeader { return (*Header)(vnet.Pointer(b)) }
func (h *Header) Write(b []byte) {
	type t struct{ data [SizeofHeader]byte }
	i := (*t)(unsafe.Pointer(h))
	copy(b[:], i.data[:])
}

func (h *Header) Parse(in *parse.Input) {
	h.VersionAndFlags = 0
	if !in.Parse("%v", &h.Type) {
		in.ParseError()
	}
}
