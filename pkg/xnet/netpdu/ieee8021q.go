// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var IEEE8021Qtypes = map[uint16]func([]byte) fmt.Formatter{
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

type IEEE8021Q []byte

func (pdu IEEE8021Q) Header() (h netph.IEEE8021, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu IEEE8021Q) Data() (d []byte) {
	if len(pdu) >= netph.IEEE8021Size {
		d = []byte(pdu)[netph.IEEE8021Size:]
	}
	return
}

func (pdu IEEE8021Q) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee8021q: ")
	ieee8021qFormat(w, verb, pdu)
}

type IEEE8021AD []byte

func (pdu IEEE8021AD) Header() (h netph.IEEE8021, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu IEEE8021AD) Data() (d []byte) {
	if len(pdu) >= netph.IEEE8021Size {
		d = []byte(pdu)[netph.IEEE8021Size:]
	}
	return
}

func (pdu IEEE8021AD) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee8021ad: ")
	ieee8021qFormat(w, verb, pdu)
}

func ieee8021qFormat(w fmt.State, verb rune, pdu interface {
	Header() (netph.IEEE8021, error)
	Data() []byte
}) {
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "pcp[%#x] dei[%v] vid [%#x]",
			h.PCP(), h.DEI(), h.VID())
		if f, ok := IEEE8021Qtypes[h.Type]; ok {
			fmt.Fprint(w, Mark, f(pdu.Data()))
		} else {
			fmt.Fprintf(w, " type[%#x]", h.Type)
		}
	}
}
