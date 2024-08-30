// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"bytes"
	"encoding/binary"
	"encoding/pem"
	"fmt"
	"io"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/box"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const (
	VPN_P_HELLO = 0x400 + iota
	VPN_P_PUBLIC_KEY
	VPN_P_WHOIS_ADDRESSED
	VPN_P_WHOIS_IDENTIFIED
	VPN_P_WHOIS_SERVICE
)

var zaddr netip.Addr

var PacketName = map[uint16]string{
	VPN_P_HELLO:            "vpn-hello",
	VPN_P_PUBLIC_KEY:       "vpn-public-key",
	VPN_P_WHOIS_ADDRESSED:  "vpn-whois-addressed",
	VPN_P_WHOIS_IDENTIFIED: "vpn-whois-identified",
	VPN_P_WHOIS_SERVICE:    "vpn-whois-service",
}

func VpnHelloTimeSpan(r io.Reader) time.Duration {
	var sent int64
	binary.Read(r, binary.BigEndian, &sent)
	now := time.Now().UnixMicro()
	return time.Microsecond * time.Duration(now-sent)
}

var VpnPublicKey = pem.Decode

func VpnWhoisAddress(r io.Reader) netip.Addr {
	var a [16]byte
	n, err := r.Read(a[:])
	if err != nil || n != 16 {
		return zaddr
	}
	return netip.AddrFrom16([16]byte(a)).Unmap()
}

func VpnWhoisId(r io.Reader) (id box.Id, ok bool) {
	err := binary.Read(r, binary.BigEndian, &id)
	ok = err == nil
	return
}

func VpnWhoisService(r io.Reader) netip.AddrPort {
	var port uint16
	addr := VpnWhoisAddress(r)
	binary.Read(r, binary.BigEndian, &port)
	return netip.AddrPortFrom(addr, port)
}

type VpnHelloPDU []byte
type VpnPublicKeyPDU []byte
type VpnWhoisAddressedPDU []byte
type VpnWhoisIdentifiedPDU []byte
type VpnWhoisServicePDU []byte

var TunPIprotos = map[uint16]func([]byte) fmt.Formatter{
	VPN_P_HELLO: func(data []byte) fmt.Formatter {
		return VpnHelloPDU(data)
	},
	VPN_P_PUBLIC_KEY: func(data []byte) fmt.Formatter {
		return VpnPublicKeyPDU(data)
	},
	VPN_P_WHOIS_ADDRESSED: func(data []byte) fmt.Formatter {
		return VpnWhoisAddressedPDU(data)
	},
	VPN_P_WHOIS_IDENTIFIED: func(data []byte) fmt.Formatter {
		return VpnWhoisIdentifiedPDU(data)
	},
	VPN_P_WHOIS_SERVICE: func(data []byte) fmt.Formatter {
		return VpnWhoisServicePDU(data)
	},
}

func (pdu VpnHelloPDU) Format(w fmt.State, verb rune) {
	if span := VpnHelloTimeSpan(bytes.NewBuffer(pdu)); span == 0 {
		fmt.Fprint(w, "underrun")
	} else {
		fmt.Fprint(w, span)
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
	fmt.Fprint(w, VpnWhoisAddress(bytes.NewBuffer(pdu)))
}

func (pdu VpnWhoisIdentifiedPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	if lbl, ok := VpnWhoisId(bytes.NewBuffer(pdu)); !ok {
		fmt.Fprint(w, "underrun")
	} else {
		fmt.Fprint(w, lbl)
	}
}

func (pdu VpnWhoisServicePDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ", VpnWhoisService(bytes.NewBuffer(pdu)))
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
