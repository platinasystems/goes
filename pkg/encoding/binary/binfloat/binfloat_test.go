// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binfloat

import (
	"math"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/binint"
)

func Test(t *testing.T) {
	for name, endian := range binint.Endians {
		t.Run(name, func(t *testing.T) {
			t.Helper()
			for _, want := range []float64{
				math.E,
				math.Pi,
				math.Phi,
				math.Sqrt2,
				math.SqrtE,
				math.SqrtPi,
				math.SqrtPhi,
				math.Ln2,
				math.Ln10,
			} {
				data := New(endian, want)
				got, _ := Pull[float64](endian, data)
				if got != want {
					t.Error(got, "!=", want)
				}
			}
		})
	}
}
