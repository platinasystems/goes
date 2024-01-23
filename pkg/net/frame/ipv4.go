// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/omni"
)

const (
	IPv4MFbit    = 13
	IPv4DFbit    = 14
	IPv4FlagMask = ((1 << IPv4MFbit) - 1)
)

// https://en.wikipedia.org/wiki/IPv4
type IPv4 struct {
	VIHL     omni.Uint8
	TOS      omni.Uint8
	TL       big.Uint16
	ID       big.Uint16
	FFO      big.Uint16
	TTL      omni.Uint8
	Protocol omni.Uint8
	Checksum big.Uint16
	SA       omni.IPv4
	DA       omni.IPv4
}

func (ipv4 *IPv4) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv4: ", ipv4.DA.Value(), " <- ", ipv4.SA.Value())
	switch proto := ipv4.Protocol.Value(); proto {
	case IPPROTO_ICMP:
		fmt.Fprint(w, ProtoMark, (*ICMP)(Data(ipv4)))
	case IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, (*TCP)(Data(ipv4)))
	case IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, (*UDP)(Data(ipv4)))
	default:
		fmt.Fprintf(w, ", proto[%#x]", proto)
	}
}

func (ipv4 *IPv4) Version() uint8 { return ipv4.VIHL.Value() >> 4 }
func (ipv4 *IPv4) IHL() uint8     { return ipv4.VIHL.Value() & 0xf }

// Set VIHL field per length of options that must be a multiple of 4.
func (ipv4 *IPv4) SetVIHL(n uint8) {
	if (n & 3) != 0 {
		panic("len(Options) must be multiple of 4")
	}
	n /= 4
	ipv4.VIHL.Put((4 << 4) | (n & 0xf))
}

func (ipv4 *IPv4) DHCP() uint8 { return ipv4.TOS.Value() >> 2 }
func (ipv4 *IPv4) ECN() uint8  { return ipv4.TOS.Value() & 0x3 }

func (ipv4 *IPv4) SetTOS(dhcp, ecn uint8) {
	ipv4.TOS.Put((dhcp << 2) | (ecn & 0x3))
}

func (ipv4 *IPv4) DF() bool { return (ipv4.FFO.Value()>>IPv4MFbit)&1 == 1 }
func (ipv4 *IPv4) MF() bool { return (ipv4.FFO.Value()>>IPv4DFbit)&1 == 1 }

func (ipv4 *IPv4) FragOffset() uint16 { return ipv4.FFO.Value() & IPv4FlagMask }

func (ipv4 *IPv4) SetFFO(df, mf bool, fragoffset uint16) {
	u := fragoffset & IPv4FlagMask
	if mf {
		u |= 1 << 13
	}
	if df {
		u |= 1 << 14
	}
	ipv4.FFO.Put(u)
}
