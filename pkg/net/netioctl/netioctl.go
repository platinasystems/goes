// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netioctl

//go:generate sh -c "go tool cgo -godefs -- sioc6.go > zsioc6_${GOOS}.go"
//go:generate stringer -output=zifcap_string_${GOOS}.go -type=IFCAP -trimprefix=IFCAP_ .
//go:generate sh -c "go doc syscall.IFF_UP | sed -n -f iff.sed | gofmt > ziff_${GOOS}.go"
//go:generate stringer -output=ziff_string_${GOOS}.go -type=IFF -trimprefix=IFF_ .
//go:generate sh -c "go doc syscall.IFT_OTHER | sed -n -f ift.sed | gofmt > zift_${GOOS}.go"
//go:generate stringer -output=zift_string_${GOOS}.go -type=IFT -trimprefix=IFT_ .

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/net/sockaddr"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const IFNAMSIZ = syscall.IFNAMSIZ

type InAlias struct {
	Addr sockaddr.In
	Dest sockaddr.In
	Mask sockaddr.In
}

type In6Alias struct {
	Addr     sockaddr.In6
	Dest     sockaddr.In6
	Mask     sockaddr.In6
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

type IfReqValue interface {
	~byte | ~uint16 | ~int32 | ~uint32 |
		Nothing |
		IfCaps |
		InAlias | In6Alias |
		sockaddr.In | sockaddr.In6
}

type IfReq[V IfReqValue] struct {
	Name  [IFNAMSIZ]byte
	Value V
}

func NewIfReq[V IfReqValue](name string) *IfReq[V] {
	ifr := new(IfReq[V])
	copy(ifr.Name[:], name)
	return ifr
}

var (
	NewIfReqUint8       = NewIfReq[byte]
	NewIfReqUint16      = NewIfReq[uint16]
	NewIfReqUint32      = NewIfReq[uint32]
	NewIfReqInt32       = NewIfReq[int32]
	NewIfReqNothing     = NewIfReq[Nothing]
	NewIfReqIfCaps      = NewIfReq[IfCaps]
	NewIfReqInAlias     = NewIfReq[InAlias]
	NewIfReqIn6Alias    = NewIfReq[In6Alias]
	NewIfReqSockaddrIn  = NewIfReq[sockaddr.In]
	NewIfReqSockaddrIn6 = NewIfReq[sockaddr.In6]
)

func Inet[R IfReq[Nothing] |
	IfReq[uint16] |
	IfReq[uint32] |
	IfReq[InAlias] |
	IfReq[sockaddr.In]](
	op uintptr, req *R,
) error {
	inet, err := af.OpenInet()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	return ioctl(inet, op, req)
}

func Inet6[R IfReq[In6Alias] | IfReq[sockaddr.In6]](
	op uintptr, req *R,
) error {
	inet6, err := af.OpenInet6()
	if err != nil {
		return err
	}
	defer af.Close(inet6)
	return ioctl(inet6, op, req)
}

func ioctl[FD ~int, R IfReq[Nothing] |
	IfReq[uint16] |
	IfReq[uint32] |
	IfReq[IfCaps] |
	IfReq[InAlias] |
	IfReq[In6Alias] |
	IfReq[sockaddr.In] |
	IfReq[sockaddr.In6]](
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
