// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/omni"
)

// https://en.wikipedia.org/wiki/Address_Resolution_Protocol
type ARP struct {
	HTYPE big.Uint16
	PTYPE big.Uint16
	HLEN  omni.Uint8
	PLEN  omni.Uint8
	OPER  big.Uint16
	SHA   omni.HardwareAddr
	SPA   omni.IPv4
	THA   omni.HardwareAddr
	TPA   omni.IPv4
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
