// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/authorized"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/cert"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var SVC = "unspecified"

// Notify exchange that this accepts consumer requests; then wait for exchange
// rendezvous to service requests.
func Accept(
	ctx context.Context,
	wg *sync.WaitGroup,
	ex string,
	f selection.Func,
) {
	defer wg.Done()

	nw, addr, cfg, err := dialcfg(ex)
	if err != nil {
		panic(err)
	}

	path := selection.Path{host.Name.String()}

	if SVC, err = program.MainReference.Value(); err != nil {
		SVC = program.Base.String()
	}

	for {
		var dl net.Dialer
		c, err := dl.DialContext(ctx, nw, addr)
		if err != nil {
			if ctx.Err() != nil {
				return
			}
			if !errors.Is(err, fs.ErrNotExist) {
				Elog(err)
			}
			t := time.NewTimer(3 * time.Second)
			select {
			case <-t.C:
			case <-ctx.Done():
				if !t.Stop() {
					<-t.C
				}
				return
			}
			continue
		}

		cl := tls.Client(c, cfg)

		if err = cl.HandshakeContext(ctx); err != nil {
			cl.Close()
			if ctx.Err() == nil {
				Elog(err)
			}
			return
		}

		got := new(strings.Builder)
		if err = Req(ctx, cl, nil, got, "accept"); err != nil {
			cl.Close()
			if ctx.Err() != nil {
				return
			}
			Elog(err)
			continue
		}

		err = service(ctx, cl, path, func(
			ctx context.Context,
			r io.Reader,
			w io.Writer,
			path selection.Path,
			args ...string,
		) (err error) {
			switch args[0] {
			case "ring":
				err = ring(ctx, cl, args[1:]...)
				if err != nil {
					err = fmt.Errorf("%s: %w", args[1],
						err)
				}
			default:
				err = fmt.Errorf("%q: %w", args[0],
					ErrUnknownCommand)
			}
			return
		})
		if err != nil {
			cl.Close()
			if ctx.Err() != nil {
				return
			}
			continue

		}
		wg.Add(1)
		go func(cl *tls.Conn) {
			defer wg.Done()
			defer cl.Close()
			for {
				err := service(ctx, cl, path, f)
				if ctx.Err() != nil ||
					errors.Is(err, net.ErrClosed) ||
					errors.Is(err, io.EOF) {
					return
				}
			}
		}(cl)
	}
}

// Confim consumer authorization.
func ring(ctx context.Context, pr *tls.Conn, args ...string) (err error) {
	enc := lv.NewEncoder(write.With(ctx, pr))
	svc := []byte(SVC)
	if len(args) != 1 {
		err = ErrNoSubjectKeyId
	} else if ski := cert.SKI.String(); ski == args[0] {
		_, err = enc.Write(svc)
	} else if val, ok := authorized.Load(args[0]); ok && val {
		_, err = enc.Write(svc)
	} else {
		err = authorized.ErrUnauthorized
	}
	return
}
