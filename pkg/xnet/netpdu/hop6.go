// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type HOP6 []byte

func (pdu HOP6) Parse() (netph.HOP6, []byte, error) {
	return netph.Parse[netph.HOP6](pdu)
}

type ChecksummingHOP6 struct {
	csr Checksummer
	pdu HOP6
}

func (x ChecksummingHOP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "hop6 ")
	h, d, err := x.pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	if s, ok := map[uint8]string{
		0:   "hop-by-hop",
		43:  "routing",
		44:  "fragment",
		50:  "ESP",
		51:  "AH",
		60:  "destination options",
		135: "mobility",
		139: "HIP",
		140: "shim",
	}[h.Type]; ok {
		fmt.Fprintf(w, "%s %d", s, h.Len)
		if h.Type != 59 {
			fmt.Fprint(w, Mark, ChecksummingHOP6{x.csr, d})
		}
		return
	}
	switch h.Type {
	case netph.IPPROTO_ICMPV6:
		fmt.Fprint(w, Mark, ChecksummingICMP6{x.csr, d})
	case netph.IPPROTO_TCP:
		fmt.Fprint(w, Mark, ChecksummingTCP{x.csr, d})
	case netph.IPPROTO_UDP:
		fmt.Fprint(w, Mark, ChecksummingUDP{x.csr, d})
	default:
		fmt.Fprintf(w, "type[%#x]", h.Type)
	}
}
