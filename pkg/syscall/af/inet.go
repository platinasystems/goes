// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !plan9

package af

import "syscall"

type Inet int

var OpenInet = Open[Inet]

func (Inet) Family() int { return syscall.AF_INET }
func (Inet) Type() int   { return syscall.SOCK_DGRAM }
func (Inet) Proto() int  { return 0 }
