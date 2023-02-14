// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package service

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/net/foreclose"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/bridge"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/input"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/lease"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/registry"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var (
	ErrIncomplete = errors.New("incomplete")
	Suppressed    = []error{
		context.Canceled,
		net.ErrClosed,
		io.EOF,
	}
)

var Selection = map[string]any{
	"approve": registry.Admin,
	"deny":    registry.Admin,
	"show": map[string]any{
		"build":       program.Build,
		"main":        program.Main,
		"registry":    registry.Show,
		"subscribers": certs.Subscribers,
		"tenants":     lease.MarshalText,
	},
}

// Listen and accept connections that are then passed to Service routine.
func Routine(ctx context.Context, wg *sync.WaitGroup, port uint) {
	defer wg.Done()
	defer style.ShortFile.Errata.Recovery()

	conch := make(chan net.Conn, 4)

	addr := &net.TCPAddr{Port: int(port)}
	ln, err := net.Listen(addr.Network(), addr.String())
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go foreclose.Routine(ctx, wg, ln)

	wg.Add(1)
	go accept.Routine(wg, conch, ln)

	for {
		select {
		case <-ctx.Done():
			return
		case conn := <-conch:
			wg.Add(1)
			go Service(ctx, wg, conn)
		}
	}
}

// Service connections by parsing TLV request arguements and input upto the
// zero length Break; then calls the selected function with a reader that LV
// decodes any input; a writer with LV data encoding to connection; and the
// decoded arguments.  If that function succeeds, this sends the zero length
// Break to the connection; otherwise, this sends an encoded Nack.
func Service(ctx context.Context, wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	defer conn.Close()
	defer style.ShortFile.Errata.Recovery()

	ra := conn.RemoteAddr().String()
	dec := lv.NewDecoder(poll.With(ctx, conn))
	enc := lv.NewEncoder(write.With(ctx, conn))

	pg := page.New()
	defer page.Free(pg)

serviceloop:
	for {
		var (
			args []string
			err  error
			n    int
			iowg sync.WaitGroup
		)

		r := io.LimitReader(nil, 0)
		w := io.Writer(enc)
		cctx, cancel := context.WithCancel(ctx)

		for i := 0; ; {
			if n, err = dec.Read(pg[i:]); err != nil {
				if suppress.Errors(err, Suppressed...) != nil {
					panic(err)
				}
				return
			} else if n == 0 {
				if len(args) == 0 {
					panic(ErrIncomplete)
				}
				break
			} else if s := string(pg[i : i+n]); s != "<<<<" {
				args = append(args, s)
				i += n
			} else if len(args) == 0 {
				panic(ErrIncomplete)
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
		}
		if len(args) == 0 {
			enc.Encode(ErrIncomplete)
			continue serviceloop
		}
		if tlsc, ok := conn.(*tls.Conn); ok {
			switch args[0] {
			case "join":
				bridge.Join(cctx, tlsc, args[1:])
				return
			case "pty":
				r = dec
			}
			err = goes.Select(cctx, r, w, append(path, ra),
				Selection, args...)
		} else if args[0] == "subscribe" {
			err = registry.Subscribe(ctx, r, w, path, args[1:]...)
		} else {
			enc.Encode(nil)
			sv, err := greet.Client(ctx, conn)
			if err != nil {
				panic(err)
			}
			dec = lv.NewDecoder(poll.With(ctx, sv))
			enc = lv.NewEncoder(write.With(ctx, sv))
			conn = sv
			continue serviceloop
		}
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
