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
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/read"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/encoding/lv"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/os/page"
)

var (
	ErrEmptyRequest    = errors.New("empty request")
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
		in   input
		t0   time.Time
	)
	dec := lv.NewDecoder(read.With(ctx, conn))
	enc := lv.NewEncoder(write.With(ctx, conn))
	pg := page.New()
	defer page.Free(pg)

	for i := 0; ; {
		err := conn.SetReadDeadline(time.Now().Add(ServiceReadTimeout))
		if err != nil {
			return err
		}
		n, err := dec.Read(pg[i:])
		if err != nil {
			if errors.Is(err, os.ErrDeadlineExceeded) {
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}
				continue
			}
			return err
		}
		if err = conn.SetReadDeadline(t0); err != nil {
			return err
		}
		if n == 0 {
			break
		} else if s := string(pg[i : i+n]); s == WithInput {
			in.r = dec
			break
		} else {
			args = append(args, s)
			i += n
		}
	}
	err := ErrEmptyRequest
	if len(args) > 0 {
		err = f(ctx, &in, enc, path, args...)
		in.flush()
	}
	switch {
	case ctx.Err() != nil:
	case errors.Is(err, net.ErrClosed):
	case errors.Is(err, io.EOF):
	case err == nil:
		err = enc.Break()
	default:
		enc.Nack(fmt.Errorf("%v: %w", path, err))
	}
	return err
}
