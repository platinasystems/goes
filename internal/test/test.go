// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/prog"
)

// Goes is set by -test.goes
var Goes bool

// DefaultTimeout for Program.{OK,Output,Failure,Done}
var DefaultTimeout = 3 * time.Second

var gdb bool

func init() {
	flag.BoolVar(&gdb, "test.gdb", false,
		"debug certain commands (e.g. vnetd)")
	flag.BoolVar(&Goes, "test.goes", false,
		"run goes command instead of test(s)")
}

// Exec will execuute given Goes().main with os.Args stripped of the
// goes-MACHINE.test program name and any leading -test/* arguments.
// This exits 0 if main returns nil; otherwise, it prints the error
// and exits 1.
//
// Usage:
//
//	func Test(t *testing.T) {
//		if Goes {
//			Exec(main.Goes().Main)
//		}
//		t.Run("Test1", func(t *testing.T) {
//			...
//		})
//		...
//	}
func Exec(main func(...string) error) {
	args := os.Args[1:]
	n := 0
	for ; n < len(args) && strings.HasPrefix(args[n], "-test."); n++ {
	}
	if n > 0 {
		copy(args[:len(args)-n], args[n:])
		args = args[:len(args)-n]
	}
	ecode := 0
	if err := main(args...); err != nil {
		fmt.Fprintln(os.Stderr, err)
		ecode = 1
	}
	syscall.Exit(ecode)
}

// Assert wraps a testing.Test or Benchmark with several assertions.
type Assert struct {
	testing.TB
}

// Program wraps an exec.Cmd to provide test results
type Program struct {
	testing.TB
	cmd   *exec.Cmd
	pargs []string
	obuf  logbuf
	ebuf  logbuf
	err   error
}

// Gdb facilitates Program debugging
type Gdb struct {
	prog *Program
	cmd  *exec.Cmd
}

// A Quitter provides a Quit method to stop it's process within a timeout
type Quitter interface {
	Quit(time.Duration)
}

// Suite of tests
type Suite []struct {
	Name string
	Func func(*testing.T)
}

type logbuf struct {
	*bytes.Buffer
}

// Nil asserts that there is no error
func (assert Assert) Nil(err error) {
	assert.Helper()
	if err != nil {
		assert.Fatal(err)
	}
}

// YoureRoot skips the calling test if EUID != 0
func (assert Assert) YoureRoot() {
	assert.Helper()
	if os.Geteuid() != 0 {
		assert.Skip("you aren't root")
	}
}

// Program creates and starts an exec.Cmd. This replaces a "goes" program name
// (pn) with {os.Args[0] -test.goes}, where os.Args[0] is the test program.
// The test program should exec the given goes command instead of the following
// tests like this:
//
//	func Test(t *testing.T) {
//		if Goes {
//			Exec(main.Goes().Main)
//		}
//		Suite{
//			{"Test1", func(t *testing.T) {
//			...
//			}},
//			...
//		}.Run(t)
//	}
//
// A Program should always end with the Done method. The Wait, OK, Output, and
// Failure methods return its Program so thet may used to describe all
// assertions in one line.
//
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Done()
//	Assert{t}.Program(NAME, ARGS...).OK().Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).OK().Done()
//	Assert{t}.Program(NAME, ARGS...).Output(EXPECT).Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Output(EXPECT).Done()
//	Assert{t}.Program(NAME, ARGS...).Failure(EXPECT).Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Failure(EXPECT).Done()
func (assert Assert) Program(stdin io.Reader, pn string,
	args ...string) *Program {
	assert.Helper()
	pargs := args
	if pn == "goes" {
		pn = prog.Name()
		args = append([]string{"-test.goes"}, args...)
	}
	obuf := logbuf{new(bytes.Buffer)}
	ebuf := logbuf{new(bytes.Buffer)}
	cmd := exec.Command(pn, args...)
	cmd.Stdin = stdin
	cmd.Stdout = obuf
	cmd.Stderr = ebuf
	assert.Nil(cmd.Start())
	return &Program{assert.TB, cmd, pargs, obuf, ebuf, nil}
}

// Done waits for Program to finish, if it hasn't already, logs the buffered
// stdout and stderr, then FailNow if Failed.
func (p *Program) Done() {
	p.Helper()
	if p.ebuf.Len() > 0 {
		if p.obuf.Len() > 0 {
			p.Log(p.pargs, p.obuf, p.ebuf)
		} else {
			p.Log(p.pargs, p.ebuf)
		}
	} else if p.obuf.Len() > 0 {
		p.Log(p.pargs, p.obuf)
	}
	p.obuf.Reset()
	p.ebuf.Reset()
	if p.Failed() {
		p.FailNow()
	}
}

