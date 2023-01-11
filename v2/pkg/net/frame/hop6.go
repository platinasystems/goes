// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/universal"
)

type HopByHop struct {
	Type    *universal.Uint8
	Len     *universal.Uint8
	Options []byte
	Data    []byte
}

func NewHopByHop(data []byte) *HopByHop {
	hop := new(HopByHop)
	hop.Write(data)
	return hop
}

func (hop *HopByHop) Format(w fmt.State, verb rune) {
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
			fmt.Fprint(w, ProtoMark, NewHopByHop(hop.Data))
		}
	} else if t == syscall.IPPROTO_ICMPV6 {
		fmt.Fprint(w, NewICMP6(hop.Data))
	} else if t == syscall.IPPROTO_TCP {
		fmt.Fprint(w, NewTCP(hop.Data))
	} else if t == syscall.IPPROTO_UDP {
		fmt.Fprint(w, NewUDP(hop.Data))
	} else {
		fmt.Fprintf(w, "proto %#x", t)
	}
}

func (hop *HopByHop) Write(data []byte) (int, error) {
	hop.Type, hop.Data = universal.NewUint8(data)
	hop.Len, hop.Data = universal.NewUint8(hop.Data)
	n := 6 + (int(hop.Len.Value()) * 8)
	hop.Options = hop.Data[:n]
	hop.Data = hop.Data[n:]
	return len(data) - len(hop.Data), nil
}
