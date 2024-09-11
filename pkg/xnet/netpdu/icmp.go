// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type ICMP []byte

func (pdu ICMP) Header() (h netph.ICMP, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP) Data() (d []byte) {
	if len(pdu) >= netph.ICMPSize {
		d = []byte(pdu)[netph.ICMPSize:]
	}
	return
}

type ChecksummingICMP struct {
	csr Checksummer
	pdu ICMP
}

func (x ChecksummingICMP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp ")
	h, err := x.pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	sum := x.csr.Checksum(netph.IPPROTO_ICMP, uint(len(x.pdu)), x.pdu)
	if sum != 0 && sum != 0xffff {
		fmt.Fprintf(w, "sum %04x, ", sum)
	}
	typename := ICMPTypeName(h.Type)
	if len(typename) == 0 {
		typename = "unknown"
	}
	fmt.Fprint(w, typename, " ")
	if f, ok := ICMPTypeCodes[h.Type]; ok {
		fmt.Fprint(w, f(h.Code))
	} else {
		fmt.Fprintf(w, "code %#x", h.Code)
	}
}

var ICMPTypeCodes = map[uint8]func(uint8) string{
	netph.ICMPTypeDestinationUnreachable: ICMPUnreachableCodeName[uint8],

	netph.ICMPTypeRedirectMessage:   ICMPRedirectCodeName[uint8],
	netph.ICMPTypeTimeExceeded:      ICMPTimeExceededCodeName[uint8],
	netph.ICMPTypeInvalidParamter:   ICMPInvalidParameterCodeName[uint8],
	netph.ICMPTypeExtendedEchoReply: ICMPExtendedEchoReplyCodeName[uint8],
}
