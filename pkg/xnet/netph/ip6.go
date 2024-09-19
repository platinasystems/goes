// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import "unsafe"

const IPv6len = 16

// https://en.wikipedia.org/wiki/IPv6
type IP6 struct {
	VCF        uint32
	LEN        uint16
	NextHeader uint8
	HopLimit   uint8
	SA         [IPv6len]byte
	DA         [IPv6len]byte
}

const IP6LenIndex = 4
const IP6AddrsIndex = 8
const IP6AddrsLen = 2 * 16
const IP6NextHeaderIndex = 6
const IP6Size = int(unsafe.Sizeof(IP6{}))
const IP6FlowMask = ((1 << 20) - 1)

func (p *IP6) Version() uint8 { return uint8(p.VCF >> 28) }
func (p *IP6) Class() uint8   { return uint8(p.VCF >> 20) }
func (p *IP6) Flow() uint32   { return p.VCF & IP6FlowMask }

func ConstructVCF(class uint8, flow uint32) uint32 {
	return ((6 << 28) | (uint32(class) << 20) | (flow & IP6FlowMask))
}
