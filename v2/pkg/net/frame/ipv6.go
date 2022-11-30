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

// https://en.wikipedia.org/wiki/IPv6
type IPv6 []byte

func (ipv6 IPv6) VCF() IPv6VCF          { return IPv6VCF(ipv6[:4]) }
func (ipv6 IPv6) LEN() big.Uint16       { return big.Uint16(ipv6[4 : 4+2]) }
func (ipv6 IPv6) NextHeader() big.Uint8 { return big.Uint8(ipv6[6 : 6+1]) }
func (ipv6 IPv6) HopLimit() big.Uint8   { return big.Uint8(ipv6[7 : 7+1]) }
func (ipv6 IPv6) SA() net.IP            { return net.IP(ipv6[8 : 8+16]) }
func (ipv6 IPv6) DA() net.IP            { return net.IP(ipv6[24 : 24+16]) }
func (ipv6 IPv6) Data() []byte          { return net.IP(ipv6[40:]) }

func (ipv6 IPv6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv6: ", ipv6.SA(), " -> ", ipv6.DA())
	switch nh := ipv6.NextHeader(); nh.Value() {
	case 0:
		fmt.Fprint(w, ProtoMark, HopByHop(ipv6.Data()))
	case syscall.IPPROTO_ICMPV6:
		fmt.Fprint(w, ProtoMark, ICMP6(ipv6.Data()))
	case syscall.IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, TCP(ipv6.Data()))
	case syscall.IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, UDP(ipv6.Data()))
	default:
		fmt.Fprintf(w, ", next[%#x]", nh)
	}
}

const IPv6FlowMask = ((1 << 20) - 1)

type IPv6VCF []byte

func (vcf IPv6VCF) Uint() uint32   { return big.Uint32(vcf).Value() >> 28 }
func (vcf IPv6VCF) Version() uint8 { return uint8(vcf.Uint() >> 28) }
func (vcf IPv6VCF) Class() uint8   { return uint8(vcf.Uint() >> 20) }
func (vcf IPv6VCF) Flow() uint32   { return vcf.Uint() & IPv6FlowMask }

func (vcf IPv6VCF) Set(ver, class uint8, flow uint32) {
	big.Uint32(vcf).Put((flow & IPv6FlowMask) |
		(uint32(class) << 20) |
		(uint32(ver) << 28))
}

type HopByHop []byte

func (h HopByHop) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "hop-by-hop ...")
}

type ICMP6 []byte

func (icmp6 ICMP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp6 ...")
}
