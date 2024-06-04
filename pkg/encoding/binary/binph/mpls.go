// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/Multiprotocol_Label_Switching
type MPLS uint32

const SizeofMPLS = int(unsafe.Sizeof(MPLS(0)))

func (v MPLS) AppendTo(data []byte) []byte {
	return binint.AppendBig(data, v)
}

func (p *MPLS) PullFrom(data []byte) []byte {
	if len(data) < SizeofMPLS {
		return data
	}
	data = binint.PullBig(data, p)
	return data
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
