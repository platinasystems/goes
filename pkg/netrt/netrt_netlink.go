// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netrt

import (
	"context"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type netrts struct{ nl }

type netrt struct {
	bits,
	index,
	line,
	expire,
	state int
	flags uint
	dst   netip.Addr
	ifa   netip.Addr
	gw    netip.Addr
	ha    net.HardwareAddr
}

func (nrt *netrt) Bits() int            { return nrt.bits }
func (nrt *netrt) Index() int           { return nrt.index }
func (nrt *netrt) Line() int            { return nrt.line }
func (nrt *netrt) Flags() uint          { return nrt.flags }
func (nrt *netrt) Dst() netip.Addr      { return nrt.dst }
func (nrt *netrt) IFA() netip.Addr      { return nrt.ifa }
func (nrt *netrt) GW() netip.Addr       { return nrt.gw }
func (nrt *netrt) HA() net.HardwareAddr { return nrt.ha }
func (nrt *netrt) Expire() int          { return nrt.expire }
func (nrt *netrt) State() int           { return nrt.state }

func Routes(ctx context.Context, family int) (NextRtCloser, error) {
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETROUTE
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	gen, req := netlink.ExpandRtGenMsg(req)
	integer.Assign(&gen.Family, family)
	sock, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	if err = sock.Request(req); err != nil {
		return nil, err
	}
	return netrts{nl{hdr.SEQ, sock}}, nil
}

func (rts netrts) NextRt(ctx context.Context) (Rt, error) {
	for {
		rsp, data, err := rts.sock.Next(ctx)
		if err != nil {
			return nil, err
		} else if rsp.SEQ != rts.seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			e, _ := netlink.ExtractMsgErr(data)
			return nil, e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWROUTE {
			continue
		}
		rtm, data := netlink.ExtractRtMsg(data)
		if rtm.Family != xnet.AF_INET && rtm.Family != xnet.AF_INET6 {
			continue
		}
		rt := new(netrt)
		integer.Assign(&rt.bits, rtm.DstLen)
		for netlink.HasAttr(data) {
			t, v, datá := netlink.ExtractAttr(data)
			data = datá
			switch rtnetlink.Rta(t) {
			case rtnetlink.RTA_UNSPEC:
			case rtnetlink.RTA_DST:
				rt.dst, _ = netlink.IP(rtm.Family, v)
			case rtnetlink.RTA_SRC:
			case rtnetlink.RTA_IIF:
			case rtnetlink.RTA_OIF:
				integer.Assign(&rt.index,
					*(netlink.Pointer[int32](v)))
				integer.Assign(&rt.line, rt.index)
			case rtnetlink.RTA_GATEWAY:
				rt.gw, _ = netlink.IP(rtm.Family, v)
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
				rt.gw, _ = netlink.Via(v)
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
		return rt, nil
	}
	return nil, nil
}

func OneRtReq(ctx context.Context, req []byte) (Rt, error) {
	sock, err := xerrors.MarkResult(netlink.Open())
	if err != nil {
		return nil, err
	}
	defer sock.Close()
	seq, err := xerrors.MarkResult(Req(sock, req))
	if err != nil {
		return nil, err
	}
	rts := netrts{nl{seq, sock}}
	return xerrors.MarkResult(rts.NextRt(ctx))
}

// If successful, this returns the request sequence number.
func Req(sock *netlink.NL, req []byte) (uint32, error) {
	hdr := netlink.PointerMsgHdr(req)
	err := sock.Request(req)
	return hdr.SEQ, err
}

func GatewayAttr(dst, gw netip.Addr) (attr uint16) {
	attr = rtnetlink.RTA_GATEWAY
	if dst.Is4() {
		if gw.Is6() {
			attr = rtnetlink.RTA_VIA
		}
	} else if gw.Is4() {
		attr = rtnetlink.RTA_VIA
	}
	return
}
