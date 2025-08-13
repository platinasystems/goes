// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import (
	"context"
	"net"
	"net/netip"
	"strconv"
	"strings"
	"unicode"
)

// Resolve or parse the following,
//
//	[<ip6>]:<port>
//	<ip6>
//	<ip4>:<port>
//	<ip4>
//	<name>:<port>
//	<name>
func ResolveAddrPort(
	ctx context.Context,
	r *net.Resolver, // if nil, PreferGo resolver
	s string,
	port uint16, // if not given in string
) (
	aps []netip.AddrPort,
	err error,
) {
	var is4a, is4ap, is6a, is6ap bool
	if r == nil {
		r = &net.Resolver{PreferGo: true}
	}
	firstColon := strings.Index(s, ":")
	lastColon := strings.LastIndex(s, ":")
	rightBracket := strings.Index(s, "]")
	if strings.HasPrefix(s, "[") {
		is6ap = rightBracket == lastColon-1
	} else if firstColon > 0 && firstColon < lastColon {
		is6a = true
	} else if unicode.IsDigit([]rune(s)[0]) {
		if lastColon >= 8 {
			is4ap = true
		} else {
			is4a = true
		}
	}
	if is4ap || is6ap {
		var ap netip.AddrPort
		if ap, err = netip.ParseAddrPort(s); err == nil {
			aps = []netip.AddrPort{ap}
		}
		return
	}
	if is4a || is6a {
		var a netip.Addr
		if a, err = netip.ParseAddr(s); err == nil {
			aps = []netip.AddrPort{
				netip.AddrPortFrom(a, port),
			}
		}
		return
	}
	if lastColon > 0 {
		var u64 uint64
		u64, err = strconv.ParseUint(s[lastColon+1:], 10, 16)
		if err != nil {
			return
		}
		port = uint16(u64)
		s = s[:lastColon]
	}
	ips, err := r.LookupNetIP(ctx, "ip", s)
	if err != nil {
		return
	}
	aps = make([]netip.AddrPort, len(ips))
	for i, ip := range ips {
		aps[i] = netip.AddrPortFrom(ip, port)
	}
	return
}
