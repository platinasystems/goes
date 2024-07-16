// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"net/netip"
)

func routeAdd(ctx context.Context, dst netip.Prefix, gw any) error {
	return route(ctx, true, dst, gw)
}

func routeDelete(ctx context.Context, dst netip.Prefix, gw any) error {
	return route(ctx, false, dst, gw)
}
