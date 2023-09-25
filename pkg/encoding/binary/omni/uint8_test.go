// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package omni

import (
	"testing"

	"github.com/platinasystems/goes/v2/pkg/encoding/binary/endiantest"
)

func TestUint8(t *testing.T) {
	for _, v := range endiantest.Uints {
		endiantest.Run[uint8](t, new(Uint8), uint8(v))
	}
}
