// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"bytes"
	"go/build"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"syscall"
	"testing"
)

// Assert wraps a testing.Test or Benchmark with several assertions.
type Assert struct {
	testing.TB
}

// Log args if -test.vv
func (assert Assert) Comment(args ...interface{}) {
	assert.Helper()
	if *VV {
		assert.Log(args...)
	}
}

// Format args if -test.vv
func (assert Assert) Commentf(format string, args ...interface{}) {
	assert.Helper()
	if *VV {
		assert.Logf(format, args...)
	}
}

// If necessary, change to the dir of the given go package.
func (assert Assert) Dir(name string) {
	wd, err := os.Getwd()
	assert.Nil(err)
	if strings.HasSuffix(wd, name) {
		return
	}
	pkg, err := build.Import(name, "", build.FindOnly)
	assert.Nil(err)
	assert.Nil(os.Chdir(pkg.Dir))
}

// Main runs the main function if given the "-test.main" flag.  With said flag,
// this strip os.Args[0] and any leading -test.* options and os.Exit(0) if the
// main returns. Otherwise, Main changes CWD to "-test.cd", if necessary, and
// returns.
func (assert Assert) Main(main func()) {
	if !*IsMain {
		return
	}
	os.Args = os.Args[1:]
	for len(os.Args) > 0 && strings.HasPrefix(os.Args[0], "-test.") {
		os.Args = os.Args[1:]
	}
	main()
	syscall.Exit(0)
}

// Nil asserts that there is no error
func (assert Assert) Nil(err error) {
	assert.Helper()
	if err != nil {
		assert.Fatal(err)
	}
}

// NonNil asserts that there is an error
func (assert Assert) NonNil(err error) {
	assert.Helper()
	if err == nil {
		assert.Fatal(err)
	}
}

// Error asserts that an error matches the given error, string, regex, or bool
// If v is true, asserts err isn't nil;
// otherwise, if false, asserts that it's nil.
func (assert Assert) Error(err error, v interface{}) {
	assert.Helper()
	switch t := v.(type) {
	case error:
		if err != t {
			assert.Fatalf("expected %q", t.Error())
		}
	case string:
		if err == nil || err.Error() != t {
			assert.Fatalf("expected %q", t)
		}
	case *regexp.Regexp:
		if err == nil || !t.MatchString(err.Error()) {
			assert.Fatalf("expected %q", t.String())
		}
	case bool:
		if t {
			if err == nil {
				assert.Fatal("not error")
			}
		} else {
			assert.Nil(err)
		}
	default:
		assert.Fatal("can't match:", t)
	}
}

// Equal asserts string equality.
func (assert Assert) Equal(s, expect string) {
	assert.Helper()
	if s != expect {
		assert.Fatalf("%q\n\t!= %q", s, expect)
	}
}

// Match asserts string pattern match.
func (assert Assert) Match(s, pattern string) {
	assert.Helper()
	if !regexp.MustCompile(pattern).MatchString(s) {
		assert.Fatalf("%q\n\t!= @(%s)", s, pattern)
	}
}

// Match asserts string pattern match.
func (assert Assert) MatchNonFatal(s, pattern string) bool {
	assert.Helper()
	if !regexp.MustCompile(pattern).MatchString(s) {
		return false
	}
	return true
}

// True asserts flag.
func (assert Assert) True(t bool) {
	assert.Helper()
	if !t {
		assert.Fatal("not true")
	}
}

// False is not True.
func (assert Assert) False(t bool) {
	assert.Helper()
	if t {
		assert.Fatal("not false")
	}
}

// YoureRoot skips the calling test if EUID != 0
func (assert Assert) YoureRoot() {
	assert.Helper()
	if os.Geteuid() != 0 {
		assert.Skip("you aren't root")
	}
}

// Verifiy that there is no listener on named Unix socket.
func (assert Assert) NoListener(atsockname string) {
	assert.Helper()
	b, err := ioutil.ReadFile("/proc/net/unix")
	if err != nil {
		return
	}
	if bytes.Index(b, []byte(atsockname)) < 0 {
		return
	}
	assert.Fatal(atsockname, "in use")
}

// Program asserts that the Program runs without error.
func (assert Assert) Program(options ...interface{}) {
	assert.Helper()
	p, err := Begin(assert.TB, options...)
	assert.Nil(err)
	assert.Nil(p.End())
}

// ProgramErr asserts that the Program returns matches (v) error (see Error).
func (assert Assert) ProgramErr(v interface{}, options ...interface{}) {
	assert.Helper()
	p, err := Begin(assert.TB, options...)
	assert.Nil(err)
	assert.Error(p.End(), true)
}

// Background Program after asserting that it starts without error.
// Usage:
//	defer Assert{t}.Background(...).Quit()
func (assert Assert) Background(options ...interface{}) *Program {
	assert.Helper()
	p, err := Begin(assert.TB, options...)
	assert.Nil(err)
	return p
}
