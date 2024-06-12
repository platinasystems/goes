// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binpdu

import (
	"bytes"
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binph"
)

var EthTypes = map[uint16]func([]byte) fmt.Formatter{
	ETH_P_8021Q: func(data []byte) fmt.Formatter {
		return IEEE8021Q(data)
	},
	ETH_P_8021AD: func(data []byte) fmt.Formatter {
		return IEEE8021AD(data)
	},
	ETH_P_ARP: func(data []byte) fmt.Formatter {
		return ARP(data)
	},
	ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
	ETH_P_MPLS_UC: func(data []byte) fmt.Formatter {
		return MPLS_UC(data)
	},
	ETH_P_MPLS_MC: func(data []byte) fmt.Formatter {
		return MPLS_MC(data)
	},
}

type Eth []byte

func (pdu Eth) Format(w fmt.State, verb rune) {
	var h binph.Eth
	buf := bytes.NewBuffer(pdu)
	fmt.Fprint(w, "eth: ")
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, net.HardwareAddr(h.DMAC[:]), " <- ",
			net.HardwareAddr(h.SMAC[:]))
		if f, ok := EthTypes[h.Type]; ok {
			fmt.Fprint(w, Mark, f(buf.Bytes()))
		} else {
			fmt.Fprintf(w, ", type[%#x]", h.Type)
		}
	}
}
