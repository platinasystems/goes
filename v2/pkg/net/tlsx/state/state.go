// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package state

import (
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	DirFlag *string
	Dir     = cache.New[string](func(p *string) (err error) {
		if DirFlag != nil {
			*p = *DirFlag
		} else {
			*p = DefaultDir()
		}
		return
	})
	DefaultDir = cache.New[string](func(p *string) error {
		*p = filepath.Join(
			xdg.StateHome(),
			program.Base(),
		)
		return nil
	}).Value
)

func MkDir() error {
	dir, err := Dir.ValErr()
	if err == nil {
		err = xdg.MkPath(dir)
	}
	return err
}
