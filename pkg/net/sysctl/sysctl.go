// Derrived from golang.org/x/net/route
// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

//go:build darwin || dragonfly || freebsd || netbsd || openbsd

package sysctl

import (
	"errors"
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

var Err3Strikes = errors.New("failed 3 times")

func Get(mib ...int32) ([]byte, error) {
	for try, b := 0, make([]byte, align.Page.Size()); try < 3; try++ {
		n := uintptr(len(b))
		err := sysctl(mib, &b[0], &n, nil, 0)
		if err == nil {
			return b[:n], nil
		}
		if err != syscall.ENOMEM {
			return nil, os.NewSyscallError("sysctl", err)
		}
		b = make([]byte, align.Page.Roundup(int(n)))
	}
	return nil, Err3Strikes
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
