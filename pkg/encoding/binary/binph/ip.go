// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binph

import (
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

// https://en.wikipedia.org/wiki/IPv4
type IP struct {
	VIHL     uint8
	TOS      uint8
	TL       uint16
	ID       uint16
	FFO      uint16
	TTL      uint8
	Protocol uint8
	Checksum uint16
	SA       [4]byte
	DA       [4]byte
}

const SizeofIP = int(unsafe.Sizeof(IP{}))

const (
	IP4MFbit    = 13
	IP4DFbit    = 14
	IP4FlagMask = ((1 << IP4MFbit) - 1)
)

func (v IP) AppendTo(data []byte) []byte {
	if v.Version() != 4 || v.IHL() < 5 {
		return data
	}
	data = append(data, v.VIHL)
	data = append(data, v.TOS)
	data = binint.AppendBig(data, v.TL)
	data = binint.AppendBig(data, v.ID)
	data = binint.AppendBig(data, v.FFO)
	data = append(data, v.TTL)
	data = append(data, v.Protocol)
	data = binint.AppendBig(data, v.Checksum)
	data = append(data, v.SA[:]...)
	data = append(data, v.DA[:]...)
	return data
}

func (p *IP) PullFrom(data []byte) []byte {
	if len(data) < SizeofIP {
		return data
	}
	if data[0] == 0 {
		data[0] = 0x45
	}
	if (data[0] >> 4) != 4 {
		return data
	}
	if ihl4 := 4 * int(data[0]&0xf); len(data) < ihl4 {
		return data
	}
	data = binint.Bite(data, &p.VIHL)
	data = binint.Bite(data, &p.TOS)
	data = binint.PullBig(data, &p.TL)
	data = binint.PullBig(data, &p.ID)
	data = binint.PullBig(data, &p.FFO)
	data = binint.Bite(data, &p.TTL)
	data = binint.Bite(data, &p.Protocol)
	data = binint.PullBig(data, &p.Checksum)
	data = Pull(data, p.SA[:])
	data = Pull(data, p.DA[:])
	return data
}

func (p *IP) Version() uint8 { return p.VIHL >> 4 }
func (p *IP) IHL() int       { return int(p.VIHL & 0xf) }

func (p *IP) SetIHL(n uint8) {
	p.VIHL = ((4 << 4) | (n & 0xf))
}

func (p *IP) DHCP() uint8 { return p.TOS >> 2 }
func (p *IP) ECN() uint8  { return p.TOS & 0x3 }

func (p *IP) SetTOS(dhcp, ecn uint8) {
	p.TOS = ((dhcp << 2) | (ecn & 0x3))
}

func (p *IP) DF() bool { return (p.FFO>>IP4MFbit)&1 == 1 }
func (p *IP) MF() bool { return (p.FFO>>IP4DFbit)&1 == 1 }

func (p *IP) FragOffset() uint16 { return p.FFO & IP4FlagMask }

func ConstructFFO(df, mf bool, fragoffset uint16) uint16 {
	u := fragoffset & IP4FlagMask
	if mf {
		u |= 1 << 13
	}
	if df {
		u |= 1 << 14
	}
	return u
}
