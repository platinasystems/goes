// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package route

import (
	"context"
	"flag"
	"net"
	"net/netip"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"golang.org/x/sys/unix"
)

type gatewayOptions struct {
	iface *bool
	flags,
	genmask,
	ifa,
	ifp,
	metrics *string
}

var Features = map[string]any{
	"add": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> <gateway> [mask]
Add a route.

{{flags .}}`)
		return opts.mod(ctx, "add", args)
	},
	"change": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> <gateway> [mask]
Change aspects of a route (such as its gateway).

{{flags .}}`)
		return opts.mod(ctx, "change", args)
	},
	"delete": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		// delete doesn't define gatewayOptions
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination>
Delete specific route.

{{flags .}}`)
		return opts.mod(ctx, "delete", args)
	},
	"flush": flush,
	"get": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		// get doesn't define gatewayOptions
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination>
Lookup and display the route for a destination.

{{flags .}}`)
		return opts.mod(ctx, "get", args)
	},
	"monitor": monitor,
}

type gwflag struct {
	rtf  uint
	name string
	val  bool
	dsc  string
}

type gwmetric struct {
	name string
	val  uint
	dsc  string
}

func newGatewayOptions() *gatewayOptions {
	opts := &gatewayOptions{
		flags: flag.String("flags", "",
			"comma separated names\n"+gwFlags),
		genmask: flag.String("genmask", "",
			"generate netmask"),
		ifa: flag.String("ifa", "",
			"A MAC address of a point-to-point peer?"),
		ifp: flag.String("ifp", "",
			"A point-to-point peer interface and MAC?"),
		iface: flag.Bool("interface", false,
			"<gateway> is a point-to-point interface name"),
		metrics: flag.String("metrics", "",
			"comma separated NAME=VALUE\n"+gwMetrics),
	}
	flag.BoolVar(opts.iface, "iface", *opts.iface, "aka -interface")
	return opts
}

func (opts *modOptions) req(
	ctx context.Context,
	op string,
	fib int,
	dst netip.Prefix,
	gw any,
) (netrt.NetRt, error) {
	msg := netrt.NewRtMsg()
	rtm := netrt.PointerRtMsghdr(msg)
	switch op {
	case "add", "change":
		if gw == nil {
			return nil, xerrors.Incomplete("gateway")
		}
		switch op {
		case "add":
			rtm.Type = unix.RTM_ADD
		case "change":
			rtm.Type = unix.RTM_CHANGE
		}
		integer.Set(&rtm.Flags, unix.RTF_UP)
		if dst.Bits() == dst.Addr().BitLen() {
			integer.Set(&rtm.Flags, unix.RTF_HOST)
		}
		for _, name := range strings.Split(*opts.gw.flags, ",") {
			if val, ok := gwFlagValues[name]; ok {
				integer.Set(&rtm.Flags, val)
			} else if val, ok = gwFlagValues["no"+name]; ok {
				integer.Reset(&rtm.Flags, val)
			}
		}
		rtmmetrics(rtm, strings.Split(*opts.gw.metrics, ","))
	case "del":
		rtm.Type = unix.RTM_DELETE
		rtmdel(rtm)
	case "get":
		rtm.Type = unix.RTM_GET
	default:
		return nil, xerrors.Invalid("command", op)
	}
	msg = xnet.SAAppend(msg, dst.Addr())
	rtm = netrt.PointerRtMsghdr(msg)
	integer.Set(&rtm.Addrs, 1<<unix.RTAX_DST)
	switch t := gw.(type) {
	case netip.Addr:
		msg = xnet.SAAppend(msg, t)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case []net.IPAddr:
		ipa := netrt.SelectGateway(t, dst.Addr().Is6())
		gwaddr, ok := netip.AddrFromSlice(ipa.IP)
		if !ok {
			return nil, xerrors.Invalid("resolved", ipa.String())
		}
		msg = xnet.SAAppend(msg, gwaddr)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case *netif.NetIf:
		msg = xnet.SAAppendDataLink(msg,
			uint16(t.Index),
			uint8(t.Type),
			t.Name,
			t.HardwareAddr,
			[]byte{})
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
	default:
		return nil, xerrors.Invalid("gateway")
	}
	if bits := dst.Bits(); bits > 0 {
		mask := net.CIDRMask(bits, dst.Addr().BitLen())
		maskaddr, ok := netip.AddrFromSlice(mask)
		if !ok {
			return nil, xerrors.Invalid("mask")
		}
		msg = xnet.SAAppend(msg, maskaddr)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_NETMASK)
	}
	if len(*opts.gw.genmask) > 0 {
		// FIXME e.g. 255.255.255.255 ?
	}
	if len(*opts.gw.ifp) > 0 {
		// FIXME e.g. eth0:1.2.3.4.5.6 ?
	}
	if len(*opts.gw.ifa) > 0 {
		// FIXME e.g. 1.2.3.4.5.6 ?
	}
	return netrt.Request(msg, fib)
}
