// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package poll

import (
	"context"
	"io"
	"net"
	"time"
)

const (
	MinInterval = 10 * time.Millisecond
	MaxInterval = 250 * time.Millisecond
)

type Reader interface {
	io.Reader
	SetReadDeadline(time.Time) error
}

// This is a contextual reader that uses the embedded SetReadDeadline to
// progressively poll the context Err and Done.
func With(ctx context.Context, r Reader) io.Reader {
	return cr{ctx, r}
}

type cr struct {
	c context.Context
	r Reader
}

func (cr cr) Read(b []byte) (int, error) {
	if cr.r == nil {
		return 0, io.EOF
	}
	defer cr.r.SetReadDeadline(time.Time{})
	for dur := MinInterval; ; {
		err := cr.c.Err()
		if err != nil {
			return 0, err
		}
		select {
		case <-cr.c.Done():
			return 0, context.Canceled
		default:
		}
		err = cr.r.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			return 0, err
		}
		n, err := cr.r.Read(b)
		if err == nil {
			return n, nil
		}
		if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				return n, operr
			}
		} else {
			return n, err
		}
		if dur < MaxInterval {
			if dur *= 2; dur > MaxInterval {
				dur = MaxInterval
			}
		}
	}
}
