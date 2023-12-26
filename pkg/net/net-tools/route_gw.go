// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"flag"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
)

func routeGateway(
	ctx context.Context,
	gwarg string,
	opts *flag.FlagSet,
) (any, error) {
	if flagset.Search[bool](opts, "interface") {
		nif := netif.Named(gwarg)
		if nif == nil {
			return nil, egress.Markf("%q %w", gwarg, ErrNotFound)
		}
		return nif, nil
	}
	if routeAddrIsNumeric(gwarg) {
		ga, err := netip.ParseAddr(gwarg)
		if err != nil {
			return nil, egress.Markf("%q %w", gwarg, err)
		}
		return ga, nil
	}
	ipas, err := net.DefaultResolver.LookupIPAddr(ctx, gwarg)
	if err != nil {
		return nil, egress.Markf("%q %w", gwarg, err)
	}
	return ipas, nil
}

func routeSelectGateway(ipas []net.IPAddr, prefer6 bool) net.IPAddr {
	ipa := ipas[0]
	l := net.IPv4len
	if prefer6 {
		l = net.IPv6len
	}
	for _, entry := range ipas {
		if len(entry.IP) == l {
			ipa = entry
		}
	}
	return ipa
}
