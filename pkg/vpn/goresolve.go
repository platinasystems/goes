// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const (
	MinResolveRetryInterval = 100 * time.Millisecond
	MaxResolveRetryInterval = 3 * time.Second
)

var Resolver = net.Resolver{
	PreferGo: true,
}

func PatientLookupIP(
	ctx context.Context, network, hostname string, timeout time.Duration,
) ([]net.IP, error) {
	var (
		dnserr *net.DNSError
		total  time.Duration
	)
	dur := MinResolveRetryInterval
	if dur >= timeout {
		dur = timeout / 3
	}
	for true {
		ips, err := Resolver.LookupIP(ctx, network, hostname)
		if err == nil {
			if total > MinResolveRetryInterval {
				verbose.Println("found", hostname, ips, total)
			}
			return ips, err
		}
		if total >= timeout ||
			!errors.As(err, &dnserr) ||
			!dnserr.IsNotFound {
			return ips, err
		}
		select {
		case <-ctx.Done():
			return ips, ctx.Err()
		case <-time.After(dur):
			if total == 0 {
				verbose.Println("wait for", hostname, "...")
			}
			total += dur
			if dur *= 2; dur > MaxResolveRetryInterval {
				dur = MaxResolveRetryInterval
			} else if total+dur > timeout {
				dur = timeout - total
			}
		}
	}
	return []net.IP{}, xerrors.Broken()
}
