// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "golang.org/x/sys/unix"

// AF - Address-Family
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
	fd, err := unix.Socket(f, t, p)
	sock = T(fd)
	return
}

func Addr[FD ~int](fd FD) (unix.Sockaddr, error) {
	return unix.Getsockname(int(fd))
}

func Close[FD ~int](fd FD) error {
	return unix.Close(int(fd))
}

func Read[FD ~int](fd FD, b []byte) (int, error) {
	return unix.Read(int(fd), b)
}

func Write[FD ~int](fd FD, b []byte) (int, error) {
	return unix.Write(int(fd), b)
}

func Recvfrom[FD ~int](fd FD, b []byte, flags int) (
	int, unix.Sockaddr, error,
) {
	return unix.Recvfrom(int(fd), b, flags)
}

func Sendto[FD ~int](fd FD, b []byte, flags int, to unix.Sockaddr) error {
	return unix.Sendto(int(fd), b, flags, to)
}
