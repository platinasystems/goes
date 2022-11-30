// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import "fmt"

type TCP []byte

func (tcp TCP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "tcp ...")
}
