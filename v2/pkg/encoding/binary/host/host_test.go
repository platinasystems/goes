// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import "testing"

func Test(t *testing.T) {
	n16 := Net16(0x1122)
	n32 := Net32(0x11223344)
	n64 := Net64(0x1122334455667788)
	x16, x32, x64 := n16, n32, n64
	if IsLittleEndian {
		x16, x32, x64 = 0x2211, 0x44332211, 0x8877665544332211
	}
	if n16 != x16 {
		t.Error(n16, "!=", x16)
	}
	if n32 != x32 {
		t.Error(n32, "!=", x32)
	}
	if n64 != x64 {
		t.Error(n64, "!=", x64)
	}
}
