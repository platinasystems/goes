// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
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

const SizeofIP6 = int64(unsafe.Sizeof(IP6{}))
const IP6FlowMask = ((1 << 20) - 1)

func (p *IP6) ReadFrom(r io.Reader) (int64, error) {
	binint.BigEndianPointer(&p.VCF).ReadFrom(r)
	binint.BigEndianPointer(&p.LEN).ReadFrom(r)
	binint.BytePointer(&p.NextHeader).ReadFrom(r)
	binint.BytePointer(&p.HopLimit).ReadFrom(r)
	r.Read(p.SA[:])
	_, err := r.Read(p.DA[:])
	return SizeofIP6, err
}

func (v IP6) WriteTo(w io.Writer) (int64, error) {
	binint.BigEndianValue(v.VCF).WriteTo(w)
	binint.BigEndianValue(v.LEN).WriteTo(w)
	binint.ByteValue(v.NextHeader).WriteTo(w)
	binint.ByteValue(v.HopLimit).WriteTo(w)
	w.Write(v.SA[:])
	_, err := w.Write(v.DA[:])
	return SizeofIP6, err
}

func (p *IP6) Version() uint8 { return uint8(p.VCF >> 28) }
func (p *IP6) Class() uint8   { return uint8(p.VCF >> 20) }
func (p *IP6) Flow() uint32   { return p.VCF & IP6FlowMask }

func ConstructVCF(class uint8, flow uint32) uint32 {
	return ((6 << 28) | (uint32(class) << 20) | (flow & IP6FlowMask))
}
