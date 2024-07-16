// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xslices

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	for i, unit := range []struct {
		s   []string
		pat string
		r   []string
	}{
		{[]string{"foo", "bar"}, "", []string{"foo", "bar"}},
		{[]string{"foo", "bar"}, "f", []string{"foo"}},
		{[]string{"foo", "bar"}, "x", []string{}},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			r := Match(unit.s, strings.HasPrefix, unit.pat)
			if !reflect.DeepEqual(r, unit.r) {
				t.Fail()
			}
		})
	}
}
