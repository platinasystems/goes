// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
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

	"github.com/platinasystems/go/command"
	"github.com/platinasystems/go/nocomment"
	"github.com/platinasystems/go/slice_args"
	"github.com/platinasystems/go/slice_string"
	"github.com/platinasystems/liner"
)

type Liner struct {
	history struct {
		buf   *bytes.Buffer
		lines []string
		i     int
	}
	state *liner.State
}

func New() *Liner {
	l := new(Liner)
	l.history.buf = new(bytes.Buffer)
	l.history.lines = make([]string, 0, 1<<6)
	return l
}

// Returns all completions of the given command line.
func complete(line string) (lines []string) {
	lsi := strings.LastIndex(line, " ")
	pl := slice_args.New("|")
	defer pl.Reset()
	pl.Slice(slice_string.New(nocomment.New(strings.TrimLeft(line,
		" \t")))...)
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
		command.Main(append([]string{"-complete"}, args...)...)
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

// Returns best available help text for the last arg of line
func help(line string) string {
	pl := slice_args.New("|")
	defer pl.Reset()
	pl.Slice(slice_string.New(nocomment.New(strings.TrimLeft(line,
		" \t")))...)
	if len(pl.Slices) == 0 {
		return ""
	}
	args := pl.Slices[len(pl.Slices)-1]
	if len(args) == 0 || pl.More {
		return "Enter command."
	}
	pr, pw, err := os.Pipe()
	if err != nil {
		return ""
	}
	go func() {
		t := os.Stdout
		defer func() { os.Stdout = t }()
		os.Stdout = pw
		command.Main(append([]string{args[0], "-help"},
			args[1:]...)...)
		pw.Close()
	}()
	buf := make([]byte, 4096)
	n, err := pr.Read(buf)
	return string(buf[:n])
}

func (l *Liner) GetLine(prompt string) (string, error) {
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
	err := command.Main("resize")
	if err != nil {
		return "", err
	}
	l.state = liner.NewLiner()
	defer l.state.Close()
	l.state = liner.NewLiner()
	l.state.SetCompleter(complete)
	l.state.SetHelper(help)
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
		l.state.ReadHistory(l.history.buf)
	}
	line, err := l.state.Prompt(prompt)
	if err == nil {
		if len(l.history.lines) < cap(l.history.lines) {
			l.history.lines = append(l.history.lines, line)
		} else {
			l.history.lines[l.history.i] = line
		}
		l.history.i++
		l.history.i &= cap(l.history.lines) - 1

	}
	return line, err
}
