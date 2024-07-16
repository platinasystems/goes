// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package nonce

const Size = 12

// BigEndian add `y` to clone of `x`.
func Sum(x, y []byte) (z []byte) {
	var carry bool
	z = append([]byte{}, x...)
	for i, j := len(z)-1, len(y)-1; i >= 0 && j >= 0; i, j = i-1, j-1 {
		z[i] += y[j]
		if carry {
			z[i] += 1
		}
		carry = z[i] < x[i]
	}
	return
}

// Xor `y` with clone of `x`.
func Xor(x, y []byte) (z []byte) {
	z = append([]byte{}, x...)
	for i, b := range x {
		if i >= len(z) {
			break
		}
		z[i] ^= b
	}
	return
}
