// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"bytes"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var IEEE8021Qtypes = map[uint16]func([]byte) fmt.Formatter{
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

type IEEE8021Q []byte
type IEEE8021AD []byte

func (pdu IEEE8021Q) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee8021q: ")
	ieee8021qFormat(w, verb, pdu)
}

func (pdu IEEE8021AD) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee8021ad: ")
	ieee8021qFormat(w, verb, pdu)
}

func ieee8021qFormat[PDU IEEE8021Q | IEEE8021AD](
	w fmt.State, verb rune, pdu PDU,
) {
	var h netph.IEEE8021Q
	buf := bytes.NewBuffer(pdu)
	if _, err := h.ReadFrom(buf); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "pcp[%#x] dei[%d] vid [%#x]",
			h.PCP(), h.DEI(), h.VID())
		if f, ok := IEEE8021Qtypes[h.Type]; ok {
			fmt.Fprint(w, Mark, f(buf.Bytes()))
		} else {
			fmt.Fprintf(w, " type[%#x]", h.Type)
		}
	}
}
