// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/hex"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

type entry struct {
	cert *x509.Certificate
	ski  string
}

var cut = slice.Cut[entry]

var reg = struct {
	mutex   sync.Mutex
	entries []entry
}{}

func Append(c *x509.Certificate) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.entries = append(reg.entries, entry{
		c, hex.EncodeToString(c.SubjectKeyId),
	})
}

// Extract calls f() with each registered certificate and it's hex encoded
// subject-key-id (SKI) until f() returns true; then deregisters and returns
// that certficate and SKI. This returns nil if f() returned false with all
// certificates.
func Extract(f func(*x509.Certificate, string) bool) *x509.Certificate {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	for i, entry := range reg.entries {
		if f(entry.cert, entry.ski) {
			reg.entries = cut(reg.entries, uint(i), 1)
			return entry.cert
		}
	}
	return nil
}

// Range calls f() with each registered certificate and SKI until f() returns
// false.
func Range(f func(*x509.Certificate, string) bool) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	for _, entry := range reg.entries {
		if !f(entry.cert, entry.ski) {
			break
		}
	}
}

func Routine(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	addr := &net.TCPAddr{Port: port.Registry.Value()}
	svch := make(chan *tls.Conn, 4)

	wg.Add(1)
	go greet.Routine(ctx, wg, svch, addr, &tls.Config{
		Certificates: []tls.Certificate{certs.Self.TLS()},
		ServerName:   certs.Self.Name(),
		ClientAuth:   tls.RequireAnyClientCert,
	})
	for sv := range svch {
		sv.Write([]byte("OK"))
		Append(sv.ConnectionState().PeerCertificates[0])
		sv.Close()
	}
}
