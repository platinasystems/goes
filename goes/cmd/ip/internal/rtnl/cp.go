// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "syscall"

func cp(dst, src []byte) (int, error) {
	if len(dst) >= len(src) {
		return copy(dst, []byte(src)), nil
	}
	return 0, syscall.EOVERFLOW
}
