// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package state

import (
	"flag"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	DirDefault = cache.New[string](func(p *string) error {
		sh, err := xdg.StateHome.ValErr()
		if err != nil {
			return err
		}
		base, err := program.Base.ValErr()
		if err == nil {
			*p = filepath.Join(sh, base)
		}
		return err
	})
	Dir = cache.New[string](func(p *string) (err error) {
		if state := flag.Lookup("state"); state != nil {
			*p = state.Value.String()
		}
		if len(*p) == 0 {
			*p, err = DirDefault.ValErr()
		}
		return
	})
)

func MkDir() error {
	dir, err := Dir.ValErr()
	if err == nil {
		err = xdg.MkPath(dir)
	}
	return err
}
