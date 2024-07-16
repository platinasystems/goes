// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xslices

import "slices"

// Prepend first [slices.Grow]'s slice “s” by the number of “args”;
// then moves the current contents before prefacing.
func Prepend[S ~[]E, E any](s S, args ...E) S {
	n := len(args)
	ś := slices.Grow(s, n)
	ś = ś[:len(s)+n]
	copy(ś[n:], s)
	copy(ś[:n], args)
	return ś
}
