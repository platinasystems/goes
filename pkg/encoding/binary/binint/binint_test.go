// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binint

import "testing"

func Test(t *testing.T) {
	for name, endian := range Endians {
		t.Run(name, func(t *testing.T) {
			t.Helper()

			var b [8]byte
			var got uint64

			for _, want := range []uint64{
				0x0000000000000000,
				0x0123456789abcdef,
				0xfedcba9876543210,
				0xffffffffffffffff,
				0xaaaaaaaaaaaaaaaa,
			} {
				Append(endian, b[:0], want)
				Pull(endian, b[:], &got)
				if got != want {
					t.Error(got, "!=", want)
				}
			}
		})
	}
}
