// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/platinasystems/goes/internal/safe"
)

// Parse a string as address of the given family, or, if family is AF_UNSPEC,
// derive family from syntax.
func Address(s string, family uint8) (Addresser, error) {
	switch family {
	case AF_MPLS:
		var err error
		labels := strings.Split(s, "/")
		mpls := make(mplsAddress, 4*len(labels))
		for i, label := range labels {
			var u32 uint32
			if _, err = fmt.Sscan(label, &u32); err != nil {
				break
			}
			u32 <<= MPLS_LS_LABEL_SHIFT
			if i == len(labels)-1 {
				u32 |= 1 << MPLS_LS_S_SHIFT
			}
			mpls[(4*i)+0] = byte(u32 >> 24)
			mpls[(4*i)+1] = byte(u32 >> 16)
			mpls[(4*i)+2] = byte(u32 >> 8)
			mpls[(4*i)+3] = byte(u32 & 0xff)
		}
		if err == nil {
			return mpls, nil
		}
	case AF_INET:
		if ip := net.ParseIP(s); ip != nil {
			if ip4 := ip.To4(); ip4 != nil {
				return ip4Address(ip4), nil
			}
		}
	case AF_INET6:
		if ip := net.ParseIP(s); ip != nil {
			if ip6 := ip.To16(); ip6 != nil {
				return ip6Address(ip6), nil
			}
		}
	case AF_UNSPEC:
		if ip := net.ParseIP(s); ip != nil {
			if ip4 := ip.To4(); ip4 != nil {
				return ip4Address(ip4), nil
			}
			return ip6Address(ip.To16()), nil
		}
		return Address(s, AF_MPLS)
	}
	return nil, fmt.Errorf("address: %q invalid", s)
}

type Addresser interface {
	Bytes() []byte
	Family() uint8
	io.Reader
	IsLoopback() bool
}

type mplsAddress []byte
type ip4Address net.IP
type ip6Address net.IP

func (mpls mplsAddress) Bytes() []byte { return mpls }
func (ip4 ip4Address) Bytes() []byte   { return ip4 }
func (ip6 ip6Address) Bytes() []byte   { return ip6 }

func (mplsAddress) Family() uint8 { return AF_MPLS }
func (ip4Address) Family() uint8  { return AF_INET }
func (ip6Address) Family() uint8  { return AF_INET6 }

func (mpls mplsAddress) Read(b []byte) (int, error) { return safe.Cp(b, mpls) }
func (ip4 ip4Address) Read(b []byte) (int, error)   { return safe.Cp(b, ip4) }
func (ip6 ip6Address) Read(b []byte) (int, error)   { return safe.Cp(b, ip6) }

func (mplsAddress) IsLoopback() bool    { return false }
func (ip4 ip4Address) IsLoopback() bool { return net.IP(ip4).IsLoopback() }
func (ip6 ip6Address) IsLoopback() bool { return net.IP(ip6).IsLoopback() }
