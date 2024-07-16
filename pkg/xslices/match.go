// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xslices

// Match returns an list of matching elements in order found.
func Match[S ~[]E, E any](
	s S, ismatch func(element, pat E) bool, pat E,
) []E {
	matches := make([]E, 0, len(s))
	for _, e := range s {
		if ismatch(e, pat) {
			matches = append(matches, e)
		}
	}
	return matches
}
