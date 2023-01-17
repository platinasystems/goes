// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"net"
	"net/netip"
	"syscall"
)

func (sock Inet6) Add(ifname string, p netip.Prefix) error {
	return sock.mod(ifname, p, syscall.SIOCSIFADDR)
}

func (sock Inet6) Del(ifname string, p netip.Prefix) error {
	return sock.mod(ifname, p, syscall.SIOCDIFADDR)
}

func (sock Inet6) mod(ifname string, p netip.Prefix, req uintptr) error {
	var ifr In6Ifreq
	if !p.IsValid() {
		return ErrInvalidPrefix
	}
	if a := p.Addr(); !a.Is6() {
		return ErrInvalidAddr
	} else {
		a16 := a.As16()
		copy(ifr.Addr.U[:], a16[:])
	}
	if bits := p.Bits(); bits < 0 {
		return ErrInvalidLength
	} else {
		ifr.Prefixlen = uint32(bits)
	}
	if itf, err := net.InterfaceByName(ifname); err == nil {
		ifr.Ifindex = int32(itf.Index)
	} else {
		return err
	}
	return sock.ioctl(req, ifr.Pointer())
}
