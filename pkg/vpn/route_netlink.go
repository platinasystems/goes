// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package vpn

import (
	"context"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

func route(
	ctx context.Context,
	add bool,
	dst netip.Prefix,
	gw any,
) error {
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	if add {
		hdr.Type = rtnetlink.RTM_NEWROUTE
		hdr.Flags |= netlink.NLM_F_CREATE | netlink.NLM_F_EXCL
	} else {
		hdr.Type = rtnetlink.RTM_DELROUTE
	}
	rtm, req := netlink.ExpandRtMsg(req)
	if dst.Addr().Is4() {
		rtm.Family = xnet.AF_INET
	} else {
		rtm.Family = xnet.AF_INET6
	}
	rtm.DstLen = uint8(dst.Bits())
	if add {
		rtm.Protocol = rtnetlink.RTPROT_STATIC
		rtm.Scope = rtnetlink.RT_SCOPE_UNIVERSE
		rtm.Table = rtnetlink.RT_TABLE_MAIN
		rtm.Type = rtnetlink.RTN_UNICAST
	}
	req = netlink.CatBytesAttr(req, rtnetlink.RTA_DST, dst.Addr().AsSlice())
	if add {
		switch t := gw.(type) {
		case netip.Addr:
			gwattr := netrt.GatewayAttr(dst.Addr(), t)
			req = netlink.CatBytesAttr(req, gwattr, t.AsSlice())
		case []net.IPAddr:
			ipa := netrt.SelectGateway(t, dst.Addr().Is6())
			gwaddr, ok := netip.AddrFromSlice(ipa.IP)
			if !ok {
				return xerrors.Invalid("resolved", ipa.String())
			}
			gwattr := netrt.GatewayAttr(dst.Addr(), gwaddr)
			gwip := gwaddr.AsSlice()
			req = netlink.CatBytesAttr(req, gwattr, gwip)
		case *netif.NetIf:
			gwi := uint32(t.Index)
			req = netlink.CatAttr(req, rtnetlink.RTA_IIF, gwi)
			req = netlink.CatAttr(req, rtnetlink.RTA_OIF, gwi)
		default:
			return xerrors.Invalid("gateway")
		}
	}
	netlink.PointerMsgHdr(req).Len = uint32(len(req))
	ack, err := xerrors.MarkResult(netrt.OneReq(ctx, req))
	_ = ack
	return err

}
