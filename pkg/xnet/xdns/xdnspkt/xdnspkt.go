// Copyright © 2024-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdnspkt

import (
	"context"
	"errors"
	"net"
	"os"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns"
)

// When asked, resends the buffered query every 1 sec until response or context
// is cancelled.
func NewTenaciousAsk(udp net.Conn) xdns.Asker {
	return tenacious{udp}.Ask
}

// When asked, resends the buffered query with deadlined context.
func NewTimeLimitedAsk(udp net.Conn, limit time.Duration) xdns.Asker {
	return timeLimited{tenacious{udp}, limit}.Ask
}

var zero time.Time

type tenacious struct {
	net.Conn
}

type timeLimited struct {
	tenacious
	time.Duration
}

func (t tenacious) Ask(ctx context.Context, buf []byte) (
	[]byte, error,
) {
	for {
		err := ctx.Err()
		if err != nil {
			return buf[:0], err
		}
		n, err := t.Write(buf)
		if err != nil {
			return buf[:0], err
		} else if n != len(buf) {
			err = xerrors.Underrun("socket-write")
			return buf[:0], err
		}
		err = t.SetReadDeadline(time.Now().Add(time.Second))
		if err != nil {
			return buf[:0], err
		}
		n, err = t.Read(buf[:cap(buf)])
		t.SetReadDeadline(zero)
		if err == nil {
			return buf[:n], err
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return buf[:0], err
		}
	}
}

func (t timeLimited) Ask(ctx context.Context, buf []byte) ([]byte, error) {
	dl := time.Now().Add(t.Duration)
	cctx, cancel := context.WithDeadline(ctx, dl)
	defer cancel()
	return t.tenacious.Ask(cctx, buf)
}
