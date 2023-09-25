// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package host

import "testing"

func TestNet16(t *testing.T) {
	testnet[uint16](t, Net16(0x1122), 0x2211)
}

func TestNet32(t *testing.T) {
	testnet[uint32](t, Net32(0x11223344), 0x44332211)
}

func TestNet64(t *testing.T) {
	testnet[uint64](t, Net64(0x1122334455667788), 0x8877665544332211)
}

func testnet[T uint16 | uint32 | uint64](t *testing.T, n, little T) {
	t.Helper()
	x := n
	if IsLittleEndian {
		x = little
	}
	if n != x {
		t.Errorf("got: %#x", x)
	}
}
