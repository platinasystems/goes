// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tuntap

import (
	"os"
	"syscall"
)

type Configuration struct {
	Unit    uint
	IsTap   bool
	Persist bool
	Owner   int
	Group   int
	Link
}

const (
	TapMin = 14         // Eth header
	TunMin = 2 + 2 + 20 // TunPI + IPv4 header
	Unset  = -1         // flag to not set user or group
)

func ioctl(fd uintptr, req uintptr, argp uintptr) error {
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, fd, req, argp)
	if errno != 0 {
		return os.NewSyscallError("ioctl", errno)
	}
	return nil
}
