// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package handshake

import (
	"context"
	"crypto/tls"
	"io"
	"net"
	"sync"
	"syscall"

	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var Suppressed = []error{
	context.Canceled,
	net.ErrClosed,
	io.EOF,
	syscall.EPIPE,
}

// Continually handshake connections from conch and forward validated peers
// until conch is closed or context is done.
func Routine(
	ctx context.Context,
	wg *sync.WaitGroup,
	svch chan<- *tls.Conn,
	conch <-chan net.Conn,
	cfg *tls.Config,
) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case c, ok := <-conch:
			if !ok {
				return
			}
			cfg.ClientCAs = certs.ClientCAs.Clone()
			sv := tls.Server(c, cfg)
			err := sv.HandshakeContext(ctx)
			if err != nil {
				style.Error(err)
				sv.Close()
			} else {
				cs := sv.ConnectionState()
				if len(cs.PeerCertificates) == 0 {
					style.Error("no peer certificate")
					sv.Close()
				} else {
					svch <- sv
				}
			}
		}
	}
}
