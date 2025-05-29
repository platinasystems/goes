// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package chunk

import "testing"

func TestPools(t *testing.T) {
	for _, want := range []uint{
		16, 64, 256, 512, 1 << 10, 2 << 10, 4 << 10, 8 << 10,
	} {
		chuck := New(want)
		if got := uint(cap(*chuck)); got != want {
			t.Fatal(got, "!=", want)
		}
	}
}
