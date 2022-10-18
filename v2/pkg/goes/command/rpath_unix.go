// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package command

import (
	"os"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
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
	RestrictedUserPath = []string{
		"/usr/local/bin",
		"/usr/bin",
		"/bin",
	}
	RestrictedPath = cache.New[[]string](func(p *[]string) (err error) {
		if program.IsSuperUser.Value() {
			*p = RestrictedSuperUserPath
		} else {
			*p = RestrictedUserPath
		}
		return
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
