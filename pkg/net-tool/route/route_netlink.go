// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package route

import (
	"context"
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

const gwFlags = ""
const gwMetrics = ""

var Features = map[string]any{
	string(Add):     Add.op,
	string(Append):  Append.op,
	string(Change):  Change.op,
	string(Delete):  Delete.op,
	string(Flush):   flush,
	string(Get):     Get.op,
	string(Monitor): monitor,
	string(Prepend): Prepend.op,
	string(Replace): Replace.op,
	string(Test):    Test.op,
}

var Usage = map[Route]string{
	Add: `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Add a route.

{{flags .}}`,
	Append: `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Change aspects of a route (such as its gateway).

{{flags .}}`,
	Change: `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Change aspects of a route (such as its gateway).

{{flags .}}`,
	Delete: `
usage: {{.Name}} [flags] <destination>
Delete a specific route.

{{flags .}}`,
	Flush: FlushUsage,
	Get: `
usage: {{.Name}} [flags] <destination>
Lookup and display the route for a destination.

{{flags .}}`,
	Monitor: MonitorUsage,
	Prepend: `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Prepend or add new route.

{{flags .}}`,
	Replace: `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Replace or add new route.

{{flags .}}`,
	Test: `
usage: {{.Name}} [flags] <destination> [options] <gateway> [mask]
Verify change or addition.

{{flags .}}`,
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

func (rt Route) defineGWFlags() {
	if rt == Delete || rt == Get {
		return
	}
	xflag.Define(&routeExpire, "expire", "Seconds from now.")
	xflag.Define(&routeHopCount, "hopcount", "FIXME")
	xflag.Define(&routeIface, "iface",
		"Inticates <gateway> is a point-to-point interface name.")
	xflag.Define(&routeIface, "interface", "aka -iface")
	xflag.Define(&routeMetric, "metric", "FIXME")
	xflag.Define(&routeMTU, "mtu", "FIXME")
	xflag.Define(&routeProtocol, "protocol",
		"{boot, kernel, redirect, static}")
	xflag.Define(&routeRTT, "rtt", "FIXME")
	xflag.Define(&routeRTTVar, "rttvar", "FIXME")
	xflag.Define(&routeScope, "scope",
		"{global, nowhere, host, link, site}")
	xflag.Define(&routeSSThresh, "ssthresh", "FIXME")
	xflag.Define(&routeTable, "table", "{compat, default, main, local}")
	xflag.Define(&routeTo, "to ", "{unicast, broadcast, blackhole, etc.}")
	xflag.Define(&routeTOS, "tos", "Set type-of-service.")
}

func (rt Route) req(
	ctx context.Context,
	fib int,
	dst netip.Prefix,
	gw any,
) (netrt.Rt, error) {
	var mx []byte
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Flags = netlink.NLM_F_REQUEST
	switch rt {
	case Add:
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_EXCL
	case Append:
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_APPEND
	case Change:
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE
	case Delete:
		hdr.Type = rtnetlink.RTM_DELROUTE
	case Get:
		hdr.Type = rtnetlink.RTM_GETROUTE
	case Prepend:
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE
	case Replace:
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE
	case Test:
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_EXCL
	default:
		return nil, xerrors.Invalid(rt.String())
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
		rtm.TOS = uint8(routeTOS)
		if v, ok := protocols[routeProtocol]; ok {
			rtm.Protocol = v
		} else {
			return nil, xerrors.Invalid("protocol")
		}
		if v, ok := scopes[routeScope]; ok {
			rtm.Scope = v
		} else {
			return nil, xerrors.Invalid("scope")
		}
		if v, ok := tables[routeTable]; ok {
			rtm.Table = v
		} else {
			return nil, xerrors.Invalid("table")
		}
		if v, ok := types[routeTo]; ok {
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
	if mtu := uint32(routeMTU); mtu != 1500 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_MTU, mtu)
	}
	if routeExpire != 0 {
		expire := uint32(time.Now().Add(routeExpire).Unix())
		req = netlink.CatAttr(req, rtnetlink.RTA_EXPIRES, expire)
	}
	if hc := uint32(routeHopCount); hc != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_HOPLIMIT, hc)
	}
	if metric := uint32(routeMetric); metric != 0 {
		req = netlink.CatAttr(req, rtnetlink.RTA_PRIORITY, metric)
	}
	if t := uint32(routeSSThresh); t != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_SSTHRESH, t)
	}
	if rtt := uint32(routeRTT); rtt != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_RTT, rtt)
	}
	if rttvar := uint32(routeRTTVar); rttvar != 0 {
		mx = netlink.CatAttr(mx, rtnetlink.RTAX_RTTVAR, rttvar)
	}
	if len(mx) > 0 {
		req = netlink.CatBytesAttr(req, rtnetlink.RTA_METRICS, mx)
	}
	netlink.PointerMsgHdr(req).Len = uint32(len(req))
	return netrt.OneRtReq(ctx, req)
}
