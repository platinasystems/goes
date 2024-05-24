// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binint

import "testing"

func Test(t *testing.T) {
	for name, endian := range Endians {
		t.Run(name, func(t *testing.T) {
			t.Helper()
			for _, want := range []uint64{
				0x0000000000000000,
				0x0123456789abcdef,
				0xfedcba9876543210,
				0xffffffffffffffff,
				0xaaaaaaaaaaaaaaaa,
			} {
				data := New(endian, want)
				got, _ := Pull[uint64](endian, data)
				if got != want {
					t.Error(got, "!=", want)
				}
			}
		})
	}
}
