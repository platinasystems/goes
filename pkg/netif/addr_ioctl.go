// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux && !darwin

package netif

import (
	"context"
	"fmt"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/netioctl"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xlog"
	"golang.org/x/sys/unix"
)

const ND6_INFINITE_LIFETIME = 0xffffffff

const AddressCommands = `
  add	Add (default) network address to network interface. (aka. alias)
  del	Remove network address to network interface. (aka. -alias)
`

const AddressParameters = ""

func (nif *NetIf) Add(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) (err error) {
	if addr := prefix.Addr(); addr.Is4() {
		args, err = nif.add4(ctx, prefix, dest, args...)
	} else if addr.Is6() {
		args, err = nif.add6(ctx, prefix, dest, args...)
	} else {
		err = fmt.Errorf("%v %w", prefix, ErrWrongFamily)
	}
	if err == nil && len(args) > 0 {
		err = nif.Config(ctx, args...)
	}
	return
}

func (nif *NetIf) add4(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) ([]string, error) {
	req := netioctl.NewIfReqInAlias(nif.Name)
	req.Value.Addr.Write(addr.AsSlice())
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
	return args, netioctl.Inet(unix.SIOCAIFADDR, req)
}

func (nif *NetIf) add6(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) ([]string, error) {
	req := netioctl.NewIfReqIn6Alias(nif.Name)
	req.Value.Addr.Write(addr.AsSlice())
	if addr.Is6() && bits == 128 {
		req.Value.Addr.Scope_id = uint32(nif.Index)
		xlog.Info.Printf("adding %v/128%%%s", addr, nif.Name)
	} else if dest.IsValid() {
		if !dest.Is6() {
			return args, fmt.Errorf("%v %w", dest, ErrWrongFamily)
		}
		req.Value.Dest.Write(dest.AsSlice())
		if bits != 128 {
			return args, fmt.Errorf("%d %w bits", bits, ErrInvalid)
		}
		xlog.Info.Printf("adding %v/%d via %v", addr, bits, dest)
	}
	req.Value.Mask.Write(net.CIDRMask(bits, 128))
	req.Value.Lifetime.Vltime = ND6_INFINITE_LIFETIME
	req.Value.Lifetime.Pltime = ND6_INFINITE_LIFETIME
	return args, netioctl.Inet6(netioctl.SIOCAIFADDR_IN6, req)
}

func (nif *NetIf) Change(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) error {
	return ErrUnsupported
}

func (nif *NetIf) Del(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) error {
	if addr.Is4() {
		req := netioctl.NewIfReqSockaddrIn(nif.Name)
		req.Value.Write(addr.AsSlice())
		return xerrors.Mark(netioctl.Inet(unix.SIOCDIFADDR, req))
	}
	req := netioctl.NewIfReqSockaddrIn6(nif.Name)
	req.Value.Write(addr.AsSlice())
	return xerrors.Mark(netioctl.Inet6(netioctl.SIOCDIFADDR_IN6, req))
}

func (nif *NetIf) Replace(
	ctx context.Context,
	addr, dest netip.Addr,
	bits int,
	args ...string,
) error {
	return ErrUnsupported
}
