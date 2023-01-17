// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux || darwin

package netif

import (
	"fmt"
	"net/netip"
)

func Add(ifname string, p netip.Prefix, bc netip.Addr) error {
	pa := p.Addr()
	if pa.Is4() {
		if sock, err := NewInet(); err == nil {
			defer sock.Close()
			return sock.Add(ifname, p, bc)
		} else {
			return err
		}
	} else if pa.Is6() {
		if sock, err := NewInet6(); err == nil {
			defer sock.Close()
			return sock.Add(ifname, p)
		} else {
			return err
		}
	}
	return fmt.Errorf("%v: invalid", pa)
}

func Del(ifname string, p netip.Prefix) error {
	pa := p.Addr()
	if pa.Is4() {
		if sock, err := NewInet(); err == nil {
			defer sock.Close()
			return sock.Del(ifname, p)
		} else {
			return err
		}
	} else if pa.Is6() {
		if sock, err := NewInet6(); err == nil {
			defer sock.Close()
			return sock.Del(ifname, p)
		} else {
			return err
		}
	}
	return fmt.Errorf("%v: invalid", pa)
}
