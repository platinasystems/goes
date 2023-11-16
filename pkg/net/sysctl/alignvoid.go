// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && (dragonfly || freebsd || openbsd)

package sysctl

import "unsafe"

var void uintptr

const AlignTo = unsafe.Sizeof(void)
