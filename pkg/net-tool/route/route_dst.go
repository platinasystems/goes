// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
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

func (rt Route) dst(ctx context.Context) (dst netip.Prefix, gw any, err error) {
	var (
		ok,
		forceHost,
		forceNet bool
		addr netip.Addr
		ipas []net.IPAddr
	)
	narg := flag.NArg()
	if narg == 0 {
		err = xerrors.Incomplete("destination")
		return
	}
	dstarg := flag.Arg(0)
	if RouteHost {
		forceHost = true
	} else if RouteNet {
		forceNet = true
	}

	if dstarg == "default" {
		if RouteInet6 {
			dst = netip.PrefixFrom(netip.IPv4Unspecified(), 0)
		} else {
			dst = netip.PrefixFrom(netip.IPv6Unspecified(), 0)
		}
		if narg > 1 {
			gw, err = lookupGW(ctx, flag.Arg(1))
		}
		return
	}
	bits := RoutePrefixlen
	if slash := strings.Index(dstarg, "/"); slash > 0 {
		if slash == len(dstarg)-1 {
			err = xerrors.Invalid("destination", dstarg)
			return
		}
		if _, err = fmt.Sscan(dstarg[slash+1:], &bits); err != nil {
			err = xerrors.Label(err, "destination", dstarg)
			return
		}
		dstarg = dstarg[:slash]
	} else if bits < 0 && narg > 1 {
		maskarg := flag.Arg(1)
		if addr, err = netip.ParseAddr(maskarg); err != nil {
			err = xerrors.Label(err, "mask", maskarg)
			return
		} else {
			bits, _ = net.IPMask(addr.AsSlice()).Size()
		}
		if narg > 2 {
			if gw, err = lookupGW(ctx, flag.Arg(2)); err != nil {
				return
			}
		}
	} else if narg > 1 {
		if gw, err = lookupGW(ctx, flag.Arg(1)); err != nil {
			return
		}
	}
	if isNumericAddr(dstarg) {
		if addr, err = netip.ParseAddr(dstarg); err != nil {
			err = xerrors.Label(err, "destrination", dstarg)
			return
		}
	} else if ipas, err = net.DefaultResolver.
		LookupIPAddr(ctx, dstarg); err != nil {
		err = xerrors.Label(err, "destination", dstarg)
		return
	} else {
		ipa := ipas[0]
		if RouteInet6 {
			for _, t := range ipas {
				if len(t.IP) == net.IPv6len {
					ipa = t
					break
				}
			}
		} else if RouteInet {
			for _, t := range ipas {
				if len(t.IP) == net.IPv4len {
					ipa = t
					break
				}
			}
		}
		if addr, ok = netip.AddrFromSlice(ipa.IP); !ok {
			err = xerrors.Invalid("ip", ipa.IP.String())
			return
		}
	}
	if bits < 0 {
		if forceHost {
			bits = addr.BitLen()
		} else if forceNet && addr.Is4() {
			bits, _ = net.IP(addr.AsSlice()).DefaultMask().Size()
		} else if rt != Get {
			err = xerrors.Incomplete("mask")
			return
		} else if bits = 32; addr.Is6() {
			bits = 128
		}
	}
	dst = netip.PrefixFrom(addr, bits)
	return
}
