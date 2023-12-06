// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"log"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/io/flusher"
	"github.com/platinasystems/goes/v2/pkg/net/accept"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/page"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var Selection = map[string]any{
	"approve": regAdmin,
	"deny":    regAdmin,
	"reserve": Reserve,
	"whois":   WhoIs,
	"show": map[string]any{
		"build":       program.Build,
		"main":        program.Main,
		"registry":    regShow,
		"subscribers": Subscribers,
		"tenants":     leaseMarshalText,
	},
}

// Listen and serve TCP/TLS connection while exchanging boxes received through
// non-nil packet listener.  This closes both listeners when context is done.
func Routine(
	ctx context.Context,
	wg *sync.WaitGroup,
	ln net.Listener,
	pc net.PacketConn,
) {
	var pcwg sync.WaitGroup

	defer wg.Done()

	if pc != nil {
		pcwg.Add(1)
		go xpacket(ctx, &pcwg, pc)
	}

	conch := make(chan net.Conn, 4)

	wg.Add(1)
	go accept.Routine(wg, conch, ln)

	for {
		select {
		case <-ctx.Done():
			ln.Close()
			if pc != nil {
				pcwg.Wait()
				pc.Close()
			}
			return
		case conn := <-conch:
			wg.Add(1)
			go service(ctx, wg, conn)
		}
	}
}

// Service connections by parsing TLV request arguements and input upto the
// zero length Break; then calls the selected function with a reader that LV
// decodes any input; a writer with LV data encoding to connection; and the
// decoded arguments.  If that function succeeds, this sends the zero length
// Break to the connection; otherwise, this sends an encoded Nack.
func service(ctx context.Context, wg *sync.WaitGroup, conn net.Conn) {
	defer wg.Done()
	defer conn.Close()

	ra := conn.RemoteAddr().String()
	dec := lv.NewDecoder(poll.WithReader(ctx, conn))
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

		ctx = ctxparm.Reader.With(ctx, io.LimitReader(nil, 0))
		ctx = ctxparm.Writer.With(ctx, io.Writer(enc))
		cctx, cancel := context.WithCancel(ctx)
		cctx = ctxparm.Strings.With(cctx, []string{host.Name()})

		for i := 0; ; {
			if n, err = dec.Read(pg[i:]); err != nil {
				log.Print(err)
				return
			} else if n == 0 {
				if len(args) == 0 {
					log.Print(goes.ErrIncomplete)
					return
				}
				break
			} else if s := string(pg[i : i+n]); s != "<<<<" {
				args = append(args, s)
				i += n
			} else if len(args) == 0 {
				log.Print(goes.ErrIncomplete)
				return
			} else if args[0] == "pty" {
				break
			} else {
				in, flush := flusher.New(dec)
				cctx = ctxparm.Reader.With(cctx, in)
				iowg.Add(1)
				go func() {
					defer iowg.Done()
					<-cctx.Done()
					flush()
				}()
				break
			}
		}

		if len(args) == 0 {
			enc.Encode(ErrIncomplete)
			continue serviceloop
		}
		if _, ok := conn.(*tls.Conn); ok {
			if args[0] == "pty " {
				cctx = ctxparm.Reader.With(cctx, dec)
			}
			cctx = ctxparm.AppendStringsIn(cctx, ra)
			cctx = ctxparm.Map.With(cctx, Selection)
			err = goes.Select(cctx, args...)
		} else {
			switch args[0] {
			case "join":
				ctxparm.AppendStringsIn(cctx, args[0])
				err = Join(cctx, conn, args[1:]...)
				if err == nil {
					return
				}
			case "subscribe":
				ctxparm.AppendStringsIn(cctx, args[0])
				err = regSubscribe(ctx, args[1:]...)
			case "tls":
				enc.Encode(nil)
				sv, err := greetClient(ctx, conn)
				if err != nil {
					log.Print(err)
					return
				}
				dec = lv.NewDecoder(poll.WithReader(ctx, sv))
				enc = lv.NewEncoder(write.With(ctx, sv))
				conn = sv
				continue serviceloop
			default:
				err = goes.ErrNotFound
			}
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
