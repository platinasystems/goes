// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xslices

import (
	"fmt"
	"reflect"
	"testing"
)

func TestPrepend(t *testing.T) {
	for i, unit := range []struct {
		init, prepend, want []string
	}{
		{[]string{}, []string{"a", "b", "c"}, []string{"a", "b", "c"}},
		{[]string{"b", "c"}, []string{"a"}, []string{"a", "b", "c"}},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			got := Prepend(unit.init, unit.prepend...)
			if !reflect.DeepEqual(got, unit.want) {
				t.Fail()
			}
		})
	}
}
