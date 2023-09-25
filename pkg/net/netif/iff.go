// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"fmt"
	"syscall"
	"unsafe"
)

const (
	IFFsize = int(unsafe.Sizeof(IFF(0)))
	IFFbits = IFFsize * 8
)

const (
	SIOCGIFFLAGS = syscall.SIOCGIFFLAGS
	SIOCSIFFLAGS = syscall.SIOCSIFFLAGS
)

func (iff IFF) Format(w fmt.State, verb rune) {
	if verb == 'x' {
		fmt.Fprintf(w, "%#04x", uint(iff))
		return
	}
	if verb != 's' && verb != 'v' {
		return
	}
	if iff == 0 {
		fmt.Fprint(w, "0")
		return
	}
	var comma string
	for bit := 0; bit < IFFbits; bit++ {
		if flag := IFF(1 << bit); (iff & flag) != 0 {
			fmt.Fprint(w, comma, flag.String())
			comma = ","
		}
	}
}

func (iff IFF) Has(f IFF) bool { return iff&f == f }
