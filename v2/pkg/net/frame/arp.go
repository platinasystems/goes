// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
)

// https://en.wikipedia.org/wiki/Address_Resolution_Protocol
type ARP []byte

func (arp ARP) HTYPE() big.Uint16 { return big.Uint16(arp[0:2]) }
func (arp ARP) PTYPE() big.Uint16 { return big.Uint16(arp[2:4]) }
func (arp ARP) HLEN() big.Uint8   { return big.Uint8(arp[4:5]) }
func (arp ARP) PLEN() big.Uint8   { return big.Uint8(arp[5:6]) }
func (arp ARP) OPER() big.Uint16  { return big.Uint16(arp[6 : 6+2]) }

func (arp ARP) SHA() net.HardwareAddr { return net.HardwareAddr(arp[8 : 8+6]) }
func (arp ARP) SPA() net.IP           { return net.IP(arp[14 : 14+4]) }
func (arp ARP) THA() net.HardwareAddr { return net.HardwareAddr(arp[18 : 18+6]) }
func (arp ARP) TPA() net.IP           { return net.IP(arp[24 : 24+2]) }

func (arp ARP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "arp: ")
	switch arp.OPER().Value() {
	case 1:
		fmt.Fprint(w, "request")
	case 2:
		fmt.Fprint(w, "reply")
	default:
		fmt.Fprint(w, "op[", arp.OPER(), "]")
	}
	fmt.Fprint(w, " ", arp.SPA())
	fmt.Fprint(w, ", ", arp.TPA())
}
