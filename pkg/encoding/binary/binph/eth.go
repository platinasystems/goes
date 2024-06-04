// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/Ethernet_frame
type Eth struct {
	DMAC [6]byte
	SMAC [6]byte
	Type uint16
}

const SizeofEth = int(unsafe.Sizeof(Eth{}))

func (v Eth) AppendTo(data []byte) []byte {
	data = append(data, v.DMAC[:]...)
	data = append(data, v.SMAC[:]...)
	data = binint.AppendBig(data, v.Type)
	return data
}

func (p *Eth) PullFrom(data []byte) []byte {
	if len(data) < SizeofEth {
		return data
	}
	data = Pull(data, p.DMAC[:])
	data = Pull(data, p.SMAC[:])
	data = binint.PullBig(data, &p.Type)
	return data
}

func (p *Eth) IsUnicast() bool   { return (p.DMAC[0] & 1) == 0 }
func (p *Eth) ShouldLearn() bool { return (p.SMAC[0] & 1) == 0 }

func EA64(ea [6]byte) uint64 {
	ea64 := uint64(ea[0]) << 40
	ea64 |= uint64(ea[1]) << 32
	ea64 |= uint64(ea[2]) << 24
	ea64 |= uint64(ea[3]) << 16
	ea64 |= uint64(ea[4]) << 8
	ea64 |= uint64(ea[5])
	return ea64
}
