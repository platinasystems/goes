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
	DirFlag    *string
	DirDefault = cache.New[string](func(p *string) error {
		sh, err := xdg.StateHome.Value()
		if err != nil {
			return err
		}
		base, err := program.Base.Value()
		if err == nil {
			*p = filepath.Join(sh, base)
		}
		return err
	})
	Dir = cache.New[string](func(p *string) (err error) {
		if DirFlag != nil && len(*DirFlag) > 0 {
			*p = *DirFlag
		} else {
			*p, err = DirDefault.Value()
		}
		return
	})
)

func MkDir() error {
	dir, err := Dir.Value()
	if err == nil {
		err = xdg.MkPath(dir)
	}
	return err
}
