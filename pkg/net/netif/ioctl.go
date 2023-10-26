// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const IFNAMSIZ = syscall.IFNAMSIZ

type SockaddrIn struct{ syscall.RawSockaddrInet4 }

func (sain *SockaddrIn) Set(addr []byte) {
	sain.Len = uint8(unsafe.Sizeof(*sain))
	sain.Family = syscall.AF_INET
	copy(sain.Addr[:], addr)
}

type SockaddrIn6 struct{ syscall.RawSockaddrInet6 }

func (sain6 *SockaddrIn6) Set(addr []byte) {
	sain6.Len = uint8(unsafe.Sizeof(*sain6))
	sain6.Family = syscall.AF_INET6
	copy(sain6.Addr[:], addr)
}

type InAlias struct {
	Addr SockaddrIn
	Dest SockaddrIn
	Mask SockaddrIn
}

type In6Alias struct {
	Addr     SockaddrIn6
	Dest     SockaddrIn6
	Mask     SockaddrIn6
	Flags    int32
	Lifetime In6AddrLifetime
	Vhid     int32
}

type In6AddrLifetime struct {
	Expire    TimeT
	Preferred TimeT
	Vltime    uint32
	Pltime    uint32
}

type Nothing struct{}

type IfreqValue interface {
	Nothing | ~byte | ~uint16 | ~int32 | ~uint32 | IFCAPS |
		InAlias | In6Alias |
		SockaddrIn | SockaddrIn6
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

func IOCTL[FD ~int, R Ifreq[Nothing] |
	Ifreq[IFF] |
	Ifreq[uint32] |
	Ifreq[IFCAPS] |
	Ifreq[InAlias] |
	Ifreq[In6Alias] |
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

func InetIOCTL[R Ifreq[Nothing] |
	Ifreq[IFF] |
	Ifreq[uint32] |
	Ifreq[InAlias] |
	Ifreq[SockaddrIn]](
	op uintptr, req *R,
) error {
	inet, err := af.OpenInet()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	return IOCTL(inet, op, req)
}

func Inet6IOCTL[R Ifreq[In6Alias] | Ifreq[SockaddrIn6]](
	op uintptr, req *R,
) error {
	inet6, err := af.OpenInet6()
	if err != nil {
		return err
	}
	defer af.Close(inet6)
	return IOCTL(inet6, op, req)
}
