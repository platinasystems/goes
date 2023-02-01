// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package dir

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

const VarRunKO = "/var/run/ko"

var (
	Default = cache.New[string](func(p *string) error {
		if _, err := os.Stat(VarRunKO); err == nil {
			*p = VarRunKO
		} else {
			filepath.Join(xdg.StateHome.Value(), program.Base())
		}
		return nil
	})
	Flag = flag.String("state", Default.Value(), "Certificate directory.")
	Name = cache.New[string](func(p *string) error {
		*p = *Flag
		return nil
	})
)

func Mk() error {
	return xdg.MkPath(Name.Value())
}
