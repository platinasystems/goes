// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package elib

import (
	"math/rand"
	"testing"
)

func slowCompress(mask uint64, value uint64) (r uint64) {
	n := uint(0)
	for i := uint(0); i < 64; i++ {
		t := uint64(1) << i
		if mask&t != 0 {
			if value&t != 0 {
				r |= 1 << n
			}
			n++
		}
	}
	return
}

func TestBitCompress(t *testing.T) {
	var b BitCompressUint64

	n := 100000

	for i := 0; i < n; i++ {
		mask := uint64(rand.Int63())
		value := uint64(rand.Int63()) & mask

		b.SetMask(mask)
		if got, want := b.Compress(value), slowCompress(mask, value); got != want {
			t.Errorf("failed %d != %d", got, want)
		}
	}
}
