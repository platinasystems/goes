// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package endian

import (
	"math"
	"testing"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/big"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/host"
	"github.com/platinasystems/goes/v2/pkg/encoding/binary/little"
)

func Test(t *testing.T) {
	b := make([]byte, 8)
	t.Run("uints", func(t *testing.T) {
		for _, v := range []uint64{
			0x0000000000000000,
			0x0123456789abcdef,
			0xfedcba9876543210,
			0xffffffffffffffff,
			0xaaaaaaaaaaaaaaaa,
		} {
			t.Run("big", func(t *testing.T) {
				ut[uint8](t, big.Uint8(b[:1]), uint8(v))
				ut[uint16](t, big.Uint16(b[:2]), uint16(v))
				ut[uint32](t, big.Uint32(b[:4]), uint32(v))
				ut[uint64](t, big.Uint64(b[:8]), uint64(v))
			})
			t.Run("little", func(t *testing.T) {
				ut[uint8](t, little.Uint8(b[:1]), uint8(v))
				ut[uint16](t, little.Uint16(b[:2]), uint16(v))
				ut[uint32](t, little.Uint32(b[:4]), uint32(v))
				ut[uint64](t, little.Uint64(b[:8]), uint64(v))
			})
			t.Run("host", func(t *testing.T) {
				ut[uint8](t, host.Uint8(b[:1]), uint8(v))
				ut[uint16](t, host.Uint16(b[:2]), uint16(v))
				ut[uint32](t, host.Uint32(b[:4]), uint32(v))
				ut[uint64](t, host.Uint64(b[:8]), uint64(v))
			})
		}
	})
	t.Run("floats", func(t *testing.T) {
		for _, v := range []float64{
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
			t.Run("big", func(t *testing.T) {
				ut[float32](t, big.Float32(b[:4]), float32(v))
				ut[float64](t, big.Float64(b[:8]), float64(v))
			})
			t.Run("little", func(t *testing.T) {
				ut[float32](t, little.Float32(b[:4]), float32(v))
				ut[float64](t, little.Float64(b[:8]), float64(v))
			})
			t.Run("host", func(t *testing.T) {
				ut[float32](t, host.Float32(b[:4]), float32(v))
				ut[float64](t, host.Float64(b[:8]), float64(v))
			})
		}
	})
}

func ut[T Number](t *testing.T, xe Endian[T], want T) {
	t.Helper()
	xe.Put(want)
	if got := xe.Value(); got != want {
		t.Errorf("%T %#x != %#x", xe, got, want)
	}
}
