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

type ReadSetReadDeadliner interface {
	io.Reader
	SetReadDeadline(time.Time) error
}

// This is a contextual reader that uses the embedded SetReadDeadline to
// progressively poll the context Err and Done.
type Poll struct {
	context.Context
	r ReadSetReadDeadliner
}

func With(ctx context.Context, r ReadSetReadDeadliner) Poll {
	return Poll{ctx, r}
}

func (p Poll) Read(b []byte) (int, error) {
	if p.r == nil {
		return 0, io.EOF
	}
	defer p.r.SetReadDeadline(time.Time{})
	for dur := MinInterval; ; {
		err := p.Err()
		if err != nil {
			return 0, err
		}
		err = p.r.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			return 0, err
		}
		n, err := p.r.Read(b)
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
