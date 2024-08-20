// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/xdg"
)

func main() {
	fmt.Printf("XDG_CACHE_HOME: %q\n", xdg.CacheHome())
	fmt.Printf("XDG_CONFIG_HOME: %q\n", xdg.ConfigHome())
	fmt.Printf("XDG_CONFIG_DIRS: %q\n", xdg.ConfigDirs())
	fmt.Printf("XDG_DATA_HOME: %q\n", xdg.DataHome())
	fmt.Printf("XDG_DATA_DIRS: %q\n", xdg.DataDirs())
	fmt.Printf("XDG_RUNTIME_DIS: %q\n", xdg.RunTimeDir())
	fmt.Printf("XDG_STATE_HOME: %q\n", xdg.StateHome())
}
