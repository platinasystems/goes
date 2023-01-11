// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

// https://en.wikipedia.org/wiki/ICMPv6
type ICMP6 struct {
	Type *universal.Uint8
	Code *universal.Uint8
	Sum  *big.Uint16
	Data []byte
}

func NewICMP6(data []byte) *ICMP6 {
	icmp6 := new(ICMP6)
	icmp6.Write(data)
	return icmp6
}

func (icmp6 *ICMP6) Format(w fmt.State, verb rune) {
	t := icmp6.Type.Value()
	c := icmp6.Code.Value()
	fmt.Fprint(w, "icmp6: ")
	switch t {
	case 1:
		s, ok := map[uint8]string{
			0: "no route to destination",
			1: "communication administratively prohibited",
			2: "beyond scope of source address",
			3: "address unreachable",
			4: "port unreachable",
			5: "source address failed ingress/egress policy",
			6: "reject route to destination",
			7: "error in source routing header",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "destination unreachable(%d)", c)
		}
	case 2:
		fmt.Fprint(w, "packet too big")
	case 3:
		s, ok := map[uint8]string{
			0: "hop limit exceeded in transit",
			1: "fragment reassembly time exceeded",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "time exceeded(%d)", c)
		}
	case 4:
		s, ok := map[uint8]string{
			0: "erroneous header field encountered",
			1: "unrecognized Next Header type encountered",
			2: "unrecognized IPv6 option encountered",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "invalid parameter(%d)", c)
		}
	case 128:
		fmt.Fprint(w, "echo request")
	case 129:
		fmt.Fprint(w, "echo reply")
	case 130:
		fmt.Fprint(w, "multicast listener query")
	case 131:
		fmt.Fprint(w, "multicast listener report")
	case 132:
		fmt.Fprint(w, "multicast listener done")
	case 133:
		fmt.Fprint(w, "router solicitation")
	case 134:
		fmt.Fprint(w, "router advertisement")
	case 135:
		fmt.Fprint(w, "neighbor solicitation")
	case 136:
		fmt.Fprint(w, "neighbor advertisement")
	case 137:
		fmt.Fprint(w, "redirect message")
	case 138:
		fmt.Fprint(w, "router renumbering", c)
		s, ok := map[uint8]string{
			0: "command",
			1: "result",
		}[c]
		if ok {
			fmt.Fprint(w, s)
		} else {
			fmt.Fprintf(w, "(%d)", c)
		}
	case 139:
		fmt.Fprintf(w, "node information query(%d)", c)
	case 140:
		fmt.Fprintf(w, "node information response(%d)", c)
	case 141:
		fmt.Fprint(w, "inverse neighbor discovery solicitation")
	case 142:
		fmt.Fprint(w, "inverse neighbor discovery advertisement")
	case 143:
		fmt.Fprint(w, "multicast listener discovery")
	case 144:
		fmt.Fprint(w, "home agent address discovery request")
	case 145:
		fmt.Fprint(w, "home agent address discovery reply")
	case 146:
		fmt.Fprint(w, "mobile prefix solicitation")
	case 147:
		fmt.Fprint(w, "mobile prefix advertisement")
	case 148:
		fmt.Fprint(w, "certification path solicitation")
	case 149:
		fmt.Fprint(w, "certification path advertisement")
	case 151:
		fmt.Fprint(w, "multicast router advertisement")
	case 152:
		fmt.Fprint(w, "multicast router solicitation")
	case 153:
		fmt.Fprint(w, "multicast router termination")
	case 155:
		fmt.Fprint(w, "RPL control message")
	default:
		fmt.Fprintf(w, "type(%d), code(%d)", t, c)
	}
}

func (icmp6 *ICMP6) Write(data []byte) (int, error) {
	icmp6.Type, icmp6.Data = universal.NewUint8(data)
	icmp6.Code, icmp6.Data = universal.NewUint8(icmp6.Data)
	icmp6.Sum, icmp6.Data = big.NewUint16(icmp6.Data)
	return len(data) - len(icmp6.Data), nil
}
