// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/netip"
	"strings"
	"syscall"
)

const Port = 8003

var ErrUnavailable = errors.New("unavailable")

func Connect(ctx context.Context, host string) (net.Conn, error) {
	var ips []net.IPAddr
	tcpa := &net.TCPAddr{
		Port: Port,
	}

	at := strings.Index(host, "@")
	colon := strings.LastIndex(host, ":")
	if at > 0 {
		s := host[at+1:]
		if colon > 0 {
			if ap, err := netip.ParseAddrPort(s); err != nil {
				return nil, fmt.Errorf("%s: %w", s, err)
			} else {
				a := ap.Addr()
				ips = append(ips, netipAddr2netIPAddr(a))
				tcpa.Port = int(ap.Port())
			}
		} else if a, err := netip.ParseAddr(s); err != nil {
			return nil, fmt.Errorf("%s: %w", s, err)
		} else if !a.IsValid() {
			return nil, fmt.Errorf("%s: invalid", s)
		} else {
			ips = append(ips, netipAddr2netIPAddr(a))
		}
	} else {
		if colon > 0 {
			s := host[colon+1:]
			host = host[:colon]
			if _, err := fmt.Sscan(s, &tcpa.Port); err != nil {
				return nil, fmt.Errorf("%s: %w", s, err)
			}
		}
		if res, err := GoResolve.LookupIPAddr(ctx, host); err != nil {
			return nil, fmt.Errorf("%s: %w", host, err)
		} else {
			ips = res
		}
	}

	var dl net.Dialer
	for _, ip := range ips {
		tcpa.IP = ip.IP
		tcpa.Zone = ip.Zone
		conn, err := dl.DialContext(ctx, tcpa.Network(), tcpa.String())
		if err == nil {
			return conn, nil
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, fmt.Errorf("%v: %w", ip.IP, err)
		}
	}
	return nil, fmt.Errorf("%s: %w", host, ErrUnavailable)
}

func netipAddr2netIPAddr(ip netip.Addr) net.IPAddr {
	if ip.Is6() {
		ip6 := ip.As16()
		return net.IPAddr{
			IP:   net.IP(ip6[:]),
			Zone: ip.Zone(),
		}
	}
	ip4 := ip.As4()
	return net.IPAddr{IP: net.IP(ip4[:])}
}
