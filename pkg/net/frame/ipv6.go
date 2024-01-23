// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/omni"
)

// https://en.wikipedia.org/wiki/IPv6
type IPv6 struct {
	VCF        IPv6VCF
	LEN        big.Uint16
	NextHeader omni.Uint8
	HopLimit   omni.Uint8
	SA         omni.IPv6
	DA         omni.IPv6
}

func (ipv6 *IPv6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ipv6: ", ipv6.DA.Value(), " <- ", ipv6.SA.Value())
	switch nh := ipv6.NextHeader.Value(); nh {
	case 0:
		fmt.Fprint(w, ProtoMark, (*Hop6)(Data(ipv6)))
	case IPPROTO_ICMPV6:
		fmt.Fprint(w, ProtoMark, (*ICMP6)(Data(ipv6)))
	case IPPROTO_TCP:
		fmt.Fprint(w, ProtoMark, (*TCP)(Data(ipv6)))
	case IPPROTO_UDP:
		fmt.Fprint(w, ProtoMark, (*UDP)(Data(ipv6)))
	default:
		fmt.Fprintf(w, ", next[%#x]", nh)
	}
}

const IPv6FlowMask = ((1 << 20) - 1)

type IPv6VCF struct{ big.Uint32 }

func (vcf IPv6VCF) Version() uint8 { return uint8(vcf.Value() >> 28) }
func (vcf IPv6VCF) Class() uint8   { return uint8(vcf.Value() >> 20) }
func (vcf IPv6VCF) Flow() uint32   { return vcf.Value() & IPv6FlowMask }

func (vcf *IPv6VCF) Set(class uint8, flow uint32) {
	vcf.Put((uint32(6) << 28) |
		(uint32(class) << 20) |
		(flow & IPv6FlowMask))
}
