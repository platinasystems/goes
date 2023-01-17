// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import (
	"fmt"
	"net"
	"net/netip"
	"syscall"
)

func (sock Inet) Add(ifname string, p netip.Prefix, bc netip.Addr) error {
	var m4 [4]byte
	copy(m4[:], net.CIDRMask(p.Bits(), 32)[:])
	a := p.Addr()
	m := netip.AddrFrom4(m4)
	err := sock.setaddr(ifname, syscall.SIOCSIFADDR, a)
	if err != nil {
		return fmt.Errorf("<address>(%v) %w", a, err)
	}
	err = sock.setaddr(ifname, syscall.SIOCSIFNETMASK, m)
	if err != nil {
		return fmt.Errorf("<netmask>(%v) %w", m, err)
	}
	if bc.IsUnspecified() {
		if p.Bits() == 32 {
			bc = a
		} else {
			return nil
		}
	}
	err = sock.setaddr(ifname, syscall.SIOCSIFBRDADDR, bc)
	if err != nil {
		return fmt.Errorf("<broadcast>(%v) %w", bc, err)
	}
	return nil
}

func (sock Inet) Del(ifname string, p netip.Prefix) error {
	pa := p.Addr()
	a, err := sock.getaddr(ifname, syscall.SIOCGIFADDR)
	if err != nil {
		return fmt.Errorf("addr: %w", err)
	}
	if pa.Compare(a) != 0 {
		return fmt.Errorf("addr: %v != %v", pa, a)
	}
	_, err = sock.getaddr(ifname, syscall.SIOCGIFNETMASK)
	if err != nil {
		return fmt.Errorf("netmask: %w", err)
	}
	_, err = sock.getaddr(ifname, syscall.SIOCGIFBRDADDR)
	if err != nil {
		return fmt.Errorf("broadcast: %w", err)
	}
	if err = sock.zaddr(ifname, syscall.SIOCSIFADDR); err != nil {
		return fmt.Errorf("addr(%v): %w", a, err)
	}
	return nil
}

func (sock Inet) getaddr(ifname string, req uintptr) (a netip.Addr, err error) {
	var ifr Ifreq
	ifr.Rename(ifname)
	if err = sock.ioctl(req, ifr.Pointer()); err == nil {
		var ifa SockaddrIn
		ifa.Write(ifr.Ifru[:])
		if af := ifa.Family.Value(); af == syscall.AF_INET {
			a = ifa.Addr.Value()
		} else {
			err = fmt.Errorf("af(%#x) != AF_INET(%#x)",
				af, syscall.AF_INET)
		}
	}
	return
}

func (sock Inet) setaddr(ifname string, req uintptr, a netip.Addr) error {
	var ifr Ifreq
	var ifa SockaddrIn
	ifr.Rename(ifname)
	ifa.Write(ifr.Ifru[:])
	ifa.Family.Put(syscall.AF_INET)
	ifa.Addr.Put(a)
	return sock.ioctl(req, ifr.Pointer())
}

func (sock Inet) zaddr(ifname string, req uintptr) error {
	var ifr Ifreq
	var ifa SockaddrIn
	ifr.Rename(ifname)
	ifa.Write(ifr.Ifru[:])
	ifa.Family.Put(syscall.AF_INET)
	return sock.ioctl(req, ifr.Pointer())
}
