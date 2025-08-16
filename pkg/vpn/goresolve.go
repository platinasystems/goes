// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"net"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
)

const ResolveRetryInterval = time.Second

var Resolver = net.Resolver{
	PreferGo: true,
}

func resolve(ctx context.Context, hn string) (netip.Addr, error) {
	var z netip.Addr
	if addr, err := netip.ParseAddr(hn); err == nil {
		return addr, err
	}
	ips, err := WaitForResolution(ctx, "ip", hn, 10*time.Second)
	if err != nil {
		return z, err
	}
	if len(ips) == 0 {
		return z, xerrors.Invalid(hn)
	}
	addr, ok := netip.AddrFromSlice(ips[0])
	if !ok {
		return z, xerrors.Invalid(hn)
	}
	return addr, nil
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
		case <-time.After(ResolveRetryInterval):
			xlog.Trace.Println("retry", hostname)
		}
	}
	return
}
