// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/netioctl"
	"golang.org/x/sys/unix"
)

func (nif *NetIf) Destroy(ctx context.Context) error {
	req := netioctl.NewIfReqNothing(nif.Name)
	return netioctl.Inet(unix.SIOCIFDESTROY, req)
}
