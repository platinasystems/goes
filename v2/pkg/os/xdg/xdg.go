// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides [XDG] or, if Unix root user, [FSHS] defined paths.
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
	if program.IsSuperUser.Value() {
		perm = 0755
	}
	return os.MkdirAll(s, perm)
}

// If SU, returns /var/cache or /var/run; otherwise, $XDG_CACHE_HOME,
// UserCacheDir or TempDir.
var CacheHome = cache.New[string](func(p *string) (err error) {
	if program.IsSuperUser.Value() {
		if *p, err = rootCacheHome.ValErr(); err == nil {
			return
		}
	}
	if *p = Getenv("XDG_CACHE_HOME"); len(*p) == 0 {
		if d, derr := UserCacheDir(); derr == nil {
			*p = d
		} else {
			*p = os.TempDir()
		}
	}
	err = nil
	return
})

var ConfigDirs = cache.New[string](func(p *string) error {
	if *p = Getenv("XDG_CONFIG_DIRS"); len(*p) == 0 {
		*p = "/etc/xdg"
	}
	return nil
})

// If SU, returns "/etc/opt" if opt program; or "/etc" otherwise.
var ConfigHome = cache.New[string](func(p *string) error {
	if program.IsSuperUser.Value() {
		if program.IsOpt.Value() {
			*p = "/etc/opt"
		} else {
			*p = "/etc"
		}
	} else if *p = Getenv("XDG_CONFIG_HOME"); len(*p) == 0 {
		if d, err := UserConfigDir(); err == nil {
			*p = d
		} else {
			*p = os.TempDir()
		}
	}
	return nil
})

var DataDirs = cache.New[string](func(p *string) error {
	if *p = Getenv("XDG_DATA_DIRS"); len(*p) == 0 {
		*p = "/usr/local/share:/usr/share"
	}
	return nil
})

// If SU, returns "/usr/local/share" if local program; "/opt/share" if opt
// program; or "/usr/share" otherwise.
var DataHome = cache.New[string](func(p *string) error {
	if program.IsSuperUser.Value() {
		if program.IsUsrLocal.Value() {
			*p = "/usr/local/share"
		} else if program.IsOpt.Value() {
			*p = "/opt/share"
		} else {
			*p = "/usr/share"
		}
	} else if *p = Getenv("XDG_DATA_HOME"); len(*p) == 0 {
		if h, err := UserHomeDir(); err == nil {
			*p = filepath.Join(h, ".local", "share")
		} else {
			*p = os.TempDir()
		}
	}
	return nil
})

// If SU, returns "/var/run", if available; otherwise, $XDG_RUNTIME_DIR or
// UserCacheDir.
var RunTimeDir = cache.New[string](func(p *string) (err error) {
	if program.IsSuperUser.Value() {
		if *p, err = rootRunTimeDir.ValErr(); err == nil {
			return
		}
	}
	if *p = Getenv("XDG_RUNTIME_DIR"); len(*p) == 0 {
		if d, derr := UserCacheDir(); derr == nil {
			*p = d
		} else {
			*p = os.TempDir()
		}
	}
	err = nil
	return
})

// If SU, returns "/var/local" if local program; "/var/opt" if opt program;
// or "/var/lib" otherwise.
var StateHome = cache.New[string](func(p *string) error {
	if program.IsSuperUser.Value() {
		if program.IsUsrLocal.Value() {
			*p = "/var/local"
		} else if program.IsOpt.Value() {
			*p = "/var/opt"
		} else {
			*p = "/var/lib"
		}
	} else if *p = Getenv("XDG_STATE_HOME"); len(*p) == 0 {
		if h, err := UserHomeDir(); err == nil {
			*p = filepath.Join(h, ".local", "state")
		} else {
			*p = os.TempDir()
		}
	}
	return nil
})

var rootCacheHome = cache.New[string](func(p *string) (err error) {
	for _, d := range []string{"/var/cache", "/var/run", "/tmp"} {
		if _, err = os.Stat(d); err == nil {
			*p = d
			break
		}
	}
	return
})

var rootRunTimeDir = cache.New[string](func(p *string) error {
	const var_run = "/var/run"
	_, err := os.Stat(var_run)
	if err == nil {
		*p = var_run
	}
	return err
})
