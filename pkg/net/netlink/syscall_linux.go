// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import "syscall"

// syscall aliases

type Errno = syscall.Errno
type IfAddrmsg = syscall.IfAddrmsg
type IfInfomsg = syscall.IfInfomsg
type NlMsghdr = syscall.NlMsghdr
type NlMsgerr = syscall.NlMsgerr
type RtAttr = syscall.RtAttr
type RtGenmsg = syscall.RtGenmsg
type RtMsg = syscall.RtMsg
type RtNexthop = syscall.RtNexthop
type Sockaddr = syscall.Sockaddr
type SockaddrNetlink = syscall.SockaddrNetlink

const (
	EINVAL        = syscall.EINVAL
	EADDRNOTAVAIL = syscall.EADDRNOTAVAIL
	MSG_PEEK      = syscall.MSG_PEEK
	NLMSG_HDRLEN  = syscall.NLMSG_HDRLEN
	NLMSG_DONE    = syscall.NLMSG_DONE
	NLMSG_ERROR   = syscall.NLMSG_ERROR
	SizeofRtAttr  = syscall.SizeofRtAttr
)
