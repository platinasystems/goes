// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build ignore

package goes

import (
	"sort"
	"strings"
)

// Return sorted map keys.
func Keys(m map[string]any) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		if k != "daemon" && !strings.HasPrefix(k, "_") {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	return keys
}
