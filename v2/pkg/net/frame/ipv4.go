// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/IPv4
type IPv4 []byte

func (ipv4 IPv4) VIHL() IPv4VIHL       { return IPv4VIHL(ipv4[:1]) }
func (ipv4 IPv4) TOS() IPv4TOS         { return IPv4TOS(ipv4[1:2]) }
func (ipv4 IPv4) TL() big.Uint16       { return big.Uint16(ipv4[2 : 2+2]) }
func (ipv4 IPv4) ID() big.Uint16       { return big.Uint16(ipv4[4 : 4+2]) }
func (ipv4 IPv4) TTL() big.Uint8       { return big.Uint8(ipv4[8 : 8+1]) }
func (ipv4 IPv4) Protocol() big.Uint8  { return big.Uint8(ipv4[9 : 9+1]) }
func (ipv4 IPv4) Checksum() big.Uint16 { return big.Uint16(ipv4[10 : 10+2]) }
func (ipv4 IPv4) SA() net.IP           { return net.IP(ipv4[12 : 12+4]) }
func (ipv4 IPv4) DA() net.IP           { return net.IP(ipv4[16 : 16+4]) }

func (ipv4 IPv4) FFO() IPv4FFO { return IPv4FFO{big.Uint16(ipv4[6 : 6+2])} }

func (ipv4 IPv4) Data() []byte {
	vihl := ipv4.VIHL()
	return ipv4[vihl.IHL()*4:]
}

func (ipv4 IPv4) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv4: ", ipv4.SA(), " -> ", ipv4.DA())
	switch proto := ipv4.Protocol().Value(); proto {
	case syscall.IPPROTO_ICMP:
		fmt.Fprint(w, ProtoMark, ICMP(ipv4.Data()))
	case syscall.IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, TCP(ipv4.Data()))
	case syscall.IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, UDP(ipv4.Data()))
	default:
		fmt.Fprintf(w, ", proto[%#x]", proto)
	}
}

type IPv4VIHL []byte

func (vihl IPv4VIHL) Version() uint8 { return vihl[0] >> 4 }
func (vihl IPv4VIHL) IHL() uint8     { return vihl[0] & 0xf }

func (vihl IPv4VIHL) Set(version, ihl uint8) {
	vihl[0] = (version << 4) | (ihl & 0xf)
}

type IPv4TOS []byte

func (tos IPv4TOS) DHCP() uint8 { return tos[0] >> 2 }
func (tos IPv4TOS) ECN() uint8  { return tos[0] & 0x3 }

func (tos IPv4TOS) Set(dhcp, ecn uint8) {
	tos[0] = (dhcp << 2) | (ecn & 0x3)
}

const (
	IPv4MFbit    = 13
	IPv4DFbit    = 14
	IPv4FlagMask = ((1 << IPv4MFbit) - 1)
)

type IPv4FFO struct{ big.Uint16 } // Flags + Fragment Offset

func (ffo IPv4FFO) DF() bool { return (ffo.Value()>>IPv4MFbit)&1 == 1 }
func (ffo IPv4FFO) MF() bool { return (ffo.Value()>>IPv4DFbit)&1 == 1 }

func (ffo IPv4FFO) FragOffset() uint16 {
	return ffo.Value() & IPv4FlagMask
}

func (ffo IPv4FFO) Set(df, mf bool, fragoffset uint16) {
	u := fragoffset & IPv4FlagMask
	if mf {
		u |= 1 << 13
	}
	if df {
		u |= 1 << 14
	}
	ffo.Put(u)
}

type ICMP []byte

func (icmp ICMP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp ...")
}
