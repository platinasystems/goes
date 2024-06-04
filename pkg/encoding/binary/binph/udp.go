// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/User_Datagram_Protocol
type UDP struct {
	SP  uint16
	DP  uint16
	Len uint16
	Sum uint16
}

const SizeofUDP = int(unsafe.Sizeof(UDP{}))

func (v UDP) AppendTo(data []byte) []byte {
	data = binint.AppendBig(data, v.SP)
	data = binint.AppendBig(data, v.DP)
	data = binint.AppendBig(data, v.Len)
	data = binint.AppendBig(data, v.Sum)
	return data
}

func (p *UDP) PullFrom(data []byte) []byte {
	if len(data) < SizeofUDP {
		return data
	}
	data = binint.PullBig(data, &p.SP)
	data = binint.PullBig(data, &p.DP)
	data = binint.PullBig(data, &p.Len)
	data = binint.PullBig(data, &p.Sum)
	return data
}
