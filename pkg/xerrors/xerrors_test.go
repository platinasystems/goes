// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package xerrors

import (
	"errors"
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	example := errors.New("example")
	for i, unit := range []struct {
		test   func() error
		expect string
	}{
		{func() error {
			return Mark(nil)
		}, ""},
		{func() error {
			return Mark(example)
		}, "xerrors_test.go:23: example"},
		{func() (err error) {
			defer Recovery(&err)
			return
		}, ""},
		{func() (err error) {
			defer Recovery(&err)
			panic(example)
			return
		}, "xerrors_test.go:31: example"},
		{func() (err error) {
			defer Recovery(&err)
			panic(example.Error())
			return
		}, "xerrors_test.go:36: example"},
		{func() (err error) {
			defer Recovery(&err)
			var a []byte
			a[0] = 0
			return
		}, "xerrors_test.go:42: runtime error: index out of range [0] with length 0"},
		{func() error {
			foo := errors.New("foo")
			return Suppress(foo, foo)
		}, ""},
		{func() error {
			foo := errors.New("foo")
			bar := errors.New("bar")
			return Suppress(foo, bar)
		}, "foo"},
		{func() error {
			return Broken("arrow")
		}, "arrow: broken"},
		{func() (err error) {
			defer Recovery(&err)
			Assert(example)
			return
		}, "xerrors_test.go:59: example"},
		{func() (err error) {
			defer Recovery(&err)
			AssertResult("foo", example)
			return
		}, "xerrors_test.go:64: example"},
		{func() (err error) {
			defer Recovery(&err)
			AssertTrue(AssertResult("foo", nil) == "bar")
			return
		}, "xerrors_test.go:69: false"},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got := unit.test()
			if len(unit.expect) == 0 {
				if got != nil {
					t.Error("unexpected: ", got)
				}
			} else if got == nil {
				t.Errorf("got nil, expected: %q", unit.expect)
			} else if s := got.Error(); s != unit.expect {
				t.Errorf("%q != %q", s, unit.expect)
			}
		})
	}
}
