// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"context"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/af"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/ifaddr"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
)

const AddressCommands = `
  add	Add (default) network address to network interface.
  del	Remove network address to network interface.
  change
  	Change <prefix> parameters.
  replace
	If necessary, create or just change <prefix> parameters.
`

const AddressParameters = ""

func (nif *NetIf) Add(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) (err error) {
	const (
		cmd   = rtnetlink.RTM_NEWADDR
		flags = netlink.NLM_F_CREATE | netlink.NLM_F_EXCL
	)
	args, err = nif.addr(ctx, cmd, flags, prefix, dest, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args)
	}
	return
}

func (nif *NetIf) Change(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) (err error) {
	const (
		cmd   = rtnetlink.RTM_NEWADDR
		flags = netlink.NLM_F_REPLACE
	)
	args, err = nif.addr(ctx, cmd, flags, prefix, dest, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args)
	}
	return
}

func (nif *NetIf) Del(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) error {
	const (
		cmd   = rtnetlink.RTM_DELADDR
		flags = 0
	)
	_, err := nif.addr(ctx, cmd, flags, prefix, dest, args)
	return err
}

func (nif *NetIf) Replace(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) (err error) {
	const (
		cmd   = rtnetlink.RTM_NEWADDR
		flags = netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE
	)
	args, err = nif.addr(ctx, cmd, flags, prefix, dest, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args)
	}
	return
}

func (nif *NetIf) addr(
	ctx context.Context,
	cmd, flags uint16,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) ([]string, error) {
	nl, err := netlink.Open()
	if err != nil {
		return args, err
	}
	defer nl.Close()

	addr, bits := prefix.Addr(), prefix.Bits()

	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = cmd
	hdr.Flags = flags | netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifa, req := netlink.ExpandIfAddrMsg(req)
	ifa.Index = uint32(nif.Index)
	ifa.PrefixLen = uint8(bits)

	if addr.Is4() {
		ifa.Family = af.INET
	} else if addr.Is6() {
		ifa.Family = af.INET6
	} else {
		return args, egress.Markf("%v %w", prefix, ErrWrongFamily)
	}

	if dest.IsValid() {
		if ifa.Family == af.INET {
			if dest.Is4() {
				ifa.PrefixLen = 32
			} else {
				return args, egress.Markf("%v %w",
					dest, ErrWrongFamily)
			}
		} else if dest.Is6() {
			ifa.PrefixLen = 128
		} else {
			return args, egress.Markf("%v %w", dest, ErrWrongFamily)
		}
		req = netlink.
			CatBytesAttr(req, ifaddr.IFA_ADDRESS, dest.AsSlice())
	}

	req = netlink.CatBytesAttr(req, ifaddr.IFA_LOCAL, addr.AsSlice())

	if err = nl.Request(req); err != nil {
		return args, egress.Mark(err)
	}
	if err = nl.Wait(ctx, hdr.SEQ); err != nil {
		return args, egress.Mark(err)
	}
	return args, nil
}
