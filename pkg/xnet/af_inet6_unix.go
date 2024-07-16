// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "golang.org/x/sys/unix"

const AF_INET6 = unix.AF_INET6

type Inet6 int

var OpenInet6 = Open[Inet6]

func (Inet6) Family() int { return AF_INET6 }
func (Inet6) Type() int   { return unix.SOCK_DGRAM }
func (Inet6) Proto() int  { return 0 }
