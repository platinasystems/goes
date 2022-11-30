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
	Cache = struct {
		Dir *cache.Cache[string]
	}{
		Dir: cache.New[string](func(p *string) error {
			*p = *Flag
			return nil
		}),
	}
	Default = filepath.Join(xdg.StateHome(), program.Base())
	Dir     = Cache.Dir.Value
	Flag    = flag.String("state", Default, "certificate directory")
)

func MkDir() error {
	return xdg.MkPath(Dir())
}
