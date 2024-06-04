// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

type TunPI struct {
	Flags,
	Proto uint16
}

const SizeofTunPI = int(unsafe.Sizeof(TunPI{}))

func (v TunPI) AppendTo(data []byte) []byte {
	data = binint.AppendBig(data, v.Flags)
	data = binint.AppendBig(data, v.Proto)
	return data
}

func (p *TunPI) PullFrom(data []byte) []byte {
	if len(data) < SizeofTunPI {
		return data
	}
	data = binint.PullBig(data, &p.Flags)
	data = binint.PullBig(data, &p.Proto)
	return data
}
