// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package kvc

import (
	"errors"
	"strings"
	"testing"
)

func Test(t *testing.T) {
	input := strings.NewReader(`
# comment
inline space    # comment
inline tab	# comment
no comment
no-values
`)
	want := map[int]struct {
		key    string
		values []string
	}{
		3: {"inline", []string{"space"}},
		4: {"inline", []string{"tab"}},
		5: {"no", []string{"comment"}},
		6: {"no-values", []string{}},
	}
	var fail = errors.New("fail")
	Range(input, func(lno int, key string, values []string) error {
		entry, ok := want[lno]
		if !ok {
			t.Error(lno)
			return fail
		}
		if key != entry.key {
			t.Errorf("%d: %q != %q", lno, key, entry.key)
			return fail
		}
		for i, v := range values {
			if i > len(entry.values) {
				t.Errorf("%d: %d", lno, i)
				return fail
			}
			if v != entry.values[i] {
				t.Errorf("%d[%d]: %q != %q", lno, i, v,
					entry.values[i])
				return fail
			}
		}
		return nil
	})
}
