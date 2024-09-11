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

type IP []byte

func (pdu IP) Header() (h netph.IP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu IP) Data() (d []byte) {
	if len(pdu) >= 1 {
		n := int(pdu[0]&0xf) * 4
		if len(pdu) > n {
			d = []byte(pdu)[n:]
		}
	}
	return
}

func (pdu IP) Options() (d []byte) {
	if len(pdu) >= 1 {
		n := int(pdu[0]&0xf) * 4
		if len(pdu) > 20 {
			d = []byte(pdu)[20:n]
		}
	}
	return
}

func (pdu IP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ip ")
	h, err := pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	d := pdu.Data()
	fmt.Fprint(w, net.IP(h.DA[:]), " <- ", net.IP(h.SA[:]))
	switch h.Protocol {
	case netph.IPPROTO_ICMP:
		fmt.Fprint(w, Mark, ChecksummingICMP{pdu, d})
	case netph.IPPROTO_TCP:
		fmt.Fprint(w, Mark, ChecksummingTCP{pdu, d})
	case netph.IPPROTO_UDP:
		fmt.Fprint(w, Mark, ChecksummingUDP{pdu, d})
	default:
		fmt.Fprintf(w, ", protocol[%#x]", h.Protocol)
	}
}

// https://www.ietf.org/rfc/rfc768.txt
// https://www.ietf.org/rfc/rfc793.txt
func (pdu IP) Checksum(prot uint8, n uint, data []byte) uint16 {
	sum := Checksum(pdu[netph.IPAddrsIndex:netph.IPSize])
	sum += Checksum([]byte{
		0,
		prot,
		byte(n >> 8),
		byte(n),
	})
	sum += Checksum(data)
	return CarryOver(sum)
}
