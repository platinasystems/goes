// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xmaps

import (
	"fmt"
	"reflect"
	"testing"
)

func TestInstall(t *testing.T) {
	for i, unit := range []struct {
		input []map[int]any
		m     map[int]any
	}{
		{
			[]map[int]any{
				map[int]any{
					0: 0,
					1: 1,
					2: 2,
				},
			},
			map[int]any{
				0: 0,
				1: 1,
				2: 2,
			},
		},
		{
			[]map[int]any{
				map[int]any{
					0: 0,
					1: 1,
					2: 2,
				},
				map[int]any{
					0: 1,
					1: 2,
					2: 3,
				},
			},
			map[int]any{
				0: 0,
				1: 1,
				2: 2,
			},
		},
		{
			[]map[int]any{
				map[int]any{
					0: 0,
					2: 2,
				},
				map[int]any{
					0: 1,
					1: 1,
					2: 3,
				},
			},
			map[int]any{
				0: 0,
				1: 1,
				2: 2,
			},
		},
		{
			[]map[int]any{
				map[int]any{
					0: 0,
					2: 2,
				},
				map[int]any{
					0: 1,
					1: map[int]any{
						0: 0,
						1: 1,
						2: 2,
					},
					2: 3,
				},
			},
			map[int]any{
				0: 0,
				1: map[int]any{
					0: 0,
					1: 1,
					2: 2,
				},
				2: 2,
			},
		},
		{
			[]map[int]any{
				map[int]any{
					0: 0,
					1: map[int]any{
						0: 0,
						2: 2,
					},
					2: 2,
				},
				map[int]any{
					0: 1,
					1: map[int]any{
						0: 1,
						1: 1,
						2: 3,
					},
					2: 3,
				},
			},
			map[int]any{
				0: 0,
				1: map[int]any{
					0: 0,
					1: 1,
					2: 2,
				},
				2: 2,
			},
		},
	} {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			m := make(map[int]any)
			Install(m, unit.input...)
			if !reflect.DeepEqual(m, unit.m) {
				t.Fail()
			}
		})
	}
}
