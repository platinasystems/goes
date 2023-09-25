// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import (
	"fmt"
	"net"
)

// Parse a string as prefix of the given family, or, if family is AF_UNSPEC,
// derive family from syntax.
func Prefix(s string, family uint8) (Prefixer, error) {
	if s == "default" || s == "any" || s == "all" {
		switch family {
		case AF_INET:
			return emptyPrefix{AF_INET}, nil
		case AF_INET6:
			return emptyPrefix{AF_INET6}, nil
		default:
			return nil, fmt.Errorf("prefix %q unsupported by "+
				"%s family", s, AfName(family))
		}
	}
	switch family {
	case AF_MPLS:
		if mpls, err := Address(s, AF_MPLS); err == nil {
			return mplsPrefix{mpls.(mplsAddress)}, nil
		}
	case AF_INET:
		if ip, ones, err := parseCIDR(s); err == nil {
			if ip4 := ip.To4(); ip4 != nil {
				return ip4Prefix{ip4Address(ip4), ones}, nil
			}
		}
	case AF_INET6:
		if ip, ones, err := parseCIDR(s); err == nil {
			if ip6 := ip.To16(); ip6 != nil {
				return ip6Prefix{ip6Address(ip6), ones}, nil
			}
		}
	case AF_UNSPEC:
		if ip, ones, err := parseCIDR(s); err == nil {
			if ip4 := ip.To4(); ip4 != nil {
				return ip4Prefix{ip4Address(ip4), ones}, nil
			}
			return ip6Prefix{ip6Address(ip.To16()), ones}, nil
		}
		if mpls, err := Address(s, AF_MPLS); err == nil {
			return mplsPrefix{mpls.(mplsAddress)}, nil
		}
	}
	return nil, fmt.Errorf("prefix: %q invalid", s)
}

type Prefixer interface {
	Addresser
	Len() uint8 // bits
}

type emptyPrefix struct{ emptyAddress }

func (emptyPrefix) Len() uint8 { return 0 }

type mplsPrefix struct{ mplsAddress }

func (mplsPrefix) Len() uint8 { return 20 }

type ip4Prefix struct {
	ip4Address
	uint8
}

func (prefix ip4Prefix) Len() uint8 { return prefix.uint8 }

type ip6Prefix struct {
	ip6Address
	uint8
}

func (prefix ip6Prefix) Len() uint8 { return prefix.uint8 }

func parseCIDR(s string) (net.IP, uint8, error) {
	ip, ipnet, err := net.ParseCIDR(s)
	if err != nil {
		return nil, 0, err
	}
	ones, _ := ipnet.Mask.Size()
	return ip, uint8(ones), nil
}
