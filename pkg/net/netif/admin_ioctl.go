// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/net/netioctl"
)

func Down(ctx context.Context, ifname string) error {
	return netioctl.Admin(ifname, 0, netioctl.IFF_UP)
}

func Up(ctx context.Context, ifname string) error {
	const iff = netioctl.IFF_UP | netioctl.IFF_RUNNING
	return netioctl.Admin(ifname, iff, 0)
}
