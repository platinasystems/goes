// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xmaps

import (
	"fmt"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestMatch(t *testing.T) {
	nothing := struct{}{}
	for i, unit := range []struct {
		m   map[string]any
		pat string
		r   []string
	}{
		{
			map[string]any{
				"foo": nothing,
				"bar": nothing,
			},
			"", []string{"foo", "bar"},
		},
		{
			map[string]any{
				"foo": nothing,
				"bar": nothing,
			},
			"f", []string{"foo"},
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			r := Match(unit.m, strings.HasPrefix, unit.pat)
			sort.Strings(r)
			sort.Strings(unit.r)
			if !reflect.DeepEqual(r, unit.r) {
				t.Fail()
			}
		})
	}
}
