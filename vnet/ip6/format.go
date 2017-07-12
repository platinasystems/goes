// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip6

import (
	"github.com/platinasystems/go/elib/parse"

	"fmt"
	"net"
)

func (a *Address) String() string {
	return (net.IP)(a[:]).String()
}

func (h *Header) String() (s string) {
	s = fmt.Sprintf("%s: %s -> %s", h.Protocol.String(), h.Src.String(), h.Dst.String())
	return
}

const DefaultTtl = 64

func (h *Header) Parse(in *parse.Input) {
	h.Ttl = DefaultTtl
	if !in.ParseLoose("%v: %v -> %v", &h.Protocol, &h.Src, &h.Dst) {
		panic(parse.ErrInput)
	}
loop:
	for {
		switch {
		case in.Parse("ttl %d", &h.Ttl):
		default:
			break loop
		}
	}
	return
}
