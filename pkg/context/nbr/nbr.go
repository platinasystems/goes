// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nbr

import (
	"bufio"
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"syscall"
	"time"

	"github.com/platinasystems/goes/v2/pkg/context/rawtty"
)

const (
	MinInterval = rawtty.MinInterval
	MaxInterval = rawtty.MaxInterval
)

type FD = rawtty.FD
type Printer interface{ Print(...any) }

// Non-Blocking Reader
type NBR struct {
	context.Context
	f *os.File
}

func With(ctx context.Context, f *os.File) NBR { return NBR{ctx, f} }

func (nbr NBR) Read(b []byte) (int, error) {
	fd := FD(nbr.f.Fd())
	syscall.SetNonblock(fd, true)
	defer syscall.SetNonblock(fd, false)
	for dur := MinInterval; ; {
		n, err := nbr.f.Read(b)
		if err == nil || !errors.Is(err, syscall.EAGAIN) {
			return n, err
		}
		t := time.NewTimer(dur)
		select {
		case <-t.C:
			if dur < MaxInterval {
				if dur *= 2; dur > MaxInterval {
					dur = MaxInterval
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
	go goredirect(wg, p, With(ctx, r))
	return w
}

func goredirect(wg *sync.WaitGroup, p Printer, r io.Reader) {
	defer wg.Done()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		p.Print(scanner.Text())
	}
}
