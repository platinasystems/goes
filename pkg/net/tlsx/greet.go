// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

func greetClient(ctx context.Context, conn net.Conn) (*tls.Conn, error) {
	self := Self()
	sv := tls.Server(conn, &tls.Config{
		Certificates: []tls.Certificate{
			self.Certificate,
		},
		ServerName: self.Name(),
		ClientAuth: tls.RequireAndVerifyClientCert,
		ClientCAs:  ClientCAs().Clone(),
	})
	return sv, sv.HandshakeContext(ctx)
}

// Send "tls" command to server then handshake and return TLS connection.
//	host: [<name>][@<dns|ip4|[ipv6]>][:<port>]	(default self)
func greetServer(ctx context.Context, host string, conn net.Conn) (
	cl *tls.Conn, err error,
) {
	var sname string
	self := Self()
	if len(host) == 0 {
		sname = self.Name()
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
		if match, merr := Match(host); merr != nil {
			err = fmt.Errorf("%s: %w", host, merr)
			return
		} else {
			sname = match.Name()
		}
	}

	ob := page.New()
	defer page.Free(ob)

	dec := lv.NewDecoder(poll.WithReader(ctx, conn))
	enc := lv.NewEncoder(write.With(ctx, conn))

	enc.Encode("tls", nil)
	if _, err = dec.Read(ob); err == nil {
		cl = tls.Client(conn, &tls.Config{
			Certificates: []tls.Certificate{self.Certificate},
			RootCAs:      RootCAs().Clone(),
			ServerName:   sname,
		})
		err = cl.HandshakeContext(ctx)
	}
	return
}
