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
	"net"
	"sync"
	"time"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/nbr"
	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	ErrEmptyRequest    = errors.New("empty request")
	ErrExit            = errors.New("exit")
	ServiceReadTimeout = 3 * time.Second
)

// Service connection by parsing TLV request arguements and input upto the zero
// length Break.  This calls f() with a reader that LV decodes any input; a
// writer that LV encodes date to connection; and the decoded arguments.  If
// f() succeeds, this sends the zero length Break to the connection; otherwise,
// this sendsan encoded Nack.
func service(
	ctx context.Context,
	conn *tls.Conn,
	path selection.Path,
	f selection.Func,
) error {
	var (
		args []string
		wg   sync.WaitGroup
	)

	defer wg.Wait()

	cctx, cancel := context.WithCancel(ctx)
	dec := lv.NewDecoder(poll.With(cctx, conn))
	enc := lv.NewEncoder(write.With(cctx, conn))
	r := io.LimitReader(nil, 0)
	w := io.Writer(enc)
	pg := page.New()
	defer page.Free(pg)

	var ispty bool
	for i := 0; ; {
		if n, err := dec.Read(pg[i:]); err != nil {
			return fmt.Errorf("decode: %w", err)
		} else if n == 0 {
			if len(args) == 0 {
				return ErrEmptyRequest
			}
			break
		} else if s := string(pg[i : i+n]); s == InputTag {
			if len(args) == 0 {
				return ErrEmptyRequest
			}
			if args[0] != "pty" {
				ir, flush := newinput(dec)
				r = ir
				wg.Add(1)
				go func() {
					defer wg.Done()
					<-cctx.Done()
					flush()
				}()
			} else if n := len(args); n < 6 {
				return fmt.Errorf("missing %s", []string{
					"pty",
					"<rows>",
					"<cols>",
					"<x-pixels>",
					"<y-pixels>",
					"<req>",
				}[n])
			} else {
				var ws pty.Winsize
				ispty = true
				args = args[1:]
				for i, pu := range []*uint16{
					&ws.Rows,
					&ws.Cols,
					&ws.X,
					&ws.Y,
				} {
					_, err = fmt.Sscan(args[i], pu)
					if err != nil {
						return fmt.Errorf("%q: %w",
							args[i], err)
					}
				}
				args = args[4:]
				ptmx, tty, err := pty.Open()
				if err != nil {
					return fmt.Errorf("pty: %w", err)
				}
				if err = pty.Setsize(tty, &ws); err != nil {
					ptmx.Close()
					return fmt.Errorf("resize: %w", err)
				}
				r, w = tty, tty
				ir, flush := newinput(dec)
				wg.Add(1)
				go func() {
					io.Copy(ptmx, ir)
					ptmx.Close()
					flush()
					// Log("ptmx<-input done")
					wg.Done()
				}()
				wg.Add(1)
				go func() {
					io.Copy(enc, nbr.With(cctx, ptmx))
					// Log("conn<-ptmx done")
					wg.Done()
				}()
			}
			break
		} else {
			args = append(args, s)
			i += n
		}
	}
	err := f(cctx, r, w, path, args...)
	switch {
	case errors.Is(err, context.Canceled):
	case errors.Is(err, net.ErrClosed):
	case errors.Is(err, io.EOF):
	case err == nil:
		if ispty {
			err = ErrExit
		}
		enc.Encode(err)
	default:
		err = fmt.Errorf("service %v: %w", path, err)
		enc.Encode(err)
	}
	cancel()
	// defer Log(os.Getpid(), err, "\r")
	return err
}
