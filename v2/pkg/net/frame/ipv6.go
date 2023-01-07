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

type IPv6Data struct {
	*IPv6
	Data []byte
}

func NewIPv6Data(data []byte) IPv6Data {
	ipv6 := (*IPv6)(unsafe.Pointer(&data[0]))
	return IPv6Data{ipv6, data[unsafe.Sizeof(ipv6):]}
}

func (ipv6 IPv6Data) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv6: ", ipv6.SA(), " -> ", ipv6.DA())
	switch nh := ipv6.NextHeader.Value(); nh {
	case 0:
		fmt.Fprint(w, ProtoMark, NewHopByHopData(ipv6.Data))
	case syscall.IPPROTO_ICMPV6:
		fmt.Fprint(w, ProtoMark, NewICMP6Data(ipv6.Data))
	case syscall.IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, NewTCPData(ipv6.Data))
	case syscall.IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, NewUDPData(ipv6.Data))
	default:
		fmt.Fprintf(w, ", next[%#x]", nh)
	}
}

// https://en.wikipedia.org/wiki/IPv6
type IPv6 struct {
	vcf        big.Uint32
	LEN        big.Uint16
	NextHeader host.Uint8
	HopLimit   host.Uint8
	sa         [16]byte
	da         [16]byte
}

func (ipv6 *IPv6) VCF() IPv6VCF { return IPv6VCF{&ipv6.vcf} }
func (ipv6 *IPv6) SA() net.IP   { return net.IP(ipv6.sa[:]) }
func (ipv6 *IPv6) DA() net.IP   { return net.IP(ipv6.da[:]) }

const IPv6FlowMask = ((1 << 20) - 1)

type IPv6VCF struct{ *big.Uint32 }

func (vcf IPv6VCF) Version() uint8 { return uint8(vcf.Value() >> 28) }
func (vcf IPv6VCF) Class() uint8   { return uint8(vcf.Value() >> 20) }
func (vcf IPv6VCF) Flow() uint32   { return vcf.Value() & IPv6FlowMask }

func (vcf IPv6VCF) Set(ver, class uint8, flow uint32) {
	vcf.Put((flow & IPv6FlowMask) |
		(uint32(class) << 20) |
		(uint32(ver) << 28))
}
