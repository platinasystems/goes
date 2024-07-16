// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && (darwin || netbsd)

package xnet

var SysctlAlign = Align(SysctlAlignTo).Roundup
