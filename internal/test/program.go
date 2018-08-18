// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/prog"
)

// Timeout is the default duration on the Program Wait timer.
const Timeout = 3 * time.Second

// Self flags Program to run itself
type Self struct{}

// Begin a Program; type options:
//
//	Self	inserts []string{os.Args[0], "-test.main}" into Program args;
//		the Test should run it's own main if said flag is set, e.g.:
//
//		func Test(t *testing.T) {
//			test.Main(main)
//			test.Suite{
//				{"Test1", func(t *testing.T) {
//					...
//				}},
//				...
//			}.Run(t)
//		}
//
//	io.Reader
//		use reader as Stdin instead of the default, /dev/null
//
//	*regexp.Regexp
//		match Stdout with compiled regex pattern
//
//	time.Duration
//		wait up to the given duration for the program to finish instead
//		of the default Timeout
func Begin(tb testing.TB, options ...interface{}) (*Program, error) {
	var (
		stdin io.Reader
		args  []string
	)
	p := &Program{
		tb:   tb,
		obuf: new(bytes.Buffer),
		ebuf: new(bytes.Buffer),
		dur:  Timeout,
	}
	for _, opt := range options {
		switch t := opt.(type) {
		case Self:
			args = append(args, prog.Name(), "-test.main")
		case io.Reader:
			stdin = t
		case *regexp.Regexp:
			p.exp = t
		case string:
			args = append(args, t)
		case []string:
			args = append(args, t...)
		case time.Duration:
			p.dur = t
		default:
			args = append(args, fmt.Sprint(t))
		}
	}
	if len(args) == 0 {
		return p, errors.New("missing command args")
	}
	// preface output with newline for pretty logging
	p.obuf.WriteRune('\n')
	p.cmd = exec.Command(args[0], args[1:]...)
	p.cmd.Stdin = stdin
	p.cmd.Stdout = p.obuf
	p.cmd.Stderr = p.ebuf
	if *VVV {
		tb.Helper()
		tb.Log(args)
	}
	return p, p.cmd.Start()
}

// Program is an exec.Cmd wrapper
type Program struct {
	cmd  *exec.Cmd
	tb   testing.TB
	obuf *bytes.Buffer
	ebuf *bytes.Buffer
	dur  time.Duration
	exp  *regexp.Regexp
}

// Quit will SIGTERM the Program then End and Log any error.
func (p *Program) Quit() {
	p.tb.Helper()
	p.cmd.Process.Signal(syscall.SIGTERM)
	if err := p.End(); err != nil {
		p.tb.Log(err)
	}
}

// End will wait for Program to finish or timeout then match and log output.
func (p *Program) End() (err error) {
	p.tb.Helper()
	tm := time.NewTimer(p.dur)
	done := make(chan error)
	go func() { done <- p.cmd.Wait() }()
	select {
	case err = <-done:
		tm.Stop()
		if s := strings.TrimSpace(p.obuf.String()); len(s) > 0 {
			p.tb.Log(s)
		}
		if s := strings.TrimSpace(p.ebuf.String()); len(s) > 0 {
			err = errors.New(p.ebuf.String())
			p.ebuf.Reset()
		}
		if err == nil && p.exp != nil && !p.exp.Match(p.obuf.Bytes()) {
			err = fmt.Errorf("mismatch %q", p.exp)
		}
	case <-tm.C:
		err = syscall.ETIME
		p.cmd.Process.Kill()
		<-done
	}
	if s := strings.TrimRight(p.obuf.String(), "\n"); len(s) > 0 {
		if err != nil || *VV {
			p.tb.Log(s)
		}
	}
	p.obuf.Reset()
	return
}

// Pid returns the program process identifier.
func (p *Program) Pid() int {
	return p.cmd.Process.Pid
}
