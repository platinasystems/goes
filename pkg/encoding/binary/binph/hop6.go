// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type HOP6 struct {
	Type uint8
	Len  uint8
}

const SizeofHOP6 = int(unsafe.Sizeof(HOP6{}))

func (v HOP6) AppendBig(data []byte) []byte {
	return append(data, v.Type, v.Len)
}

func (p *HOP6) PullFrom(data []byte) []byte {
	if len(data) < SizeofHOP6 {
		return data
	}
	data = binint.Bite(data, &p.Type)
	data = binint.Bite(data, &p.Len)
	return data
}
