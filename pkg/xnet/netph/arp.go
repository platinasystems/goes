// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/xnet"
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

func (p *ARP) ReadFrom(r io.Reader) (int64, error) {
	xnet.BigEndianPointer(&p.HTYPE).ReadFrom(r)
	xnet.BigEndianPointer(&p.PTYPE).ReadFrom(r)
	xnet.BytePointer(&p.HLEN).ReadFrom(r)
	xnet.BytePointer(&p.PLEN).ReadFrom(r)
	xnet.BigEndianPointer(&p.OPER).ReadFrom(r)
	r.Read(p.SHA[:])
	r.Read(p.SPA[:])
	r.Read(p.THA[:])
	_, err := r.Read(p.TPA[:])
	return SizeofARP, err
}

func (v ARP) WriteTo(w io.Writer) (int64, error) {
	xnet.BigEndianValue(v.HTYPE).WriteTo(w)
	xnet.BigEndianValue(v.PTYPE).WriteTo(w)
	xnet.ByteValue(v.HLEN).WriteTo(w)
	xnet.ByteValue(v.PLEN).WriteTo(w)
	xnet.BigEndianValue(v.OPER).WriteTo(w)
	w.Write(v.SHA[:])
	w.Write(v.SPA[:])
	w.Write(v.THA[:])
	_, err := w.Write(v.TPA[:])
	return SizeofARP, err
}
