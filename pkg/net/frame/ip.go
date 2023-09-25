// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import (
	"fmt"
	"unsafe"
)

// https://en.wikipedia.org/wiki/IPv4
// https://en.wikipedia.org/wiki/IPv6
type IP []byte

func (ip IP) Version() uint8 { return ip[0] >> 4 }

func (ip *IP) Format(w fmt.State, verb rune) {
	switch ver := ip.Version(); ver {
	case 4:
		fmt.Fprint(w, *(*IPv4)(unsafe.Pointer(ip)))
	case 6:
		fmt.Fprint(w, *(*IPv6)(unsafe.Pointer(ip)))
	default:
		fmt.Fprintf(w, "ipv%d", ver)
	}
}
