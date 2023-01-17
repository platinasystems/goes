// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

import "syscall"

type Inet int
type Inet6 int

func NewInet() (Inet, error) {
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_DGRAM, 0)
	if err != nil {
		fd = -1
	}
	return Inet(fd), err
}

func NewInet6() (Inet6, error) {
	fd, err := syscall.Socket(syscall.AF_INET6, syscall.SOCK_DGRAM, 0)
	if err != nil {
		fd = -1
	}
	return Inet6(fd), err
}

func (sock Inet) Close() error {
	if sock < 0 {
		return syscall.EBADF
	}
	return syscall.Close(int(sock))
}

func (sock Inet6) Close() error {
	if sock < 0 {
		return syscall.EBADF
	}
	return syscall.Close(int(sock))
}

func (sock Inet) IsOpen() bool  { return sock != -1 }
func (sock Inet6) IsOpen() bool { return sock != -1 }

func (sock Inet) ioctl(req, argp uintptr) error {
	return ioctl(uintptr(sock), req, argp)
}

func (sock Inet6) ioctl(req, argp uintptr) error {
	return ioctl(uintptr(sock), req, argp)
}
