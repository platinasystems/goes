// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netfloater

import (
	"bytes"
	"encoding/binary"
	"math"
	"testing"
)

func ByteOrderTest(bo binary.ByteOrder, t *testing.T) {
	var got float64
	b := new(bytes.Buffer)
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
		b.Reset()
		ByteOrderValue(bo, want).WriteTo(b)
		ByteOrderPointer(bo, &got).ReadFrom(b)
		if got != want {
			t.Errorf("%#x != %#x", got, want)
		}
	}
}

func TestBig(t *testing.T)    { ByteOrderTest(binary.BigEndian, t) }
func TestLittle(t *testing.T) { ByteOrderTest(binary.LittleEndian, t) }
func TestNative(t *testing.T) { ByteOrderTest(binary.NativeEndian, t) }
