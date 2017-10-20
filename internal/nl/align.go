// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nl

import "syscall"

const (
	PAGE   Align = 4 << 10
	NLMSG  Align = syscall.NLMSG_ALIGNTO
	NLATTR Align = syscall.RTA_ALIGNTO
)

type Align int

func (to Align) Align(i int) int {
	return (i + to.Size() - 1) & ^(to.Size() - 1)
}

func (to Align) Size() int { return int(to) }
