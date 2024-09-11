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

var EthTypes = map[uint16]func([]byte) fmt.Formatter{
	netph.ETH_P_8021Q: func(data []byte) fmt.Formatter {
		return IEEE8021Q(data)
	},
	netph.ETH_P_8021AD: func(data []byte) fmt.Formatter {
		return IEEE8021AD(data)
	},
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

type Eth []byte

func (pdu Eth) Header() (h netph.Eth, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu Eth) Data() (d []byte) {
	if len(pdu) >= netph.EthSize {
		d = []byte(pdu)[netph.EthSize:]
	}
	return
}

func (pdu Eth) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "eth: ")
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, net.HardwareAddr(h.DMAC[:]), " <- ",
			net.HardwareAddr(h.SMAC[:]))
		if f, ok := EthTypes[h.Type]; ok {
			fmt.Fprint(w, Mark, f(pdu.Data()))
		} else {
			fmt.Fprintf(w, ", type[%#x]", h.Type)
		}
	}
}
