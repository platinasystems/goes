// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package route

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

type destinationOptions struct {
	inet,
	inet6,
	host,
	net *bool
	prefixlen *int
}

func newDestinationOptions() *destinationOptions {
	opts := &destinationOptions{
		inet: flag.Bool("inet", false,
			"Address hint or filter."),
		inet6: flag.Bool("inet6", false,
			"Address hint or filter."),
		host: flag.Bool("host", false,
			"Host <destination>."),
		net: flag.Bool("net", false,
			"Network <destination>."),
		prefixlen: flag.Int("prefixlen", -1,
			"If >= 0, use instead of 1st arg/<suffix> or 3rd arg."),
	}
	flag.BoolVar(opts.inet, "4", *opts.inet, "aka -inet")
	flag.BoolVar(opts.inet6, "6", *opts.inet6, "aka -inet6")
	return opts
}

func (opts *destinationOptions) prefix(
	ctx context.Context,
	dstarg, maskarg string,
) (netip.Prefix, error) {
	var (
		err error
		ok,
		forceHost,
		forceNet bool
		addr netip.Addr
		ipas []net.IPAddr
		zero netip.Prefix
	)
	if *opts.host {
		forceHost = true
	} else if *opts.net {
		forceNet = true
	}

	if dstarg == "default" {
		if *opts.inet6 {
			return netip.
				PrefixFrom(netip.IPv6Unspecified(), 0), nil
		}
		return netip.PrefixFrom(netip.IPv4Unspecified(), 0), nil
	}
	bits := *opts.prefixlen
	if slash := strings.Index(dstarg, "/"); slash > 0 {
		if slash == len(dstarg)-1 {
			return zero, xerrors.Invalid("destination", dstarg)
		}
		if _, err := fmt.Sscan(dstarg[slash+1:], &bits); err != nil {
			return zero, xerrors.Label(err, "destination", dstarg)
		}
		dstarg = dstarg[:slash]
	} else if bits < 0 && len(maskarg) > 0 {
		if addr, err = netip.ParseAddr(maskarg); err != nil {
			return zero, xerrors.Label(err, "mask", maskarg)
		} else {
			bits, _ = net.IPMask(addr.AsSlice()).Size()
		}
	}
	if isNumericAddr(dstarg) {
		if addr, err = netip.ParseAddr(dstarg); err != nil {
			return zero, xerrors.Label(err, "destrination", dstarg)
		}
	} else if ipas, err = net.DefaultResolver.
		LookupIPAddr(ctx, dstarg); err != nil {
		return zero, xerrors.Label(err, "destination", dstarg)
	} else {
		ipa := ipas[0]
		if *opts.inet6 {
			for _, t := range ipas {
				if len(t.IP) == net.IPv6len {
					ipa = t
					break
				}
			}
		} else if *opts.inet {
			for _, t := range ipas {
				if len(t.IP) == net.IPv4len {
					ipa = t
					break
				}
			}
		}
		if addr, ok = netip.AddrFromSlice(ipa.IP); !ok {
			return zero, xerrors.Invalid("ip", ipa.IP.String())
		}
	}
	if bits < 0 {
		if forceHost {
			bits = addr.BitLen()
		} else if forceNet && addr.Is4() {
			bits, _ = net.IP(addr.AsSlice()).DefaultMask().Size()
		} else {
			return zero, xerrors.Incomplete("mask")
		}
	}
	return netip.PrefixFrom(addr, bits), nil
}
