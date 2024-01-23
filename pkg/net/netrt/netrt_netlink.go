// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netrt

import (
	"context"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/net/af"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
)

type netrt struct {
	index int32
	line  int32
	flags int32
	bits  int
	dst   netip.Addr
	ifa   netip.Addr
	gw    netip.Addr
	ha    net.HardwareAddr
}

func (nrt *netrt) Bits() int            { return nrt.bits }
func (nrt *netrt) Index() int           { return int(nrt.index) }
func (nrt *netrt) Line() int            { return int(nrt.line) }
func (nrt *netrt) Flags() uint          { return uint(nrt.flags) }
func (nrt *netrt) Dst() netip.Addr      { return nrt.dst }
func (nrt *netrt) IFA() netip.Addr      { return nrt.ifa }
func (nrt *netrt) GW() netip.Addr       { return nrt.gw }
func (nrt *netrt) HA() net.HardwareAddr { return nrt.ha }

type stream struct {
	nl  *netlink.NL
	seq uint32
}

func NewList(ctx context.Context) (Streamer, error) {
	var family uint8
	if goes.SearchContextFlags[bool](ctx, "4") {
		family = af.INET
	} else if goes.SearchContextFlags[bool](ctx, "6") {
		family = af.INET6
	} else {
		family = af.UNSPEC
	}
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETROUTE
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	gen, req := netlink.ExpandRtGenMsg(req)
	gen.Family = family
	nl, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	if err = nl.Request(req); err != nil {
		return nil, err
	}
	return &stream{nl, hdr.SEQ}, nil
}

func (strm *stream) Close() error {
	return strm.nl.Close()
}

func (strm *stream) Next(ctx context.Context) (NetRt, error) {
	for {
		rsp, data, err := strm.nl.Next(ctx)
		if err != nil {
			return nil, err
		} else if rsp.SEQ != strm.seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			e, _ := netlink.ExtractMsgErr(data)
			return nil, e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWROUTE {
			continue
		}
		m, data := netlink.ExtractRtMsg(data)
		if m.Family != af.INET && m.Family != af.INET6 {
			continue
		}
		nrt := &netrt{
			bits: int(m.DstLen),
		}
		for netlink.HasAttr(data) {
			t, v, datá := netlink.ExtractAttr(data)
			data = datá
			switch rtnetlink.Rta(t) {
			case rtnetlink.RTA_UNSPEC:
			case rtnetlink.RTA_DST:
				nrt.dst, _ = netlink.IP(m.Family, v)
			case rtnetlink.RTA_SRC:
			case rtnetlink.RTA_IIF:
			case rtnetlink.RTA_OIF:
				nrt.index = *(netlink.Pointer[int32](v))
				nrt.line = nrt.index
			case rtnetlink.RTA_GATEWAY:
				nrt.gw, _ = netlink.IP(m.Family, v)
			case rtnetlink.RTA_PRIORITY:
			case rtnetlink.RTA_PREFSRC:
			case rtnetlink.RTA_METRICS:
			case rtnetlink.RTA_MULTIPATH:
				// FIXME
			case rtnetlink.RTA_PROTOINFO:
			case rtnetlink.RTA_FLOW:
			case rtnetlink.RTA_CACHEINFO:
			case rtnetlink.RTA_SESSION:
			case rtnetlink.RTA_MP_ALGO:
			case rtnetlink.RTA_TABLE:
			case rtnetlink.RTA_MARK:
			case rtnetlink.RTA_MFC_STATS:
			case rtnetlink.RTA_VIA:
				nrt.gw, _ = netlink.Via(v)
			case rtnetlink.RTA_NEWDST:
			case rtnetlink.RTA_PREF:
			case rtnetlink.RTA_ENCAP_TYPE:
			case rtnetlink.RTA_ENCAP:
			case rtnetlink.RTA_EXPIRES:
			case rtnetlink.RTA_PAD:
			case rtnetlink.RTA_UID:
			case rtnetlink.RTA_TTL_PROPAGATE:
			case rtnetlink.RTA_IP_PROTO:
			case rtnetlink.RTA_SPORT:
			case rtnetlink.RTA_DPORT:
			case rtnetlink.RTA_NH_ID:
			}
		}
		return nrt, nil
	}
	return nil, nil
}

func OneReq(ctx context.Context, req []byte) (NetRt, error) {
	nl, err := egress.MarkResult(netlink.Open())
	if err != nil {
		return nil, err
	}
	defer nl.Close()
	seq, err := egress.MarkResult(Req(nl, req))
	if err != nil {
		return nil, err
	}
	strm := stream{nl, seq}
	return egress.MarkResult(strm.Next(ctx))
}

// If successful, this returns the request sequence number.
func Req(nl *netlink.NL, req []byte) (uint32, error) {
	hdr := netlink.Pointer[netlink.MsgHdr](req)
	err := nl.Request(req)
	return hdr.SEQ, err
}
