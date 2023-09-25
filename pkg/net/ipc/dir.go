// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"flag"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/os/xdg"
)

var (
	Dir = sync.OnceValue(func() string {
		if len(dir) > 0 {
			return dir
		}
		return xdg.RunTimeDir()
	})
)

var dir string

func FlagDir(fs *flag.FlagSet) {
	fs.StringVar(&dir, "ipc", "", "socket directory"+
		"(default $XDG_RUNTIME_DIR)")
}
