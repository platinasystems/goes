// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type TCP []byte

func (pdu TCP) Parse() (netph.TCP, []byte, error) {
	return netph.Parse[netph.TCP](pdu)
}

func (pdu TCP) SetSum(sum uint16) {
	pdu[netph.TCPSumIndex] = byte(sum >> 8)
	pdu[netph.TCPSumIndex+1] = byte(sum)
}

type ChecksummingTCP struct {
	csr Checksummer
	pdu TCP
}

func (x ChecksummingTCP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tcp ")
	h, _, err := x.pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, "sum ")
	sum := x.csr.Checksum(netph.IPPROTO_TCP, x.pdu)
	if h.Sum == 0 {
		fmt.Fprint(w, "zero")
	} else if sum != 0 {
		fmt.Fprint(w, "bad")
	} else {
		fmt.Fprint(w, "ok")
	}
	fmt.Fprint(w, Mark, "port ", h.DP, " <- ", h.SP)
}
