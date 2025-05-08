// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcontext

import (
	"context"
	"errors"
	"os"
	"time"

	"golang.org/x/sys/unix"
)

// Non-Blocking Reader
type NBR struct {
	context.Context
	f *os.File
}

func WithNBR(ctx context.Context, f *os.File) NBR { return NBR{ctx, f} }

func (nbr NBR) Read(b []byte) (int, error) {
	fd := FD(nbr.f.Fd())
	unix.SetNonblock(fd, true)
	defer unix.SetNonblock(fd, false)
	for dur := RawTTYMinInterval; ; {
		n, err := nbr.f.Read(b)
		if err == nil || !errors.Is(err, unix.EAGAIN) {
			return n, err
		}
		t := time.NewTimer(dur)
		select {
		case <-t.C:
			if dur < RawTTYMaxInterval {
				if dur *= 2; dur > RawTTYMaxInterval {
					dur = RawTTYMaxInterval
				}
			}
		case <-nbr.Done():
			if !t.Stop() {
				<-t.C
			}
			return 0, context.Canceled
		}
	}
}
