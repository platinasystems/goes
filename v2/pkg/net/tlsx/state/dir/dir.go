// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dir

import (
	"flag"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	Cached = struct {
		Name *cache.Cache[string]
	}{
		Name: cache.New[string](func(p *string) error {
			*p = *Flag
			return nil
		}),
	}
	Default = filepath.Join(xdg.StateHome(), program.Base())
	Flag    = flag.String("state", Default, "certificate directory")
	Name    = Cached.Name.Value
)

func Mk() error {
	return xdg.MkPath(Name())
}
