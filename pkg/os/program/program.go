// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides the program full name, base name, and main module
// reference.
package program

import (
	"fmt"
	"golang/buildid"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
)

var Base = sync.OnceValue(func() string {
	return filepath.Base(Executable())
})

var Executable = sync.OnceValue(func() string {
	if s, err := os.Executable(); err == nil {
		return s
	}
	return os.Args[0]
})

var BuildId = sync.OnceValue(func() string {
	s, err := buildid.ReadFile(Executable())
	if err != nil {
		s = err.Error()
	}
	return s
})

var BuildInfo = sync.OnceValue(func() (bi *debug.BuildInfo) {
	bi, _ = debug.ReadBuildInfo()
	return
})

var Build = map[string]any{
	"id": BuildId,
	"info": func() fmt.Stringer {
		return BuildInfo()
	},
}

var MainModule = sync.OnceValue(func() *debug.Module {
	if bi := BuildInfo(); bi != nil {
		if m := &bi.Main; m.Replace != nil {
			return m.Replace
		} else {
			return m
		}
	}
	return nil
})

// The main module reference in the form of PATH@SYMVER or empty if the
// main module is unavailable, as with GO tests.
var MainReference = sync.OnceValue(func() string {
	m := MainModule()
	if m != nil && len(m.Path) > 0 && len(m.Version) > 0 {
		return fmt.Sprint(m.Path, "@", m.Version)
	}
	return ""
})

// The main module version if available; empty otherwise.
var MainVersion = sync.OnceValue(func() string {
	if m := MainModule(); m != nil && len(m.Version) > 0 {
		return m.Version
	}
	return ""
})

var Main = map[string]any{
	"reference": MainReference,
	"version":   MainVersion,
}

var (
	Etc      = filepath.FromSlash("/etc")
	EtcOpt   = filepath.FromSlash("/etc/opt")
	KoApp    = filepath.FromSlash("/ko-app")
	Opt      = filepath.FromSlash("/opt")
	UsrLocal = filepath.FromSlash("/usr/local")
	VarCache = filepath.FromSlash("/var/cache")
	VarLocal = filepath.FromSlash("/var/local")
	VarLib   = filepath.FromSlash("/var/lib")
	VarOpt   = filepath.FromSlash("/var/opt")
	VarRun   = filepath.FromSlash("/var/run")
)

var (
	IsKoApp = sync.OnceValue(func() bool {
		return strings.HasPrefix(Executable(), KoApp)
	})
	IsOpt = sync.OnceValue(func() bool {
		return strings.HasPrefix(Executable(), Opt)
	})
	IsUsrLocal = sync.OnceValue(func() bool {
		return strings.HasPrefix(Executable(), UsrLocal)
	})
)

var IsSuperUser = sync.OnceValue(func() bool {
	return os.Geteuid() == 0
})

// $XDG_CACHE_HOME, os.UserCacheDir()/Base(), /var/cache/Base(),
// /var/run/Base() or os.TempDir()/Base()
var CacheHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		return s
	}
	if !IsKoApp() && !IsSuperUser() {
		if s, err := os.UserCacheDir(); err == nil {
			return filepath.Join(s, Base())
		}
	}
	for _, s := range []string{VarCache, VarRun} {
		if _, err := os.Stat(s); err == nil {
			return filepath.Join(s, Base())
		}
	}
	return filepath.Join(os.TempDir(), Base())
})

// $XDG_CONFIG_DIRS, os.UserConfigDir()/Base():/etc/Base(), or /etc/Base()
var ConfigDirs = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_CONFIG_DIRS"); len(s) == 0 {
		return s
	}
	var etc string
	if IsOpt() {
		etc = filepath.Join(EtcOpt, Base())
	} else {
		etc = filepath.Join(EtcOpt, Base())
	}
	if s, err := os.UserConfigDir(); err == nil {
		cfg := filepath.Join(s, Base())
		return fmt.Sprint(cfg, os.PathListSeparator, etc)
	}
	return etc
})

