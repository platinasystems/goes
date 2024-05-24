// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import "github.com/platinasystems/goes/v2/pkg/encoding/binary/endian"

const (
	TUNPI_P_IP   = 0x800
	TUNPI_P_IPV6 = 0x86dd
)

const (
	TUNPI_P_VPN_HELLO = 0x400
	TUNPI_P_VPN_WHOIS = 0x401
)

const (
	VPN_HELLO_F_UNIX_MICRO = iota
)

const (
	VPN_WHOIS_F_RESPONSE = iota
	VPN_WHOIS_F_ADDRESSED
	VPN_WHOIS_F_LABELLED
	VPN_WHOIS_F_SERVICE
)

type TunPI struct {
	Flags,
	Proto uint16
	Data []byte
}

func NewTunPI(data []byte) *TunPI {
	flags, data := endian.PullBigInteger[uint16](data)
	proto, data := endian.PullBigInteger[uint16](data)
	return &TunPI{flags, proto, data}
}

func (pi TunPI) Append(data []byte) []byte {
	data = endian.AppendBigInteger(data, pi.Flags)
	data = endian.AppendBigInteger(data, pi.Proto)
	if pi.Data != nil && len(pi.Data) > 0 {
		data = append(data, pi.Data...)
	}
	return data
}
