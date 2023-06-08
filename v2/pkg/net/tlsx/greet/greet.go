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
//	host: [<name>][@<dns|ip4|[ipv6]>][:<port>]	(default self)
func Server(ctx context.Context, host string, conn net.Conn) (
	cl *tls.Conn, err error,
) {
	var sname string
	if len(host) == 0 {
		if sname, err = certs.Self.Name(); err != nil {
			return
		}
	} else {
		at := strings.Index(host, "@")
		colon := strings.LastIndex(host, ":")
		if at < 0 {
			if colon > 0 { // <name>:<port>
				host = host[:colon]
			}
		} else if at > 0 { // <name>@...
			host = host[:at]
		} else if colon > 0 { // @<dns>:<port>
			host = host[1:colon]
		} else { // @<dns>
			host = host[1:]
		}
		if match, merr := certs.Match(host); merr != nil {
			err = fmt.Errorf("%s: %w", host, merr)
			return
		} else {
			sname = match.Name()
		}
	}

	self, err := certs.Self.TLS()
	if err != nil {
		return
	}

	ob := page.New()
	defer page.Free(ob)

	dec := lv.NewDecoder(poll.With(ctx, conn))
	enc := lv.NewEncoder(write.With(ctx, conn))

	enc.Encode("hello", nil)
	if _, err = dec.Read(ob); err == nil {
		cl = tls.Client(conn, &tls.Config{
			Certificates: []tls.Certificate{self},
			RootCAs:      certs.RootCAs.Clone(),
			ServerName:   sname,
		})
		err = cl.HandshakeContext(ctx)
	}
	return
}
