// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "syscall"

const NLMSG_ALIGNTO = syscall.NLMSG_ALIGNTO

func Align(i int) int {
	return (i + NLMSG_ALIGNTO - 1) & ^(NLMSG_ALIGNTO - 1)
}
