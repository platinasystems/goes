// Copyright © 2015-2017 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package test

import (
	"os"
	"regexp"
	"testing"
)

// Assert wraps a testing.Test or Benchmark with several assertions.
type Assert struct {
	testing.TB
}

// Nil asserts that there is no error
func (assert Assert) Nil(err error) {
	assert.Helper()
	if err != nil {
		assert.Fatal(err)
	}
}

// Error asserts that an error matches the given error, string, or regex
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

// Program asserts that the Program runs without error.
func (assert Assert) Program(options ...interface{}) {
	assert.Helper()
	p, err := Begin(assert.TB, options...)
	assert.Nil(err)
	assert.Nil(p.End())
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
