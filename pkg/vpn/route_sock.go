// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package vpn

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/integer"
	"github.com/platinasystems/goes/v2/pkg/netif"
	"github.com/platinasystems/goes/v2/pkg/netrt"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
	"golang.org/x/sys/unix"
)

func route(
	ctx context.Context,
	add bool,
	dst netip.Prefix,
	gw any,
) error {
	msg := netrt.NewRtMsg()
	rtm := netrt.PointerRtMsghdr(msg)
	if add {
		rtm.Type = unix.RTM_ADD
		integer.Set(&rtm.Flags, unix.RTF_UP)
		integer.Set(&rtm.Flags, unix.RTF_STATIC)
	} else {
		rtm.Type = unix.RTM_DELETE
		integer.Set(&rtm.Flags, RTF_PINNED)
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
		return xerrors.Invalid("gateway", fmt.Sprintf("%T", gw))
	}
	mask := net.CIDRMask(dst.Bits(), dst.Addr().BitLen())
	maskaddr, ok := netip.AddrFromSlice(mask)
	if !ok {
		return xerrors.Invalid("mask")
	}
	msg = xnet.SAAppend(msg, maskaddr)
	rtm = netrt.PointerRtMsghdr(msg)
	integer.Set(&rtm.Addrs, 1<<unix.RTAX_NETMASK)
	_, err := netrt.Request(msg, -1)
	return err
}
