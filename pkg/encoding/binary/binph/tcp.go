// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/Transmission_Control_Protocol
type TCP struct {
	SP    uint16
	DP    uint16
	Seq   uint32
	Ack   uint32
	Flags uint32
	Sum   uint16
	UP    uint16
}

const SizeofTCP = int(unsafe.Sizeof(TCP{}))

func (v TCP) AppendTo(data []byte) []byte {
	data = binint.AppendBig(data, v.SP)
	data = binint.AppendBig(data, v.DP)
	data = binint.AppendBig(data, v.Seq)
	data = binint.AppendBig(data, v.Ack)
	data = binint.AppendBig(data, v.Flags)
	data = binint.AppendBig(data, v.Sum)
	data = binint.AppendBig(data, v.UP)
	return data
}

func (p *TCP) PullFrom(data []byte) []byte {
	if len(data) < SizeofTCP {
		return data
	}
	data = binint.PullBig(data, &p.SP)
	data = binint.PullBig(data, &p.DP)
	data = binint.PullBig(data, &p.Seq)
	data = binint.PullBig(data, &p.Ack)
	data = binint.PullBig(data, &p.Flags)
	data = binint.PullBig(data, &p.Sum)
	data = binint.PullBig(data, &p.UP)
	return data
}

func (p *TCP) DataOffset() uint8 { return uint8(p.Flags >> 28) }

func (p *TCP) NS() bool  { return (p.Flags & (1 << 24)) != 0 }
func (p *TCP) CWR() bool { return (p.Flags & (1 << 23)) != 0 }
func (p *TCP) ECE() bool { return (p.Flags & (1 << 22)) != 0 }
func (p *TCP) URG() bool { return (p.Flags & (1 << 21)) != 0 }
func (p *TCP) ACK() bool { return (p.Flags & (1 << 20)) != 0 }
func (p *TCP) PSH() bool { return (p.Flags & (1 << 19)) != 0 }
func (p *TCP) RST() bool { return (p.Flags & (1 << 18)) != 0 }
func (p *TCP) SYN() bool { return (p.Flags & (1 << 17)) != 0 }
func (p *TCP) FIN() bool { return (p.Flags & (1 << 16)) != 0 }

func (p *TCP) Window() uint16 { return uint16(p.Flags) }
