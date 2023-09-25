// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import "github.com/platinasystems/goes/v2/pkg/syscall/af"

const (
	IN6_IFF_ANYCAST = 1 << iota
	IN6_IFF_TENTATIVE
	IN6_IFF_DUPLICATED
	IN6_IFF_DETACHED
	IN6_IFF_DEPRECATED
	IN6_IFF_NODAD
	IN6_IFF_AUTOCONF
	IN6_IFF_TEMPORARY
	IN6_IFF_PREFER_SOURCE
	IN6_IFF_OPTIMISTIC
	IN6_IFF_SECURED
	_
	IN6_IFF_CLAT46
	_
	_
	IN6_IFF_NOPFX
)

const (
	IN6_IFF_DYNAMIC     = IN6_IFF_PREFER_SOURCE
	IN6_IFF_DADPROGRESS = IN6_IFF_TENTATIVE | IN6_IFF_OPTIMISTIC
	IN6_IFF_NOTREADY    = IN6_IFF_TENTATIVE | IN6_IFF_DUPLICATED
	IN6_IFF_NOTMANUAL   = IN6_IFF_AUTOCONF | IN6_IFF_DYNAMIC
)

func Admin(ifname string, with, without IFF) error {
	inet, err := af.Open[af.Inet]()
	if err != nil {
		return err
	}
	defer af.Close(inet)
	req := NewIfreq[IFF](ifname)
	if IOCTL(inet, SIOCGIFFLAGS, req); err != nil {
		return err
	}
	req.Value |= with
	req.Value &^= without
	return IOCTL(inet, SIOCSIFFLAGS, req)
}

func Down(ifname string) error {
	return Admin(ifname, 0, IFF_UP)
}

func Up(ifname string) error {
	return Admin(ifname, IFF_UP|IFF_RUNNING, 0)
}
