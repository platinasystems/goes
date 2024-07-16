// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "testing"

func (a Align) test(t *testing.T, tuples ...[2]int) {
	t.Helper()
	for _, tuple := range tuples {
		v, want := tuple[0], tuple[1]
		if got := a.Roundup(v); got != want {
			t.Errorf("%d: %d != %d", v, got, want)
		}
	}
}

func TestAlign(t *testing.T) {
	t.Run("Word", func(t *testing.T) {
		Align(4).test(t,
			[2]int{0, 0},
			[2]int{1, 4},
			[2]int{3, 4},
			[2]int{4, 4},
			[2]int{5, 8},
		)
	})
	t.Run("Block", func(t *testing.T) {
		Align(512).test(t,
			[2]int{0, 0},
			[2]int{1, 512},
			[2]int{511, 512},
			[2]int{512, 512},
			[2]int{513, 1024},
		)
	})
	t.Run("Page", func(t *testing.T) {
		Align(4<<10).test(t,
			[2]int{0, 0},
			[2]int{1, 4096},
			[2]int{4095, 4096},
			[2]int{4096, 4096},
			[2]int{4097, 8192},
		)
	})
	t.Run("IP", func(t *testing.T) {
		Align(5*4).test(t,
			[2]int{0, 0},
			[2]int{1, 20},
			[2]int{19, 20},
			[2]int{20, 20},
			[2]int{21, 40},
			[2]int{39, 40},
			[2]int{40, 40},
			[2]int{41, 60},
		)
	})
}
