// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type ICMP6 []byte

func (pdu ICMP6) Header() (h netph.ICMP6, err error) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6Size {
		d = []byte(pdu)[netph.ICMP6Size:]
	}
	return
}

func (pdu ICMP6) SetSum(sum uint16) {
	pdu[netph.ICMP6SumIndex] = byte(sum >> 8)
	pdu[netph.ICMP6SumIndex+1] = byte(sum)
}

type ChecksummingICMP6 struct {
	csr Checksummer
	pdu ICMP6
}

func (x ChecksummingICMP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp6 ")
	h, err := x.pdu.Header()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	d := x.pdu.Data()
	fmt.Fprint(w, "sum ")
	sum := x.csr.Checksum(netph.IPPROTO_ICMPV6, x.pdu)
	if h.Sum == 0 {
		fmt.Fprint(w, "zero")
	} else if sum != 0 {
		fmt.Fprint(w, "bad")
	} else {
		fmt.Fprint(w, "ok")
	}
	fmt.Fprint(w, Mark, ICMP6TypeName(h.Type), " ")
	switch h.Type {
	case netph.ICMP6TypeEchoRequest:
		fmt.Fprint(w, ICMP6EchoRequest(d))
	case netph.ICMP6TypeEchoReply:
		fmt.Fprint(w, ICMP6EchoReply(d))
	case netph.ICMP6TypeRouterSolicitation:
		ICMP6RouterSolicitation(d).Format(w, verb)
	case netph.ICMP6TypeRouterAdvertisement:
		ICMP6RouterAdvertisement(d).Format(w, verb)
	case netph.ICMP6TypeNeighborSolicitation:
		ICMP6NeighborSolicitation(d).Format(w, verb)
	case netph.ICMP6TypeNeighborAdvertisement:
		ICMP6NeighborAdvertisement(d).Format(w, verb)
	case netph.ICMP6TypeRedirectMessage:
		ICMP6RedirectMessage(d).Format(w, verb)
	case netph.ICMP6TypeDestinationUnreachable:
		fmt.Fprint(w, ICMP6UnreachableCodeName(h.Code))
	case netph.ICMP6TypeTimeExceeded:
		fmt.Fprint(w, ICMP6TimeExceededCodeName(h.Code))
	case netph.ICMP6TypeInvalidParameter:
		fmt.Fprint(w, ICMP6InvalidParameterCodeName(h.Code))
	case netph.ICMP6TypeRouterRenumbering:
		fmt.Fprint(w, ICMP6RouterRenumberingCodeName(h.Code))
	default:
		fmt.Fprintf(w, "code %#x", h.Code)
	}
}

type ICMP6EchoRequest []byte

func (pdu ICMP6EchoRequest) Header() (
	h netph.ICMP6EchoRequest, err error,
) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6EchoRequest) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6EchoRequestSize {
		d = []byte(pdu)[netph.ICMP6EchoRequestSize:]
	}
	return
}

func (pdu ICMP6EchoRequest) Format(w fmt.State, verb rune) {
	if h, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "id %#04x, seq %d", h.Identifier, h.Sequence)
	}
}

type ICMP6EchoReply = ICMP6EchoRequest

type ICMP6RouterSolicitation []byte

func (pdu ICMP6RouterSolicitation) Header() (
	h netph.ICMP6RouterSolicitation, err error,
) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6RouterSolicitation) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6RouterSolicitationSize {
		d = []byte(pdu)[netph.ICMP6RouterSolicitationSize:]
	}
	return
}

func (pdu ICMP6RouterSolicitation) Format(w fmt.State, verb rune) {
	if _, err := pdu.Header(); err != nil {
		fmt.Fprint(w, err)
		return
	}
	d := pdu.Data()
	if n := len(d); n > 0 {
		fmt.Fprintf(w, "options[%d]", n)
	}
}

type ICMP6RouterAdvertisement []byte

func (pdu ICMP6RouterAdvertisement) Header() (
	h netph.ICMP6RouterSolicitation, err error,
) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6RouterAdvertisement) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6RouterAdvertisementSize {
		d = []byte(pdu)[netph.ICMP6RouterAdvertisementSize:]
	}
	return
}

func (pdu ICMP6RouterAdvertisement) Format(w fmt.State, verb rune) {
	// FIXME
}

type ICMP6NeighborSolicitation []byte

func (pdu ICMP6NeighborSolicitation) Header() (
	h netph.ICMP6NeighborSolicitation, err error,
) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6NeighborSolicitation) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6NeighborSolicitationSize {
		d = []byte(pdu)[netph.ICMP6NeighborSolicitationSize:]
	}
	return
}

func (pdu ICMP6NeighborSolicitation) Format(w fmt.State, verb rune) {
	// FIXME
}

type ICMP6NeighborAdvertisement []byte

func (pdu ICMP6NeighborAdvertisement) Header() (
	h netph.ICMP6NeighborSolicitation, err error,
) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6NeighborAdvertisement) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6NeighborAdvertisementSize {
		d = []byte(pdu)[netph.ICMP6NeighborAdvertisementSize:]
	}
	return
}

func (pdu ICMP6NeighborAdvertisement) Format(w fmt.State, verb rune) {
	// FIXME
}

type ICMP6RedirectMessage []byte

func (pdu ICMP6RedirectMessage) Header() (
	h netph.ICMP6RedirectMessage, err error,
) {
	_, err = xnet.Subtract(pdu, &h)
	return
}

func (pdu ICMP6RedirectMessage) Data() (d []byte) {
	if len(pdu) >= netph.ICMP6RedirectMessageSize {
		d = []byte(pdu)[netph.ICMP6RedirectMessageSize:]
	}
	return
}

func (pdu ICMP6RedirectMessage) Format(w fmt.State, verb rune) {
	// FIXME
}
