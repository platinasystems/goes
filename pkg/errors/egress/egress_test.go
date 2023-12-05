// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the BSD-style license
// described in the golang/LICENSE file.

package egress

import (
	"errors"
	"testing"
)

func Test(t *testing.T) {
	example := errors.New("example")
	for _, unit := range []struct {
		title  string
		test   func() error
		expect string
	}{
		{"MarkNil", func() error {
			return Mark(nil)
		}, ""},
		{"MarkExample", func() error {
			return Mark(example)
		}, "egress_test.go:23: example"},
		{"RecoveryNil", func() (err error) {
			defer Recovery(&err)
			return
		}, ""},
		{"RecoveryExample", func() (err error) {
			defer Recovery(&err)
			panic(example)
			return
		}, "egress_test.go:31: example"},
		{"RecoveryString", func() (err error) {
			defer Recovery(&err)
			panic(example.Error())
			return
		}, "egress_test.go:36: example"},
		{"RecoveryRange", func() (err error) {
			defer Recovery(&err)
			var a []byte
			a[0] = 0
			return
		}, "egress_test.go:42: runtime error: index out of range [0] with length 0"},
		{"SuppressedFoo", func() error {
			foo := errors.New("foo")
			return Suppress(foo, foo)
		}, ""},
		{"UnsuppressedFoo", func() error {
			foo := errors.New("foo")
			bar := errors.New("bar")
			return Suppress(foo, bar)
		}, "foo"},
	} {
		t.Run(unit.title, func(t *testing.T) {
			got := unit.test()
			if len(unit.expect) == 0 {
				if got != nil {
					t.Error("unexpected: ", got)
				}
			} else if got == nil {
				t.Errorf("expected: %q", unit.expect)
			} else if s := got.Error(); s != unit.expect {
				t.Errorf("%q != %q", s, unit.expect)
			}
		})
	}
}
