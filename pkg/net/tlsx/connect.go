// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
)

const Port = 8003

// host: [<name>][@<dns|ip4|[ipv6]>][:<port>]
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

	for _, ip := range ips {
		tcpa.IP = ip.IP
		conn, err := Dial(ctx, tcpa.Network(), tcpa.String())
		if err == nil {
			return conn, nil
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			err = fmt.Errorf("%v: %w", ip.IP, err)
			return nil, egress.Marked(err)
		}
	}
	return nil, egress.Marked(fmt.Errorf("%s: %w", host, ErrUnavailable))
}

func DNS0(conn net.Conn) string {
	if tlsc, ok := conn.(*tls.Conn); ok {
		cs := tlsc.ConnectionState()
		if len(cs.PeerCertificates) > 0 &&
			len(cs.PeerCertificates[0].DNSNames) > 0 {
			return cs.PeerCertificates[0].DNSNames[0]
		}
	}
	return "anonymous"
}

func parseIP(s string) (net.IP, error) {
	if ip := net.ParseIP(s); ip != nil {
		return ip, nil
	}
	return nil, fmt.Errorf("%q: %w", s, ErrInvalid)
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
