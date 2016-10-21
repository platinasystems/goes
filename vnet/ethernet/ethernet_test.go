// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ethernet

import (
	"testing"
)

func TestAddressBlock(t *testing.T) {
	b := &AddressBlock{
		Base:  Address{0xa, 0xb, 0xc, 0x0, 0x0, 0x0},
		Count: 257,
	}
	as := make([]Address, b.Count)
	for i := range as {
		as[i] = b.Base
		as[i][AddressBytes-1] = uint8(i)
		if i >= 256 {
			as[i][AddressBytes-2] = uint8(i >> 8)
		}
	}
	for i := range as {
		if a, ok := b.Alloc(); !ok || !a.Equal(as[i]) {
			t.Errorf("alloc %d: got %v %v want %v", i, &a, ok, &as[i])
		}
	}
	b.Free(&as[0])
	if a, ok := b.Alloc(); !ok || !a.Equal(as[0]) {
		t.Errorf("alloc 3: got %v %v want %v", &a, ok, &as[0])
	}
	b.Free(&as[0])
}
