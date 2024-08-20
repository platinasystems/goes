// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix

package xdg

import (
	"os"
	"os/user"
	"sync"
)

// If “$SUDO_USER” isn't empty, return its home instead of current user.
var SudoUserHome = sync.OnceValue(func() string {
	if uname := os.Getenv("SUDO_USER"); len(uname) > 0 {
		if u, err := user.Lookup(uname); err == nil {
			return u.HomeDir
		}
	} else if u, err := user.Current(); err == nil {
		return u.HomeDir
	}
	return os.Getenv("HOME")
})
