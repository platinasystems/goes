// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type UDP []byte

func (pdu UDP) Parse() (netph.UDP, []byte, error) {
	h, d, err := netph.Parse[netph.UDP](pdu)
	if err == nil {
		if n := int(h.Len - netph.UDPSize); n < len(d) {
			d = d[:n]
		}
	}
	return h, d, err
}

func (pdu UDP) SetLen() {
	n := len(pdu)
	pdu[netph.UDPLenIndex] = byte(n >> 8)
	pdu[netph.UDPLenIndex+1] = byte(n)
}

func (pdu UDP) SetSum(sum uint16) {
	pdu[netph.UDPSumIndex] = byte(sum >> 8)
	pdu[netph.UDPSumIndex+1] = byte(sum)
}

type ChecksummingUDP struct {
	csr Checksummer
	pdu UDP
}

func (x ChecksummingUDP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp ")
	h, _, err := x.pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, "sum ")
	sum := x.csr.Checksum(netph.IPPROTO_UDP, x.pdu[:h.Len])
	if h.Sum == 0 {
		fmt.Fprint(w, "zero")
	} else if sum != 0 {
		fmt.Fprint(w, "bad")
	} else {
		fmt.Fprint(w, "ok")
	}
	fmt.Fprint(w, Mark, "port ", h.DP, " <- ", h.SP)
}
