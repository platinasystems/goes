// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package binint

import (
	"bytes"
	"encoding/binary"
	"testing"
)

func ByteOrderTest(bo binary.ByteOrder, t *testing.T) {
	var got uint64
	b := new(bytes.Buffer)
	t.Helper()
	for _, want := range []uint64{
		0x0000000000000000,
		0x0123456789abcdef,
		0xfedcba9876543210,
		0xffffffffffffffff,
		0xaaaaaaaaaaaaaaaa,
	} {
		b.Reset()
		ByteOrderValue(bo, want).WriteTo(b)
		t.Logf("%#x", b.Bytes())
		ByteOrderPointer(bo, &got).ReadFrom(b)
		if got != want {
			t.Errorf("%#x != %#x", got, want)
		}
	}
}

func TestBig(t *testing.T)    { ByteOrderTest(binary.BigEndian, t) }
func TestLittle(t *testing.T) { ByteOrderTest(binary.LittleEndian, t) }
func TestNative(t *testing.T) { ByteOrderTest(binary.NativeEndian, t) }
