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
	"github.com/platinasystems/goes/v2/pkg/net/netif"
	"github.com/platinasystems/goes/v2/pkg/net/netrt"
	"github.com/platinasystems/goes/v2/pkg/net/sockaddr"
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
	msg = sockaddr.Append(msg, dst.Addr())
	rtm = netrt.PointerRtMsghdr(msg)
	integer.Set(&rtm.Addrs, 1<<unix.RTAX_DST)
	switch t := gw.(type) {
	case netip.Addr:
		msg = sockaddr.Append(msg, t)
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
		integer.Set(&rtm.Flags, unix.RTF_GATEWAY)
	case *netif.NetIf:
		msg = sockaddr.AppendDl(msg,
			uint16(t.Index),
			uint8(t.Type),
			t.Name,
			t.HardwareAddr,
			[]byte{})
		rtm = netrt.PointerRtMsghdr(msg)
		integer.Set(&rtm.Addrs, 1<<unix.RTAX_GATEWAY)
	default:
		return fmt.Errorf("<gateway>: %T %w", gw, ErrInvalid)
	}
	mask := net.CIDRMask(dst.Bits(), dst.Addr().BitLen())
	maskaddr, ok := netip.AddrFromSlice(mask)
	if !ok {
		return fmt.Errorf("%w mask or prefixlen", ErrInvalid)
	}
	msg = sockaddr.Append(msg, maskaddr)
	rtm = netrt.PointerRtMsghdr(msg)
	integer.Set(&rtm.Addrs, 1<<unix.RTAX_NETMASK)
	_, err := netrt.Request(msg, -1)
	return err
}
