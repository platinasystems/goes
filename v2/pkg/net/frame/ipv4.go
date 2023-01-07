// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
)

type IPv4Data struct {
	*IPv4
	Data []byte
}

func NewIPv4Data(data []byte) IPv4Data {
	ipv4 := (*IPv4)(unsafe.Pointer(&data[0]))
	n := uint16(ipv4.VIHL().IHL()) * 4
	if min := uint16(unsafe.Sizeof(ipv4)); n < min {
		n = min
	}
	return IPv4Data{ipv4, data[n:]}
}

func (ipv4 IPv4Data) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv4: ", ipv4.SA(), " -> ", ipv4.DA())
	switch proto := ipv4.Protocol.Value(); proto {
	case syscall.IPPROTO_ICMP:
		fmt.Fprint(w, ProtoMark, NewICMPData(ipv4.Data))
	case syscall.IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, NewTCPData(ipv4.Data))
	case syscall.IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, NewUDPData(ipv4.Data))
	default:
		fmt.Fprintf(w, ", proto[%#x]", proto)
	}
}

// https://en.wikipedia.org/wiki/IPv4
type IPv4 struct {
	vihl     host.Uint8
	tos      host.Uint8
	TL       big.Uint16
	ID       big.Uint16
	ffo      big.Uint16
	TTL      host.Uint8
	Protocol host.Uint8
	Checksum big.Uint16
	sa       [4]byte
	da       [4]byte
	Data     []byte
}

func (ipv4 *IPv4) VIHL() IPv4VIHL { return IPv4VIHL{&ipv4.vihl} }
func (ipv4 *IPv4) TOS() IPv4TOS   { return IPv4TOS{&ipv4.tos} }
func (ipv4 *IPv4) FFO() IPv4FFO   { return IPv4FFO{&ipv4.ffo} }
func (ipv4 *IPv4) SA() net.IP     { return net.IP(ipv4.sa[:]) }
func (ipv4 *IPv4) DA() net.IP     { return net.IP(ipv4.da[:]) }

type IPv4VIHL struct{ *host.Uint8 }

func (vihl IPv4VIHL) Version() uint8 { return vihl.Value() >> 4 }
func (vihl IPv4VIHL) IHL() uint8     { return vihl.Value() & 0xf }

func (vihl IPv4VIHL) Set(version, ihl uint8) {
	vihl.Put((version << 4) | (ihl & 0xf))
}

type IPv4TOS struct{ *host.Uint8 }

func (tos IPv4TOS) DHCP() uint8 { return tos.Value() >> 2 }
func (tos IPv4TOS) ECN() uint8  { return tos.Value() & 0x3 }

func (tos IPv4TOS) Set(dhcp, ecn uint8) { tos.Put((dhcp << 2) | (ecn & 0x3)) }

const (
	IPv4MFbit    = 13
	IPv4DFbit    = 14
	IPv4FlagMask = ((1 << IPv4MFbit) - 1)
)

type IPv4FFO struct{ *big.Uint16 } // Flags + Fragment Offset

func (ffo IPv4FFO) DF() bool { return (ffo.Value()>>IPv4MFbit)&1 == 1 }
func (ffo IPv4FFO) MF() bool { return (ffo.Value()>>IPv4DFbit)&1 == 1 }

func (ffo IPv4FFO) FragOffset() uint16 { return ffo.Value() & IPv4FlagMask }

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
