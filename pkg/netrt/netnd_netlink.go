// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netrt

import (
	"bytes"
	"context"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

type netnds struct{ nl }

func Neighbors(ctx context.Context, family int) (NextNdCloser, error) {
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETNEIGH
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
	return netnds{nl{hdr.SEQ, sock}}, nil
}

type netnd struct {
	dst,
	gw netip.Addr
	ha net.HardwareAddr
	expire,
	index,
	line,
	state int
	flags,
	probes,
	tipe uint
	cacheinfo []uint
}

func (nd *netnd) Dst() netip.Addr      { return nd.dst }
func (nd *netnd) GW() netip.Addr       { return nd.gw }
func (nd *netnd) HA() net.HardwareAddr { return nd.ha }
func (nd *netnd) Expire() int          { return nd.expire }
func (nd *netnd) Index() int           { return nd.index }
func (nd *netnd) Line() int            { return nd.line }
func (nd *netnd) Flags() uint          { return nd.flags }
func (nd *netnd) State() int           { return nd.state }
func (nd *netnd) Type() uint           { return nd.tipe }
func (nd *netnd) CacheInfo() []uint    { return nd.cacheinfo }

func (nds netnds) NextNd(ctx context.Context) (Nd, error) {
	for {
		rsp, data, err := nds.sock.Next(ctx)
		if err != nil {
			return nil, err
		} else if rsp.SEQ != nds.seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			e, _ := netlink.ExtractMsgErr(data)
			return nil, e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWNEIGH {
			continue
		}
		ndm, data := netlink.ExtractNdMsg(data)
		if ndm.Family != xnet.AF_INET && ndm.Family != xnet.AF_INET6 {
			continue
		}
		nd := new(netnd)
		integer.Assign(&nd.index, ndm.IfIndex)
		integer.Assign(&nd.state, ndm.State)
		integer.Assign(&nd.flags, ndm.Flags)
		integer.Assign(&nd.tipe, ndm.Type)
		for netlink.HasAttr(data) {
			t, v, d := netlink.ExtractAttr(data)
			data = d
			switch rtnetlink.Rta(t) {
			case rtnetlink.NDA_DST:
				nd.dst, _ = netlink.IP(ndm.Family, v)
			case rtnetlink.NDA_LLADDR:
				nd.ha = net.HardwareAddr(bytes.Clone(v))
			case rtnetlink.NDA_CACHEINFO:
				for ; len(v) >= 4; v = v[4:] {
					u := uint(*(netlink.Pointer[uint32](v)))
					nd.cacheinfo = append(nd.cacheinfo, u)
				}
			case rtnetlink.NDA_PROBES:
				integer.Assign(&nd.probes,
					*(netlink.Pointer[uint32](v)))
			case rtnetlink.NDA_VLAN:
			case rtnetlink.NDA_PORT:
			case rtnetlink.NDA_VNI:
			case rtnetlink.NDA_IFINDEX:
				integer.Assign(&nd.line,
					*(netlink.Pointer[int32](v)))
			case rtnetlink.NDA_MASTER:
			case rtnetlink.NDA_LINK_NETNSID:
			}
		}
		return nd, nil
	}
	return nil, nil
}
