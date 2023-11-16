// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/iflink"
)

func Down(ifname string) error {
	return netlink.Admin(ifname, 0, iflink.IFF_UP)
}

func Up(ifname string) error {
	return netlink.Admin(ifname, iflink.IFF_UP, 0)
}
