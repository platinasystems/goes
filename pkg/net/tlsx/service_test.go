// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

// FIXME
package tlsx

import (
	"context"
	"encoding/pem"
	"errors"
	"io"
	"net"
	"os/signal"
	"sync"
	"testing"
	"time"

	"github.com/platinasystems/goes/v2/pkg/crypto/cipher/box"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func Test(t *testing.T) {
	const (
		duration = 30 * time.Second
		period   = 3 * time.Second
	)
	var (
		clients [5]*tlsx.Client
		lc      net.ListenConfig
		wg      sync.WaitGroup
	)

	ctx, cancel := context.WithCancel(context.Background())
	ctx, stop := signal.NotifyContext(ctx, termination.Signals...)
	defer stop()
	defer wg.Wait()

	ln, err := lc.Listen(ctx, "tcp", "127.0.0.1:8003")
	if err != nil {
		panic(err)
	}
	pc, err := lc.ListenPacket(ctx, "udp", "127.0.0.1:8003")
	if err != nil {
		ln.Close()
		panic(err)
	}

	wg.Add(1)
	go Routine(ctx, &wg, ln, pc)

	for i := range clients[:] {
		cl, err := tlsx.NewClient()
		if err != nil {
			panic(err)
		}
		req := pem.EncodeToMemory(cl.Exchange.PublicKey.Local)
		rsp, err := exchange.reserve(uint32(i), req)
		if err != nil {
			panic(err)
		}
		if err = cl.Peer(rsp); err != nil {
			panic(err)
		}
		clients[i] = cl
	}

	t.Log("run for", duration, "...")
	rt := time.NewTimer(duration)

	for i, cl := range clients {
		network := map[int]string{0: "udp", 1: "tcp"}[i&1]
		conn, err := tlsx.Dial(ctx, network, "127.0.0.1:8003")
		if err != nil {
			panic(err)
		}
		if network == "tcp" {
			err = tlsx.Exec(ctx, conn, nil, io.Discard,
				"join", cl.Confirmation)
			if err != nil {
				panic(err)
			}
		} else if _, err = cl.Confirmation.WriteTo(conn); err != nil {
			panic(err)
		}
		wg.Add(1)
		go testrx(t, ctx, &wg, conn, cl)
		wg.Add(1)
		go cl.KeepAlive(ctx, &wg, conn, period)
	}

	select {
	case <-ctx.Done():
		if !rt.Stop() {
			<-rt.C
		}
	case <-rt.C:
		cancel()
	}
}

func testrx(
	t *testing.T,
	ctx context.Context,
	wg *sync.WaitGroup,
	conn net.Conn,
	cl *tlsx.Client,
) {
	var err error

	defer wg.Done()
	defer conn.Close()

	bx := box.New()
	defer bx.Recycle()

	for {
		bx, err = bx.Receive(ctx, conn, cl.Exchange)
		if ctx.Err() != nil {
			break
		}
		if err != nil {
			if errors.Is(err, context.Canceled) {
				break
			} else {
				t.Error(err)
				break
			}
		}
		from := bx.FromWhom()
		if len(bx.Contents()) == 0 {
			t.Logf("%d<-%d: keepalive", cl.ID, from)
			if remote, ok := cl.Remote(from); !ok {
				m, err := exchange.whois(from)
				if err != nil {
					t.Error(err)
					continue
				}
				remote, err = cl.AddRemote(from, m.PublicKey)
				if err != nil {
					t.Error(err)
					continue
				}
			} else {
				bx = bx.Empty().Append("hello").
					To(from).From(cl.ID).
					Close(remote).Seal(cl.Exchange)
				conn.Write(bx)
			}
		} else if remote, ok := cl.Remote(from); !ok {
			t.Errorf("%d<-%d: indecipherable", cl.ID, from)
		} else if bx, err = bx.Open(remote); err != nil {
			t.Errorf("%d<-%d: %v", cl.ID, from, err)
		} else {
			t.Logf("%d<-%d: %q", cl.ID, from, bx.Contents())
		}
	}
}
