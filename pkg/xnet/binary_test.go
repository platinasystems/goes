// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xnet

import "testing"

func TestNetUint64(t *testing.T) {
	var err error
	var got uint64
	samples := []uint64{
		0x0000000000000000,
		0x0123456789abcdef,
		0xfedcba9876543210,
		0xffffffffffffffff,
		0xaaaaaaaaaaaaaaaa,
	}
	b := make([]byte, 0, len(samples)*8)
	for _, want := range samples {
		b = b[:0]
		if b, err = Add(b, want); err != nil {
			t.Fatalf("%#x: %v", want, err)
		}
		if _, err = Subtract(b, &got); err != nil {
			t.Fatalf("%#x: %v", want, err)
		}
		if got != want {
			t.Fatalf("%#x != %#x", got, want)
		}
	}
	b = b[:0]
	for _, want := range samples {
		if b, err = Add(b, want); err != nil {
			t.Fatalf("%#x: %v", want, err)
		}
	}
	for i, want := range samples {
		if b, err = Subtract(b, &got); err != nil {
			t.Fatalf("%d:%#x: %v", i, want, err)
		}
		if got != want {
			t.Fatalf("%d:%#x != %#x", i, got, want)
		}
	}
}

func TestNetStruct(t *testing.T) {
	var err error
	var got, want struct {
		U16 uint16
		_   uint16
		U32 uint32
		_   uint32
		U64 uint64
	}
	want.U16 = 0x0123
	want.U32 = 0x456789ab
	want.U64 = 0xcdef0123
	b := make([]byte, 0, 2+2+4+4+8)
	if b, err = Add(b, want); err != nil {
		t.Fatal(err)
	}
	if _, err = Subtract(b, &got); err != nil {
		t.Fatal(err)
	}
	if got.U16 != want.U16 {
		t.Fatalf("%#x != %#x", got.U16, want.U16)
	}
	if got.U32 != want.U32 {
		t.Fatalf("%#x != %#x", got.U32, want.U32)
	}
	if got.U64 != want.U64 {
		t.Fatalf("%#x != %#x", got.U64, want.U64)
	}
}
