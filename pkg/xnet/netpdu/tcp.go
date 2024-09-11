// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type TCP []byte

func (pdu TCP) Header() (h netph.TCP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu TCP) Data() (d []byte) {
	if len(pdu) >= netph.TCPSize {
		d = []byte(pdu)[netph.TCPSize:]
	}
	return
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
	h, err := x.pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	sum := x.csr.Checksum(netph.IPPROTO_TCP, uint(len(x.pdu)), x.pdu)
	if sum != 0 && sum != 0xffff {
		fmt.Fprintf(w, "sum %04x, ", sum)
	}
	fmt.Fprint(w, "port ", h.DP, " <- ", h.SP)
}
