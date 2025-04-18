// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import (
	"context"
	"io"
	"net"
	"net/netip"
)

type Rt interface {
	Bits() int
	Dst() netip.Addr
	Flags() uint
	GW() netip.Addr
	HA() net.HardwareAddr
	Index() int
	Line() int
}

type Nd interface {
	Dst() netip.Addr
	Flags() uint
	Expire() int
	GW() netip.Addr
	HA() net.HardwareAddr
	Index() int
	Line() int
	State() int
}

type IFAer interface{ IFA() netip.Addr }
type MTUer interface{ MTU() uint }
type HopCounter interface{ HopCount() uint }
type RecvPiper interface{ RecvPipe() uint }
type SendPiper interface{ SendPipe() uint }
type SSThresher interface{ SSThresh() uint }
type RTTer interface{ RTT() uint }
type RTTVarer interface{ RTTVar() uint }
type Probes interface{ Probes() uint }

type NextNdCloser interface {
	NextNd(context.Context) (Nd, error)
	io.Closer
}

type NextRtCloser interface {
	NextRt(context.Context) (Rt, error)
	io.Closer
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
