// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type IP6 []byte

func (pdu IP6) Header() (h netph.IP6, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu IP6) Data() (d []byte) {
	if len(pdu) >= netph.IP6Size {
		d = []byte(pdu)[netph.IP6Size:]
	}
	return
}

func (pdu IP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ip6 ")
	h, err := pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "%d bytes ", h.LEN)
	d := pdu.Data()
	fmt.Fprint(w, net.IP(h.DA[:]), " <- ", net.IP(h.SA[:]), Mark)
	switch h.NextHeader {
	case netph.IPPROTO_HOPOPTS:
		fmt.Fprint(w, ChecksummingHOP6{pdu, d})
	case netph.IPPROTO_ICMPV6:
		fmt.Fprint(w, ChecksummingICMP6{pdu, d})
	case netph.IPPROTO_TCP:
		fmt.Fprint(w, ChecksummingTCP{pdu, d})
	case netph.IPPROTO_UDP:
		fmt.Fprint(w, ChecksummingUDP{pdu, d})
	default:
		fmt.Fprintf(w, "next[%#x]", h.NextHeader)
	}
}

// https://datatracker.ietf.org/doc/html/rfc2460#section-8
func (pdu IP6) Checksum(prot uint8, data []byte) uint16 {
	n := uint(len(data))
	sum := Checksum(pdu[netph.IP6AddrsIndex:netph.IP6Size])
	sum += Checksum([]byte{
		byte(n >> 24),
		byte(n >> 16),
		byte(n >> 8),
		byte(n),
	})
	sum += Checksum([]byte{0, 0, 0, prot})
	sum += Checksum(data)
	return CarryOver(sum) ^ 0xffff
}

func (pdu IP6) SetLen() {
	n := len(pdu.Data())
	b := pdu[netph.IP6LenIndex:]
	b[0] = byte(n >> 8)
	b[1] = byte(n)
}
