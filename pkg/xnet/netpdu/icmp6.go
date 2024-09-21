// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netpdu

import (
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

type ICMP6 []byte

func (pdu ICMP6) Parse() (netph.ICMP6, []byte, error) {
	return netph.Parse[netph.ICMP6](pdu)
}

func (pdu ICMP6) SetSum(sum uint16) {
	pdu[netph.ICMP6SumIndex] = byte(sum >> 8)
	pdu[netph.ICMP6SumIndex+1] = byte(sum)
}

func (pdu ICMP6) Type() (t uint8) {
	if len(pdu) >= netph.ICMP6Size {
		t = pdu[netph.ICMP6TypeIndex]
	}
	return
}

type ChecksummingICMP6 struct {
	csr Checksummer
	pdu ICMP6
}

func (x ChecksummingICMP6) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "icmp6 ")
	h, _, err := x.pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, "sum ")
	sum := x.csr.Checksum(netph.IPPROTO_ICMPV6, x.pdu)
	if h.Sum == 0 {
		fmt.Fprint(w, "zero")
	} else if sum != 0 {
		fmt.Fprint(w, "bad")
	} else {
		fmt.Fprint(w, "ok")
	}
	switch h.Type {
	case netph.ICMP6TypeEchoRequest:
		fmt.Fprint(w, Mark, ICMP6EchoRequest(x.pdu))
	case netph.ICMP6TypeEchoReply:
		fmt.Fprint(w, Mark, ICMP6EchoReply(x.pdu))
	case netph.ICMP6TypeRouterSolicitation:
		fmt.Fprint(w, Mark, ICMP6RouterSolicitation(x.pdu))
	case netph.ICMP6TypeRouterAdvertisement:
		fmt.Fprint(w, Mark, ICMP6RouterAdvertisement(x.pdu))
	case netph.ICMP6TypeNeighborSolicitation:
		fmt.Fprint(w, Mark, ICMP6NeighborSolicitation(x.pdu))
	case netph.ICMP6TypeNeighborAdvertisement:
		fmt.Fprint(w, Mark, ICMP6NeighborAdvertisement(x.pdu))
	case netph.ICMP6TypeRedirectMessage:
		fmt.Fprint(w, Mark, ICMP6RedirectMessage(x.pdu))
	case netph.ICMP6TypeDestinationUnreachable:
		fmt.Fprint(w, Mark, ICMP6UnreachableCodeName(h.Code))
	case netph.ICMP6TypeTimeExceeded:
		fmt.Fprint(w, Mark, ICMP6TimeExceededCodeName(h.Code))
	case netph.ICMP6TypeInvalidParameter:
		fmt.Fprint(w, Mark, ICMP6InvalidParameterCodeName(h.Code))
	case netph.ICMP6TypeRouterRenumbering:
		fmt.Fprint(w, Mark, ICMP6RouterRenumberingCodeName(h.Code))
	default:
		fmt.Fprint(w, Mark, ICMP6TypeName(h.Type))
		fmt.Fprintf(w, " code %#x", h.Code)
	}
}

type ICMP6Options []byte

func (opt ICMP6Options) Type() (t uint8) {
	if len(opt) >= netph.ICMP6OptionSize {
		t = opt[0]
	}
	return
}

func (opt ICMP6Options) Length() (l int) {
	if len(opt) >= netph.ICMP6OptionSize {
		l = int(opt[1])
	}
	return
}

func (opt ICMP6Options) Next() (next ICMP6Options) {
	if l := opt.Length(); l > 0 {
		next = opt[l:]
	}
	return
}

func (opt ICMP6Options) Format(w fmt.State, verb rune) {
	for len(opt) >= netph.ICMP6OptionSize {
		switch opt.Type() {
		case netph.ICMP6OptionTypeSourceLinkLayerAddress:
			fmt.Fprint(w, Mark, ICMP6SourceLinkLayerAddress(opt))
		case netph.ICMP6OptionTypeTargetLinkLayerAddress:
			fmt.Fprint(w, Mark, ICMP6TargetLinkLayerAddress(opt))
		case netph.ICMP6OptionTypePrefixInformation:
			fmt.Fprint(w, Mark, ICMP6PrefixInformation(opt))
		case netph.ICMP6OptionTypeRedirectedHeader:
			fmt.Fprint(w, Mark, ICMP6RedirectedHeader(opt))
		case netph.ICMP6OptionTypeMTU:
			fmt.Fprint(w, Mark, ICMP6MTU(opt))
		case netph.ICMP6OptionTypeRDNSS:
			fmt.Fprint(w, Mark, ICMP6RDNSS(opt))
		case netph.ICMP6OptionTypeDNSSL:
			fmt.Fprint(w, Mark, ICMP6DNSSL(opt))
		}
		opt = opt.Next()
	}
}

type ICMP6EchoRequest []byte

func (pdu ICMP6EchoRequest) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "echo-request ")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "id %#04x, seq %d, %#x",
			h.Identifier, h.Sequence, d)
	}
}

func (pdu ICMP6EchoRequest) Parse() (netph.ICMP6EchoRequest, []byte, error) {
	return netph.Parse[netph.ICMP6EchoRequest](pdu)
}

type ICMP6EchoReply []byte

func (pdu ICMP6EchoReply) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "echo-reply ")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "id %#04x, seq %d, %#x",
			h.Identifier, h.Sequence, d)
	}
}

func (pdu ICMP6EchoReply) Parse() (netph.ICMP6EchoReply, []byte, error) {
	return netph.Parse[netph.ICMP6EchoReply](pdu)
}

type ICMP6RouterSolicitation []byte

