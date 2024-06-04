// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"encoding/pem"
	"fmt"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box/label"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

const (
	VPN_P_HELLO = 0x400 + iota
	VPN_P_PUBLIC_KEY
	VPN_P_WHOIS_ADDRESSED
	VPN_P_WHOIS_LABELLED
	VPN_P_WHOIS_SERVICE
)

var PacketName = map[uint16]string{
	VPN_P_HELLO:           "vpn-hello",
	VPN_P_PUBLIC_KEY:      "vpn-public-key",
	VPN_P_WHOIS_ADDRESSED: "vpn-whois-addressed",
	VPN_P_WHOIS_LABELLED:  "vpn-whois-labelled",
	VPN_P_WHOIS_SERVICE:   "vpn-whois-service",
}

func AppendVpnHelloUnixMicro(data []byte, v int64) []byte {
	return binint.AppendBig(data, v)
}

func AppendVpnPublicKey(data []byte, p *pem.Block) []byte {
	return append(data, pem.EncodeToMemory(p)...)
}

func AppendVpnWhoisAddress(data []byte, v netip.Addr) []byte {
	a16 := v.As16()
	return append(data, a16[:]...)
}

func AppendVpnWhoisLabel(data []byte, v label.Label) []byte {
	return binint.AppendBig(data, v)
}

func AppendVpnWhoisService(data []byte, v netip.AddrPort) []byte {
	a16 := v.Addr().As16()
	data = append(data, a16[:]...)
	data = binint.AppendBig(data, v.Port())
	return data
}

func VpnHelloTimeSpan(data []byte) time.Duration {
	var sent int64
	payload := binint.PullBig(data, &sent)
	if len(payload) == len(data) {
		return 0
	}
	now := time.Now().UnixMicro()
	return time.Microsecond * time.Duration(now-sent)
}

var VpnPublicKey = pem.Decode

func VpnWhoisAddress(data []byte) netip.Addr {
	if len(data) >= 16 {
		return netip.AddrFrom16([16]byte(data)).Unmap()
	}
	return netip.Addr{}
}

func VpnWhoisLabel(data []byte) (lbl label.Label, ok bool) {
	ok = len(binint.PullBig(data, &lbl)) < len(data)
	return
}

func VpnWhoisService(data []byte) netip.AddrPort {
	var addr netip.Addr
	var port uint16
	if len(data) >= 16+2 {
		addr = netip.AddrFrom16([16]byte(data)).Unmap()
		binint.PullBig(data[16:], &port)
	}
	return netip.AddrPortFrom(addr, port)
}

type VpnHelloPDU []byte
type VpnPublicKeyPDU []byte
type VpnWhoisAddressedPDU []byte
type VpnWhoisLabelledPDU []byte
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
	VPN_P_WHOIS_LABELLED: func(data []byte) fmt.Formatter {
		return VpnWhoisLabelledPDU(data)
	},
	VPN_P_WHOIS_SERVICE: func(data []byte) fmt.Formatter {
		return VpnWhoisServicePDU(data)
	},
}

func (pdu VpnHelloPDU) Format(w fmt.State, verb rune) {
	if span := VpnHelloTimeSpan(pdu); span == 0 {
		fmt.Fprint(w, ErrUnderrun)
	} else {
		fmt.Fprint(w, span)
	}
}

func (pdu VpnPublicKeyPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "pubkey ")
	if blk, _ := VpnPublicKey(pdu); blk == nil {
		fmt.Fprint(w, ErrUnderrun)
	} else {
		fmt.Fprint(w, FIXME)
	}
}

func (pdu VpnWhoisAddressedPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	fmt.Fprint(w, VpnWhoisAddress(pdu))
}

func (pdu VpnWhoisLabelledPDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	if lbl, ok := VpnWhoisLabel(pdu); !ok {
		fmt.Fprint(w, ErrUnderrun)
	} else {
		fmt.Fprint(w, lbl)
	}
}

func (pdu VpnWhoisServicePDU) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "whois ")
	fmt.Fprint(w, VpnWhoisService(pdu))
}
