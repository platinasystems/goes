// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/native"
	"github.com/platinasystems/goes/v2/pkg/net/af"
)

type TunTapPI struct {
	Flags native.Uint16
	Proto big.Uint16
}

type TapPI struct{ TunTapPI }
type TunPI struct{ TunTapPI }

func (pi *TapPI) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tap")
	if flags := pi.Flags.Value(); flags != 0 {
		fmt.Fprintf(w, ": %#x", flags)
	}
	fmt.Fprint(w, ProtoMark, (*ETH)(Data(pi)))
}

// Returns AF_INET or AF_INET6 if the top 4 bits of the first data byte are 4
// or 6 respectively; otherwise 0.
func (pi *TunPI) AF() uint16 {
	verlen := *(*uint8)(Data(pi))
	switch verlen >> 4 {
	case 4:
		return af.INET
	case 6:
		return af.INET6
	}
	return 0
}

func (pi *TunPI) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tun")
	if flags := pi.Flags.Value(); flags != 0 {
		fmt.Fprintf(w, ": %#x", flags)
	}
	switch proto := pi.Proto.Value(); proto {
	case af.INET:
		fmt.Fprint(w, ProtoMark, (*IPv4)(Data(pi)))
	case af.INET6:
		fmt.Fprint(w, ProtoMark, (*IPv6)(Data(pi)))
	default:
		fmt.Fprintf(w, ", proto[%#x]", proto)
	}
}
