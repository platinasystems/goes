// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package poll

import (
	"context"
	"errors"
	"net"
	"os"
	"time"
)

const (
	MinInterval = 10 * time.Millisecond
	MaxInterval = 250 * time.Millisecond
)

type Reader interface {
	Read([]byte) (int, error)
	SetReadDeadline(time.Time) error
}

// This is a contextual reader that uses the embedded SetReadDeadline to
// progressively poll the context Err and Done.
type ContextReader struct {
	context.Context
	r Reader
}

func WithReader(ctx context.Context, v Reader) ContextReader {
	return ContextReader{ctx, v}
}

func (v ContextReader) Read(b []byte) (int, error) {
	defer v.r.SetReadDeadline(time.Time{})
	for dur := MinInterval; ; {
		err := v.Err()
		if err != nil {
			return 0, err
		}
		err = v.r.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			return 0, err
		}
		n, err := v.r.Read(b)
		if err == nil {
			return n, nil
		}
		if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				return n, operr
			}
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return n, err
		}
		if dur < MaxInterval {
			if dur *= 2; dur > MaxInterval {
				dur = MaxInterval
			}
		}
	}
}

type ReadFromer interface {
	ReadFrom([]byte) (int, net.Addr, error)
	SetReadDeadline(time.Time) error
}

// This is a contextual packet reader that uses the embedded SetReadDeadline to
// progressively poll the context Err and Done.
type ContextReadFromer struct {
	context.Context
	r ReadFromer
}

func WithReadFromer(ctx context.Context, v ReadFromer) ContextReadFromer {
	return ContextReadFromer{ctx, v}
}

func (v ContextReadFromer) ReadFrom(b []byte) (int, net.Addr, error) {
	defer v.r.SetReadDeadline(time.Time{})
	for dur := MinInterval; ; {
		err := v.Err()
		if err != nil {
			return 0, nil, err
		}
		err = v.r.SetReadDeadline(time.Now().Add(dur))
		if err != nil {
			return 0, nil, err
		}
		n, a, err := v.r.ReadFrom(b)
		if err == nil {
			return n, a, nil
		}
		if operr, ok := err.(*net.OpError); ok {
			if !operr.Timeout() {
				return n, a, operr
			}
		} else if !errors.Is(err, os.ErrDeadlineExceeded) {
			return n, nil, err
		}
		if dur < MaxInterval {
			if dur *= 2; dur > MaxInterval {
				dur = MaxInterval
			}
		}
	}
}
