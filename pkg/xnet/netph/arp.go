// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"encoding/binary"
	"io"
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
	return SizeofARP, binary.Read(r, binary.BigEndian, p)
}

func (v ARP) WriteTo(w io.Writer) (int64, error) {
	return SizeofARP, binary.Write(w, binary.BigEndian, v)
}
