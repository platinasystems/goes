// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xmaps

// Install is a deep copy of each source w/o overwriting existing entries.
func Install[M ~map[K]V, K comparable, V any](m M, sources ...M) {
	for _, src := range sources {
		for k, v := range src {
			if mV, ok := m[k]; ok {
				if vM, ok := any(v).(M); ok {
					if mM, ok := any(mV).(M); ok {
						Install(mM, vM)
					}
				}
			} else {
				m[k] = v
			}
		}
	}
}
