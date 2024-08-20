// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// See https://specifications.freedesktop.org/basedir-spec/latest/
package xdg

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var (
	Etc    = filepath.FromSlash("/etc")
	EtcOpt = filepath.FromSlash("/etc/opt")

	Opt      = filepath.FromSlash("/opt")
	OptShare = filepath.FromSlash("/opt/share")

	UsrLocal      = filepath.FromSlash("/usr/local")
	UsrLocalShare = filepath.FromSlash("/usr/local/share")

	UsrShare = filepath.FromSlash("/usr/share")

	VarCache = filepath.FromSlash("/var/cache")
	VarLocal = filepath.FromSlash("/var/local")
	VarLib   = filepath.FromSlash("/var/lib")
	VarOpt   = filepath.FromSlash("/var/opt")
	VarRun   = filepath.FromSlash("/var/run")

	PathListSeparatorString = string(os.PathListSeparator)
)

func IsOpt() bool {
	return strings.HasPrefix(xprogram.Path(), Opt)
}

func IsUsrLocal() bool {
	return strings.HasPrefix(xprogram.Path(), UsrLocal)
}

func IsSuperUser() bool {
	return os.Geteuid() == 0
}

// Returns [os.TempDir] if none of the “dirs” exist.
func FirstExistingDir(dirs ...string) string {
	for _, dir := range dirs {
		if len(dir) == 0 {
			continue
		}
		if fi, err := os.Stat(dir); err == nil && fi.IsDir() {
			return dir
		}
	}
	return os.TempDir()
}

// CacheHome has non-essential, ephemeral data.
func CacheHome() string {
	s := os.Getenv("XDG_CACHE_HOME")
	if len(s) == 0 {
		s = goosCacheHome()
	}
	return s
}

func ConfigDirs() (dirs []string) {
	if s := os.Getenv("XDG_CONFIG_DIRS"); len(s) > 0 {
		dirs = strings.Split(s, PathListSeparatorString)
	} else {
		dirs = goosConfigDirs()
	}
	return
}

// ConfigHome has persistent configuration.
func ConfigHome() string {
	s := os.Getenv("XDG_CONFIG_HOME")
	if len(s) == 0 {
		s = goosConfigHome()
	}
	return s
}

func DataDirs() (dirs []string) {
	if s := os.Getenv("XDG_DATA_DIRS"); len(s) > 0 {
		dirs = strings.Split(s, PathListSeparatorString)
	} else {
		dirs = goosDataDirs()
	}
	return
}

// DataHome has essential, persistent data.
func DataHome() string {
	s := os.Getenv("XDG_DATA_HOME")
	if len(s) == 0 {
		s = goosDataHome()
	}
	return s
}

// RunTimeDir has non-essential, ephemeral runtime files and other objects,
// such as sockets and named pipes.
func RunTimeDir() string {
	s := os.Getenv("XDG_RUNTIME_DIR")
	if len(s) == 0 {
		s = goosRunTimeDir()
	}
	return s
}

// StateHome persists between application restarts,
//
//   - actions history (logs, history, recently used files, …)
//
//   - current state of the application that can be reused on a restart
//     (view, layout, open files, undo history, …)
func StateHome() string {
	s := os.Getenv("XDG_STATE_HOME")
	if len(s) == 0 {
		s = goosStateHome()
	}
	return s
}
