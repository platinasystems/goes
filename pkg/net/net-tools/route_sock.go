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

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/net/sockaddr"
	"golang.org/x/sys/unix"
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

type routeGatewayFlag struct {
	rtf  uint
	name string
	val  bool
	dsc  string
}

type routeGatewayMetric struct {
	name string
	val  uint
	dsc  string
}

func routeGatewayOptions() *flag.FlagSet {
	opts := new(flag.FlagSet)
	opts.String("genmask", "", "FIXME")
	opts.String("ifa", "", "A MAC address of a point-to-point peer?")
	opts.String("ifp", "", "A point-to-point peer interface and MAC?")
	iface := opts.Bool("interface", false,
		"<gateway> is a point-to-point interface name")
	opts.BoolVar(iface, "iface", *iface, "aka -interface")
	for _, o := range routeGatewayFlags {
		opts.Bool(o.name, o.val, o.dsc)
	}
	for _, o := range routeGatewayMetrics {
		opts.Uint(o.name, o.val, o.dsc)
	}
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
			rtm.Type = unix.RTM_ADD
		case "change":
			rtm.Type = unix.RTM_CHANGE
		}
		integer.Set(&rtm.Flags, unix.RTF_UP)
		if dst.Bits() == dst.Addr().BitLen() {
			integer.Set(&rtm.Flags, unix.RTF_HOST)
		}
		for _, o := range routeGatewayFlags {
			if flagset.Search[bool](opts, o.name) {
				integer.Set(&rtm.Flags, o.rtf)
			} else {
				integer.Reset(&rtm.Flags, o.rtf)
			}
		}
		routeSetMetrics(rtm, opts)
	case "del":
		rtm.Type = unix.RTM_DELETE
		routeSetDelFlag(rtm)
	case "get":
		rtm.Type = unix.RTM_GET
	default:
		return nil, fmt.Errorf("%q %w", cmd, ErrUnsupported)
	}
	msg = sockaddr.Append(msg, dst.Addr())
	rtm = netrt.PointerRtMsghdr(msg)
	integer.Set(&rtm.Addrs, 1<<unix.RTAX_DST)
	switch t := gw.(type) {
	case netip.Addr:
		msg = sockaddr.Append(msg, t)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case []net.IPAddr:
		ipa := netrt.SelectGateway(t, dst.Addr().Is6())
		gwaddr, ok := netip.AddrFromSlice(ipa.IP)
		if !ok {
			return nil, fmt.Errorf("%w resolved address (%v)",
				ErrInvalid, ipa)
		}
		msg = sockaddr.Append(msg, gwaddr)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case *netif.NetIf:
		msg = sockaddr.AppendDl(msg,
			uint16(t.Index),
			uint8(t.Type),
			t.Name,
			t.HardwareAddr,
			[]byte{})
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
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
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_NETMASK)
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
