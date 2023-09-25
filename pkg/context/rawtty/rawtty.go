// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rawtty

import (
	"context"
	"errors"
	"io"
	"os"
	"syscall"
	"time"

	"golang.org/x/term"
)

const (
	MinReadBufferLen = 2
	MinInterval      = 10 * time.Millisecond
	MaxInterval      = 250 * time.Millisecond
)

const ctrlZ = 26

const help = "\n\rSupported escape sequences:" +
	"\n\r ~.   - EOF" +
	"\n\r ~B   - BREAK (zero length read)" +
	"\n\r ~^Z  - suspend" +
	"\n\r ~?   - this message" +
	"\n\r ~~   - send the escape character by typing it twice" +
	"\n\r(Note that escapes are only recognized immediately after newline.)" +
	"\n\r"

var (
	ErrNotTerminal = errors.New("non-terminal input")
	ErrTooShort    = errors.New("short read buffer")
)

type TTY struct {
	context.Context
	io.Writer
	esc  uint
	prev *term.State
}

// Change to terminal raw mode and Read until "\r~." or context done.
func With(ctx context.Context) (tty *TTY, err error) {
	tty = &TTY{
		Context: ctx,
		Writer:  os.Stdout,
	}
	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		tty.prev, err = term.MakeRaw(fd)
	} else {
		err = ErrNotTerminal
	}
	return
}

func (tty *TTY) Close() error {
	return term.Restore(int(os.Stdin.Fd()), tty.prev)
}

func (tty *TTY) Read(b []byte) (n int, err error) {
	if len(b) < MinReadBufferLen {
		err = ErrTooShort
		return
	}
	fd := FD(os.Stdin.Fd())
	syscall.SetNonblock(fd, true)
	defer syscall.SetNonblock(fd, false)
	for dur := MinInterval; ; {
		n, err = os.Stdin.Read(b[:1])
		if err == nil {
			dur = MinInterval
		} else if errors.Is(err, syscall.EAGAIN) {
			t := time.NewTimer(dur)
			select {
			case <-t.C:
				if dur < MaxInterval {
					if dur *= 2; dur > MaxInterval {
						dur = MaxInterval
					}
				}
			case <-tty.Done():
				if !t.Stop() {
					<-t.C
				}
				n, err = 0, context.Canceled
				return
			}
			continue
		} else {
			return
		}
		switch tty.esc {
		case 2:
			switch b[0] {
			case '~':
			case '.':
				n, err = 0, io.EOF
			case '?':
				tty.Write([]byte(help))
				tty.esc = 0
				continue
			case 'B':
				n = 0
			case ctrlZ:
				tty.Write([]byte("\n\rFIXME\n\r"))
				tty.esc = 0
				continue
			default:
				b[1] = b[0]
				b[0] = '~'
				n = 2
			}
			tty.esc = 0
			return
		case 1:
			if b[0] == '~' {
				tty.esc = 2
				continue
			}
			tty.esc = 0
			return
		case 0:
			if b[0] == '\r' {
				tty.esc = 1
			}
			return
		}
	}
	return
}
