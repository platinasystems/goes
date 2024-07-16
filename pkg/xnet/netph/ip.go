// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netph

import (
	"io"

	"github.com/platinasystems/goes/v2/pkg/xnet"
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

const (
	IP4MFbit    = 13
	IP4DFbit    = 14
	IP4FlagMask = ((1 << IP4MFbit) - 1)
)

func (p *IP) ReadFrom(r io.Reader) (int64, error) {
	xnet.BytePointer(&p.VIHL).ReadFrom(r)
	xnet.BytePointer(&p.TOS).ReadFrom(r)
	xnet.BigEndianPointer(&p.TL).ReadFrom(r)
	xnet.BigEndianPointer(&p.ID).ReadFrom(r)
	xnet.BigEndianPointer(&p.FFO).ReadFrom(r)
	xnet.BytePointer(&p.TTL).ReadFrom(r)
	xnet.BytePointer(&p.Protocol).ReadFrom(r)
	xnet.BigEndianPointer(&p.Checksum).ReadFrom(r)
	r.Read(p.SA[:])
	_, err := r.Read(p.DA[:])
	return SizeofIP, err
}

func (v IP) WriteTo(w io.Writer) (int64, error) {
	xnet.ByteValue(v.VIHL).WriteTo(w)
	xnet.ByteValue(v.TOS).WriteTo(w)
	xnet.BigEndianValue(v.TL).WriteTo(w)
	xnet.BigEndianValue(v.ID).WriteTo(w)
	xnet.BigEndianValue(v.FFO).WriteTo(w)
	xnet.ByteValue(v.TTL).WriteTo(w)
	xnet.ByteValue(v.Protocol).WriteTo(w)
	xnet.BigEndianValue(v.Checksum).WriteTo(w)
	w.Write(v.SA[:])
	_, err := w.Write(v.DA[:])
	return SizeofIP, err
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
