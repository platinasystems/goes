// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package net_tools

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/net/sockaddr"
)

var routeModBranch = map[string]any{
	"add":     routeModCmd,
	"change":  routeModCmd,
	"delete":  routeModCmd,
	"flush":   routeFlush,
	"get":     routeModCmd,
	"monitor": routeMonitor,
}

var routeModUsage = map[string]string{
	"add": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Add a route.

Destination Options{{dstopts}}
Gateway Options{{gwopts}}`,
	"change": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Change aspects of a route (such as its gateway).

Destination Options{{dstopts}}
Gateway Options{{gwopts}}`,
	"delete": `
usage: {{branch .}} [<options>] <destination>
Delete a specific route.",
{{dstopts}}`,
	"get": `
usage: {{branch .}} [<options>] <destination>
Lookup and display the route for a destination.
{{dstopts}}`,
}

func routeGatewayOptions() *flag.FlagSet {
	opts := new(flag.FlagSet)
	opts.String("genmask", "", "FIXME")
	opts.String("ifa", "", "A MAC address of a point-to-point peer?")
	opts.String("ifp", "", "A point-to-point peer interface and MAC?")
	netrt.AddFlags(opts)
	netrt.AddMetricFlags(opts)
	return opts
}

func routeModReq(
	ctx context.Context,
	cmd string,
	fib int,
	dst netip.Prefix,
	gw any,
	opts *flag.FlagSet,
) (netrt.NetRt, error) {
	msg := netrt.NewRtMsg()
	rtm := netrt.PointerRtMsghdr(msg)
	switch cmd {
	case "add", "change":
		if gw == nil {
			return nil, ErrNoGW
		}
		switch cmd {
		case "add":
			rtm.Type = syscall.RTM_ADD
		case "change":
			rtm.Type = syscall.RTM_CHANGE
		}
		integer.Set(&rtm.Flags, netrt.RTF_UP)
		if dst.Bits() == dst.Addr().BitLen() {
			integer.Set(&rtm.Flags, netrt.RTF_HOST)
		}
		netrt.SetFlags(rtm, opts)
		netrt.SetMetrics(rtm, opts)
	case "del":
		rtm.Type = syscall.RTM_DELETE
		integer.Set(&rtm.Flags, netrt.RTF_PINNED)
	case "get":
		rtm.Type = syscall.RTM_GET
	default:
		return nil, fmt.Errorf("%q %w", cmd, ErrUnsupported)
	}
	msg = sockaddr.Append(msg, dst.Addr())
	rtm = netrt.PointerRtMsghdr(msg)
	integer.Set(&rtm.Addrs, 1<<syscall.RTAX_DST)
	switch t := gw.(type) {
	case netip.Addr:
		msg = sockaddr.Append(msg, t)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<syscall.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, netrt.RTF_GATEWAY)
	case []net.IPAddr:
		ipa := routeSelectGateway(t, dst.Addr().Is6())
		gwaddr, ok := netip.AddrFromSlice(ipa.IP)
		if !ok {
			return nil, fmt.Errorf("%w resolved address (%v)",
				ErrInvalid, ipa)
		}
		msg = sockaddr.Append(msg, gwaddr)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<syscall.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, netrt.RTF_GATEWAY)
	case *netif.NetIf:
		msg = sockaddr.AppendDl(msg,
			uint16(t.Index),
			uint8(t.Type),
			t.Name,
			t.HardwareAddr,
			[]byte{})
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<syscall.RTAX_GATEWAY)
	default:
		return nil, fmt.Errorf("<gateway>: %w", ErrInvalid)
	}
	if bits := dst.Bits(); bits > 0 {
		mask := net.CIDRMask(bits, dst.Addr().BitLen())
		maskaddr, ok := netip.AddrFromSlice(mask)
		if !ok {
			return nil,
				fmt.Errorf("%w mask or prefixlen", ErrInvalid)
		}
		msg = sockaddr.Append(msg, maskaddr)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<syscall.RTAX_NETMASK)
	}
	if s := flagset.Search[string](opts, "genmask"); len(s) > 0 {
		// FIXME e.g. 255.255.255.255 ?
	}
	if s := flagset.Search[string](opts, "ifp"); len(s) > 0 {
		// FIXME e.g. eth0:1.2.3.4.5.6 ?
	}
	if s := flagset.Search[string](opts, "ifa"); len(s) > 0 {
		// FIXME e.g. 1.2.3.4.5.6 ?
	}
	return netrt.Request(msg, fib)
}
