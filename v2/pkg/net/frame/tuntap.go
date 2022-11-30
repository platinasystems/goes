// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
)

type TunPI []byte

func (pi TunPI) Flags() host.Uint16 { return host.Uint16(pi[:2]) }
func (pi TunPI) Proto() big.Uint16  { return big.Uint16(pi[2 : 2+2]) }
func (pi TunPI) Data() []byte       { return []byte(pi[4:]) }

// Returns AF_INET or AF_INET6 if the top 4 bits of the first data byte are 4
// or 6 respectively; otherwise 0.
func (pi TunPI) AF() (af uint16) {
	switch pi.Data()[0] >> 4 {
	case 4:
		af = syscall.AF_INET
	case 6:
		af = syscall.AF_INET6
	}
	return
}

func (pi TunPI) Format(w fmt.State, verb rune) {
	fmt.Fprintf(w, "tun: %#x", pi.Flags().Value())
	switch proto := pi.Proto().Value(); proto {
	case syscall.AF_INET:
		fmt.Fprint(w, ProtoMark, IPv4(pi.Data()))
	case syscall.AF_INET6:
		fmt.Fprint(w, ProtoMark, IPv6(pi.Data()))
	default:
		fmt.Fprintf(w, ", proto[%#x]", proto)
	}
}
