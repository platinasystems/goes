// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package endian

import (
	"math"
	"testing"
)

var names = []string{"big", "little", "native"}
var endians = []Endian{Big, Little, Native}

func TestIntegers(t *testing.T) {
	for i, endian := range endians {
		t.Run(names[i], func(t *testing.T) {
			t.Helper()
			for _, want := range []uint64{
				0x0000000000000000,
				0x0123456789abcdef,
				0xfedcba9876543210,
				0xffffffffffffffff,
				0xaaaaaaaaaaaaaaaa,
			} {
				data := NewInteger(endian, want)
				got, _ := PullInteger[uint64](endian, data)
				if got != want {
					t.Error(got, "!=", want)
				}
			}
		})
	}
}

func TestFloats(t *testing.T) {
	for i, endian := range endians {
		t.Run(names[i], func(t *testing.T) {
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
				data := NewFloat(endian, want)
				got, _ := PullFloat[float64](endian, data)
				if got != want {
					t.Error(got, "!=", want)
				}
			}
		})
	}
}