// $XDG_CONFIG_HOME, os.UserConfigDir()/Base(), /etc/opt/Base(), or
// /etc/Base()
var ConfigHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_CONFIG_HOME"); len(s) > 0 {
		return s
	}
	if !IsKoApp() && !IsSuperUser() {
		if s, err := os.UserConfigDir(); err == nil {
			return filepath.Join(s, Base())
		}
	}
	if IsOpt() {
		return filepath.Join(EtcOpt, Base())
	}
	return filepath.Join(Etc, Base())
})

// $XDG_DATA_DIRS,
// os.UserHomeDir()/.local/share/Base():/usr/local/share/Base():/usr/share/Base(),
// or /usr/local/share/Base():/usr/share/Base()
var DataDirs = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_DATA_DIRS"); len(s) > 0 {
		return s
	}
	usrLocalShare := filepath.Join("/usr/local/share", Base())
	usrShare := filepath.Join("/usr/share", Base())
	sysShare := fmt.Sprint(usrLocalShare, os.PathListSeparator, usrShare)
	if s, err := os.UserHomeDir(); err == nil {
		homeShare := filepath.Join(s, ".local", "share", Base())
		return fmt.Sprint(homeShare, os.PathListSeparator, sysShare)
	}
	return sysShare
})

// essential, persistent data: $XDG_DATA_HOME, os.UserHomeDir()/Base(),
// /usr/local/share; /opt/share, /usr/share, or /tmp.
var DataHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_DATA_HOME"); len(s) > 0 {
		return s
	}
	if IsKoApp() || IsSuperUser() {
		if IsUsrLocal() {
			return filepath.Join("/usr/local/share", Base())
		}
		if IsOpt() {
			return filepath.Join("/opt/share", Base())
		}
		return filepath.Join("/usr/share", Base())
	}
	if s, err := os.UserHomeDir(); err == nil {
		return filepath.Join(s, ".local", "share", Base())
	}
	return os.TempDir()
})

// non-essential, ephemeral runtime files and other file objects
// (such as sockets, named pipes, ...)
// $XDG_RUNTIME_DIR, os.UserHomeeDir()/.local/run/Base(), /var/run, or /tmp.
var RunTimeDir = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_RUNTIME_DIR"); len(s) > 0 {
		return s
	}
	if !IsKoApp() && !IsSuperUser() {
		if s, err := os.UserHomeDir(); err == nil {
			return filepath.Join(s, ".local", "run", Base())
		}
	}
	if fi, err := os.Stat(VarRun); err == nil && fi.IsDir() {
		return filepath.Join(VarRun, Base())
	}
	return os.TempDir()
})

// os.UserHomeDir()/.Base(), /etc/opt/.Base(), /etc/.Base()
var SecretHome = sync.OnceValue(func() string {
	dotbase := fmt.Sprint(".", Base())
	if !IsKoApp() && !IsSuperUser() {
		if s, err := os.UserHomeDir(); err == nil {
			return filepath.Join(s, dotbase)
		}
	}
	if IsOpt() {
		return filepath.Join(EtcOpt, dotbase)
	}
	return filepath.Join(Etc, dotbase)
})

// persists between application restarts,
//
//   - actions history (logs, history, recently used files, …)
//
//   - current state of the application that can be reused on a restart
//     (view, layout, open files, undo history, …)
//
// $XDG_STATE_HOME, os.UserHomeDir()/.local/state/Base(),
// /var/local/Base(), /var/opt/Base(), /var/lib/Base() or os.TempDir()
var StateHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_STATE_HOME"); len(s) > 0 {
		return s
	}
	if !IsKoApp() && !IsSuperUser() {
		if s, err := os.UserHomeDir(); err == nil {
			return filepath.Join(s, ".local", "state", Base())
		}
	}
	if IsUsrLocal() {
		return filepath.Join(VarLocal, Base())
	}
	if IsOpt() {
		return filepath.Join(VarOpt, Base())
	}
	if fi, err := os.Stat(VarLib); err == nil && fi.IsDir() {
		return filepath.Join(VarLib, Base())
	}
	return os.TempDir()
})
