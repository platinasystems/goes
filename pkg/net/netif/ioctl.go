// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"net"
	"os"
	"syscall"
	"unsafe"
)

const IFNAMSIZ = syscall.IFNAMSIZ

type SockaddrIn = syscall.RawSockaddrInet4

func SockaddrInInit(sain *SockaddrIn, addr []byte) {
	sain.Len = uint8(unsafe.Sizeof(*sain))
	sain.Family = syscall.AF_INET
	copy(sain.Addr[:], addr)
}

type SockaddrIn6 = syscall.RawSockaddrInet6

func SockaddrIn6Init(sain6 *SockaddrIn6, addr []byte) {
	sain6.Len = uint8(unsafe.Sizeof(*sain6))
	sain6.Family = syscall.AF_INET6
	copy(sain6.Addr[:], addr)
}

type InAliasreq struct {
	Name      [IFNAMSIZ]byte
	Addr      SockaddrIn
	Broadaddr SockaddrIn
	Mask      SockaddrIn
}

func NewInAliasreq(
	name string,
	addr, dest net.IP,
	mask net.IPMask,
) *InAliasreq {
	req := new(InAliasreq)
	copy(req.Name[:], name)
	if addr != nil {
		SockaddrInInit(&req.Addr, addr)
	}
	if dest != nil {
		SockaddrInInit(&req.Broadaddr, dest)
	}
	if mask != nil {
		SockaddrInInit(&req.Mask, mask)
	}
	return req
}

type In6Aliasreq struct {
	Name       [IFNAMSIZ]byte
	Addr       SockaddrIn6
	Broadaddr  SockaddrIn6
	Prefixmask SockaddrIn6
	Flags      int32
	Lifetime   In6Addrlifetime
}

type In6Addrlifetime struct {
	Expire    TimeT
	Preferred TimeT
	Vltime    uint32
	Pltime    uint32
}

func NewIn6Aliasreq(
	name string,
	addr, dest net.IP,
	mask net.IPMask,
) *In6Aliasreq {
	req := new(In6Aliasreq)
	copy(req.Name[:], name)
	req.Lifetime.Vltime = ND6_INFINITE_LIFETIME
	req.Lifetime.Pltime = ND6_INFINITE_LIFETIME
	if addr != nil {
		SockaddrIn6Init(&req.Addr, addr)
	}
	if dest != nil {
		SockaddrIn6Init(&req.Broadaddr, dest)
	}
	if mask != nil {
		SockaddrIn6Init(&req.Prefixmask, mask)
	}
	return req
}

type Nothing struct{}

type IfreqValue interface {
	~byte | ~uint16 | ~int32 | ~uint32 | Nothing |
		IFCAPS | SockaddrIn | SockaddrIn6
}

type Ifreq[V IfreqValue] struct {
	Name  [IFNAMSIZ]byte
	Value V
}

func NewIfreq[V IfreqValue](name string) *Ifreq[V] {
	ifr := new(Ifreq[V])
	copy(ifr.Name[:], name)
	return ifr
}

func IOCTL[FD ~int, R InAliasreq | In6Aliasreq |
	Ifreq[Nothing] |
	Ifreq[IFF] |
	Ifreq[uint32] |
	Ifreq[SockaddrIn] |
	Ifreq[SockaddrIn6]](
	fd FD, op uintptr, req *R,
) error {
	if op == 0 {
		return syscall.EOPNOTSUPP
	}
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, uintptr(fd), op,
		uintptr(unsafe.Pointer(req)))
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}
