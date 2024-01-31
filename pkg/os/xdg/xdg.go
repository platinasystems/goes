// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides [XDG] or, if Unix super user, [FSHS] defined paths.
//
//	XDG  https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
//	FSHS https://en.wikipedia.org/wiki/Filesystem_Hierarchy_Standard
package xdg

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// test overrides
var Getenv = os.Getenv
var UserCacheDir = os.UserCacheDir
var UserConfigDir = os.UserConfigDir
var UserHomeDir = os.UserHomeDir

// flag overrides
var Flag = struct {
	XdgCacheHome,
	XdgConfigDirs,
	XdgConfigHome,
	XdgDataDirs,
	XdgDataHome,
	XdgRunTimeDir,
	XdgStateHome *string
}{
	XdgCacheHome: flag.String("xdg-cache-home", "",
		"Default $XDG_CACHE_HOME or ~/.local/cache."),
	XdgConfigDirs: flag.String("xdg-config-dirs", "",
		"Default $XDG_CONFIG_DIRS or ~/.config:/etc/xdg."),
	XdgConfigHome: flag.String("xdg-config-home", "",
		"Default $XDG_CONFIG_HOME or ~/.config."),
	XdgDataDirs: flag.String("xdg-data-dirs", "",
		"Default $XDG_DATA_DIRS or "+
			"~/.local/share:/usr/local/share:/usr/share."),
	XdgDataHome: flag.String("xdg-data-home", "",
		"Default $XDG_DATA_HOME or ~/.local/share."),
	XdgRunTimeDir: flag.String("xdg-runtime-dir", "",
		"Default $XDG_RUNTIME_DIR or ~/.local/cache."),
	XdgStateHome: flag.String("xdg-state-home", "",
		"Default $XDG_STATE_HOME or ~/.local/state."),
}

var SlashOpt = filepath.FromSlash("/opt")

var IsOpt = sync.OnceValue(func() bool {
	return strings.HasPrefix(os.Args[0], SlashOpt)
})

var UsrLocal = filepath.FromSlash("/usr/local")

var IsUsrLocal = sync.OnceValue(func() bool {
	return strings.HasPrefix(os.Args[0], UsrLocal)
})

var IsSuperUser = sync.OnceValue(func() bool {
	return os.Geteuid() == 0
})

var SuperUser = struct {
	CacheHome, RunTimeDir func() string
}{
	CacheHome: sync.OnceValue(func() string {
		for _, s := range []string{
			filepath.FromSlash("/var/cache"),
			filepath.FromSlash("/var/run"),
		} {
			if _, err := os.Stat(s); err == nil {
				return s
			}
		}
		return os.TempDir()
	}),
	RunTimeDir: sync.OnceValue(func() string {
		var_run := filepath.FromSlash("/var/run")
		if _, err := os.Stat(var_run); err == nil {
			return var_run
		}
		return os.TempDir()
	}),
}

// If available, returns the -xdg-cache-home command line flag or
// $XDG_CACHE_HOME environment variable; or if super user, /var/cache or
// /var/run; otherwise, if not super user, ~/.local/cache or /tmp).
var CacheHome = sync.OnceValue(getCacheHome)

// If available, returns the -xdg-config-dirs command line flag or
// $XDG_CONFIG_DIRS environment variable; otherwise, ~/.config:/etc/xdg.
var ConfigDirs = sync.OnceValue(getConfigDirs)

// If available, returns the -xdg-config-home command line flag or
// $XDG_CONFIG_HOME environment variable; or if super user, /etc/opt if opt
// program; or /etc; otherwise, if not super user, returns ~/.config or
// /tmp.
var ConfigHome = sync.OnceValue(getConfigHome)

// If available, returns the -xdg-data-dirs command line flag or $XDG_DATA_DIRS
// environment variable; otherwise,
// ~/.local/share:/usr/local/share:/usr/share.
var DataDirs = sync.OnceValue(getDataDirs)

