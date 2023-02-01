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

	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var (
	// test overwrides
	Getenv = os.Getenv

	UserCacheDir = os.UserCacheDir

	UserConfigDir = os.UserConfigDir

	UserHomeDir = os.UserHomeDir
)

// If SU, make an XDG path w/ 0755 permissions or 0700 otherwise.
func MkPath(s string) error {
	var perm os.FileMode = 0700
	if program.IsSuperUser() {
		perm = 0755
	}
	return os.MkdirAll(s, perm)
}

// If available, returns $XDG_CACHE_HOME.  If SU, returns /var/cache or
// /var/run; otherwise, if not SU, returns UserCacheDir or TempDir.
var CacheHome = cache.NewReadOnly[string](
	func(p *string) error {
		if *p = Getenv("XDG_CACHE_HOME"); len(*p) > 0 {
		} else if program.IsSuperUser() {
			*p = SU.CacheHome.Value()
		} else if d, derr := UserCacheDir(); derr == nil {
			*p = d
		} else {
			*p = os.TempDir()
		}
		return nil
	})

// If available, returns $XDG_CONFIG_DIRS; otherwise, returns /etc/xdg.
var ConfigDirs = cache.New[string](
	func(p *string) error {
		if *p = Getenv("XDG_CONFIG_DIRS"); len(*p) == 0 {
			*p = "/etc/xdg"
		}
		return nil
	})

// If available, returns $XDG_CONFIG_HOME.  If SU, returns /etc/opt if opt
// program; or /etc; otherwise, if not SU, returns UserConfigDir or TempDir.
var ConfigHome = cache.New[string](
	func(p *string) error {
		if *p = Getenv("XDG_CONFIG_HOME"); len(*p) > 0 {
		} else if program.IsSuperUser() {
			if program.IsOpt() {
				*p = "/etc/opt"
			} else {
				*p = "/etc"
			}
		} else if d, err := UserConfigDir(); err == nil {
			*p = d
		} else {
			*p = os.TempDir()
		}
		return nil
	})

// If available, returns $XDG_DATA_DIRS; otherwise,
// /usr/local/share:/usr/share.
var DataDirs = cache.New[string](
	func(p *string) error {
		if *p = Getenv("XDG_DATA_DIRS"); len(*p) == 0 {
			*p = "/usr/local/share:/usr/share"
		}
		return nil
	})

// If available, returns $XDG_DATA_HOME.  If SU, returns /usr/local/share if
// local program; /opt/share if opt program; or /usr/share; otherwise, if not
// SU, returns UserHomeDir or TempDir.
var DataHome = cache.New[string](
	func(p *string) error {
		if *p = Getenv("XDG_DATA_HOME"); len(*p) > 0 {
		} else if program.IsSuperUser() {
			if program.IsUsrLocal() {
				*p = "/usr/local/share"
			} else if program.IsOpt() {
				*p = "/opt/share"
			} else {
				*p = "/usr/share"
			}
		} else if h, err := UserHomeDir(); err == nil {
			*p = filepath.Join(h, ".local", "share")
		} else {
			*p = os.TempDir()
		}
		return nil
	})

// If available, returns $XDG_RUNTIME_DIR; or if SU, "/var/run"; otherwise,
// UserCacheDir or TempDir.
var RunTimeDir = cache.New[string](
	func(p *string) error {
		if *p = Getenv("XDG_RUNTIME_DIR"); len(*p) > 0 {
		} else if program.IsSuperUser() {
			*p = SU.RunTimeDir.Value()
		} else if d, err := UserCacheDir(); err == nil {
			*p = d
		} else {
			*p = os.TempDir()
		}
		return nil
	})

// If available, returns $XDG_STATE_HOME. If SU, returns "/var/local"
// if local program; "/var/opt" if opt program; or "/var/lib"
// otherwise. if not SU and no $XDG_STATE_HOME, returns
// UserHomeDir()/.local/state.
var StateHome = cache.New[string](
	func(p *string) error {
		if *p = Getenv("XDG_STATE_HOME"); len(*p) > 0 {
		} else if program.IsSuperUser() {
			if program.IsUsrLocal() {
				*p = "/var/local"
			} else if program.IsOpt() {
				*p = "/var/opt"
			} else {
				*p = "/var/lib"
			}
		} else if h, err := UserHomeDir(); err == nil {
			*p = filepath.Join(h, ".local", "state")
		} else {
			*p = os.TempDir()
		}
		return nil
	})

var SU = struct {
	CacheHome,
	RunTimeDir *cache.Cache[string]
}{
	CacheHome: cache.New[string](func(p *string) error {
		for _, *p = range []string{"/var/cache", "/var/run"} {
			if _, err := os.Stat(*p); err == nil {
				return nil
			}
		}
		*p = os.TempDir()
		return nil
	}),
	RunTimeDir: cache.New[string](func(p *string) error {
		const var_run = "/var/run"
		if _, err := os.Stat(var_run); err == nil {
			*p = var_run
		} else {
			*p = os.TempDir()
		}
		return nil
	}),
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
