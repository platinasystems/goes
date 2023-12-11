// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netrt

import (
	"net"
	"net/netip"
	"unicode"
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
	Next() (NetRt, error)
	Close() error
}

type NetstatFlagCode struct {
	Flag uint
	Code rune
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

func isnumeric(s string) bool {
	r := []rune(s)[0]
	return unicode.IsNumber(r) || r == ':'
}
