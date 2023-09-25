// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"io"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/context/poll"
	"github.com/platinasystems/goes/v2/pkg/context/rawtty"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

// Send request through connection; then write response and return any error.
//
// The request is made with successive link-value encoded args; then any
// "<<<<" prefaced input data; and concludes with an empty value (break).
//
// The response output is the decoded received data up to break; then returns
// following error or nil if break.
//
//	> REQ [ARG]... ["<<<<" INPUT...] ""
//	< OUTPUT... {NACK | ""}
func Exec(
	ctx context.Context,
	conn net.Conn,
	r io.Reader,
	w io.Writer,
	args ...any,
) error {
	var wg sync.WaitGroup
	defer wg.Wait()

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	dec := lv.NewDecoder(poll.WithReader(cctx, conn))
	enc := lv.NewEncoder(write.With(cctx, conn))

	if args[0] == "pty" {
		tty, err := rawtty.With(cctx)
		if err != nil {
			return err
		}
		defer tty.Close()
		r, w = tty, tty
	}

	if _, err := enc.Encode(args...); err != nil {
		return err
	}

	if r != nil {
		if _, err := enc.Encode("<<<<"); err != nil {
			return err
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			pg := page.New()
			defer page.Free(pg)
			defer enc.Encode(nil)
			for {
				n, err := r.Read(pg)
				if err != nil || n == 0 {
					break
				}
				if _, err = enc.Write(pg[:n]); err != nil {
					break
				}
				if cctx.Err() != nil {
					break
				}
			}
		}()
	} else if _, err := enc.Encode(nil); err != nil {
		return err
	}

	ob := page.New()
	defer page.Free(ob)

	for {
		if n, err := dec.Read(ob); err != nil {
			return err
		} else if n == 0 {
			break
		} else {
			w.Write(ob[:n])
		}
	}

	return nil
}
