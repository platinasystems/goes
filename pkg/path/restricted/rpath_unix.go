// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package restricted

import (
	"os"
	"sync"
)

var (
	SuperUserPath = []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
	}
	OrdinaryUserPath = []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
	}
	Path = sync.OnceValue(func() []string {
		if os.Geteuid() == 0 {
			return SuperUserPath
		}
		return OrdinaryUserPath
	})
)

func IsExecutable(full string) bool {
	fi, err := os.Stat(full)
	if err != nil {
		return false
	}
	m := fi.Mode()
	return !m.IsDir() && m&0111 != 0
}
