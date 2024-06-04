// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/IPv6
type IP6 struct {
	VCF        uint32
	LEN        uint16
	NextHeader uint8
	HopLimit   uint8
	SA         [16]byte
	DA         [16]byte
}

const SizeofIP6 = int(unsafe.Sizeof(IP6{}))
const IP6FlowMask = ((1 << 20) - 1)

func (v IP6) AppendTo(data []byte) []byte {
	data = binint.AppendBig(data, v.VCF)
	data = binint.AppendBig(data, v.LEN)
	data = append(data, v.NextHeader)
	data = append(data, v.HopLimit)
	data = append(data, v.SA[:]...)
	data = append(data, v.DA[:]...)
	return data
}

func (p *IP6) PullFrom(data []byte) []byte {
	if len(data) < SizeofIP6 {
		return data
	}
	data = binint.PullBig(data, &p.VCF)
	data = binint.PullBig(data, &p.LEN)
	data = binint.Bite(data, &p.NextHeader)
	data = binint.Bite(data, &p.HopLimit)
	data = Pull(data, p.SA[:])
	data = Pull(data, p.DA[:])
	return data
}

func (p *IP6) Version() uint8 { return uint8(p.VCF >> 28) }
func (p *IP6) Class() uint8   { return uint8(p.VCF >> 20) }
func (p *IP6) Flow() uint32   { return p.VCF & IP6FlowMask }

func ConstructVCF(class uint8, flow uint32) uint32 {
	return ((6 << 28) | (uint32(class) << 20) | (flow & IP6FlowMask))
}
