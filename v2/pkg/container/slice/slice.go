// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package slice

// Remove the i'th through i+n element(s) from slice.
func Cut[T any](l []T, i, n uint) []T {
	copy(l[i:], l[i+n:])
	return l[:uint(len(l))-n]
}

// Remove all matching elements from slice.
func Filter[T comparable](l []T, v T) []T {
	for i, lv := range l {
		if lv == v {
			n := len(l) - 1
			f := make([]T, n, n)
			if i > 0 {
				copy(f[:i], l[:i])
			}
			if i < n {
				copy(f[i:], Filter[T](l[i+1:], v))
			}
			return f
		}
	}
	return l
}

// Returns the first positive index of value w/in slice.
func Index[T comparable](l []T, v T) int {
	for i, lv := range l {
		if lv == v {
			return i
		}
	}
	return -1
}
