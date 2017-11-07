// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
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

var gdb, Goes, Loopback bool

func init() {
	flag.BoolVar(&gdb, "test.gdb", false,
		"debug certain commands (e.g. vnetd)")
	flag.BoolVar(&Goes, "test.goes", false,
		"run goes command instead of test(s)")
	flag.BoolVar(&Loopback, "test.loopback", false,
		"run goes loopback test(s)")
}

// Execute given Goes().main with os.Args stripped of the goes-MACHINE.test
// program name and any leading -test/* arguments. This exits 0 if main returns
// nil; otherwise, it prints the error and exits 1.
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

// Usage:
//	Assert{t}.Program(NAME, ARGS...).Output(Equal(EXPECT))
func Equal(expect string) Comparator {
	return func(b []byte) error {
		if s := string(b); s != expect {
			return fmt.Errorf("%q != %q", s, expect)
		}
		return nil
	}
}

// Usage:
//	Assert{t}.Program(NAME, ARGS...).Output(Match(PATTERN))
//	Assert{t}.Program(NAME, ARGS...).Failure(Match(PATTERN))
func Match(pattern string) Comparator {
	return func(b []byte) error {
		x, err := regexp.Compile(pattern)
		if err != nil {
			return err
		}
		if !x.Match(b) {
			return fmt.Errorf("Mismatched: %q", pattern)
		}
		return nil
	}
}

type Assert struct {
	testing.TB
}

type Program struct {
	testing.TB
	cmd   *exec.Cmd
	pargs []string
	obuf  logbuf
	ebuf  logbuf
}

type logbuf struct {
	*bytes.Buffer
}

type Gdb struct {
	prog *Program
	cmd  *exec.Cmd
}

type Quitter interface {
	Quit(time.Duration)
}

type Comparator func([]byte) error

func (assert Assert) Nil(err error) {
	assert.Helper()
	if err != nil {
		assert.Fatal(err)
	}
}

func (assert Assert) YoureRoot() {
	assert.Helper()
	if os.Geteuid() != 0 {
		assert.Skip("you aren't root")
	}
}

// Create and Start a Program. This replaces a "goes" program name (pn) with
// {os.Args[0] -test.goes}, where os.Args[0] is the test program.  The test
// program should exec the given goes command instead of the following tests
// like this:
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
	return &Program{assert.TB, cmd, pargs, obuf, ebuf}
}

// With -test.gdb flag, run gdb on the Program's pid.
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

// Send SIGTERM to Program then SIGKILL if incomplete by timeout.
// Usage:
//	defer Assert{t}.Program(NAME, ARGS...).Quit(TIMEOUT)
func (p *Program) Quit(timeout time.Duration) {
	p.Helper()
	p.cmd.Process.Signal(syscall.SIGTERM)
	tm := time.NewTimer(timeout)
	done := make(chan struct{})
	go func() {
		p.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		tm.Stop()
	case <-tm.C:
		p.cmd.Process.Kill()
	}
	if p.ebuf.Len() > 0 {
		if p.obuf.Len() > 0 {
			p.Log(p.pargs, p.obuf, p.ebuf)
			p.obuf.Reset()
			p.ebuf.Reset()
		} else {
			p.Log(p.pargs, p.ebuf)
			p.ebuf.Reset()
		}
		p.Fail()
	} else if p.obuf.Len() > 0 {
		p.Log(p.pargs, p.obuf)
		p.obuf.Reset()
	}
}

// Wait for Program finish then assert there were no errors and log output.
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Ok()
func (p *Program) Ok() {
	p.Helper()
	if err := p.cmd.Wait(); p.ebuf.Len() == 0 && err != nil {
		p.ebuf.WriteString(err.Error())
	}
	if p.ebuf.Len() > 0 {
		if p.obuf.Len() > 0 {
			p.Fatal(p.pargs, p.obuf, p.ebuf)
		}
		p.Fatal(p.pargs, p.ebuf)
	}
	if p.obuf.Len() > 0 {
		p.Log(p.pargs, p.obuf)
		p.obuf.Reset()
	}
}

// Wait for Program finish then assert there were no errors and it had the
// expected output.
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Output(Equal(EXPECT))
//	Assert{t}.Program(NAME, ARGS...).Output(Match(PATTERN))
func (p *Program) Output(f Comparator) {
	p.Helper()
	if err := p.cmd.Wait(); p.ebuf.Len() == 0 {
		if err == nil {
			err = f(p.obuf.Bytes())
		}
		if err != nil {
			p.ebuf.WriteString(err.Error())
		}
	}
	if p.ebuf.Len() > 0 {
		if p.obuf.Len() > 0 {
			p.Fatal(p.pargs, p.obuf, p.ebuf)
		}
		p.Fatal(p.pargs, p.ebuf)
	}
	if p.obuf.Len() > 0 {
		p.Log(p.pargs, p.obuf)
		p.obuf.Reset()
	}
}

// Wait for Program finish then assert that it had the expected error.
// Usage:
//	Assert{t}.Program(NAME, ARGS...).Failure(Match(PATTERN))
func (p *Program) Failure(f Comparator) {
	p.Helper()
	p.cmd.Wait()
	if p.obuf.Len() > 0 {
		p.Log(p.pargs, p.obuf, p.ebuf)
	} else {
		p.Log(p.pargs, p.ebuf)
	}
	if err := f(p.ebuf.Bytes()); err != nil {
		p.Fatal(p.pargs, "\n"+err.Error())
	}
	p.obuf.Reset()
	p.ebuf.Reset()
}

// Quit Program then wait for user to quit gdb.
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

// Suite of tests
type Suite []struct {
	Name string
	Func func(*testing.T)
}

// Run test suite
func (suite Suite) Run(t *testing.T) {
	for _, x := range suite {
		t.Run(x.Name, x.Func)
	}
}
