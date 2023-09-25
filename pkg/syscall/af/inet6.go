// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !plan9

package af

import "syscall"

type Inet6 int

var OpenInet6 = Open[Inet6]

func (Inet6) Family() int { return syscall.AF_INET6 }
func (Inet6) Type() int   { return syscall.SOCK_DGRAM }
func (Inet6) Proto() int  { return 0 }
