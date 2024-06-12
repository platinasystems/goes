// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"io"
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

const SizeofTCP = int64(unsafe.Sizeof(TCP{}))

func (p *TCP) ReadFrom(r io.Reader) (int64, error) {
	binint.BigEndianPointer(&p.SP).ReadFrom(r)
	binint.BigEndianPointer(&p.DP).ReadFrom(r)
	binint.BigEndianPointer(&p.Seq).ReadFrom(r)
	binint.BigEndianPointer(&p.Ack).ReadFrom(r)
	binint.BigEndianPointer(&p.Flags).ReadFrom(r)
	binint.BigEndianPointer(&p.Sum).ReadFrom(r)
	_, err := binint.BigEndianPointer(&p.UP).ReadFrom(r)
	return SizeofTCP, err
}

func (v TCP) WriteTo(w io.Writer) (int64, error) {
	binint.BigEndianValue(v.SP).WriteTo(w)
	binint.BigEndianValue(v.DP).WriteTo(w)
	binint.BigEndianValue(v.Seq).WriteTo(w)
	binint.BigEndianValue(v.Ack).WriteTo(w)
	binint.BigEndianValue(v.Flags).WriteTo(w)
	binint.BigEndianValue(v.Sum).WriteTo(w)
	_, err := binint.BigEndianValue(v.UP).WriteTo(w)
	return SizeofTCP, err
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
