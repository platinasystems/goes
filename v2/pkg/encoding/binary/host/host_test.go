// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import (
	"testing"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/endiantest"
)

func Test(t *testing.T) {
	b := make([]byte, 8)
	for _, v := range endiantest.Uints {
		endiantest.Run[uint16](t, (*Uint16)(b), uint16(v))
		endiantest.Run[uint32](t, (*Uint32)(b), uint32(v))
		endiantest.Run[uint64](t, (*Uint64)(b), uint64(v))
	}
	for _, v := range endiantest.Floats {
		endiantest.Run[float32](t, (*Float32)(b), float32(v))
		endiantest.Run[float64](t, (*Float64)(b), float64(v))
	}
	t.Run("net16", func(t *testing.T) {
		run[uint16](t, Net16(0x1122), 0x2211)
	})
	t.Run("net32", func(t *testing.T) {
		run[uint32](t, Net32(0x11223344), 0x44332211)
	})
	t.Run("net64", func(t *testing.T) {
		run[uint64](t, Net64(0x1122334455667788), 0x8877665544332211)
	})
}

func run[T uint16 | uint32 | uint64](t *testing.T, n, little T) {
	t.Helper()
	x := n
	if IsLittleEndian {
		x = little
	}
	if n != x {
		t.Errorf("got: %#x", x)
	}
}
