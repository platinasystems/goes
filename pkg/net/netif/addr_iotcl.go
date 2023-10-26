// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"fmt"
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

const AddressCommands = `
  add	Add (default) network address to network interface. (aka. alias)
  del	Remove network address to network interface. (aka. -alias)
`

const AddressParameters = ""

func (nif *Netif) Add(addr, dest netip.Addr, bits int, args []string) (
	err error,
) {
	if addr.Is4() {
		args, err = nif.add4(addr, dest, bits, args)
	} else if addr.Is6() {
		args, err = nif.add6(addr, dest, bits, args)
	} else {
		err = fmt.Errorf("%v %w", addr, ErrWrongFamily)
	}
	if err == nil && len(args) > 0 {
		err = nif.Config(args)
	}
	return
}

func (nif *Netif) add4(addr, dest netip.Addr, bits int, args []string) (
	[]string, error,
) {
	req := NewIfreq[InAlias](nif.Name)
	req.Value.Addr.Set(addr.AsSlice())
	if dest.IsValid() {
		if !dest.Is4() {
			return args, fmt.Errorf("%v %w", dest, ErrWrongFamily)
		}
		req.Value.Dest.Set(dest.AsSlice())
		bits = 32
	}
	req.Value.Mask.Set(net.CIDRMask(bits, 32))
	return args, InetIOCTL(SIOCAIFADDR, req)
}

func (nif *Netif) add6(addr, dest netip.Addr, bits int, args []string) (
	[]string, error,
) {
	req := NewIfreq[In6Alias](nif.Name)
	req.Value.Addr.Set(addr.AsSlice())
	if dest.IsValid() {
		if !dest.Is6() {
			return args, fmt.Errorf("%v %w", dest, ErrWrongFamily)
		}
		req.Value.Dest.Set(dest.AsSlice())
		bits = 128
	}
	req.Value.Mask.Set(net.CIDRMask(bits, 128))
	req.Value.Lifetime.Vltime = ND6_INFINITE_LIFETIME
	req.Value.Lifetime.Pltime = ND6_INFINITE_LIFETIME
	return args, Inet6IOCTL(SIOCAIFADDR_IN6, req)
}

func (nif *Netif) Change(addr, dest netip.Addr, bits int, args []string) error {
	return ErrUnsupported
}

func (nif *Netif) Del(addr, dest netip.Addr, bits int, args []string) error {
	if addr.Is4() {
		req := NewIfreq[SockaddrIn](nif.Name)
		req.Value.Set(addr.AsSlice())
		return egress.Marked(InetIOCTL(SIOCDIFADDR, req))
	}
	req := NewIfreq[SockaddrIn6](nif.Name)
	req.Value.Set(addr.AsSlice())
	return egress.Marked(Inet6IOCTL(SIOCDIFADDR_IN6, req))
}

func (nif *Netif) Replace(addr, dest netip.Addr, bits int, args []string) error {
	return ErrUnsupported
}
