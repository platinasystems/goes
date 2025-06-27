// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"net"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xlog"
)

const (
	MinResolveRetryInterval = 100 * time.Millisecond
	MaxResolveRetryInterval = 3 * time.Second
)

var Resolver = net.Resolver{
	PreferGo: true,
}

func WaitForResolution(
	ctx context.Context, network, hostname string, timeout time.Duration,
) (ips []net.IP, err error) {
	begin := time.Now()
	for {
		ips, err = Resolver.LookupIP(ctx, network, hostname)
		if err == nil {
			break
		}
		if time.Now().Sub(begin) > timeout {
			return
		}
		select {
		case <-ctx.Done():
			err = ctx.Err()
			return
		case <-time.After(time.Second):
			xlog.Info.Println("retry", hostname, "...")
		}
	}
	xlog.Info.Println("after", time.Now().Sub(begin), "found", hostname,
		"with", ips)
	return
}
