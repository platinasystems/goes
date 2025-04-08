// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"github.com/platinasystems/goes/v2/pkg/xnet/netpdu"
	"github.com/platinasystems/goes/v2/pkg/xnet/netph"
)

const (
	VPN_P_HELLO = 0x400 + iota
	VPN_P_PUBLIC_KEY
	VPN_P_WHOIS_ADDRESSED
	VPN_P_WHOIS_IDENTIFIED
	VPN_P_WHOIS_SERVICE
)

const (
	VPN_P_IP  = netph.ETH_P_IP
	VPN_P_IP6 = netph.ETH_P_IPV6
)

var zaddr netip.Addr

var PacketName = map[uint16]string{
	VPN_P_HELLO:            "vpn-hello",
	VPN_P_PUBLIC_KEY:       "vpn-public-key",
	VPN_P_WHOIS_ADDRESSED:  "vpn-whois-addressed",
	VPN_P_WHOIS_IDENTIFIED: "vpn-whois-identified",
	VPN_P_WHOIS_SERVICE:    "vpn-whois-service",
}

func WhoisAddress(d []byte) (addr netip.Addr) {
	if len(d) >= netph.IPv6len {
		addr = netip.AddrFrom16([netph.IPv6len]byte(d)).Unmap()
	}
	return
}

func WhoisId(d []byte) (id box.Id, err error) {
	_, err = xnet.Remove(d, &id)
	return
}

func WhoisService(d []byte) netip.AddrPort {
	var port uint16
	addr := WhoisAddress(d)
	if len(d) >= 16+2 {
		port = binary.BigEndian.Uint16(d[16:])
	}
	return netip.AddrPortFrom(addr, port)
}

type PDU []byte
type HelloPDU []byte
type PublicKeyPDU []byte
type WhoisAddressedPDU []byte
type WhoisIdentifiedPDU []byte
type WhoisServicePDU []byte

type Box struct{ *box.Box }

func (vpn Box) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, vpn.Box, netpdu.Mark, vpn.PDU())
}

func (vpn Box) PDU() PDU {
	return PDU(vpn.Contents)
}

func (pdu PDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "vpn")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, " ", err)
		return
	}
	switch h.Proto {
	case VPN_P_HELLO:
		fmt.Fprint(w, netpdu.Mark, HelloPDU(d))
	case VPN_P_PUBLIC_KEY:
		fmt.Fprint(w, netpdu.Mark, PublicKeyPDU(d))
	case VPN_P_WHOIS_ADDRESSED:
		fmt.Fprint(w, netpdu.Mark, WhoisAddressedPDU(d))
	case VPN_P_WHOIS_IDENTIFIED:
		fmt.Fprint(w, netpdu.Mark, WhoisIdentifiedPDU(d))
	case VPN_P_WHOIS_SERVICE:
		fmt.Fprint(w, netpdu.Mark, WhoisServicePDU(d))
	case VPN_P_IP:
		fmt.Fprint(w, netpdu.Mark, netpdu.IP(d))
	case VPN_P_IP6:
		fmt.Fprint(w, netpdu.Mark, netpdu.IP6(d))
	default:
		fmt.Fprintf(w, " proto %#x", h.Proto)
	}
}

func (pdu PDU) Parse() (netph.TunPI, []byte, error) {
	return netph.Parse[netph.TunPI](pdu)
}

func (pdu HelloPDU) Format(w fmt.State, verb rune) {
	var t int64
	fmt.Fprint(w, "hello ")
	if _, err := xnet.Remove(pdu, &t); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, time.UnixMicro(t))
	}
}

func (pdu PublicKeyPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "pubkey ")
	if blk, _ := pem.Decode(pdu); blk == nil {
		fmt.Fprint(w, "underrun")
	} else if pub, err := ecdhPublicKey(blk); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, pub.Curve())
	}
}

func (pdu WhoisAddressedPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ", WhoisAddress(pdu))
}

func (pdu WhoisIdentifiedPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	if lbl, err := WhoisId(pdu); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, lbl)
	}
}

func (pdu WhoisServicePDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ", WhoisService(pdu))
}

func WriteAddrTo(w io.Writer, addr netip.Addr) (int64, error) {
	a16 := addr.As16()
	n, err := w.Write(a16[:])
	return int64(n), err
}

func WriteAddrPortTo(w io.Writer, ap netip.AddrPort) (int64, error) {
	n, err := WriteAddrTo(w, ap.Addr())
	if err == nil {
		err = binary.Write(w, binary.BigEndian, ap.Port())
		if err == nil {
			n += 2
		}
	}
	return n, err
}
