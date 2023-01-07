// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package little

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
}
