// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "syscall"

const (
	PAGE  align = 4 << 10
	NLMSG align = syscall.NLMSG_ALIGNTO
	RTA   align = syscall.RTA_ALIGNTO
)

type align int

func (to align) Align(i int) int {
	return (i + to.Size() - 1) & ^(to.Size() - 1)
}

func (to align) Size() int { return int(to) }
