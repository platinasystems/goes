// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"net"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/address"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

func DialAndHandshake(ctx context.Context, ex string) (
	cl *tls.Conn, err error,
) {
	addr, cfg, err := AddrCfg(ex)
	if err != nil {
		return
	}
	var dl net.Dialer
	c, err := dl.DialContext(ctx, addr.Network(), addr.String())
	if err != nil {
		return
	}
	cl = tls.Client(c, cfg)
	if err = cl.HandshakeContext(ctx); err != nil {
		cl.Close()
		cl = nil
	}
	return
}

func AddrCfg(ex string) (addr net.Addr, cfg *tls.Config, err error) {
	var ok bool
	name, ski, exc := Lookup(ex)
	if exc == nil {
		err = certs.Unsubscribed(ex)
		return
	}
	if addr, ok = address.Load(ski); !ok {
		err = address.Unaddressed(ski, name)
		return
	}
	cfg = &tls.Config{
		Certificates: []tls.Certificate{cert.Value()},
		ServerName:   name,
		RootCAs:      certs.RootCAs.Clone(),
	}
	return
}
