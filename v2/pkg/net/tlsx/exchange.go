// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/ipc"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var (
	ErrNoSubjectKeyId    = errors.New("missing subject-key-id")
	ErrNoPeer            = errors.New("no peer certificates")
	ErrNameOrSKINotFound = errors.New("name or subject-key-id not found")
	ErrSKINotFound       = errors.New("subject-key-id not found")
	ErrUnknownCommand    = errors.New("unknown command")

	ClientHeaders = make(map[string]string)

	IPC = ipc.Preface(program.Base.String())
)

func dialcfg(ex string) (nw, addr string, cfg *tls.Config, err error) {
	var sn string
	hn, err := host.Name.Value()
	if err != nil {
		return
	}
	nw = "tcp"
	if len(ex) == 0 {
		nw = ipc.Network
		sn = hn
		if addr, err = IPC.Address(); err != nil {
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
	tlsc, err := cert.Value()
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

func Exchange(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
) {
	defer wg.Done()

	tlsc, err := cert.Value()
	if err != nil {
		panic(err)
	}
	tlscs := []tls.Certificate{tlsc}
	if tlsc.Leaf == nil {
		panic("nil leaf")
	}
	if len(tlsc.Leaf.DNSNames) == 0 {
		panic("no DNS names")
	}
	sn := tlsc.Leaf.DNSNames[0]
	clients, err := certs.Clients.Pool()
	if err != nil {
		panic(err)
	}
	hn, err := host.Name.Value()
	if err != nil {
		panic(err)
	}
	path := selection.Path{hn}

	for c := range accept.With(ctx, ln, make(chan net.Conn, 4)) {
		wg.Add(1)
		sv := tls.Server(c, &tls.Config{
			Certificates: tlscs,
			ServerName:   sn,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clients,
		})
		go func(sv *tls.Conn) {
			wg.Done()
			defer sv.Close()

			if err := sv.HandshakeContext(ctx); err != nil {
				if suppress.Errors(err,
					context.Canceled,
					net.ErrClosed,
					io.EOF,
				) != nil {
					Elog(err)
				}
				return
			}
			service(ctx, sv, path, func(
				ctx context.Context,
				r io.Reader,
				w io.Writer,
				path selection.Path,
				args ...string,
			) (err error) {
				switch args[0] {
				case "accept":
					err = exAccept(ctx, sv)
				case "approve":
					err = exRegistry(ctx, sv, args)
				case "clients":
					err = exClients(ctx, sv)
				case "connect":
					args = args[1:]
					err = exConnect(ctx, sv, args)
				case "deny":
					err = exRegistry(ctx, sv, args)
				default:
					err = ErrUnknownCommand
				}
				return
			})
		}(sv)
	}
}

// Ack provider accepting consumer requests then wait for ring with consumer
// SubjectKeyID consumer before copying from consumer to the provider.
func exAccept(ctx context.Context, pr *tls.Conn) error {
	cs := pr.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		return ErrNoPeer
	}
	enc := lv.NewEncoder(pr)
	ack(enc, "OK")
	ski := hex.EncodeToString(cs.PeerCertificates[0].SubjectKeyId)
	cn, err := WaitForConsumer(ctx, pr, ski)
	if err != nil {
		if suppress.Errors(err,
			context.Canceled,
			net.ErrClosed,
			io.EOF,
		) != nil {
			Elog(err)
		}
		return err
	}
	io.Copy(pr, cn)
	return context.Canceled
}

// Ring provider with consumer SubjectKeyId then copy the provider to consumer.
func exConnect(ctx context.Context, cn *tls.Conn, args []string) error {
	cs := cn.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		return ErrNoPeer
	}
	ski := hex.EncodeToString(cs.PeerCertificates[0].SubjectKeyId)
	if len(args) == 0 {
		return ErrNoSubjectKeyId
	}
	pr, err := WaitForProvider(ctx, cn, args[0])
	if err != nil {
		return err
	}
	defer pr.Close()
	enc := lv.NewEncoder(pr)
	req(enc, "ring", ski, nil)
	io.Copy(cn, pr)
	return context.Canceled
}

func exClients(ctx context.Context, c *tls.Conn) error {
	enc := lv.NewEncoder(c)
	certs.Clients.Range(func(
		headers certs.Headers,
		client *x509.Certificate,
	) bool {
		certs.Fsequent(enc, headers, client)
		return true
	})
	return nil
}

func exRegistry(ctx context.Context, c *tls.Conn, args []string) error {
	Reg.mutex.Lock()
	defer Reg.mutex.Unlock()
	if len(args) <= 1 {
		enc := lv.NewEncoder(c)
		for _, cert := range Reg.l {
			fmt.Fprint(enc, hex.EncodeToString(cert.SubjectKeyId),
				": ", cert.DNSNames, "\n")
		}
		enc.Break()
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
				if err := certs.Clients.Add(
					ClientHeaders,
					cert,
				); err != nil {
					return err
				}
			}
			copy(Reg.l[i:], Reg.l[i+1:])
			Reg.l = Reg.l[:len(Reg.l)-1]
			continue skiloop
		}
		return fmt.Errorf("%s: %w", ski, ErrSKINotFound)
	}
	return nil
}
