// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package net_tools

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
)

func routeDestinationOptions() *flag.FlagSet {
	opts := new(flag.FlagSet)
	inet := opts.Bool("inet", false, "Address hint or filter.")
	opts.BoolVar(inet, "4", *inet, "aka -inet")
	inet6 := opts.Bool("inet6", false, "Address hint or filter.")
	opts.BoolVar(inet6, "6", *inet6, "aka -inet6")
	opts.Bool("host", false, "Host <destination>.")
	opts.Bool("net", false, "Network <destination>.")
	opts.Int("prefixlen", -1,
		"If >= 0, use this instead of 1st arg /<suffix> or 3rd arg.")
	return opts
}

func routeDestination(
	ctx context.Context,
	dstarg, maskarg string,
	opts *flag.FlagSet,
) (netip.Prefix, error) {
	var err error
	var ok bool
	var forceHost, forceNet bool
	var addr netip.Addr
	var ipas []net.IPAddr
	var zero netip.Prefix
	if flagset.Search[bool](opts, "host") {
		forceHost = true
	} else if flagset.Search[bool](opts, "net") {
		forceNet = true
	}

	if dstarg == "default" {
		if flagset.Search[bool](opts, "inet6") {
			return netip.
				PrefixFrom(netip.IPv6Unspecified(), 0), nil
		}
		return netip.PrefixFrom(netip.IPv4Unspecified(), 0), nil
	}
	bits := flagset.Search[int](opts, "prefixlen")
	if slash := strings.Index(dstarg, "/"); slash > 0 {
		if slash == len(dstarg)-1 {
			return zero, egress.Markf("%q %w", dstarg, ErrInvalid)
		}
		if _, err := fmt.Sscan(dstarg[slash+1:], &bits); err != nil {
			return zero, egress.Markf("%q %w", dstarg, err)
		}
		dstarg = dstarg[:slash]
	} else if bits < 0 && len(maskarg) > 0 {
		if addr, err = netip.ParseAddr(maskarg); err != nil {
			return zero, egress.Markf("%q %w", maskarg, err)
		} else {
			bits, _ = net.IPMask(addr.AsSlice()).Size()
		}
	}
	if routeAddrIsNumeric(dstarg) {
		if addr, err = netip.ParseAddr(dstarg); err != nil {
			return zero, egress.Markf("%q %w", dstarg, err)
		}
	} else if ipas, err = net.DefaultResolver.
		LookupIPAddr(ctx, dstarg); err != nil {
		return zero, egress.Markf("%q %w", dstarg, err)
	} else {
		ipa := ipas[0]
		if flagset.Search[bool](opts, "inet6") {
			for _, t := range ipas {
				if len(t.IP) == net.IPv6len {
					ipa = t
					break
				}
			}
		} else if flagset.Search[bool](opts, "inet") {
			for _, t := range ipas {
				if len(t.IP) == net.IPv4len {
					ipa = t
					break
				}
			}
		}
		if addr, ok = netip.AddrFromSlice(ipa.IP); !ok {
			return zero, egress.Markf("%v %w", ipa.IP, ErrInvalid)
		}
	}
	if bits < 0 {
		if forceHost {
			bits = addr.BitLen()
		} else if forceNet && addr.Is4() {
			bits, _ = net.IP(addr.AsSlice()).DefaultMask().Size()
		} else {
			return zero, ErrNoMask
		}
	}
	return netip.PrefixFrom(addr, bits), nil
}
