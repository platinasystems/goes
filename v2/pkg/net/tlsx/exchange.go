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
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
)

var (
	ErrNoSubjectKeyId    = errors.New("missing subject-key-id")
	ErrNoPeer            = errors.New("no peer certificates")
	ErrNameOrSKINotFound = errors.New("name or subject-key-id not found")
	ErrSKINotFound       = errors.New("subject-key-id not found")
	ErrUnknownCommand    = errors.New("unknown command")

	ClientHeaders = make(map[string]string)
)

var Suppressed = []error{
	context.Canceled,
	net.ErrClosed,
	io.EOF,
	syscall.EPIPE,
}

func Exchange(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
) {
	defer wg.Done()
	tlsc, err := cert.ValErr()
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
	hn, err := host.Name.ValErr()
	if err != nil {
		panic(err)
	}
	path := selection.Path{hn}

	for c := range accept.With(ctx, ln, make(chan net.Conn, 4)) {
		wg.Add(1)
		tlsc := tls.Server(c, &tls.Config{
			Certificates: tlscs,
			ServerName:   sn,
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    clients,
		})
		go func(tlsc *tls.Conn) {
			wg.Done()
			defer tlsc.Close()

			if err = tlsc.HandshakeContext(ctx); err != nil {
				if suppress.Errors(err, Suppressed...) != nil {
					Elog(err)
				}
				return
			} else if true {
				// skip following log(s) if true
			} else if s := tlsc.RemoteAddr().String(); len(s) > 0 {
				Log(s)
			} else {
				Log(tlsc.LocalAddr())
			}
			err = service(ctx, tlsc, path, func(
				ctx context.Context,
				r io.Reader,
				w io.Writer,
				path selection.Path,
				args ...string,
			) error {
				switch args[0] {
				case "accept":
					return exAccept(ctx, tlsc)
				case "approve":
					return exRegistry(ctx, tlsc, args)
				case "clients":
					return exClients(ctx, tlsc)
				case "connect":
					args = args[1:]
					return exConnect(ctx, tlsc, args)
				case "deny":
					return exRegistry(ctx, tlsc, args)
				}
				return ErrUnknownCommand
			})
			if suppress.Errors(err, Suppressed...) != nil {
				Elog(err)
			}
		}(tlsc)
	}
}

// Ack provider accepting consumer requests then wait for ring with consumer
// SubjectKeyID consumer before copying from consumer to the provider.
func exAccept(ctx context.Context, host *tls.Conn) error {
	cs := host.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		return fmt.Errorf("accept: %w", ErrNoPeer)
	}
	enc := lv.NewEncoder(write.With(ctx, host))
	_, err := enc.Encode("OK", nil)
	if err != nil {
		return fmt.Errorf("accept ack: %w", err)
	}
	ski := hex.EncodeToString(cs.PeerCertificates[0].SubjectKeyId)
	guest, err := waitForGuest(ctx, host, ski)
	if err != nil {
		return fmt.Errorf("accept guest: %w", err)
	}
	io.Copy(host, guest)
	// Log("host<-guest done")
	return nil
}

// Ring provider with consumer SubjectKeyId then copy the provider to consumer.
func exConnect(ctx context.Context, guest *tls.Conn, args []string) (
	err error,
) {
	cs := guest.ConnectionState()
	if len(cs.PeerCertificates) == 0 {
		return ErrNoPeer
	}
	ski := hex.EncodeToString(cs.PeerCertificates[0].SubjectKeyId)
	if len(args) == 0 {
		return ErrNoSubjectKeyId
	}
	host, err := waitForHost(ctx, guest, args[0])
	if err != nil {
		return fmt.Errorf("host: %w", err)
	}
	if _, err := lv.NewEncoder(host).Encode("ring", ski, nil); err != nil {
		return fmt.Errorf("ring:%s: %w", ski, err)
	}
	io.Copy(guest, host)
	// Log("guest<-host done")
	return nil
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
				if err := certs.Clients.Add(
					ClientHeaders,
					cert,
				); err != nil {
					return fmt.
						Errorf("registry approve: %w",
							err)
				}
			}
			copy(Reg.l[i:], Reg.l[i+1:])
			Reg.l = Reg.l[:len(Reg.l)-1]
			continue skiloop
		}
		return fmt.Errorf("registry %s: %w", ski, ErrSKINotFound)
	}
	return nil
}
