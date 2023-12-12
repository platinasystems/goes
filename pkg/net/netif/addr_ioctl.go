// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"context"
	"fmt"
	"net"
	"net/netip"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netioctl"
)

const ND6_INFINITE_LIFETIME = 0xffffffff

const AddressCommands = `
  add	Add (default) network address to network interface. (aka. alias)
  del	Remove network address to network interface. (aka. -alias)
`

const AddressParameters = ""

func (nif *NetIf) Add(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) (err error) {
	if addr := prefix.Addr(); addr.Is4() {
		args, err = nif.add4(ctx, prefix, dest, args)
	} else if addr.Is6() {
		args, err = nif.add6(ctx, prefix, dest, args)
	} else {
		err = fmt.Errorf("%v %w", prefix, ErrWrongFamily)
	}
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args)
	}
	return
}

func (nif *NetIf) add4(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) ([]string, error) {
	bits := prefix.Bits()
	req := netioctl.NewIfReqInAlias(nif.Name)
	req.Value.Addr.Write(prefix.Addr().AsSlice())
	if dest.IsValid() {
		if !dest.Is4() {
			return args, fmt.Errorf("%v %w", dest, ErrWrongFamily)
		}
		if bits != 32 {
			return args, fmt.Errorf("%d %w bits", bits, ErrInvalid)
		}
		req.Value.Dest.Write(dest.AsSlice())
	}
	req.Value.Mask.Write(net.CIDRMask(prefix.Bits(), bits))
	return args, netioctl.Inet(syscall.SIOCAIFADDR, req)
}

func (nif *NetIf) add6(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) ([]string, error) {
	bits := prefix.Bits()
	req := netioctl.NewIfReqIn6Alias(nif.Name)
	req.Value.Addr.Write(prefix.Addr().AsSlice())
	if dest.IsValid() {
		if !dest.Is6() {
			return args, fmt.Errorf("%v %w", dest, ErrWrongFamily)
		}
		req.Value.Dest.Write(dest.AsSlice())
		if bits != 128 {
			return args, fmt.Errorf("%d %w bits", bits, ErrInvalid)
		}
	}
	req.Value.Mask.Write(net.CIDRMask(prefix.Bits(), 128))
	req.Value.Lifetime.Vltime = ND6_INFINITE_LIFETIME
	req.Value.Lifetime.Pltime = ND6_INFINITE_LIFETIME
	return args, netioctl.Inet6(netioctl.SIOCAIFADDR_IN6, req)
}

func (nif *NetIf) Change(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) error {
	return ErrUnsupported
}

func (nif *NetIf) Del(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) error {
	addr := prefix.Addr()
	if addr.Is4() {
		req := netioctl.NewIfReqSockaddrIn(nif.Name)
		req.Value.Write(prefix.Addr().AsSlice())
		return egress.Mark(netioctl.Inet(syscall.SIOCDIFADDR, req))
	}
	req := netioctl.NewIfReqSockaddrIn6(nif.Name)
	req.Value.Write(addr.AsSlice())
	return egress.Mark(netioctl.Inet6(netioctl.SIOCDIFADDR_IN6, req))
}

func (nif *NetIf) Replace(
	ctx context.Context,
	prefix netip.Prefix,
	dest netip.Addr,
	args []string,
) error {
	return ErrUnsupported
}
