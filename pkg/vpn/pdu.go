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
	"github.com/platinasystems/goes/v2/pkg/xerrors"
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

var VpnPublicKey = pem.Decode

func VpnWhoisAddress(d []byte) (addr netip.Addr) {
	if len(d) >= netph.IPv6len {
		addr = netip.AddrFrom16([netph.IPv6len]byte(d)).Unmap()
	}
	return
}

func VpnWhoisId(d []byte) (id box.Id, err error) {
	_, err = xnet.Subtract(d, &id)
	return
}

func VpnWhoisService(d []byte) netip.AddrPort {
	var port uint16
	addr := VpnWhoisAddress(d)
	if len(d) >= 16+2 {
		port = binary.BigEndian.Uint16(d[16:])
	}
	return netip.AddrPortFrom(addr, port)
}

type VpnPDU []byte
type VpnHelloPDU []byte
type VpnPublicKeyPDU []byte
type VpnWhoisAddressedPDU []byte
type VpnWhoisIdentifiedPDU []byte
type VpnWhoisServicePDU []byte

func (pdu VpnPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "vpn")
	h, d, err := pdu.Parse()
	if err != nil {
		fmt.Fprint(w, " ", err)
		return
	}
	switch h.Proto {
	case VPN_P_HELLO:
		fmt.Fprint(w, netpdu.Mark, VpnHelloPDU(d))
	case VPN_P_PUBLIC_KEY:
		fmt.Fprint(w, netpdu.Mark, VpnPublicKeyPDU(d))
	case VPN_P_WHOIS_ADDRESSED:
		fmt.Fprint(w, netpdu.Mark, VpnWhoisAddressedPDU(d))
	case VPN_P_WHOIS_IDENTIFIED:
		fmt.Fprint(w, netpdu.Mark, VpnWhoisIdentifiedPDU(d))
	case VPN_P_WHOIS_SERVICE:
		fmt.Fprint(w, netpdu.Mark, VpnWhoisServicePDU(d))
	case VPN_P_IP:
		fmt.Fprint(w, netpdu.Mark, netpdu.IP(d))
	case VPN_P_IP6:
		fmt.Fprint(w, netpdu.Mark, netpdu.IP6(d))
	default:
		fmt.Fprintf(w, " proto %#x", h.Proto)
	}
}

func (pdu VpnPDU) Parse() (netph.TunPI, []byte, error) {
	return netph.Parse[netph.TunPI](pdu)
}

func (pdu VpnHelloPDU) Format(w fmt.State, verb rune) {
	var then int64
	if _, err := xnet.Subtract(pdu, &then); err != nil {
		fmt.Fprint(w, err)
	} else {
		now := time.Now().UnixMicro()
		dur := time.Microsecond * time.Duration(now-then)
		fmt.Fprint(w, dur)
	}
}

func (pdu VpnPublicKeyPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "pubkey ")
	if blk, _ := VpnPublicKey(pdu); blk == nil {
		fmt.Fprint(w, "underrun")
	} else {
		fmt.Fprint(w, xerrors.FIXME("print key type"))
	}
}

func (pdu VpnWhoisAddressedPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	fmt.Fprint(w, VpnWhoisAddress(pdu))
}

func (pdu VpnWhoisIdentifiedPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	if lbl, err := VpnWhoisId(pdu); err != nil {
		fmt.Fprint(w, err)
	} else {
		fmt.Fprint(w, lbl)
	}
}

func (pdu VpnWhoisServicePDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ", VpnWhoisService(pdu))
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
