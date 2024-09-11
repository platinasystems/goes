// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

var MPLStypes = map[uint16]func([]byte) fmt.Formatter{
	netph.ETH_P_IP: func(data []byte) fmt.Formatter {
		return IP(data)
	},
	netph.ETH_P_IPV6: func(data []byte) fmt.Formatter {
		return IP6(data)
	},
}

type MPLS_UC []byte

func (pdu MPLS_UC) Header() (h netph.MPLS, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu MPLS_UC) Data() (d []byte) {
	if len(pdu) >= netph.MPLSSize {
		d = []byte(pdu)[netph.MPLSSize:]
	}
	return
}

func (pdu MPLS_UC) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls-uc: ")
	mplsFormat(w, verb, pdu)
}

type MPLS_MC []byte

func (pdu MPLS_MC) Header() (h netph.MPLS, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu MPLS_MC) Data() (d []byte) {
	if len(pdu) >= netph.MPLSSize {
		d = []byte(pdu)[netph.MPLSSize:]
	}
	return
}

func (pdu MPLS_MC) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mpls-mc: ")
	mplsFormat(w, verb, pdu)
}

func mplsFormat(w fmt.State, verb rune, pdu interface {
	Header() (netph.MPLS, error)
	Data() []byte
}) {
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
	} else {
		var t uint16
		fmt.Fprintf(w, "label[%#x] tc[%#x] ttl[%d]",
			h.Label(), h.TC(), h.TTL())
		d := pdu.Data()
		_, err = xnet.Subtract(d, &t)
		if err != nil {
			fmt.Fprint(w, " type ", err)
		} else if !h.IsBOS() {
			fmt.Fprint(w, Mark, d)
		} else if f, ok := MPLStypes[t]; ok {
			fmt.Fprint(w, Mark, f(d))
		} else {
			fmt.Fprintf(w, " type[%#x]", t)
		}
	}
}
