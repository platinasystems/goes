// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/log/style"
)

const Port = 8003

var (
	ErrInvalidAddr = errors.New("invalid address")
	ErrUnavailable = errors.New("unavailable")
)

// [<name>][@<dns|ip4|[ipv6]>][:<port>]
func Connect(ctx context.Context, host string) (net.Conn, error) {
	var ips []net.IPAddr
	var err error
	tcpa := &net.TCPAddr{
		Port: Port,
	}

	at := strings.Index(host, "@")
	colon := strings.LastIndex(host, ":")
	if at < 0 {
		if colon > 0 { // <name>:<port>
			tcpa.Port, err = portscan(host[colon+1:])
			if err != nil {
				return nil, err
			}
			host = host[:colon]
		} // else <name>
		if ips, err = resolve(ctx, host); err != nil {
			return nil, err
		}
	} else if strings.HasPrefix(host[at+1:], "[") {
		rb := strings.LastIndex(host, "]")
		if rb <= 0 {
			return nil, fmt.Errorf("%q: invalid", host)
		}
		if colon > rb { // <name>@[<ipv6>]:<port>
			tcpa.Port, err = portscan(host[colon+1:])
			if err != nil {
				return nil, err
			}
		} // else <name>@[<ipv6>]
		if ip, err := parseIP(host[at+2 : rb]); err == nil {
			ips = append(ips, net.IPAddr{IP: ip})
		} else {
			return nil, err
		}
	} else {
		if colon > 0 { // <name>@<dns|ip4>:<port>
			tcpa.Port, err = portscan(host[colon+1:])
			if err != nil {
				return nil, err
			}
			host = host[at+1 : colon]
		} else { // <name>@<dns|ip4>
			host = host[at+1:]
		}
		if ip, err := parseIP(host); err == nil {
			ips = append(ips, net.IPAddr{IP: ip})
		} else if ips, err = resolve(ctx, host); err != nil {
			return nil, err
		}
	}

	style.Noteln("ips", ips)
	var dl net.Dialer
	for _, ip := range ips {
		tcpa.IP = ip.IP
		style.Noteln("connect", tcpa.IP, "...")
		conn, err := dl.DialContext(ctx, tcpa.Network(), tcpa.String())
		if err == nil {
			style.Noteln("connected", tcpa.IP)
			return conn, nil
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			return nil, fmt.Errorf("%v: %w", ip.IP, err)
		}
	}
	style.Error(host, ": ", ErrUnavailable)
	return nil, fmt.Errorf("%s: %w", host, ErrUnavailable)
}

func parseIP(s string) (net.IP, error) {
	if ip := net.ParseIP(s); ip != nil {
		return ip, nil
	}
	return nil, fmt.Errorf("%q: %w", s, ErrInvalidAddr)
}

func portscan(s string) (port int, err error) {
	if _, err = fmt.Sscan(s, &port); err != nil {
		err = fmt.Errorf("%q: %w", s, err)
	}
	return
}

func resolve(ctx context.Context, s string) (ips []net.IPAddr, err error) {
	if ips, err = GoResolve.LookupIPAddr(ctx, s); err != nil {
		err = fmt.Errorf("%q: %w", s, err)
	}
	return
}
