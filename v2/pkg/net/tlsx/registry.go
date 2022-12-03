// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/foreclose"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	Reg = struct {
		mutex sync.Mutex
		l     []*x509.Certificate
	}{}
	SubscriptionHeaders = make(map[string]string)
)

func Registry(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
) {
	defer wg.Done()

	tlsc := certs.Self.TLS()
	if tlsc.Leaf == nil {
		panic("nil leaf")
	}
	if len(tlsc.Leaf.DNSNames) == 0 {
		panic("no DNS names")
	}
	sn := tlsc.Leaf.DNSNames[0]

	wg.Add(1)
	go foreclose.With(ctx, wg, ln)

	conch := make(chan net.Conn, 4)

	wg.Add(1)
	go accept.With(wg, ln, conch)

	for c := range conch {
		sv := tls.Server(c, &tls.Config{
			Certificates: []tls.Certificate{tlsc},
			ServerName:   sn,
			ClientAuth:   tls.RequireAnyClientCert,
		})
		wg.Add(1)
		go func(sv *tls.Conn) {
			wg.Done()
			defer sv.Close()

			err := sv.HandshakeContext(ctx)
			if err != nil {
				if err = suppress.Errors(err,
					context.Canceled,
					net.ErrClosed,
				); err != nil {
					style.Error(err)
				}
				return
			}
			cs := sv.ConnectionState()
			if len(cs.PeerCertificates) == 0 {
				style.Error(ErrNoPeer)
				return
			}
			sv.Write([]byte("OK"))
			Reg.mutex.Lock()
			Reg.l = append(Reg.l, cs.PeerCertificates[0])
			Reg.mutex.Unlock()
		}(sv)
	}
}

func Subscribe(ctx context.Context, args []string) (err error) {
	var name, addr string
	var dl net.Dialer
	nw := "tcp"
	switch len(args) {
	case 0:
		return ErrMissingAddress
	case 1:
		if i := strings.LastIndex(args[0], ":"); i < 0 {
			err = ErrNotNamePort
			return
		} else if i > 0 {
			addr = args[0]
			name = args[0][:i]
		} else if name, err = host.Name.ValErr(); err != nil {
			return
		} else {
			addr = args[0]
		}
	case 2:
		name = args[0]
		addr = args[1]
	default:
		err = fmt.Errorf("%w: %v", ErrUnexpectedArgs, args[2:])
		return
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
	return certs.Subscriptions.Add(SubscriptionHeaders, p0)
}
