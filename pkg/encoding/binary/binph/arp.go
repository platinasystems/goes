// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/Address_Resolution_Protocol
type ARP struct {
	HTYPE uint16
	PTYPE uint16
	HLEN  uint8
	PLEN  uint8
	OPER  uint16
	SHA   [6]byte
	SPA   [4]byte
	THA   [6]byte
	TPA   [4]byte
}

const SizeofARP = int64(unsafe.Sizeof(ARP{}))

func (p *ARP) ReadFrom(r io.Reader) (int64, error) {
	binint.BigEndianPointer(&p.HTYPE).ReadFrom(r)
	binint.BigEndianPointer(&p.PTYPE).ReadFrom(r)
	binint.BytePointer(&p.HLEN).ReadFrom(r)
	binint.BytePointer(&p.PLEN).ReadFrom(r)
	binint.BigEndianPointer(&p.OPER).ReadFrom(r)
	r.Read(p.SHA[:])
	r.Read(p.SPA[:])
	r.Read(p.THA[:])
	_, err := r.Read(p.TPA[:])
	return SizeofARP, err
}

func (v ARP) WriteTo(w io.Writer) (int64, error) {
	binint.BigEndianValue(v.HTYPE).WriteTo(w)
	binint.BigEndianValue(v.PTYPE).WriteTo(w)
	binint.ByteValue(v.HLEN).WriteTo(w)
	binint.ByteValue(v.PLEN).WriteTo(w)
	binint.BigEndianValue(v.OPER).WriteTo(w)
	w.Write(v.SHA[:])
	w.Write(v.SPA[:])
	w.Write(v.THA[:])
	_, err := w.Write(v.TPA[:])
	return SizeofARP, err
}
