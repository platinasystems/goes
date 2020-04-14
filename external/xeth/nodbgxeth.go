// Copyright © 2018-2019 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build nodbgxeth

package xeth

import "fmt"

func (bits EthtoolLinkModeBits) Format(f fmt.State, c rune) {
	fmt.Fprint(f, "0b")
	for i := len(bits) - 1; i >= 0; i-- {
		fmt.Fprintf(f, "%0b", bits[i])
	}
}

func (flags RtnhFlags) Format(f fmt.State, c rune) {
	if flags != 0 {
		fmt.Fprintf(f, "0b%b", uint32(flags))
	} else {
		fmt.Fprint(f, "nene")
	}
}
