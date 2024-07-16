// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import (
	"context"
	"net"
	"net/netip"
)

type NetRt interface {
	Index() int
	Line() int
	Flags() uint
	Bits() int
	Dst() netip.Addr
	IFA() netip.Addr
	GW() netip.Addr
	HA() net.HardwareAddr
}

type Streamer interface {
	Next(context.Context) (NetRt, error)
	Close() error
}

func FirstNonNil(args ...any) any {
	for _, arg := range args {
		if arg != nil {
			return arg
		}
	}
	return nil
}

func Zero[T comparable]() (t T) { return }

func FirstNonZero[T comparable](args ...T) T {
	var zero T
	for _, arg := range args {
		if arg != zero {
			return arg
		}
	}
	return zero
}

func SelectGateway(ipas []net.IPAddr, prefer6 bool) net.IPAddr {
	ipa := ipas[0]
	l := net.IPv4len
	if prefer6 {
		l = net.IPv6len
	}
	for _, entry := range ipas {
		if len(entry.IP) == l {
			ipa = entry
		}
	}
	return ipa
}
