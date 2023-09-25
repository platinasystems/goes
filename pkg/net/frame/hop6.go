// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/omni"
)

type Hop6 struct {
	Type omni.Uint8
	Len  omni.Uint8
}

func (hop *Hop6) Format(w fmt.State, verb rune) {
	t := hop.Type.Value()
	s, ok := map[uint8]string{
		0:   "hop-by-hop",
		43:  "routing",
		44:  "fragment",
		50:  "ESP",
		51:  "AH",
		60:  "destination options",
		135: "mobility",
		139: "HIP",
		140: "shim",
	}[t]
	if ok {
		fmt.Fprintf(w, "%s[%d]", s, hop.Len.Value())
		if t != 59 {
			fmt.Fprint(w, ProtoMark, (*Hop6)(Data(hop)))
		}
	} else if t == syscall.IPPROTO_ICMPV6 {
		fmt.Fprint(w, (*ICMP6)(Data(hop)))
	} else if t == syscall.IPPROTO_TCP {
		fmt.Fprint(w, (*TCP)(Data(hop)))
	} else if t == syscall.IPPROTO_UDP {
		fmt.Fprint(w, (*UDP)(Data(hop)))
	} else {
		fmt.Fprintf(w, "proto %#x", t)
	}
}
