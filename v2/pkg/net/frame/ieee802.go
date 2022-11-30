// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import "fmt"

type IEEE802_1Q []byte

func (IEEE802_1Q) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee802_1Q ...")
}

type IEEE802_1AD []byte

func (IEEE802_1AD) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "ieee802_1ad ...")
}
