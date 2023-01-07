// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package endiantest

import (
	"fmt"
	"math"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/endian"
)

var Uints = []uint64{
	0x0000000000000000,
	0x0123456789abcdef,
	0xfedcba9876543210,
	0xffffffffffffffff,
	0xaaaaaaaaaaaaaaaa,
}

var Floats = []float64{
	math.E,
	math.Pi,
	math.Phi,
	math.Sqrt2,
	math.SqrtE,
	math.SqrtPi,
	math.SqrtPhi,
	math.Ln2,
	math.Ln10,
}

func Run[T endian.Numeric](t *testing.T, x endian.Number[T], want T) {
	t.Helper()
	t.Run(fmt.Sprintf("%T(%#x)", x, want), func(t *testing.T) {
		t.Helper()
		x.Put(want)
		if got := x.Value(); got != want {
			t.Errorf("got: %#x", got)
		}
	})
}
