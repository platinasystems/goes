// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package route

import (
	"context"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

func lookupGW(ctx context.Context, arg string) (any, error) {
	if IfaceFlag.Value() {
		return netif.Named(ctx, arg)
	}
	if isNumericAddr(arg) {
		ga, err := netip.ParseAddr(arg)
		if err != nil {
			return nil, xerrors.
				Label(err, "gateway", "address", arg)
		}
		return ga, nil
	}
	ipas, err := net.DefaultResolver.LookupIPAddr(ctx, arg)
	if err != nil {
		return nil, xerrors.Label(err, "gateway", "lookup", arg)
	}
	return ipas, nil
}
