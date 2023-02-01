// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exchange

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/netip"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/exchange/bridge"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/input"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/port"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/remote"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	ErrMisconfigured = errors.New("misconfigured")
	ErrEmptyRequest  = errors.New("empty request")
	Suppressed       = []error{
		context.Canceled,
		net.ErrClosed,
		io.EOF,
	}
)

type T struct {
	cfg       tls.Config
	Bridge    bridge.T
	Selection map[string]any
}

func (t *T) Configure(args []string) ([]string, error) {
	var leasing netip.Prefix
	fs := flags.New()
	bflag := fs.Bool("b", false, "bridge")
	fs.TextVar(&leasing, "l", leasing, "lease prefix [ip/bits]")
	if t.Selection == nil {
		return args, ErrMisconfigured
	}
	err := fs.Parse(args)
	if err != nil {
		return nil, err
	}
	if *bflag {
		t.Bridge.Configure(leasing)
	}
	t.cfg.Certificates = []tls.Certificate{certs.Self.TLS()}
	t.cfg.ServerName = certs.Self.Name()
	t.cfg.ClientAuth = tls.RequireAndVerifyClientCert
	return fs.Args(), nil
}

func (t *T) Routine(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	addr := &net.TCPAddr{Port: port.Exchange.Value()}

	svch := make(chan *tls.Conn, 4)

	wg.Add(1)
	go greet.Routine(ctx, wg, svch, addr, &t.cfg)

	if t.Bridge.Enabled {
		wg.Add(1)
		go t.Bridge.Routine(ctx, wg)
	}

	for sv := range svch {
		wg.Add(1)
		go t.service(cctx, wg, sv)
	}
}

// Service connection by parsing TLV request arguements and input upto the zero
// length Break.  This calls f() with a reader that LV decodes any input; a
// writer with LV data encoding to connection; and the decoded arguments.  If
// f() succeeds, this sends the zero length Break to the connection; otherwise,
// this sends an encoded Nack.
func (t *T) service(
	ctx context.Context,
	wg *sync.WaitGroup,
	sv *tls.Conn,
) {
	wg.Done()
	defer sv.Close()

	ra := remote.Addr(sv)
	dec := lv.NewDecoder(poll.With(ctx, sv))
	enc := lv.NewEncoder(write.With(ctx, sv))
	r := io.LimitReader(nil, 0)
	w := io.Writer(enc)
	pg := page.New()
	defer page.Free(pg)

serviceLoop0:
	for {
		var (
			args []string
			err  error
			n    int
			iowg sync.WaitGroup
		)

		cctx, cancel := context.WithCancel(ctx)

		for i := 0; ; {
			if n, err = dec.Read(pg[i:]); err != nil {
				if suppress.Errors(err, Suppressed...) != nil {
					style.Error(err)
				}
				return
			} else if n == 0 {
				if len(args) == 0 {
					style.Error(ErrEmptyRequest)
					return
				}
				break
			} else if s := string(pg[i : i+n]); s != "<<<<" {
				args = append(args, s)
				i += n
			} else if len(args) == 0 {
				style.Error(ErrEmptyRequest)
				return
			} else if args[0] == "pty" {
				break
			} else {
				ir, flush := input.New(dec)
				r = ir
				iowg.Add(1)
				go func() {
					defer iowg.Done()
					<-cctx.Done()
					flush()
				}()
				break
			}
		}

		style.Note(ra, args)

		path := []string{host.Name.Value()}

		switch args[0] {
		case "complete":
			path = append(path, args[0])
			args = args[1:]
			if len(args) > 0 && args[0] == "help" {
				args = args[1:]
			}
		case "help":
			path = append(path, args[0])
			args = args[1:]
		case "join":
			t.Bridge.Join(ctx, sv, args[1:])
			continue serviceLoop0
		case "pty":
			r = dec
		}
		err = goes.Select(cctx, r, w, append(path, ra), t.Selection,
			args...)
		cancel()
		iowg.Wait()
		switch {
		case errors.Is(err, context.Canceled):
		case errors.Is(err, net.ErrClosed):
		case errors.Is(err, io.EOF):
		default:
			enc.Encode(err)
		}
	}
}
