// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package af

import "syscall"

const NETLINK = syscall.AF_NETLINK

type Netlink int

var OpenNetlink = Open[Netlink]

func (Netlink) Family() int { return NETLINK }
func (Netlink) Type() int   { return syscall.SOCK_RAW | syscall.SOCK_CLOEXEC }
func (Netlink) Proto() int  { return syscall.NETLINK_ROUTE }

func (nl Netlink) Bind() (syscall.Sockaddr, error) {
	sa := &syscall.SockaddrNetlink{Family: uint16(nl.Family())}
	err := syscall.Bind(int(nl), sa)
	return sa, err
}
