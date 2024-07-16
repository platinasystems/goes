// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xcontext

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type Printer interface{ Print(...any) }

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

// Redirect opens an os.Pipe and, after starting a goroutine to Print the pipe
// input, returns the pipe output. The goroutine continues until context done
// and the returned writer is an *os.File that may reassign os.Stdout or
// os.Strerr.
func Redirect(ctx context.Context, wg *sync.WaitGroup, p Printer) *os.File {
	r, w, err := os.Pipe()
	if err != nil {
		panic(err)
	}

	wg.Add(1)
	go goredirect(wg, p, WithNBR(ctx, r))
	return w
}

func goredirect(wg *sync.WaitGroup, p Printer, r io.Reader) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		p.Print(scanner.Text())
	}
}
