// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !linux && !darwin

package netif

import "net/netip"

func Add(ifname string, p netip.Prefix, bc netip.Addr) error {
	return ErrUnavailable
}

func Del(ifname string, p netip.Prefix) error {
	return ErrUnavailable
}
