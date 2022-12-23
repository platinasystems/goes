// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package greet

import (
	"context"
	"crypto/tls"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/foreclose"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/handshake"
)

func Routine(
	ctx context.Context,
	wg *sync.WaitGroup,
	svch chan<- *tls.Conn,
	addr *net.TCPAddr,
	cfg *tls.Config,
) {
	defer wg.Done()
	defer close(svch)

	ln, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		style.Error(err)
		return
	}

	conch := make(chan net.Conn, 4)

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(1)
	go foreclose.Routine(cctx, wg, ln)

	wg.Add(1)
	go accept.Routine(wg, conch, ln)

	wg.Add(1)
	go handshake.Routine(cctx, wg, svch, conch, cfg)

	<-ctx.Done()
}
