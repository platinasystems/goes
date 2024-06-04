// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol
type ICMP struct {
	Type uint8
	Code uint8
	Sum  uint16
}

const SizeofICMP = int(unsafe.Sizeof(ICMP{}))

func (v ICMP) AppendTo(data []byte) []byte {
	data = append(data, v.Type)
	data = append(data, v.Code)
	data = binint.AppendBig(data, v.Sum)
	return data
}

func (p *ICMP) PullFrom(data []byte) []byte {
	if len(data) < SizeofICMP {
		return data
	}
	data = binint.Bite(data, &p.Type)
	data = binint.Bite(data, &p.Code)
	data = binint.PullBig(data, &p.Sum)
	return data
}
