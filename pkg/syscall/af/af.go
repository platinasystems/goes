// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !plan9

package af

import (
	"syscall"
)

type AF interface {
	~int
	Family() int
	Type() int
	Proto() int
}

// usage: Open[AF]([<type>[, <proto>]])
//
// Where <type> and <proto> override the Address Family's default socket Type
// and Proto.
func Open[T AF](opt ...int) (sock T, err error) {
	f, t, p := sock.Family(), sock.Type(), sock.Proto()
	if len(opt) > 0 {
		t = opt[0]
		if len(opt) > 1 {
			p = opt[1]
		}
	}
	fd, err := syscall.Socket(f, t, p)
	sock = T(fd)
	return
}

func Addr[FD ~int](fd FD) (syscall.Sockaddr, error) {
	return syscall.Getsockname(int(fd))
}

func Close[FD ~int](fd FD) error {
	return syscall.Close(int(fd))
}

func Recvfrom[FD ~int](fd FD, b []byte, flags int) (
	int, syscall.Sockaddr, error,
) {
	return syscall.Recvfrom(int(fd), b, flags)
}

func Sendto[FD ~int](fd FD, b []byte, flags int, to syscall.Sockaddr) error {
	return syscall.Sendto(int(fd), b, flags, to)
}
