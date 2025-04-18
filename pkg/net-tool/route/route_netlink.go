// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package route

import (
	"context"
	"flag"
	"net"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type gatewayOptions struct {
	iface  *bool
	expire *int
	protocol,
	scope,
	table,
	to *string
	hopcount,
	metric,
	mtu,
	rtt,
	rttvar,
	ssthresh,
	tos *uint
}

var Features = map[string]any{
	"add": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Add a route.

{{flags .}}`)
		return opts.mod(ctx, "add", args)
	},
	"append": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Change aspects of a route (such as its gateway).

{{flags .}}`)
		return opts.mod(ctx, "append", args)
	},
	"change": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
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
Delete a specific route.

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
	"prepend": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Prepend or add new route.

{{flags .}}`)
		return opts.mod(ctx, "prepend", args)
	},
	"replace": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Replace or add new route.

{{flags .}}`)
		return opts.mod(ctx, "replace", args)
	},
	"test": func(ctx context.Context, args []string) error {
		opts := newModOptions()
		opts.dst = newDestinationOptions()
		opts.gw = newGatewayOptions()
		xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Verify change or addition.

{{flags .}}`)
		return opts.mod(ctx, "test", args)
	},
}

var protocols = map[string]uint8{
	"redirect": rtnetlink.RTPROT_REDIRECT,
	"kernel":   rtnetlink.RTPROT_KERNEL,
	"boot":     rtnetlink.RTPROT_BOOT,
	"static":   rtnetlink.RTPROT_STATIC,
}

var scopes = map[string]rtnetlink.RtScope{
	"global":   rtnetlink.RT_SCOPE_UNIVERSE,
	"universe": rtnetlink.RT_SCOPE_UNIVERSE,
	"nowhere":  rtnetlink.RT_SCOPE_NOWHERE,
	"host":     rtnetlink.RT_SCOPE_HOST,
	"link":     rtnetlink.RT_SCOPE_LINK,
	"site":     rtnetlink.RT_SCOPE_SITE,
}

var tables = map[string]rtnetlink.RtTable{
	"compat":  rtnetlink.RT_TABLE_COMPAT,
	"default": rtnetlink.RT_TABLE_DEFAULT,
	"main":    rtnetlink.RT_TABLE_MAIN,
	"local":   rtnetlink.RT_TABLE_LOCAL,
}

var types = map[string]uint8{
	"unicast":     rtnetlink.RTN_UNICAST,
	"local":       rtnetlink.RTN_LOCAL,
	"broadcast":   rtnetlink.RTN_BROADCAST,
	"anycast":     rtnetlink.RTN_ANYCAST,
	"multicast":   rtnetlink.RTN_MULTICAST,
	"blackhole":   rtnetlink.RTN_BLACKHOLE,
	"unreachable": rtnetlink.RTN_UNREACHABLE,
	"prohibit":    rtnetlink.RTN_PROHIBIT,
	"throw":       rtnetlink.RTN_THROW,
	"nat":         rtnetlink.RTN_NAT,
	"xresolve":    rtnetlink.RTN_XRESOLVE,
	"cnt":         rtnetlink.RTN_CNT,
}

func newGatewayOptions() *gatewayOptions {
	opts := &gatewayOptions{
		expire: flag.Int("expire", 0, "Seconds from now."),
		iface: flag.Bool("interface", false,
			"<gateway> is a point-to-point interface name"),
		protocol: flag.String("protocol", "boot",
			"{boot, kernel, redirect, static}"),
		scope: flag.String("scope", "global",
			"{global, nowhere, host, link, site}"),
		table: flag.String("table", "main",
			"{compat, default, main, local}"),
		to: flag.String("to", "unicast",
			"{unicast, broadcast, blackhole, etc.}"),
		hopcount: flag.Uint("hopcount", 0, "FIXME"),
		metric:   flag.Uint("metric", 0, "FIXME"),
		mtu:      flag.Uint("mtu", 1500, "FIXME"),
		rtt:      flag.Uint("rtt", 0, "FIXME"),
		rttvar:   flag.Uint("rttvar", 0, "FIXME"),
		ssthresh: flag.Uint("ssthresh", 0, "FIXME"),
		tos:      flag.Uint("tos", 0, "type-of-service"),
	}
	flag.BoolVar(opts.iface, "iface", *opts.iface, "aka -interface")
	return opts
}

func (opts *modOptions) req(
	ctx context.Context,
	cmd string,
	fib int,
	dst netip.Prefix,
	gw any,
) (netrt.Rt, error) {
	var mx []byte
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Flags = netlink.NLM_F_REQUEST
	switch cmd {
	case "add":
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_EXCL
	case "append":
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_APPEND
	case "change":
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE
	case "delete":
		hdr.Type = rtnetlink.RTM_DELROUTE
	case "get":
		hdr.Type = rtnetlink.RTM_GETROUTE
	case "prepend":
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE
	case "replace":
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE
	case "test":
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_EXCL
	default:
		return nil, xerrors.Invalid(cmd)
	}
	if hdr.Type != rtnetlink.RTM_GETROUTE {
		hdr.Flags |= netlink.NLM_F_ACK
	}
	rtm, req := netlink.ExpandRtMsg(req)
	rtm.Family = xnet.AF_INET
	if dst.Addr().Is6() {
		rtm.Family = xnet.AF_INET6
	}
	rtm.DstLen = uint8(dst.Bits())
	rtm.Table = rtnetlink.RT_TABLE_MAIN
	rtm.Scope = rtnetlink.RT_SCOPE_NOWHERE
	if hdr.Type != rtnetlink.RTM_DELROUTE {
		rtm.TOS = uint8(*opts.gw.tos)
		if v, ok := protocols[*opts.gw.protocol]; ok {
			rtm.Protocol = v
		} else {
			return nil, xerrors.Invalid("protocol")
		}
		if v, ok := scopes[*opts.gw.scope]; ok {
			rtm.Scope = v
		} else {
			return nil, xerrors.Invalid("scope")
		}
		if v, ok := tables[*opts.gw.table]; ok {
			rtm.Table = v
		} else {
			return nil, xerrors.Invalid("table")
		}
		if v, ok := types[*opts.gw.to]; ok {
			rtm.Type = v
		} else {
			return nil, xerrors.Invalid("to")
		}
	}
	req = netlink.CatBytesAttr(req, rtnetlink.RTA_DST, dst.Addr().AsSlice())
	if hdr.Type == rtnetlink.RTM_NEWROUTE {
		if gw == nil {
			return nil, xerrors.Incomplete("gateway")
		}
		switch t := gw.(type) {
		case netip.Addr:
			gwattr := netrt.GatewayAttr(dst.Addr(), t)
			req = netlink.CatBytesAttr(req, gwattr, t.AsSlice())
		case []net.IPAddr:
			ipa := netrt.SelectGateway(t, dst.Addr().Is6())
			gwaddr, ok := netip.AddrFromSlice(ipa.IP)
			if !ok {
				return nil, xerrors.
					Invalid("resolved", ipa.String())
			}
			gwattr := netrt.GatewayAttr(dst.Addr(), gwaddr)
			gwip := gwaddr.AsSlice()
			req = netlink.CatBytesAttr(req, gwattr, gwip)
		case *netif.NetIf:
			gwi := uint32(t.Index)
			req = netlink.CatAttr(req, rtnetlink.RTA_IIF, gwi)
			req = netlink.CatAttr(req, rtnetlink.RTA_OIF, gwi)
		default:
			return nil, xerrors.Invalid("gateway")
		}
	}
	if mtu := uint32(*opts.gw.mtu); mtu != 1500 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_MTU, mtu)
	}
	if secs := *opts.gw.expire; secs != 0 {
		elapse := time.Second * time.Duration(secs)
		expire := uint32(time.Now().Add(elapse).Unix())
		req = netlink.CatAttr(req, rtnetlink.RTA_EXPIRES, expire)
	}
	if hopcount := uint32(*opts.gw.hopcount); hopcount != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_HOPLIMIT, hopcount)
	}
	if metric := uint32(*opts.gw.metric); metric != 0 {
		req = netlink.CatAttr(req, rtnetlink.RTA_PRIORITY, metric)
	}
	if t := uint32(*opts.gw.ssthresh); t != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_SSTHRESH, t)
	}
	if rtt := uint32(*opts.gw.rtt); rtt != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_RTT, rtt)
	}
	if rttvar := uint32(*opts.gw.rttvar); rttvar != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_RTTVAR, rttvar)
	}
	if len(mx) > 0 {
		req = netlink.CatBytesAttr(req, rtnetlink.RTA_METRICS, mx)
	}
	netlink.PointerMsgHdr(req).Len = uint32(len(req))
	return netrt.OneRtReq(ctx, req)
}
