// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

// https://en.wikipedia.org/wiki/Internet_Control_Message_Protocol
type ICMP struct {
	Type *universal.Uint8
	Code *universal.Uint8
	Sum  *big.Uint16
	Data []byte
}

func NewICMP(data []byte) *ICMP {
	icmp := new(ICMP)
	icmp.Write(data)
	return icmp
}

func (icmp *ICMP) Format(w fmt.State, verb rune) {
	t := icmp.Type.Value()
	c := icmp.Code.Value()
	fmt.Fprint(w, "icmp: ")
	switch t {
	case 0:
		fmt.Fprint(w, "echo reply")
	case 3:
		s, ok := map[uint8]string{
			0:  "network unreachable",
			1:  "host unreachable",
			2:  "protocol unreachable",
			3:  "port unreachable",
			4:  "fragmentation required",
			5:  "source route failed",
			6:  "network unknown",
			7:  "host unknown",
			8:  "source host isolated",
			9:  "network administratively prohibited",
			10: "host administratively prohibited",
			11: "tps Network unreachable",
			12: "tps Host unreachable",
			13: "communication administratively prohibited",
			14: "host precedence violation",
			15: "precedence cutoff in effect",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "unreachable(%d)", c)
		}
	case 5:
		s, ok := map[uint8]string{
			0: "network redirect",
			1: "host redirect",
			2: "tos network redirect",
			3: "tos host redirect",
		}[c]
		if ok {
			fmt.Fprint(w, s, ", ", net.IP(icmp.Data[:4]))
		} else {
			fmt.Fprintf(w, "redirect(%d) %v", c,
				net.IP(icmp.Data[:4]))
		}
	case 8:
		fmt.Fprint(w, "echo request")
	case 9:
		fmt.Fprint(w, "router advertisement")
	case 10:
		fmt.Fprint(w, "router solicitation")
	case 11:
		s, ok := map[uint8]string{
			0: "TTL expired in transit",
			1: "fragment reassembly time exceeded",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "time exceeded(%d)", c)
		}
	case 12:
		s, ok := map[uint8]string{
			0: "pointer indicates the error",
			1: "missing a required option",
			2: "bad length",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "invalid parameter(%d)", c)
		}
	case 13:
		fmt.Fprint(w, "timestamp")
	case 14:
		fmt.Fprint(w, "reply timestamp")
	case 42:
		fmt.Fprint(w, "extended echo request")
	case 43:
		s, ok := map[uint8]string{
			0: "no error",
			1: "malformed query",
			2: "no such interface",
			3: "no such table entry",
			4: "multiple interfaces satisfy query",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "extended echo reply(%d)", c)
		}
	default:
		fmt.Fprintf(w, "type(%d), code(%d)", t, c)
	}
}

func (icmp *ICMP) Write(data []byte) (int, error) {
	icmp.Type, icmp.Data = universal.NewUint8(data)
	icmp.Code, icmp.Data = universal.NewUint8(icmp.Data)
	icmp.Sum, icmp.Data = big.NewUint16(icmp.Data)
	return len(data) - len(icmp.Data), nil
}
