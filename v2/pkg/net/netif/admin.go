// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux || darwin

package netif

import (
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
)

func Up(ifname string) error {
	return Admin(ifname, syscall.IFF_UP|syscall.IFF_RUNNING, 0)
}

func Down(ifname string) error {
	return Admin(ifname, 0, syscall.IFF_UP)
}

func Admin(ifname string, with, without uint16) error {
	var ifr Ifreq
	ifr.Rename(ifname)
	sock, err := NewInet()
	if err != nil {
		return err
	}
	defer sock.Close()
	err = sock.ioctl(syscall.SIOCGIFFLAGS, ifr.Pointer())
	if err != nil {
		return err
	}
	iff := (*host.Uint16)(ifr.Ifru[:]).Value()
	iff |= with
	iff &^= without
	(*host.Uint16)(ifr.Ifru[:]).Put(iff)
	return sock.ioctl(syscall.SIOCSIFFLAGS, ifr.Pointer())
}
