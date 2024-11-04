// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnsmessage

import (
	_ "unsafe"
)

//go:linkname runtime_rand runtime.rand
func runtime_rand() uint64

func NewID() uint16 {
	return uint16(runtime_rand())
}
