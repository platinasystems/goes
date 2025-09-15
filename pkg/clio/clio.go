// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package clio (Command Line Input/Output) provides a [term.Terminal] wrapper
// that includes a method to interject output over the current prompt.
package clio

import (
	"fmt"
	"io"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xutf8"
	"golang.org/x/term"
)

type CLIO struct {
	*term.Terminal

	fd int

	cooked *term.State

	mutex sync.Mutex
}

// Returns a wrapped raw mode [term.Terminal] with the given “prompt” and
// non-nil [term.Terminal.AutoCompleteCallback] function.
func New(
	tty interface {
		io.Reader
		io.Writer
		Fd() uintptr
	},
	prompt string,
	accb func(string, int, rune) (string, int, bool),
) (clio *CLIO, err error) {
	clio = &CLIO{
		Terminal: term.NewTerminal(tty, prompt),
		fd:       int(tty.Fd()),
	}
	if accb != nil {
		clio.Terminal.AutoCompleteCallback = accb
	}
	clio.cooked, err = term.MakeRaw(clio.fd)
	return
}

// Restore cooked tty (aka not-raw).
func (clio *CLIO) Close() error {
	return term.Restore(clio.fd, clio.cooked)
}

// Restore cooked tty (aka not-raw) before calling “f” the return to raw mode.
func (clio *CLIO) Cook(f func()) error {
	raw, err := term.GetState(clio.fd)
	if err == nil {
		if err = term.Restore(clio.fd, clio.cooked); err == nil {
			f()
			err = term.Restore(clio.fd, raw)
		}
	}
	return err
}

// Clear the current line then:
// write a byte slice arg;
// copy from a reader arg;
// or [fmt.Fprintln] all other args
// and assure that there is one and only on trailing newline..
func (clio *CLIO) Interject(args ...any) (total int, err error) {
	var n int
	var n64 int64
	if len(args) == 0 {
		return
	}
	clio.mutex.Lock()
	defer clio.mutex.Unlock()
	clio.Write([]byte{'\r', 27, '[', 'K'})
	w := xutf8.NewLastRuneWrapper(clio)
	for _, arg := range args {
		switch t := arg.(type) {
		case []byte:
			if n, err = w.Write(t); err == nil {
				total += n
			} else {
				return
			}
		case io.Reader:
			if n64, err = io.Copy(w, t); err == nil {
				total += int(n64)
			} else {
				return
			}
		default:
			if n, err = fmt.Fprint(w, t); err == nil {
				total += n
			} else {
				return
			}
		}
	}
	if w.LastWrittenRune() != '\n' {
		io.WriteString(clio, "\n")
		n += 1
	}
	return
}
