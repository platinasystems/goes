// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package recovered

import (
	"bytes"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/platinasystems/go/log"
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
		func(...string) error {
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
		func(...string) error {
			return io.EOF
		},
	}).Main()
	if err != io.EOF {
		t.Error("unexpected err:", err)
	}
}

func TestUnexpectedEOF(t *testing.T) {
	expect := "test: " + io.ErrUnexpectedEOF.Error()
	err := New(&x{
		"test",
		func(...string) error {
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
		func(...string) error {
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
		func(...string) error {
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

func TestGo(t *testing.T) {
	buf := new(bytes.Buffer)
	log.Writer = buf
	go Go(func(...interface{}) {
		panic("oops")
	})
	for i := 0; i < 3; i++ {
		if buf.Len() > 0 {
			t.Log("\n" + buf.String())
			return
		}
		time.Sleep(time.Second)
	}
	t.Error("didn't receive expected panic message")
}

func TestGoDiv0(t *testing.T) {
	buf := new(bytes.Buffer)
	log.Writer = buf
	go Go(func(...interface{}) {
		var i, j int
		_ = i / j
	})
	for i := 0; i < 3; i++ {
		if buf.Len() > 0 {
			t.Log("\n" + buf.String())
			return
		}
		time.Sleep(time.Second)
	}
	t.Error("didn't receive expected panic message")
}
