// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package little

import (
	"testing"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/endiantest"
)

func Test(t *testing.T) {
	for _, v := range endiantest.Uints {
		endiantest.Run[uint16](t, new(Uint16), uint16(v))
		endiantest.Run[uint32](t, new(Uint32), uint32(v))
		endiantest.Run[uint64](t, new(Uint64), uint64(v))
	}
	for _, v := range endiantest.Floats {
		endiantest.Run[float32](t, new(Float32), float32(v))
		endiantest.Run[float64](t, new(Float64), float64(v))
	}
}
