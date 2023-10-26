// Derrived from golang.org/x/net/route
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package sysctl

import (
	"os"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

type MsgSubHdr struct {
	Msglen  uint16
	Version uint8
	Type    uint8
}

type MsgTypes interface {
	MsgSubHdr | IfMsghdr | IfaMsghdr | IfmaMsghdr | IfmaMsghdr2 | RtMsghdr
}

type SockaddrSubHdr struct {
	Len    uint8
	Family uint8
	Index  uint16
}

type SockaddrTypes interface {
	SockaddrSubHdr | SockaddrDatalink | SockaddrIn | SockaddrIn6
}

var trysize = []uintptr{
	4 << 10, // 4KB
	64 << 10,
	256 << 10,
	1 << 20, // 1MB
}

func Get(mib ...int32) ([]byte, error) {
	for try := 0; try < len(trysize); try++ {
		n := trysize[try]
		b := make([]byte, n)
		if err := sysctl(mib, &b[0], &n, nil, 0); err == nil {
			return b[:n], nil
		} else if err != syscall.ENOMEM {
			return nil, os.NewSyscallError("sysctl", err)
		}
	}
	return nil, syscall.ENOMEM
}

// Return the beginning type along with any attached body and remainder.
func Extract[T MsgTypes](data []byte) (p *T, body, rem []byte) {
	h := Pointer[MsgSubHdr](data)
	p = Pointer[T](data)
	body = data[Sizeof(p):]
	rem = data[align.Sysctl.Roundup(int(h.Msglen)):]
	return
}

var (
	ExtractIfMsghdr    = Extract[IfMsghdr]
	ExtractIfaMsghdr   = Extract[IfaMsghdr]
	ExtractIfmaMsghdr  = Extract[IfmaMsghdr]
	ExtractIfmaMsghdr2 = Extract[IfmaMsghdr2]
	ExtractRtMsghdr    = Extract[RtMsghdr]
)

func ExtractSockaddr[T SockaddrTypes](data []byte) (p *T, body, rem []byte) {
	h := Pointer[SockaddrSubHdr](data)
	p = Pointer[T](data)
	body = data[Sizeof(p):]
	rem = data[align.Sysctl.Roundup(int(h.Len)):]
	return
}

var (
	ExtractSockaddrDatalink = ExtractSockaddr[SockaddrDatalink]
	ExtractSockaddrIn       = ExtractSockaddr[SockaddrIn]
	ExtractSockaddrIn6      = ExtractSockaddr[SockaddrIn6]
)

// Return type at beginning of data.
func Pointer[T MsgTypes | SockaddrTypes](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

func Sizeof[T MsgTypes | SockaddrTypes](p *T) int {
	return int(unsafe.Sizeof(*p))
}
