// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// Package fhs (Filesystem Hierarchy Standard) returns GOOS directories.
// See,
//
//   - <https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard>
//   - <https://developer.apple.com/library/archive/documentation/FileManagement/Conceptual/FileSystemProgrammingGuide/FileSystemOverview/FileSystemOverview.html#//apple_ref/doc/uid/TP40010672-CH2-SW14>
//   - <https://man.freebsd.org/cgi/man.cgi?hier>
package fhs

import (
	"runtime"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

// Cache has non-essential, ephemeral data.
func Cache() string {
	switch runtime.GOOS {
	case "darwin", "ios":
		return "/Library/Caches"
	}
	return "/var/cache"
}

// Config has persistent configuration.
func Config() string {
	switch runtime.GOOS {
	case "darwin", "ios":
		return "/Library/Application Support"
	case "freebsd":
		if xprogram.IsUsrLocal() {
			return "/usr/local/etc"
		}
		return "/etc"

	}
	if xprogram.IsOpt() {
		return "/etc/opt"
	}
	return "/etc"
}

// Data has essential, persistent data.
func Data() string {
	switch runtime.GOOS {
	case "darwin", "ios":
		return "/Library/Application Support"
	case "freebsd":
		if xprogram.IsUsrLocal() {
			return "/usr/local"
		}
		return "/usr/share"
	}
	if xprogram.IsUsrLocal() {
		return "/usr/local"
	} else if xprogram.IsOpt() {
		return "/opt/share"
	}
	return "/usr/share"
}

// RunTime has non-essential, ephemeral runtime files and other objects,
// such as sockets and named pipes.
func RunTime() string {
	switch runtime.GOOS {
	case "darwin", "ios":
		return "/Library/Application Support"
	}
	return "/var/run"
}

// State persists between application restarts,
//
//   - actions history (logs, history, recently used files, …)
//
//   - current state of the application that can be reused on a restart
//     (view, layout, open files, undo history, …)
func State() string {
	switch runtime.GOOS {
	case "darwin", "ios", "freebsd":
		return "/var/lib"

	}
	if xprogram.IsUsrLocal() {
		return "/var/local"
	} else if xprogram.IsOpt() {
		return "/var/opt"
	}
	return "/var/lib"
}
