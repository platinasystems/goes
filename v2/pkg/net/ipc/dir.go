// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"flag"

	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	Dir = cache.New[string](func(p *string) error {
		if len(*DirFlag) > 0 {
			*p = *DirFlag
		} else {
			*p = xdg.RunTimeDir.Value()
		}
		return nil
	}).Value
	DirFlag = flag.String("ipc", "", "socket directory"+
		"(default $XDG_RUNTIME_DIR)")
)
