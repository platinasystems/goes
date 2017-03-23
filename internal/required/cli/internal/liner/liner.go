// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package liner is a wrapper to Peter Harris' <pharris@opentext.com>
// "Go line editor" <github.com:peterh/liner>.
package liner

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"strings"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/internal/fields"
	"github.com/platinasystems/go/internal/goes"
	"github.com/platinasystems/go/internal/nocomment"
	"github.com/platinasystems/go/internal/pizza"
	"github.com/platinasystems/go/internal/required/cli/internal/notliner"
	"github.com/platinasystems/liner"
)

const woliner = false

type Liner struct {
	history struct {
		buf   *bytes.Buffer
		lines []string
		i     int
	}
	fallback *notliner.Prompter
	byName   goes.ByName
}

func New(byName goes.ByName) *Liner {
	l := new(Liner)
	l.history.buf = new(bytes.Buffer)
	l.history.lines = make([]string, 0, 1<<6)
	if woliner {
		l.fallback = notliner.New(os.Stdin, os.Stdout)
	}
	l.byName = byName
	return l
}

// Returns all completions of the given command line.
func (l *Liner) complete(line string) (lines []string) {
	lsi := strings.LastIndex(line, " ")
	pl := pizza.New("|")
	defer pl.Reset()
	pl.Slice(fields.New(nocomment.New(strings.TrimLeft(line, " \t")))...)
	if len(pl.Slices) == 0 {
		return
	}
	args := pl.Slices[len(pl.Slices)-1]
	pr, pw, err := os.Pipe()
	if err != nil {
		return
	}
	go func() {
		t := os.Stdout
		defer func() { os.Stdout = t }()
		os.Stdout = pw
		l.byName.Main(append([]string{"complete"}, args...)...)
		pw.Close()
	}()
	prs := bufio.NewScanner(pr)
	for prs.Scan() {
		if lsi < 1 {
			lines = append(lines, prs.Text())
		} else {
			lines = append(lines, line[:lsi+1]+prs.Text())
		}
	}
	pr.Close()
	if len(lines) == 1 {
		lines[0] += " "
	}
	return
}

// Prints the best available help text for the last arg of line
func (l *Liner) help(line string) {
	pl := pizza.New("|")
	defer pl.Reset()
	pl.Slice(fields.New(nocomment.New(strings.TrimLeft(line, " \t")))...)
	if len(pl.Slices) == 0 {
		return
	}
	args := pl.Slices[len(pl.Slices)-1]
	if len(args) == 0 || pl.More {
		fmt.Println("Enter command.")
	} else {
		l.byName.Main(append([]string{"help"}, args...)...)
	}
}

func (l *Liner) Prompt(prompt string) (string, error) {
	if l.fallback != nil {
		return l.fallback.Prompt(prompt)
	}
	t := &syscall.Termios{}

	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCGETS),
		uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return "", fmt.Errorf("TCGETS: %v", errno)
	}
	t.Iflag |= syscall.BRKINT
	t.Iflag |= syscall.IMAXBEL
	t.Iflag |= syscall.IUTF8
	t.Lflag &^= syscall.IEXTEN

	_, _, errno = syscall.Syscall(syscall.SYS_IOCTL,
		uintptr(syscall.Stdin),
		uintptr(syscall.TCSETS),
		uintptr(unsafe.Pointer(t)))
	if errno != 0 {
		return "", fmt.Errorf("TCSETS: %v", errno)
	}
	err := l.byName.Main("resize")
	if err != nil {
		return "", err
	}

	state := liner.NewLiner()
	state.SetCompleter(l.complete)
	state.SetHelper(l.help)
	if len(l.history.lines) > 0 {
		l.history.buf.Reset()
		if len(l.history.lines) < cap(l.history.lines) {
			for i := 0; i < l.history.i; i++ {
				fmt.Fprintln(l.history.buf, l.history.lines[i])
			}
		} else {
			for i := l.history.i + 1; ; i++ {
				i &= cap(l.history.lines) - 1
				if i == l.history.i {
					break
				}
				fmt.Fprintln(l.history.buf, l.history.lines[i])
			}
		}
		state.ReadHistory(l.history.buf)
	}
	line, err := state.Prompt(prompt)
	state.Close()
	if err == nil {
		if len(l.history.lines) < cap(l.history.lines) {
			l.history.lines = append(l.history.lines, line)
		} else {
			l.history.lines[l.history.i] = line
		}
		l.history.i++
		l.history.i &= cap(l.history.lines) - 1
	} else if err == liner.ErrNotTerminalOutput {
		l.fallback = notliner.New(os.Stdin, os.Stdout)
		line, err = l.fallback.Prompt(prompt)
	}
	return line, err
}