// If available, returns the -xdg-data-home command line flag or $XDG_DATA_HOME
// environment variable; or if super user, /usr/local/share if local program;
// /opt/share if opt program; or /usr/share; otherwise, if not super user,
// ~/.local/share or /tmp.
var DataHome = sync.OnceValue(getDataHome)

// If available, returns the -xdg-runtime-dir command line flag or
// $XDG_RUNTIME_DIR environment variable; or if superuser, "/var/run";
// otherwise, ~/.local/cache or /tmp.
var RunTimeDir = sync.OnceValue(getRunTimeDir)

// If available, returns the -xdg-state-home command line flag or
// $XDG_STATE_HOME environment variable; or is superuser, /var/local if local
// program; /var/opt if opt program; or /var/lib; otherwise. if not superuser,
// ~/.local/state.
var StateHome = sync.OnceValue(getStateHome)

var Dirs = map[string]any{
	"cache":   CacheHome,
	"config":  ConfigHome,
	"data":    DataHome,
	"runtime": RunTimeDir,
	"state":   StateHome,
}

func getCacheHome() string {
	if s := *Flag.XdgCacheHome; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		return SuperUser.CacheHome()
	} else if d, err := UserCacheDir(); err == nil {
		return d
	} else {
		return os.TempDir()
	}
	return ""
}

func getConfigDirs() string {
	sysconfigdirs := filepath.FromSlash("/etc/xdg")
	if s := *Flag.XdgConfigDirs; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_CONFIG_DIRS"); len(s) == 0 {
		return s
	} else if s, err := UserConfigDir(); err == nil {
		return fmt.Sprint(s, os.PathListSeparator, sysconfigdirs)
	}
	return sysconfigdirs
}

func getConfigHome() string {
	if s := *Flag.XdgConfigHome; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_CONFIG_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		if IsOpt() {
			return filepath.FromSlash("/etc/opt")
		} else {
			return filepath.FromSlash("/etc")
		}
	} else if s, err := UserConfigDir(); err == nil {
		return s
	}
	return os.TempDir()
}

func getDataDirs() string {
	sysdatadirs := fmt.Sprint(filepath.FromSlash("/usr/local/share"),
		os.PathListSeparator, filepath.FromSlash("/usr/share"))
	if s := *Flag.XdgDataDirs; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_DATA_DIRS"); len(s) > 0 {
		return s
	} else if s, err := os.UserHomeDir(); err == nil {
		return fmt.Sprint(filepath.Join(s, ".local", "share"),
			os.PathListSeparator, sysdatadirs)
	}
	return sysdatadirs
}

func getDataHome() string {
	if s := *Flag.XdgDataHome; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_DATA_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		if IsUsrLocal() {
			return "/usr/local/share"
		} else if IsOpt() {
			return "/opt/share"
		}
		return "/usr/share"
	} else if s, err := UserHomeDir(); err == nil {
		return filepath.Join(s, ".local", "share")
	}
	return os.TempDir()
}

func getRunTimeDir() string {
	if s := *Flag.XdgRunTimeDir; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_RUNTIME_DIR"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		return SuperUser.RunTimeDir()
	} else if s, err := UserCacheDir(); err == nil {
		return s
	}
	return os.TempDir()
}

func getStateHome() string {
	if s := *Flag.XdgStateHome; len(s) > 0 {
		return s
	} else if s = Getenv("XDG_STATE_HOME"); len(s) > 0 {
		return s
	} else if IsSuperUser() {
		if IsUsrLocal() {
			return filepath.FromSlash("/var/local")
		} else if IsOpt() {
			return filepath.FromSlash("/var/opt")
		}
		return filepath.FromSlash("/var/lib")
	} else if h, err := UserHomeDir(); err == nil {
		return filepath.Join(h, ".local", "state")
	}
	return os.TempDir()
}
