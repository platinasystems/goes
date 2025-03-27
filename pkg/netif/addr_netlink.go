// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"context"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/netlink"
	"github.com/platinasystems/goes/v2/pkg/netlink/ifaddr"
	"github.com/platinasystems/goes/v2/pkg/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet"
)

const AddressCommands = `
  add	Add (default) network address to network interface.
  del	Remove network address to network interface.
  change
	Change address parameters.
  replace
	If necessary, create or just change address parameters.
`

const AddressParameters = ""

func (nif *NetIf) Add(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) (err error) {
	const (
		cmd   = rtnetlink.RTM_NEWADDR
		flags = netlink.NLM_F_CREATE | netlink.NLM_F_EXCL
	)
	args, err = nif.addr(ctx, cmd, flags, addr, dest, bits, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args...)
	}
	return
}

func (nif *NetIf) Change(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) (err error) {
	const (
		cmd   = rtnetlink.RTM_NEWADDR
		flags = netlink.NLM_F_REPLACE
	)
	args, err = nif.addr(ctx, cmd, flags, addr, dest, bits, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args...)
	}
	return
}

func (nif *NetIf) Del(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) error {
	const (
		cmd   = rtnetlink.RTM_DELADDR
		flags = 0
	)
	_, err := nif.addr(ctx, cmd, flags, addr, dest, bits, args)
	return err
}

func (nif *NetIf) Replace(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) (err error) {
	const (
		cmd   = rtnetlink.RTM_NEWADDR
		flags = netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE
	)
	args, err = nif.addr(ctx, cmd, flags, addr, dest, bits, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args...)
	}
	return
}

func (nif *NetIf) addr(
	ctx context.Context,
	cmd, flags uint16,
	addr, dest netip.Addr,
	bits int,
	args []string,
) ([]string, error) {
	nl, err := netlink.Open()
	if err != nil {
		return args, err
	}
	defer nl.Close()

	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = cmd
	hdr.Flags = flags | netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifa, req := netlink.ExpandIfAddrMsg(req)
	ifa.Index = uint32(nif.Index)
	ifa.PrefixLen = uint8(bits)

	if addr.Is4() {
		ifa.Family = xnet.AF_INET
	} else if addr.Is6() {
		ifa.Family = xnet.AF_INET6
	} else {
		return args, xerrors.Label(ErrWrongFamily, addr.String())
	}

	if dest.IsValid() {
		if ifa.Family == xnet.AF_INET {
			if dest.Is4() {
				ifa.PrefixLen = 32
			} else {
				return args, xerrors.Markf("%v %w",
					dest, ErrWrongFamily)
			}
		} else if dest.Is6() {
			ifa.PrefixLen = 128
		} else {
			return args, xerrors.
				Label(ErrWrongFamily, dest.String())
		}
		req = netlink.
			CatBytesAttr(req, ifaddr.IFA_ADDRESS, dest.AsSlice())
	}

	req = netlink.CatBytesAttr(req, ifaddr.IFA_LOCAL, addr.AsSlice())

	if err = nl.Request(req); err != nil {
		return args, xerrors.Mark(err)
	}
	if err = nl.Wait(ctx, hdr.SEQ); err != nil {
		return args, xerrors.Mark(err)
	}
	return args, nil
}
