// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xmaps

// Returns an unsorted list of all, or with patterns, matching map keys.
func Match[M ~map[K]V, K comparable, V any](
	m M, ismatch func(key, pat K) bool, patterns ...K,
) []K {
	matches := make([]K, 0, len(m))
	for k := range m {
		if len(patterns) == 0 {
			matches = append(matches, k)
		} else {
			for _, pat := range patterns {
				if ismatch(k, pat) {
					matches = append(matches, k)
				}
			}
		}
	}
	return matches
}
