// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package registry

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var reg = struct {
	mutex sync.Mutex
	certs []*x509.Certificate
}{}

func Append(c *x509.Certificate) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	reg.certs = append(reg.certs, c)
}

// Extract calls f() with each registered certificate until f() returns true;
// then deregisters and returns that certficate. This returns nil if f()
// returned false with all certificates.
func Extract(f func(*x509.Certificate) bool) (x *x509.Certificate) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	for i, c := range reg.certs {
		if f(c) {
			x = c
			copy(reg.certs[i:], reg.certs[i+1:])
			reg.certs = reg.certs[:len(reg.certs)-1]
			break
		}
	}
	return
}

// Range calls f() with each registered certificate until f() returns false.
func Range(f func(*x509.Certificate) bool) {
	reg.mutex.Lock()
	defer reg.mutex.Unlock()
	for _, c := range reg.certs {
		if !f(c) {
			break
		}
	}
}

func Routine(
	ctx context.Context,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	addr := &net.TCPAddr{Port: port.Registry.ValueContext(ctx)}
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
