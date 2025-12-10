// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package route

import (
	"context"
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

var Features = map[string]any{
	string(Add):     Add.op,
	string(Change):  Change.op,
	string(Delete):  Delete.op,
	string(Flush):   flush,
	string(Get):     Get.op,
	string(Monitor): monitor,
}

var Usage = map[Route]string{
	Add: `
usage: {{.Name}} [flags] <destination> <gateway> [mask]
Add a route.

{{flags .}}`,
	Change: `
usage: {{.Name}} [flags] <destination> <gateway> [mask]
Change aspects of a route (such as its gateway).

{{flags .}}`,
	Delete: `
usage: {{.Name}} [flags] <destination>
Delete specific route.

{{flags .}}`,
	Flush: FlushUsage,
	Get: `
usage: {{.Name}} [flags] <destination>
Lookup and display the route for a destination.

{{flags .}}`,
	Monitor: MonitorUsage,
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

var GatewayFlags = xflag.Labels{
	{"flags", "A comma separated list." + gwFlags, &RouteFlags},
	{"genmask", "Generate netmask.", &RouteGenMask},
	{"ifa", "A MAC address of a point-to-point peer?", &RouteIfa},
	{"iface", "Inticates <gateway> is point-to-point interface.",
		&RouteIface},
	{"interface", "aka -iface", &RouteIface},
	{"ifp", "A point-to-point peer interface and MAC.", &RouteIfp},
	{"metrics", "A comma separated NAME=VALUE." + gwMetrics,
		&RouteMetrics},
}

func (rt Route) req(
	ctx context.Context,
	fib int,
	dst netip.Prefix,
	gw any,
) (netrt.Rt, error) {
	msg := netrt.NewRtMsg()
	rtm := netrt.PointerRtMsghdr2(msg)
	switch rt {
	case Add, Change:
		if gw == nil {
			return nil, xerrors.Incomplete("gateway")
		}
		switch rt {
		case Add:
			rtm.Type = unix.RTM_ADD
		case Change:
			rtm.Type = unix.RTM_CHANGE
		}
		integer.Set(&rtm.Flags, unix.RTF_UP)
		if dst.Bits() == dst.Addr().BitLen() {
			integer.Set(&rtm.Flags, unix.RTF_HOST)
		}
		ff := strings.Split(RouteFlags, ",")
		for _, name := range ff {
			if val, ok := gwFlagValues[name]; ok {
				integer.Set(&rtm.Flags, val)
			} else if val, ok = gwFlagValues["no"+name]; ok {
				integer.Reset(&rtm.Flags, val)
			}
		}
		metrics := strings.Split(RouteMetrics, ",")
		rtmmetrics(rtm, metrics)
	case Delete:
		rtm.Type = unix.RTM_DELETE
		rtmdel(rtm)
	case Get:
		rtm.Type = unix.RTM_GET
	default:
		return nil, xerrors.Invalid("command", string(rt))
	}
	msg = xnet.SAAppend(msg, dst.Addr())
	rtm = netrt.PointerRtMsghdr2(msg)
	integer.Set(&rtm.Addrs, 1<<unix.RTAX_DST)
	switch t := gw.(type) {
	case nil:
	case netip.Addr:
		msg = xnet.SAAppend(msg, t)
		rtm = netrt.PointerRtMsghdr2(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case []net.IPAddr:
		ipa := netrt.SelectGateway(t, dst.Addr().Is6())
		gwaddr, ok := netip.AddrFromSlice(ipa.IP)
		if !ok {
			return nil, xerrors.Invalid("resolved", ipa.String())
		}
		msg = xnet.SAAppend(msg, gwaddr)
		rtm = netrt.PointerRtMsghdr2(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case *netif.NetIf:
		msg = xnet.SAAppendDataLink(msg,
			uint16(t.Index),
			uint8(t.Type),
			t.Name,
			t.HardwareAddr,
			[]byte{})
		rtm = netrt.PointerRtMsghdr2(msg)
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
		rtm = netrt.PointerRtMsghdr2(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_NETMASK)
	}
	if RouteGenMask {
		// FIXME e.g. 255.255.255.255 ?
	}
	if len(RouteIfp) > 0 {
		// FIXME e.g. eth0:1.2.3.4.5.6 ?
	}
	if len(RouteIfa) > 0 {
		// FIXME e.g. 1.2.3.4.5.6 ?
	}
	return netrt.Request(msg, fib)
}
