// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netif

// Copy from string to int8 or uint8 (aka. byte) slice and null remainder.
func Rename[B uint8 | int8](dst []B, name string) {
	n := len(dst[:])
	for i, b := range []byte(name) {
		if i < n {
			dst[i] = B(b)
		}
	}
	for i := len(name); i < n; i++ {
		dst[i] = 0
	}
}
