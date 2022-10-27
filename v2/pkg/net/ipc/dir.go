// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"github.com/platinasystems/goes/v2/pkg/os/xdg"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	Dir = cache.New[string](func(p *string) error {
		if DirFlag != nil {
			*p = *DirFlag
		} else {
			*p = DefaultDir()
		}
		return nil
	}).Value
	DirFlag    *string
	DefaultDir = xdg.RunTimeDir
)
