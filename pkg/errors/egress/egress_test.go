// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package egress

import (
	"errors"
	"testing"
)

func Test(t *testing.T) {
	example := errors.New("example")
	for _, unit := range []struct {
		title  string
		expect string
		test   func() error
	}{
		{"MarkNil", "", func() error { return Marked(nil) }},
		{"MarkedExample", "egress_test.go:21: example",
			func() error { return Marked(example) }},
		{"RecoverNil", "",
			func() (err error) {
				defer Recovery(&err)
				return
			}},
		{"RecoverExample", "egress_test.go:30: example",
			func() (err error) {
				defer Recovery(&err)
				panic(example)
				return
			}},
		{"RecoverString", "egress_test.go:36: example",
			func() (err error) {
				defer Recovery(&err)
				panic(example.Error())
				return
			}},
		{"RecoverRange", "egress_test.go:43: runtime error: index out of range [0] with length 0",
			func() (err error) {
				defer Recovery(&err)
				var a []byte
				a[0] = 0
				return
			}},
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
