// Copyright 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style license described in the
// LICENSE file.

package recovered

import (
	"io"
	"strings"
	"testing"
)

type x struct {
	name string
	main func(...string) error
}

func (x *x) Main(args ...string) error { return x.main(args...) }

func (x *x) String() string { return x.name }

func TestNoErr(t *testing.T) {
	err := New(&x{
		"test",
		func(_ ...string) error {
			return nil
		},
	}).Main()
	if err != nil {
		t.Error("unexpected err:", err)
	}
}

func TestIoEOF(t *testing.T) {
	err := New(&x{
		"test",
		func(_ ...string) error {
			return io.EOF
		},
	}).Main()
	if err != nil {
		t.Error("unexpected err:", err)
	}
}

func TestUnexpectedEOF(t *testing.T) {
	expect := "test: " + io.ErrUnexpectedEOF.Error()
	err := New(&x{
		"test",
		func(_ ...string) error {
			return io.ErrUnexpectedEOF
		},
	}).Main()
	if err.Error() != expect {
		t.Error("unexpected err:", err)
	}
}

func TestPanic(t *testing.T) {
	err := New(&x{
		"test",
		func(_ ...string) error {
			panic(42)
			return nil
		},
	}).Main()
	if err.Error() != "test: 42" {
		t.Error("unexpected err:", err)
	}
}

func TestDivby0(t *testing.T) {
	err := New(&x{
		"test",
		func(_ ...string) error {
			var i, j int
			_ = i / j
			return nil
		},
	}).Main()
	if !strings.HasPrefix(err.Error(),
		"test: runtime error: integer divide by zero") {
		t.Error("unexpected err:", err)
	} else {
		t.Log(err)
	}
}
