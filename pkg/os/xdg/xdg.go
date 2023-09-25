// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides [XDG] or, if Unix super user, [FSHS] defined paths.
//
//	XDG  https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
//	FSHS https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard
package xdg

import (
	"os"
	"path/filepath"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/os/program"
)

var (
	// test overwrides
	Getenv = os.Getenv

	UserCacheDir = os.UserCacheDir

	UserConfigDir = os.UserConfigDir

	UserHomeDir = os.UserHomeDir
)

var IsSuperUser = sync.OnceValue(func() bool {
	return os.Geteuid() == 0
})

// If SU, make an XDG path w/ 0755 permissions or 0700 otherwise.
func MkPath(s string) error {
	var perm os.FileMode = 0700
	if IsSuperUser() {
		perm = 0755
	}
	return os.MkdirAll(s, perm)
}

// If available, returns $XDG_CACHE_HOME.  If SU, returns /var/cache or
// /var/run; otherwise, if not SU, returns UserCacheDir or TempDir.
var CacheHome = sync.OnceValue(cacheHome)

func cacheHome() string {
	if s := Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		return SU.CacheHome()
	} else if d, err := UserCacheDir(); err == nil {
		return d
	} else {
		return os.TempDir()
	}
	return ""
}

// If available, returns $XDG_CONFIG_DIRS; otherwise, returns /etc/xdg.
var ConfigDirs = sync.OnceValue(configDirs)

func configDirs() string {
	if s := Getenv("XDG_CONFIG_DIRS"); len(s) == 0 {
		return s
	}
	return "/etc/xdg"
}

// If available, returns $XDG_CONFIG_HOME.  If SU, returns /etc/opt if opt
// program; or /etc; otherwise, if not SU, returns UserConfigDir or TempDir.
var ConfigHome = sync.OnceValue(configHome)

func configHome() string {
	if s := Getenv("XDG_CONFIG_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		if program.IsOpt() {
			return "/etc/opt"
		} else {
			return "/etc"
		}
	} else if d, err := UserConfigDir(); err == nil {
		return d
	}
	return os.TempDir()
}

// If available, returns $XDG_DATA_DIRS; otherwise,
// /usr/local/share:/usr/share.
var DataDirs = sync.OnceValue(dataDirs)

func dataDirs() string {
	if s := Getenv("XDG_DATA_DIRS"); len(s) > 0 {
		return s
	}
	return "/usr/local/share:/usr/share"
}

// If available, returns $XDG_DATA_HOME.  If SU, returns /usr/local/share if
// local program; /opt/share if opt program; or /usr/share; otherwise, if not
// SU, returns UserHomeDir or TempDir.
var DataHome = sync.OnceValue(dataHome)

func dataHome() string {
	if s := Getenv("XDG_DATA_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		if program.IsUsrLocal() {
			return "/usr/local/share"
		} else if program.IsOpt() {
			return "/opt/share"
		}
		return "/usr/share"
	} else if h, err := UserHomeDir(); err == nil {
		return filepath.Join(h, ".local", "share")
	}
	return os.TempDir()
}

// If available, returns $XDG_RUNTIME_DIR; or if SU, "/var/run"; otherwise,
// UserCacheDir or TempDir.
var RunTimeDir = sync.OnceValue(runTimeDir)

func runTimeDir() string {
	if s := Getenv("XDG_RUNTIME_DIR"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		return SU.RunTimeDir()
	} else if d, err := UserCacheDir(); err == nil {
		return d
	}
	return os.TempDir()
}

// If available, returns $XDG_STATE_HOME. If SU, returns "/var/local"
// if local program; "/var/opt" if opt program; or "/var/lib"
// otherwise. if not SU and no $XDG_STATE_HOME, returns
// UserHomeDir()/.local/state.
var StateHome = sync.OnceValue(stateHome)

func stateHome() string {
	if s := Getenv("XDG_STATE_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		if program.IsUsrLocal() {
			return "/var/local"
		} else if program.IsOpt() {
			return "/var/opt"
		}
		return "/var/lib"
	} else if h, err := UserHomeDir(); err == nil {
		return filepath.Join(h, ".local", "state")
	}
	return os.TempDir()
}

var SU = struct {
	CacheHome, RunTimeDir func() string
}{
	CacheHome:  sync.OnceValue(suCacheHome),
	RunTimeDir: sync.OnceValue(suRunTimeDir),
}

func suCacheHome() string {
	for _, s := range []string{"/var/cache", "/var/run"} {
		if _, err := os.Stat(s); err == nil {
			return s
		}
	}
	return os.TempDir()
}

func suRunTimeDir() string {
	const var_run = "/var/run"
	if _, err := os.Stat(var_run); err == nil {
		return var_run
	}
	return os.TempDir()
}

var Dirs = map[string]any{
	"cache-home":   CacheHome,
	"config-dirs":  ConfigDirs,
	"config-home":  ConfigHome,
	"data-dirs":    DataDirs,
	"data-home":    DataHome,
	"run-time-dir": RunTimeDir,
	"state-home":   StateHome,
}
