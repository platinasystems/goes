// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"reflect"
	"regexp"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/platinasystems/go/internal/prog"
)

var gdb, Goes bool

func init() {
	flag.BoolVar(&gdb, "test.gdb", false,
		"debug certain commands (e.g. vnetd)")
	flag.BoolVar(&Goes, "test.goes", false,
		"run goes command instead of test(s)")
}

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

// Assert's Backgound, Ok, OutputEqual and OutputMatch methods swap a "goes"
// program name (pn) with {os.Args[0] -test.goes}.
// The test program should exec the given goes command instead of the following
// tests like this:
//
//	func Test(t *testing.T) {
//		if test.Goes {
//			test.Exec(Goes().Main)
//		}
//		t.Run("Test1", func(t *testing.T) {
//			...
//		})
//		...
//	}
type Assert struct {
	testing.TB
}

type Quitter interface {
	Quit(time.Duration)
}

type Background struct {
	testing.TB
	cmd *exec.Cmd
	log *bytes.Buffer
}

type Gdb struct {
	bg  *Background
	cmd *exec.Cmd
}

func (assert Assert) Background(pn string, args ...string) *Background {
	assert.Helper()
	pn, args = goesRename(pn, args...)
	log := new(bytes.Buffer)
	cmd := exec.Command(pn, args...)
	cmd.Stdout = log
	cmd.Stderr = log
	fmt.Fprintln(log, cmd.Args)
	assert.Nil(cmd.Start())
	return &Background{assert.TB, cmd, log}
}

// if -test.gdb, run gdb on the background process's pid
func (assert Assert) Gdb(pn string, args ...string) Quitter {
	assert.Helper()
	bg := assert.Background(pn, args...)
	if !gdb {
		return bg
	}
	cmd := exec.Command("gdb", fmt.Sprintf("--pid=%d", bg.cmd.Process.Pid))
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	assert.Nil(cmd.Start())
	return &Gdb{bg, cmd}
}

// Send SIGTERM to backgroud process then SIGKILL if incomplete by timeout
func (bg *Background) Quit(timeout time.Duration) {
	bg.Helper()
	bg.cmd.Process.Signal(syscall.SIGTERM)
	tm := time.NewTimer(timeout)
	done := make(chan struct{})
	go func() {
		bg.cmd.Wait()
		close(done)
	}()
	select {
	case <-done:
		tm.Stop()
	case <-tm.C:
		bg.cmd.Process.Kill()
	}
	if bg.log == nil {
		return
	}
	bg.Log(strings.TrimSpace(bg.log.String()))
	bg.log.Reset()
}

// quit background process with SIGTERM/SIGKILL then wait for user to quit gdb
func (gdb *Gdb) Quit(timeout time.Duration) {
	gdb.bg.Quit(timeout)
	gdb.cmd.Wait()
}

func (assert Assert) Nil(err error) {
	assert.Helper()
	if err != nil {
		assert.Fatal(err)
	}
}

func (assert Assert) Ok(pn string, args ...string) {
	assert.Helper()
	pn, args = goesRename(pn, args...)
	cmd := exec.Command(pn, args...)
	bout := new(bytes.Buffer)
	berr := new(bytes.Buffer)
	cmd.Stdout = bout
	cmd.Stderr = berr
	bout.WriteRune('\n')
	berr.WriteRune('\n')
	assert.Nil(cmd.Start())
	err := cmd.Wait()
	if berr.Len() > 1 {
		if bout.Len() > 1 {
			assert.Fatal(cmd.Args, bout.String(),
				strings.TrimRight(berr.String(), "\n"))
		}
		assert.Fatal(cmd.Args, strings.TrimRight(berr.String(), "\n"))
	}
	assert.Nil(err)
	if bout.Len() > 1 {
		assert.Log(cmd.Args, strings.TrimRight(bout.String(), "\n"))
	}
}

func (assert Assert) OutputEqual(expect string, pn string, args ...string) {
	assert.Helper()
	pn, args = goesRename(pn, args...)
	b, err := exec.Command(pn, args...).CombinedOutput()
	assert.Equal(string(b), expect)
	assert.Nil(err)
}

func (assert Assert) Equal(x, y interface{}) {
	assert.Helper()
	if !reflect.DeepEqual(x, y) {
		assert.Fatalf("%#v != %#v", x, y)
	}
}

func (assert Assert) OutputMatch(pattern string, pn string, args ...string) {
	assert.Helper()
	pn, args = goesRename(pn, args...)
	b, err := exec.Command(pn, args...).CombinedOutput()
	assert.Match(pattern, b)
	assert.Nil(err)
}

func (assert Assert) Match(pattern string, b []byte) {
	assert.Helper()
	x, err := regexp.Compile(pattern)
	assert.Nil(err)
	if !x.Match(b) {
		assert.Fatalf("%q mismatched %q", string(b), pattern)
	}
}

func (assert Assert) YoureRoot() {
	assert.Helper()
	if os.Geteuid() != 0 {
		assert.Fatal("you aren't root")
	}
}

func goesRename(pn string, args ...string) (string, []string) {
	if pn == "goes" {
		pn = prog.Name()
		args = append([]string{"-test.goes"}, args...)
	}
	return pn, args
}
