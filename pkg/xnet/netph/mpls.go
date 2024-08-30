// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"encoding/binary"
	"io"
)

// https://en.wikipedia.org/wiki/Multiprotocol_Label_Switching
type MPLS uint32

func (p *MPLS) ReadFrom(r io.Reader) (int64, error) {
	return SizeofMPLS, binary.Read(r, binary.BigEndian, p)
}

func (v MPLS) WriteTo(w io.Writer) (int64, error) {
	return SizeofMPLS, binary.Write(w, binary.BigEndian, v)
}

func (v MPLS) Label() uint32 { return uint32(v) >> (3 + 1 + 8) }
func (v MPLS) TC() uint8     { return uint8(v>>9) & 7 }
func (v MPLS) IsBOS() bool   { return (v & (1 << 8)) != 0 }
func (v MPLS) TTL() uint8    { return uint8(v) }

func ConstructMPLS(label uint32, tc uint8, isbos bool, ttl uint8) MPLS {
	v := MPLS(ttl)
	if isbos {
		v |= (1 << 8)
	}
	v |= MPLS(tc & 7)
	v |= MPLS(label&((1<<20)-1)) << (3 + 1 + 8)
	return v
}