// Gdb Program if given a -test.gdb flag
//
// Usage:
//	defer Assert{t}.Program(NAME, ARGS...).Gdb().Quit(TIMEOUT)
func (p *Program) Gdb() Quitter {
	if !gdb {
		return p
	}
	cmd := exec.Command("gdb", fmt.Sprintf("--pid=%d", p.cmd.Process.Pid))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Fprintln(os.Stderr, "gdb:", err)
		return p
	}
	return &Gdb{p, cmd}
}

// Quit the Program with SIGTERM then SIGKILL if incomplete by TIMEOUT.
//
// Usage:
//	defer Assert{t}.Program(NAME, ARGS...).Quit(TIMEOUT)
func (p *Program) Quit(timeout time.Duration) {
	p.Helper()
	p.cmd.Process.Signal(syscall.SIGTERM)
	tm := time.NewTimer(timeout)
	done := make(chan error)
	go func() { done <- p.cmd.Wait() }()
	select {
	case err := <-done:
		tm.Stop()
		if p.ebuf.Len() == 0 && err != nil {
			p.ebuf.WriteString(err.Error())
		}
	case <-tm.C:
		p.ebuf.WriteString("process not responding, sending SIGKILL")
		p.cmd.Process.Kill()
	}
	p.Done()
}

// Wait asserts that Program finishes within the given timeout.
//
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Done()
func (p *Program) Wait(timeout time.Duration) *Program {
	p.Helper()
	if p.cmd == nil {
		return p
	}
	tm := time.NewTimer(timeout)
	done := make(chan error)
	go func() { done <- p.cmd.Wait() }()
	select {
	case p.err = <-done:
		tm.Stop()
		if p.ebuf.Len() == 0 && p.err != nil {
			p.ebuf.WriteString(p.err.Error())
		}
	case <-tm.C:
		p.cmd.Process.Kill()
		<-done
		p.err = syscall.ETIME
		p.ebuf.WriteString(p.err.Error())
		p.Fail()
	}
	p.cmd = nil
	return p
}

// Ok asserts that Program finishes w/o error.
//
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Ok().Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Ok().Done()
func (p *Program) Ok() *Program {
	p.Helper()
	if p.Wait(DefaultTimeout); p.err != nil {
		p.Fail()
	}
	return p
}

// Output asserts that Program finished Ok with the expected output.
//
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Output(EXPECT).Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Output(EXPECT).Done()
func (p *Program) Output(expect string) *Program {
	p.Helper()
	if p.Ok(); !p.Failed() {
		p.Expect(expect, p.obuf.Bytes())
	}
	return p
}

// Failure asserts that Program finish the expected error.
//
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Failure(EXPECT).Done()
//	Assert{t}.Program(NAME, ARGS...).Wait(TIMEOUT).Failure(EXPECT).Done()
func (p *Program) Failure(expect string) *Program {
	p.Helper()
	if p.Wait(DefaultTimeout); !p.Failed() {
		p.Expect(expect, p.ebuf.Bytes())
	}
	return p
}

// Expect asserts buffer content.  If EXPECT is /PATTERN/, the output is
// regexp matched with PATTERN; otherwise, it's an exact match.
func (p *Program) Expect(expect string, b []byte) {
	var match bool
	if strings.HasPrefix(expect, "/") && strings.HasPrefix(expect, "/") {
		expect = strings.TrimPrefix(strings.TrimSuffix(expect, "/"),
			"/")
		match = regexp.MustCompile(expect).Match(b)
	} else {
		match = string(b) == expect
	}
	if !match {
		fmt.Fprintf(p.ebuf, "mismatch %q", expect)
		p.Fail()
	}
}

// Quit Program then wait for user to quit gdb.
//
// Usage:
//	defer Assert{t}.Gdb(NAME, ARGS...).Quit(TIMEOUT)
func (gdb *Gdb) Quit(timeout time.Duration) {
	gdb.prog.Quit(timeout)
	gdb.cmd.Wait()
}

// reformat log buffer for pretty print with (*testing.T).Log
func (lb logbuf) String() string {
	return "\n" + strings.TrimRight(lb.Buffer.String(), "\n")
}

// Run test suite
func (suite Suite) Run(t *testing.T) {
	for _, x := range suite {
		t.Run(x.Name, x.Func)
	}
}
