// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xexec

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// Match PATH program prefix.  Returns a sorted list.
func Match(prefix string) (c []string) {
	if strings.Contains(prefix, "/") {
		return
	}
	path := os.Getenv("PATH")
	if len(path) == 0 {
		return
	}
	defer sort.Strings(c)
	for _, dn := range filepath.SplitList(path) {
		dir, err := os.ReadDir(dn)
		if err != nil {
			continue
		}
		for _, de := range dir {
			fn := de.Name()
			if de.IsDir() {
				continue
			}
			if !strings.HasPrefix(fn, prefix) {
				continue
			}
			pn := filepath.Join(dn, fn)
			if err = Access(pn, X_OK); err == nil {
				c = append(c, fn)
			}
		}
	}
	return
}
