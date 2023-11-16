// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import "github.com/platinasystems/goes/v2/pkg/net/netioctl"

func Down(ifname string) error {
	return netioctl.Admin(ifname, 0, netioctl.IFF_UP)
}

func Up(ifname string) error {
	return netioctl.Admin(ifname, netioctl.IFF_UP|netioctl.IFF_RUNNING, 0)
}
