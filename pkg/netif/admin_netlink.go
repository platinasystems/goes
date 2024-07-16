// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/iflink"
)

func Down(ctx context.Context, ifname string) error {
	return netlink.Admin(ctx, ifname, 0, iflink.IFF_UP)
}

func Up(ctx context.Context, ifname string) error {
	return netlink.Admin(ctx, ifname, iflink.IFF_UP, 0)
}
