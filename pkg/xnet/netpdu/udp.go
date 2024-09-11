// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type UDP []byte

func (pdu UDP) Header() (h netph.UDP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu UDP) Data() (d []byte) {
	if len(pdu) >= netph.UDPSize {
		d = []byte(pdu)[netph.UDPSize:]
	}
	return
}

func (pdu UDP) SetSum(sum uint16) {
	if len(pdu) >= netph.UDPSize {
		pdu[netph.UDPSumIndex] = byte(sum >> 8)
		pdu[netph.UDPSumIndex+1] = byte(sum)
	}
}

type ChecksummingUDP struct {
	csr Checksummer
	pdu UDP
}

func (x ChecksummingUDP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp ")
	h, err := x.pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	sum := x.csr.Checksum(netph.IPPROTO_UDP, uint(h.Len), x.pdu)
	if sum != 0 && sum != 0xffff {
		fmt.Fprintf(w, "sum %04x, ", sum)
	}
	fmt.Fprint(w, "port ", h.DP, " <- ", h.SP)
}
