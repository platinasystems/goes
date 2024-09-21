// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type IEEE8021 []byte
type IEEE8021Q []byte
type IEEE8021AD []byte

func (pdu IEEE8021Q) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee8021q ", IEEE8021(pdu))
}

func (pdu IEEE8021AD) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee8021ad ", IEEE8021(pdu))
}

func (pdu IEEE8021) Format(w fmt.State, verb rune) {
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "pcp %#x, dei %v, vid %#x", h.PCP(), h.DEI(), h.VID())
	if f, ok := IEEE8021types[h.Type]; ok {
		fmt.Fprint(w, Mark, f(d))
	} else {
		fmt.Fprintf(w, ", type %#x", h.Type)
	}
}

func (pdu IEEE8021) Parse() (netph.IEEE8021, []byte, error) {
	return netph.Parse[netph.IEEE8021](pdu)
}

var IEEE8021types = map[uint16]func([]byte) fmt.Formatter{
	netph.ETH_P_ARP: func(data []byte) fmt.Formatter {
		return ARP(data)
	},
	netph.ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	netph.ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
	netph.ETH_P_MPLS_UC: func(data []byte) fmt.Formatter {
		return MPLS_UC(data)
	},
	netph.ETH_P_MPLS_MC: func(data []byte) fmt.Formatter {
		return MPLS_MC(data)
	},
}
