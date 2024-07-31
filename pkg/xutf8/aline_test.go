// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xutf8

import (
	"fmt"
	"slices"
	"testing"
)

func TestAline(t *testing.T) {
	for i, unit := range []struct{ input, want string }{
		{"", ""},
		{"\n", "\n"},
		{"\n\n", "\n"},
		{"\n\n\n", "\n"},
		{"foo", "foo\n"},
		{"foo\n", "foo\n"},
		{"foo\n\n", "foo\n"},
		{"foo\n\n\n", "foo\n"},
		{"foo\n\n\n\n", "foo\n"},
		{"foo\nbar", "foo\nbar\n"},
		{"foo\nbar\n", "foo\nbar\n"},
		{"foo\nbar\n\n", "foo\nbar\n"},
		{"foo\nb", "foo\nb\n"},
		{"foo\nb\n", "foo\nb\n"},
		{"foo\nb\n\n", "foo\nb\n"},
	} {
		t.Run(fmt.Sprintf("Bytes[%d]", i), func(t *testing.T) {
			got := AlineBytes([]byte(unit.input))
			if !slices.Equal(got, []byte(unit.want)) {
				t.Errorf("%q != %q", got, unit.want)
			}
		})
		t.Run(fmt.Sprintf("Runes[%d]", i), func(t *testing.T) {
			got := AlineRunes([]rune(unit.input))
			if !slices.Equal(got, []rune(unit.want)) {
				t.Errorf("%q != %q", got, unit.want)
			}
		})
		t.Run(fmt.Sprintf("String[%d]", i), func(t *testing.T) {
			got := AlineString(unit.input)
			if got != unit.want {
				t.Errorf("%q != %q", got, unit.want)
			}
		})
	}
}
