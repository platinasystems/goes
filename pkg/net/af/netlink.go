// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package af

import "golang.org/x/sys/unix"

type Netlink int

var OpenNetlink = Open[Netlink]

func (Netlink) Family() int { return NETLINK }
func (Netlink) Type() int   { return unix.SOCK_RAW | unix.SOCK_CLOEXEC }
func (Netlink) Proto() int  { return NETLINK_ROUTE }

func (nl Netlink) Bind() (unix.Sockaddr, error) {
	sa := &unix.SockaddrNetlink{Family: uint16(nl.Family())}
	err := unix.Bind(int(nl), sa)
	return sa, err
}
