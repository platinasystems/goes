// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package subscribe

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	ErrMissingAddress = errors.New("missing <address>:<port>")
	ErrNoPeer         = errors.New("no peer certificates")
	ErrNotNamePort    = errors.New("expect <name>:<port>")
	ErrNotOK          = errors.New("unsuccessful")
	ErrUnexpectedArgs = errors.New("unexpected argument(s)")
)

func Req(ctx context.Context, headers map[string]string, args []string) error {
	const unixurl = "unix://"
	var name, addr string
	var dl net.Dialer
	nw := "tcp"
	switch len(args) {
	case 0:
		return ErrMissingAddress
	case 1:
		if strings.HasPrefix(args[0], unixurl) {
			nw = "unix"
			addr = strings.TrimPrefix(args[0], unixurl)
			name = host.Name.Value()
		} else if i := strings.LastIndex(args[0], ":"); i < 0 {
			return ErrNotNamePort
		} else if i > 0 {
			addr = args[0]
			name = args[0][:i]
		} else {
			addr = args[0]
			name = host.Name.Value()
		}
	case 2:
		name = args[0]
		addr = args[1]
	default:
		return fmt.Errorf("%w: %v", ErrUnexpectedArgs, args[2:])
	}
	c, err := dl.DialContext(ctx, nw, addr)
	if err != nil {
		return err
	}
	defer c.Close()
	cl := tls.Client(c, &tls.Config{
		Certificates:       []tls.Certificate{certs.Self.TLS()},
		ServerName:         name,
		InsecureSkipVerify: true,
	})
	if err = cl.HandshakeContext(ctx); err != nil {
		return err
	}
	cs := cl.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		return ErrNoPeer
	}
	pg := page.New()
	defer page.Free(pg)
	n, err := cl.Read(pg)
	if err != nil {
		return ErrNotOK
	}
	if string(pg[:n]) != "OK" {
		return ErrNotOK
	}
	p0 := cs.PeerCertificates[0]
	certs.RootCAs.Add(p0)
	certs.ClientCAs.Add(p0)
	return certs.Subscriptions.Add(headers, p0)
}
