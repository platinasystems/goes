// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package slice

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

func Index[T comparable](l []T, v T) int {
	for i, lv := range l {
		if lv == v {
			return i
		}
	}
	return -1
}
