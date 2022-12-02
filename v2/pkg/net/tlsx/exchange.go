// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"sync"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/foreclose"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/bridge"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

var SubscriberHeaders = make(map[string]string)

var Suppressed = []error{
	context.Canceled,
	net.ErrClosed,
	io.EOF,
	syscall.EPIPE,
}

func Lookup(ex string) (name, ski string, c *x509.Certificate) {
	c = cert.Value().Leaf
	ski = cert.SKI()
	if len(ex) == 0 || ex == ski {
		name = c.DNSNames[0]
		return
	}
	for _, s := range c.DNSNames {
		if ex == s {
			name = s
			return
		}
	}
	name, ski, c = certs.Subscriptions.Lookup(ex)
	return
}

func Exchange(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
	m selection.Map,
) {
	defer wg.Done()

	excert := cert.Value()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go foreclose.With(cctx, wg, ln)

	conch := make(chan net.Conn, 4)

	wg.Add(1)
	go accept.With(wg, ln, conch)

	br := bridge.New(cctx, wg, excert.Leaf.DNSNames[0])

	for c := range conch {
		sv := tls.Server(c, &tls.Config{
			Certificates: []tls.Certificate{excert},
			ServerName:   excert.Leaf.DNSNames[0],
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certs.ClientCAs.Clone(),
		})
		wg.Add(1)
		go exService(cctx, wg, sv, m, br)
	}
}

func exService(
	ctx context.Context,
	wg *sync.WaitGroup,
	c *tls.Conn,
	m selection.Map,
	br *bridge.Bridge,
) {
	wg.Done()
	defer c.Close()

	err := c.HandshakeContext(ctx)
	if err != nil {
		if suppress.Errors(err, Suppressed...) != nil {
			style.Error(err)
		}
		return
	}
	err = service(ctx, c, func(
		ctx context.Context,
		r io.Reader,
		w io.Writer,
		args ...string,
	) error {
		cs := c.ConnectionState()
		if len(cs.PeerCertificates) == 0 {
			return ErrNoPeer
		}

		path := []string{host.Name.Value()}

		if len(args) == 0 {
			return selection.ErrIncomplete
		}
		switch args[0] {
		case "approve":
			return exRegistry(ctx, c, args)
		case "complete":
			path = append(path, args[0])
			args = args[1:]
			if len(args) > 0 && args[0] == "help" {
				args = args[1:]
			}
		case "deny":
			return exRegistry(ctx, c, args)
		case "help":
			path = append(path, args[0])
			args = args[1:]
		case "join":
			enc := lv.NewEncoder(write.With(ctx, c))
			enc.Encode("OK", nil)
			br.Join(ctx, c)
			return net.ErrClosed
		case "subscribers":
			return exSubscribers(ctx, c)
		}
		return m.Select(ctx, r, w, append(path, "host"), args...)
	})
	if suppress.Errors(err, Suppressed...) != nil {
		style.Error(err)
	}
}

func exRegistry(ctx context.Context, c *tls.Conn, args []string) error {
	Reg.mutex.Lock()
	defer Reg.mutex.Unlock()
	if len(args) <= 1 {
		enc := lv.NewEncoder(c)
		for _, cert := range Reg.l {
			enc.Encode(hex.EncodeToString(cert.SubjectKeyId),
				": ", cert.DNSNames, "\n")
		}
		enc.Encode(nil)
		return nil
	}
	approve := args[0] == "approve"
skiloop:
	for _, ski := range args[1:] {
		for i, cert := range Reg.l {
			if hex.EncodeToString(cert.SubjectKeyId) != ski {
				continue
			}
			if approve {
				if err := certs.Subscribers.Add(
					SubscriberHeaders,
					cert,
				); err != nil {
					return fmt.Errorf("approve: %w", err)
				}
				certs.ClientCAs.Add(cert)
			}
			copy(Reg.l[i:], Reg.l[i+1:])
			Reg.l = Reg.l[:len(Reg.l)-1]
			continue skiloop
		}
		return fmt.Errorf("registry %s: %w", ski, ErrSKINotFound)
	}
	return nil
}

func exSubscribers(ctx context.Context, c *tls.Conn) error {
	enc := lv.NewEncoder(c)
	fmt.Fprint(enc, certs.Subscribers)
	return nil
}
