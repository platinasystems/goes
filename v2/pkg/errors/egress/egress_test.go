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
	for _, x := range []struct {
		title, want string
		err         error
	}{
		{"UnmarkedNil", "",
			Unmarked(nil),
		},
		{"MarkedNil", "",
			Marked(nil),
		},
		{"UnmarkedNilFuncNil", "",
			Unmarked(nil, func() error { return nil }),
		},
		{"MarkedNilFuncNil", "",
			Marked(nil, func() error { return nil }),
		},
		{"UnmarkedExample", "example",
			Unmarked(example),
		},
		{"MarkedExample", "egress_test.go:34: example",
			Marked(example),
		},
		{"UnmarkedNilFuncExample", "example",
			Unmarked(nil, func() error { return example }),
		},
		{"MarkedNilFuncExample", "egress_test.go:40: example",
			Marked(nil, func() error { return example }),
		},
		{"UnmarkedExampleFuncNil", "example",
			Unmarked(example, func() error { return nil }),
		},
		{"MarkedExampleFuncNil", "egress_test.go:46: example",
			Marked(example, func() error { return nil }),
		},
	} {
		t.Run(x.title, func(t *testing.T) {
			if len(x.want) == 0 {
				if x.err != nil {
					t.Error(x.err)
				}
			} else if x.err == nil {
				t.Errorf("<nil> != %q", x.want)
			} else if got := x.err.Error(); got != x.want {
				t.Errorf("%q != %q", got, x.want)
			}
		})
	}
}
