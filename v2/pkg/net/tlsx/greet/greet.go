// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package greet

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

func Client(ctx context.Context, conn net.Conn) (*tls.Conn, error) {
	self, err := certs.Self.TLS()
	if err != nil {
		return nil, err
	}
	name, err := certs.Self.Name()
	if err != nil {
		return nil, err
	}
	sv := tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{
			self,
		},
		ServerName: name,
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  certs.ClientCAs.Clone(),
	})
	return sv, sv.HandshakeContext(ctx)
}

// Send hello to server then handshake and return TLS connection.
func Server(ctx context.Context, name string, conn net.Conn) (
	*tls.Conn, error,
) {
	self, err := certs.Self.TLS()
	if err != nil {
		return nil, err
	}
	if name, err = certs.Self.Name(); err != nil {
		return nil, err
	}

	ob := page.New()
	defer page.Free(ob)

	dec := lv.NewDecoder(poll.With(ctx, conn))
	enc := lv.NewEncoder(write.With(ctx, conn))

	enc.Encode("hello", nil)
	if _, err := dec.Read(ob); err != nil {
		return nil, err
	}

	cfg := &tls.Config{
		Certificates: []tls.Certificate{self},
		RootCAs:      certs.RootCAs.Clone(),
	}
	if i := strings.Index(name, "@"); i == 0 {
		cfg.ServerName = name
	} else {
		if i > 0 {
			name = name[:i]
		} else if i = strings.Index(name, ":"); i > 0 {
			name = name[:i]
		}
		if match, err := certs.Match(name); err == nil {
			cfg.ServerName = match.Name()
		} else {
			return nil, fmt.Errorf("%s: %w", name, err)
		}
	}
	cl := tls.Client(conn, cfg)
	return cl, cl.HandshakeContext(ctx)
}
