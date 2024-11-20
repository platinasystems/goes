// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnspkt

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/chunk"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
)

const Cap = 1232

var Pool = chunk.Pool{
	Cap: Cap,
}

var ResetDeadline time.Time

// TenaciousAsk resends the buffered query every 1 sec until it receives a
// response or context is cancelled.
// If successful, it returns the raw binary response within the same,
// probably expanded buffer.
func TenaciousAsk(ctx context.Context, conn net.Conn, buf []byte) (
	[]byte, error,
) {
	for {
		err := ctx.Err()
		if err != nil {
			return buf[:0], err
		}
		n, err := conn.Write(buf)
		if err != nil {
			return buf[:0], err
		} else if n != len(buf) {
			err = xerrors.Underrun("socket-write")
			return buf[:0], err
		}
		err = conn.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			return buf[:0], err
		}
		n, err = conn.Read(buf[:cap(buf)])
		conn.SetReadDeadline(ResetDeadline)
		if err == nil {
			return buf[:n], err
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return buf[:0], err
		}
	}
}

// TimeLimitedAsk is a `TenaciousAsk` with a deadlined context.
func TimeLimitedAsk(
	ctx context.Context, conn net.Conn, buf []byte, dur time.Duration,
) ([]byte, error) {
	dl := time.Now().Add(dur)
	cctx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()
	return TenaciousAsk(cctx, conn, buf)
}
