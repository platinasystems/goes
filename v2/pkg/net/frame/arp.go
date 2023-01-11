// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

// https://en.wikipedia.org/wiki/Address_Resolution_Protocol
type ARP struct {
	HTYPE *big.Uint16
	PTYPE *big.Uint16
	HLEN  *universal.Uint8
	PLEN  *universal.Uint8
	OPER  *big.Uint16
	SHA   *universal.HardwareAddr
	SPA   *universal.IPv4
	THA   *universal.HardwareAddr
	TPA   *universal.IPv4
	Data  []byte
}

func NewARP(data []byte) *ARP {
	arp := new(ARP)
	arp.Write(data)
	return arp
}

func (arp *ARP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "arp: ")
	switch oper := arp.OPER.Value(); oper {
	case 1:
		fmt.Fprint(w, arp.SPA.Value(), " request ", arp.TPA.Value())
	case 2:
		fmt.Fprint(w, arp.TPA.Value(), " reply ", arp.THA.Value())
	default:
		fmt.Fprint(w, "op[", oper, "]")
	}
}

func (arp *ARP) Write(data []byte) (int, error) {
	arp.HTYPE, arp.Data = big.NewUint16(data)
	arp.PTYPE, arp.Data = big.NewUint16(arp.Data)
	arp.HLEN, arp.Data = universal.NewUint8(arp.Data)
	arp.PLEN, arp.Data = universal.NewUint8(arp.Data)
	arp.OPER, arp.Data = big.NewUint16(arp.Data)
	arp.SHA, arp.Data = universal.NewHardwareAddr(arp.Data)
	arp.SPA, arp.Data = universal.NewIPv4(arp.Data)
	arp.THA, arp.Data = universal.NewHardwareAddr(arp.Data)
	arp.TPA, arp.Data = universal.NewIPv4(arp.Data)
	return len(data) - len(arp.Data), nil
}
