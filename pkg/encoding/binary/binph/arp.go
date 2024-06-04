// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
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

const SizeofARP = int(unsafe.Sizeof(ARP{}))

func (v ARP) AppendTo(data []byte) []byte {
	data = binint.AppendBig(data, v.HTYPE)
	data = binint.AppendBig(data, v.PTYPE)
	data = append(data, v.HLEN)
	data = append(data, v.PLEN)
	data = binint.AppendBig(data, v.OPER)
	data = append(data, v.SHA[:]...)
	data = append(data, v.SPA[:]...)
	data = append(data, v.THA[:]...)
	data = append(data, v.TPA[:]...)
	return data
}

func (p *ARP) PullFrom(data []byte) []byte {
	if len(data) < SizeofARP {
		return data
	}
	data = binint.PullBig(data, &p.HTYPE)
	data = binint.PullBig(data, &p.PTYPE)
	data = binint.Bite(data, &p.HLEN)
	data = binint.Bite(data, &p.PLEN)
	data = binint.PullBig(data, &p.OPER)
	data = Pull(data, p.SHA[:])
	data = Pull(data, p.SPA[:])
	data = Pull(data, p.THA[:])
	data = Pull(data, p.TPA[:])
	return data
}