func (pdu ICMP6RouterSolicitation) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "router-solicitation")
	_, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, " ", err)
	} else {
		fmt.Fprint(w, ICMP6Options(d))
	}
}

func (pdu ICMP6RouterSolicitation) Parse() (
	netph.ICMP6RouterSolicitation, []byte, error,
) {
	return netph.Parse[netph.ICMP6RouterSolicitation](pdu)
}

type ICMP6RouterAdvertisement []byte

func (pdu ICMP6RouterAdvertisement) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "router-advertisement")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "hop-limit %d, ", h.CurHopLimit)
	fmt.Fprintf(w, "flags %#x, ", h.Flags)
	fmt.Fprintf(w, "lifetime %v, ",
		time.Second*time.Duration(h.Lifetime))
	fmt.Fprintf(w, "reachable %d, ",
		time.Millisecond*time.Duration(h.ReachableTime))
	fmt.Fprintf(w, "retrans %d",
		time.Millisecond*time.Duration(h.RetransTimer))
	fmt.Fprint(w, ICMP6Options(d))
}

func (pdu ICMP6RouterAdvertisement) Parse() (
	netph.ICMP6RouterAdvertisement, []byte, error,
) {
	return netph.Parse[netph.ICMP6RouterAdvertisement](pdu)
}

type ICMP6NeighborSolicitation []byte

func (pdu ICMP6NeighborSolicitation) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "neighbor-solicitation ")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "target %v", netip.AddrFrom16(h.TargetAddress))
	fmt.Fprint(w, ICMP6Options(d))
}

func (pdu ICMP6NeighborSolicitation) Parse() (
	netph.ICMP6NeighborSolicitation, []byte, error,
) {
	return netph.Parse[netph.ICMP6NeighborSolicitation](pdu)
}

type ICMP6NeighborAdvertisement []byte

func (pdu ICMP6NeighborAdvertisement) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "neighbor-advertisement ")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "flags %#x", h.Flags)
	fmt.Fprintf(w, ", target %v", netip.AddrFrom16(h.TargetAddress))
	fmt.Fprint(w, ICMP6Options(d))
}

func (pdu ICMP6NeighborAdvertisement) Parse() (
	netph.ICMP6NeighborAdvertisement, []byte, error,
) {
	return netph.Parse[netph.ICMP6NeighborAdvertisement](pdu)
}

type ICMP6RedirectMessage []byte

func (pdu ICMP6RedirectMessage) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "redirect-message ")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprintf(w, "target %v", netip.AddrFrom16(h.TargetAddress))
	fmt.Fprintf(w, ", destrination %v",
		netip.AddrFrom16(h.DestinationAddress))
	fmt.Fprint(w, ICMP6Options(d))
}

func (pdu ICMP6RedirectMessage) Parse() (
	netph.ICMP6RedirectMessage, []byte, error,
) {
	return netph.Parse[netph.ICMP6RedirectMessage](pdu)
}

type ICMP6SourceLinkLayerAddress []byte
type ICMP6TargetLinkLayerAddress []byte
type ICMP6PrefixInformation []byte
type ICMP6RedirectedHeader []byte
type ICMP6MTU []byte
type ICMP6RDNSS []byte
type ICMP6DNSSL []byte

func (pdu ICMP6SourceLinkLayerAddress) Parse() (
	netph.ICMP6SourceLinkLayerAddress, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6SourceLinkLayerAddress](pdu)
}

func (pdu ICMP6TargetLinkLayerAddress) Parse() (
	netph.ICMP6TargetLinkLayerAddress, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6TargetLinkLayerAddress](pdu)
}

func (pdu ICMP6PrefixInformation) Parse() (
	netph.ICMP6PrefixInformation, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6PrefixInformation](pdu)
}

func (pdu ICMP6RedirectedHeader) Parse() (
	netph.ICMP6RedirectedHeader, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6RedirectedHeader](pdu)
}

func (pdu ICMP6MTU) Parse() (
	netph.ICMP6MTU, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6MTU](pdu)
}

func (pdu ICMP6RDNSS) Parse() (
	netph.ICMP6RDNSS, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6RDNSS](pdu)
}

func (pdu ICMP6DNSSL) Parse() (
	netph.ICMP6DNSSL, []byte, []byte, error,
) {
	return netph.ParseICMP6Option[netph.ICMP6DNSSL](pdu)
}

func (pdu ICMP6SourceLinkLayerAddress) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "source ")
	_, d, _, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, net.HardwareAddr(d))
	}
}

func (pdu ICMP6TargetLinkLayerAddress) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "target ")
	_, d, _, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, net.HardwareAddr(d))
	}
}

func (pdu ICMP6PrefixInformation) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "prefix-information ")
	h, _, _, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
		return
	}
	fmt.Fprint(w, "length %d, ", h.PrefixLength)
	fmt.Fprintf(w, "flags %#x, ", h.Flags)
	fmt.Fprintf(w, "valid-lifetime %v, ",
		time.Second*time.Duration(h.ValidLifetime))
	fmt.Fprintf(w, "preferred-lifetime %v, ",
		time.Second*time.Duration(h.PreferredLifetime))
	fmt.Fprintf(w, "prefix %v", netip.AddrFrom16(h.Prefix))
}

func (pdu ICMP6RedirectedHeader) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "redirected-header")
}

func (pdu ICMP6MTU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "mtu ")
	h, _, _, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprintf(w, "%d", h.MTU)
	}
}

func (pdu ICMP6RDNSS) Format(w fmt.State, verb rune) {
	// FIXME
}

func (pdu ICMP6DNSSL) Format(w fmt.State, verb rune) {
	// FIXME
}
