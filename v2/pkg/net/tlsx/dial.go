// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/ipc"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

func DialAndHandshake(ctx context.Context, ex string) (
	cl *tls.Conn, err error,
) {
	nw, addr, cfg, err := dialcfg(ex)
	if err != nil {
		return
	}
	var dl net.Dialer
	c, err := dl.DialContext(ctx, nw, addr)
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

func dialcfg(ex string) (nw, addr string, cfg *tls.Config, err error) {
	var sn string
	hn, err := host.Name.ValErr()
	if err != nil {
		return
	}
	nw = "tcp"
	if len(ex) == 0 {
		sn = hn
		na := ipc.Exchange()
		nw = na.Network()
		if addr, err = na.Address(); err != nil {
			return
		}
	} else if dns, port := certs.Exchanges.NamePort(ex); len(dns) > 0 {
		if sn = dns; len(port) == 0 {
			err = fmt.Errorf("%s: port empty", dns)
		} else if port == "0" {
			nw, addr, cfg, err = dialcfg("")
			return
		} else {
			addr = fmt.Sprint(dns, ":", port)
		}
	} else {
		err = fmt.Errorf("%s: %w", ex, ErrNameOrSKINotFound)
		return
	}
	tlsc, err := cert.ValErr()
	if err != nil {
		return
	}
	rootCAs, err := certs.Exchanges.Pool()
	if err != nil {
		return
	}
	cfg = &tls.Config{
		Certificates: []tls.Certificate{tlsc},
		ServerName:   sn,
		RootCAs:      rootCAs,
	}
	return
}
