// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && (darwin || netbsd)

package sysctl

func Align(i int) int {
	return i + (AlignTo-i%AlignTo)%AlignTo
}
