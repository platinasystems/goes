// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xexec

import (
	"os"
	"sync"
)

var (
	RestrictedSuperUserPath = []string{
		"/usr/local/sbin",
		"/usr/local/bin",
		"/usr/sbin",
		"/usr/bin",
		"/sbin",
		"/bin",
	}
	RestrictedOrdinaryUserPath = []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
	}
	RestrictedCurrentUserPath = sync.OnceValue(func() []string {
		rp := RestrictedOrdinaryUserPath
		if os.Geteuid() == 0 {
			rp = RestrictedSuperUserPath
		}
		return rp
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
