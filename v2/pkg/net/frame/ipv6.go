// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

// https://en.wikipedia.org/wiki/IPv6
type IPv6 struct {
	VCF        *big.Uint32
	LEN        *big.Uint16
	NextHeader *universal.Uint8
	HopLimit   *universal.Uint8
	SA         *universal.IPv6
	DA         *universal.IPv6
	Data       []byte
}

func NewIPv6(data []byte) *IPv6 {
	ipv6 := new(IPv6)
	ipv6.Write(data)
	return ipv6
}

func (ipv6 *IPv6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv6: ", ipv6.DA.Value(), " <- ", ipv6.SA.Value())
	switch nh := ipv6.NextHeader.Value(); nh {
	case 0:
		fmt.Fprint(w, ProtoMark, NewHopByHop(ipv6.Data))
	case syscall.IPPROTO_ICMPV6:
		fmt.Fprint(w, ProtoMark, NewICMP6(ipv6.Data))
	case syscall.IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, NewTCP(ipv6.Data))
	case syscall.IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, NewUDP(ipv6.Data))
	default:
		fmt.Fprintf(w, ", next[%#x]", nh)
	}
}

const IPv6FlowMask = ((1 << 20) - 1)

type IPv6VCF struct{ *big.Uint32 }

func (ipv6 *IPv6) Version() uint8 { return uint8(ipv6.VCF.Value() >> 28) }
func (ipv6 *IPv6) Class() uint8   { return uint8(ipv6.VCF.Value() >> 20) }
func (ipv6 *IPv6) Flow() uint32   { return ipv6.VCF.Value() & IPv6FlowMask }

func (ipv6 *IPv6) SetVCF(class uint8, flow uint32) {
	ipv6.VCF.Put((uint32(6) << 28) |
		(uint32(class) << 20) |
		(flow & IPv6FlowMask))
}

func (ipv6 *IPv6) Write(data []byte) (int, error) {
	ipv6.VCF, ipv6.Data = big.NewUint32(data)
	ipv6.LEN, ipv6.Data = big.NewUint16(ipv6.Data)
	ipv6.NextHeader, ipv6.Data = universal.NewUint8(ipv6.Data)
	ipv6.HopLimit, ipv6.Data = universal.NewUint8(ipv6.Data)
	ipv6.SA, ipv6.Data = universal.NewIPv6(ipv6.Data)
	ipv6.DA, ipv6.Data = universal.NewIPv6(ipv6.Data)
	return len(data) - len(ipv6.Data), nil
}
