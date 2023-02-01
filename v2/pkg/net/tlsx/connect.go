// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var GoResolve = &net.Resolver{PreferGo: true}

type Connector struct{ *cache.Cache[int] }

var (
	Exchange = Connector{port.Exchange}
	Registry = Connector{port.Registry}
	RPC      = Connector{port.RPC}
)

func (c Connector) Connect(ctx context.Context, host string) (
	conn net.Conn,
	err error,
) {
	cfg := &tls.Config{
		Certificates: []tls.Certificate{
			certs.Self.TLS(),
		},
		RootCAs: certs.RootCAs.Clone(),
	}
	if match, merr := certs.Match(host); merr == nil {
		cfg.ServerName = match.Name
	}

	ipas, err := GoResolve.LookupIPAddr(ctx, host)
	if err != nil {
		return
	}

	tcpa := &net.TCPAddr{
		Port: c.Value(),
	}

	var dl net.Dialer
	for _, ipa := range ipas {
		tcpa.IP = ipa.IP
		tcpa.Zone = ipa.Zone
		conn, err = dl.DialContext(ctx, tcpa.Network(), tcpa.String())
		if err == nil {
			cl := tls.Client(conn, cfg)
			if err = cl.HandshakeContext(ctx); err != nil {
				cl.Close()
				conn = nil
			} else {
				conn = cl
			}
			break
		} else if errors.Is(err, syscall.ECONNREFUSED) {
			style.Error(ipa.IP, err)
			break
		}
	}
	return
}
