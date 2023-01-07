// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
)

type ARPData struct {
	*ARP
	Data []byte
}

func NewARPData(data []byte) ARPData {
	arp := (*ARP)(unsafe.Pointer(&data[0]))
	return ARPData{arp, data[unsafe.Sizeof(arp):]}
}

func (arp ARPData) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "arp: ")
	switch oper := arp.OPER.Value(); oper {
	case 1:
		fmt.Fprint(w, "request")
	case 2:
		fmt.Fprint(w, "reply")
	default:
		fmt.Fprint(w, "op[", oper, "]")
	}
	fmt.Fprint(w, " ", arp.SPA())
	fmt.Fprint(w, ", ", arp.TPA())
}

// https://en.wikipedia.org/wiki/Address_Resolution_Protocol
type ARP struct {
	HTYPE big.Uint16
	PTYPE big.Uint16
	HLEN  host.Uint8
	PLEN  host.Uint8
	OPER  big.Uint16
	sha   [6]byte
	spa   [4]byte
	tha   [6]byte
	tpa   [4]byte
}

func (arp *ARP) SHA() net.HardwareAddr { return net.HardwareAddr(arp.sha[:]) }
func (arp *ARP) SPA() net.IP           { return net.IP(arp.spa[:]) }
func (arp *ARP) THA() net.HardwareAddr { return net.HardwareAddr(arp.tha[:]) }
func (arp *ARP) TPA() net.IP           { return net.IP(arp.tpa[:]) }
