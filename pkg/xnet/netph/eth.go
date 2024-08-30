// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"encoding/binary"
	"io"
)

// https://en.wikipedia.org/wiki/Ethernet_frame
type Eth struct {
	DMAC [6]byte
	SMAC [6]byte
	Type uint16
}

func (p *Eth) ReadFrom(r io.Reader) (int64, error) {
	return SizeofEth, binary.Read(r, binary.BigEndian, p)
}

func (v Eth) WriteTo(w io.Writer) (int64, error) {
	return SizeofEth, binary.Write(w, binary.BigEndian, v)
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
