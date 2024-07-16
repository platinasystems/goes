// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcontext

import (
	. "context"
	"errors"
	"net"
	"os"
	"time"
)

const (
	PollMinInterval = 10 * time.Millisecond
	PollMaxInterval = 250 * time.Millisecond
)

type ReadDeadliner interface {
	Read([]byte) (int, error)
	SetReadDeadline(time.Time) error
}

// This is a contextual reader that uses its SetReadDeadline to progressively
// poll the context Err and Done.
type ContextReadDeadliner struct {
	Context
	r ReadDeadliner
}

func WithReadDeadliner(ctx Context, v ReadDeadliner) ContextReadDeadliner {
	return ContextReadDeadliner{ctx, v}
}

func (v ContextReadDeadliner) Read(b []byte) (int, error) {
	defer v.r.SetReadDeadline(time.Time{})
	for dur := PollMinInterval; ; {
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
		if dur < PollMaxInterval {
			if dur *= 2; dur > PollMaxInterval {
				dur = PollMaxInterval
			}
		}
	}
}

type ReadFromDeadliner interface {
	ReadFrom([]byte) (int, net.Addr, error)
	SetReadDeadline(time.Time) error
}

// This is a contextual packet reader that uses the embedded SetReadDeadline to
// progressively poll the context Err and Done.
type ContextReadFromDeadliner struct {
	Context
	r ReadFromDeadliner
}

func WithReadFromDeadliner(
	ctx Context,
	v ReadFromDeadliner,
) ContextReadFromDeadliner {
	return ContextReadFromDeadliner{ctx, v}
}

func (v ContextReadFromDeadliner) ReadFrom(b []byte) (int, net.Addr, error) {
	defer v.r.SetReadDeadline(time.Time{})
	for dur := PollMinInterval; ; {
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
		if dur < PollMaxInterval {
			if dur *= 2; dur > PollMaxInterval {
				dur = PollMaxInterval
			}
		}
	}
}
