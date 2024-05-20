// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package net_tools

import (
	"context"
	"flag"
	"fmt"
	"net"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/flag/flagset"
	"github.com/platinasystems/goes/v2/pkg/net/af"
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
)

var routeModBranch = map[string]any{
	"add":     routeModCmd,
	"append":  routeModCmd,
	"change":  routeModCmd,
	"delete":  routeModCmd,
	"flush":   routeFlush,
	"get":     routeModCmd,
	"monitor": routeMonitor,
	"prepend": routeModCmd,
	"replace": routeModCmd,
	"test":    routeModCmd,
}

var routeModUsage = map[string]string{
	"add": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Add a route.

Destination Options{{dstopts}},
Gateway Options{{gwopts}}`,
	"append": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Change aspects of a route (such as its gateway).

Destination Options{{dstopts}},
Gateway Options{{gwopts}}`,
	"change": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Change aspects of a route (such as its gateway).

Destination Options{{dstopts}},
Gateway Options{{gwopts}}`,
	"delete": `
usage: {{branch .}} [<options>] <destination>
Delete a specific route.",
{{dstopts}}`,
	"get": `
usage: {{branch .}} [<options>] <destination>
Lookup and display the route for a destination.
{{dstopts}}`,
	"prepend": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Change or add new route.

Destination Options{{dstopts}},
Gateway Options{{gwopts}}`,
	"replace": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Change or add new route.

Destination Options{{dstopts}},
Gateway Options{{gwopts}}`,
	"test": `
usage: {{branch .}} [<options>] <destination> [<options>] <gateway> [<mask>]
Verify change or addition.

Destination Options{{dstopts}},
Gateway Options{{gwopts}}`,
}

var routeProtocol = map[string]uint8{
	"redirect": rtnetlink.RTPROT_REDIRECT,
	"kernel":   rtnetlink.RTPROT_KERNEL,
	"boot":     rtnetlink.RTPROT_BOOT,
	"static":   rtnetlink.RTPROT_STATIC,
}

var routeScope = map[string]rtnetlink.RtScope{
	"global":   rtnetlink.RT_SCOPE_UNIVERSE,
	"universe": rtnetlink.RT_SCOPE_UNIVERSE,
	"nowhere":  rtnetlink.RT_SCOPE_NOWHERE,
	"host":     rtnetlink.RT_SCOPE_HOST,
	"link":     rtnetlink.RT_SCOPE_LINK,
	"site":     rtnetlink.RT_SCOPE_SITE,
}

var routeTable = map[string]rtnetlink.RtTable{
	"compat":  rtnetlink.RT_TABLE_COMPAT,
	"default": rtnetlink.RT_TABLE_DEFAULT,
	"main":    rtnetlink.RT_TABLE_MAIN,
	"local":   rtnetlink.RT_TABLE_LOCAL,
}

var routeTo = map[string]uint8{
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

func routeGatewayOptions() *flag.FlagSet {
	opts := new(flag.FlagSet)
	opts.Int("expire", 0, "Seconds from now.")
	iface := opts.String("interface", "",
		"<gw> is a point-to-point interface name")
	opts.StringVar(iface, "iface", *iface, "aka -interface")
	opts.String("protocol", "boot", "{boot, kernel, redirect, static}")
	opts.String("scope", "global", "{global, nowhere, host, link, site}")
	opts.String("table", "main", "{compat, default, main, local}")
	opts.String("to", "unicast", "{unicast, broadcast, blackhole, etc.}")
	opts.Uint("hopcount", 0, "FIXME")
	opts.Uint("metric", 0, "FIXME")
	opts.Uint("mtu", 1500, "FIXME")
	opts.Uint("rtt", 0, "FIXME")
	opts.Uint("rttvar", 0, "FIXME")
	opts.Uint("ssthresh", 0, "FIXME")
	opts.Uint("tos", 0, "type-of-service")
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
		return nil, fmt.Errorf("%q %w", cmd, ErrUnsupported)
	}
	if hdr.Type != rtnetlink.RTM_GETROUTE {
		hdr.Flags |= netlink.NLM_F_ACK
	}
	rtm, req := netlink.ExpandRtMsg(req)
	rtm.Family = af.INET
	if dst.Addr().Is6() {
		rtm.Family = af.INET6
	}
	rtm.DstLen = uint8(dst.Bits())
	rtm.Table = rtnetlink.RT_TABLE_MAIN
	rtm.Scope = rtnetlink.RT_SCOPE_NOWHERE
	if hdr.Type != rtnetlink.RTM_DELROUTE {
		rtm.TOS = uint8(flagset.Search[uint](opts, "tos"))
		if v, ok := routeProtocol[flagset.Search[string](opts,
			"protocol")]; ok {
			rtm.Protocol = v
		} else {
			return nil, fmt.Errorf("%w protocol", ErrInvalid)
		}
		if v, ok := routeScope[flagset.Search[string](opts,
			"scope")]; ok {
			rtm.Scope = v
		} else {
			return nil, fmt.Errorf("%w scope", ErrInvalid)
		}
		if v, ok := routeTable[flagset.Search[string](opts,
			"table")]; ok {
			rtm.Table = v
		} else {
			return nil, fmt.Errorf("%w table", ErrInvalid)
		}
		if v, ok := routeTo[flagset.Search[string](opts, "to")]; ok {
			rtm.Type = v
		} else {
			return nil, fmt.Errorf("%w to", ErrInvalid)
		}
	}
	req = netlink.CatBytesAttr(req, rtnetlink.RTA_DST, dst.Addr().AsSlice())
	if hdr.Type == rtnetlink.RTM_NEWROUTE {
		if gw == nil {
			return nil, ErrNoGW
		}
		switch t := gw.(type) {
		case netip.Addr:
			gwattr := netrt.GatewayAttr(dst.Addr(), t)
			req = netlink.CatBytesAttr(req, gwattr, t.AsSlice())
		case []net.IPAddr:
			ipa := netrt.SelectGateway(t, dst.Addr().Is6())
			gwaddr, ok := netip.AddrFromSlice(ipa.IP)
			if !ok {
				return nil, fmt.
					Errorf("%w resolved address (%v)",
						ErrInvalid, ipa)
			}
			gwattr := netrt.GatewayAttr(dst.Addr(), gwaddr)
			gwip := gwaddr.AsSlice()
			req = netlink.CatBytesAttr(req, gwattr, gwip)
		case *netif.NetIf:
			gwi := uint32(t.Index)
			req = netlink.CatAttr(req, rtnetlink.RTA_IIF, gwi)
			req = netlink.CatAttr(req, rtnetlink.RTA_OIF, gwi)
		default:
			return nil, fmt.Errorf("%w <gateway>", ErrInvalid)
		}
	}
	if u := flagset.Search[uint](opts, "mtu"); u != 1500 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_MTU, uint32(u))
	}
	if secs := flagset.Search[int](opts, "expire"); secs != 0 {
		elapse := time.Second * time.Duration(secs)
		expire := uint32(time.Now().Add(elapse).Unix())
		req = netlink.CatAttr(req, rtnetlink.RTA_EXPIRES, expire)
	}
	if u := flagset.Search[uint](opts, "hopcount"); u != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_HOPLIMIT, uint32(u))
	}
	if u := flagset.Search[uint](opts, "metric"); u != 0 {
		req = netlink.CatAttr(req, rtnetlink.RTA_PRIORITY, uint32(u))
	}
	if u := flagset.Search[uint](opts, "ssthresh"); u != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_SSTHRESH, uint32(u))
	}
	if u := flagset.Search[uint](opts, "rtt"); u != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_RTT, uint32(u))
	}
	if u := flagset.Search[uint](opts, "rttvar"); u != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_RTTVAR, uint32(u))
	}
	if len(mx) > 0 {
		req = netlink.CatBytesAttr(req, rtnetlink.RTA_METRICS, mx)
	}
	netlink.PointerMsgHdr(req).Len = uint32(len(req))
	return netrt.OneReq(ctx, req)
}
