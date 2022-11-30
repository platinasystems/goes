// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package frame

import "fmt"

type UDP []byte

func (udp UDP) Format(w fmt.State, verb rune) {
	fmt.Fprint(w, "udp ...")
}
