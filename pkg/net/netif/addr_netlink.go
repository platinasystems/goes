// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
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

func (nif *Netif) Add(addr, dest netip.Addr, bits int, args []string) (
	err error,
) {
	args, err = nif.addr(netlink.RTM_NEWADDR,
		netlink.NLM_F_CREATE|netlink.NLM_F_EXCL,
		addr, dest, bits, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(args)
	}
	return
}

func (nif *Netif) Change(addr, dest netip.Addr, bits int, args []string) (
	err error,
) {
	args, err = nif.addr(netlink.RTM_NEWADDR, netlink.NLM_F_REPLACE,
		addr, dest, bits, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(args)
	}
	return
}

func (nif *Netif) Del(addr, dest netip.Addr, bits int, args []string) error {
	_, err := nif.addr(netlink.RTM_DELADDR, 0, addr, dest, bits, args)
	return err
}

func (nif *Netif) Replace(addr, dest netip.Addr, bits int, args []string) (
	err error,
) {
	args, err = nif.addr(netlink.RTM_NEWADDR,
		netlink.NLM_F_CREATE|netlink.NLM_F_REPLACE,
		addr, dest, bits, args)
	if err == nil && len(args) > 0 {
		err = nif.Config(args)
	}
	return
}

func (nif *Netif) addr(
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

	req, msg := netlink.Expand[netlink.NlMsghdr](nil)
	req.Type = cmd
	req.Flags = flags | netlink.NLM_F_REQUEST | netlink.NLM_F_ACK
	ifa, msg := netlink.Expand[netlink.IfAddrmsg](msg)
	ifa.Index = uint32(nif.Index)
	ifa.Prefixlen = uint8(bits)

	if addr.Is4() {
		ifa.Family = netlink.AF_INET
	} else if addr.Is6() {
		ifa.Family = netlink.AF_INET6
	} else {
		return args, egress.Markf("%v %w", addr, ErrWrongFamily)
	}

	if dest.IsValid() {
		if ifa.Family == netlink.AF_INET {
			if dest.Is4() {
				ifa.Prefixlen = 32
			} else {
				return args, egress.Markf("%v %w",
					dest, ErrWrongFamily)
			}
		} else if dest.Is6() {
			ifa.Prefixlen = 128
		} else {
			return args, egress.Markf("%v %w", dest, ErrWrongFamily)
		}
		msg = netlink.
			CatBytesAttr(msg, netlink.IFA_ADDRESS, dest.AsSlice())
	}

	msg = netlink.CatBytesAttr(msg, netlink.IFA_LOCAL, addr.AsSlice())

	if err := nl.Request(msg); err != nil {
		return args, egress.Marked(err)
	}
	if err = nl.Wait(req.Seq); err != nil {
		return args, egress.Marked(err)
	}
	if dest.IsValid() && (ifa.Family == netlink.AF_INET && bits < 32) ||
		(ifa.Family == netlink.AF_INET6 && bits < 128) {
		// FIXME add route
	}
	return args, nil
}
