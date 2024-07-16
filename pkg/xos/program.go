// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xos

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

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

var Arg0Base = sync.OnceValue(func() string {
	return filepath.Base(os.Args[0])
})

// Precedense:
//   - [os.Executable]
//   - [path/filepath.EvalSymlinks]([os.Args][0])
//   - os.Args[0]
func Program() string {
	s, err := os.Executable()
	if err == nil {
		return s
	}
	s, err = filepath.EvalSymlinks(os.Args[0])
	if err == nil {
		return s
	}
	return os.Args[0]
}

func ProgramIsKoApp() bool {
	return strings.HasPrefix(Program(), KoApp)
}

func ProgramIsOpt() bool {
	return strings.HasPrefix(Program(), Opt)
}

func ProgramIsUsrLocal() bool {
	return strings.HasPrefix(Program(), UsrLocal)
}

func UserIsRoot() bool {
	return os.Geteuid() == 0
}

// CacheHome has user-specific non-essential data files
//
// Precedence:
//   - $XDG_CACHE_HOME
//   - if not koapp and not super user, ~/.cache/Arg0Base
//   - /var/cache/Arg0Base
//   - /var/run/Arg0Base
//   - /tmp/Arg0Base
var CacheHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		return s
	}
	if !ProgramIsKoApp() && !UserIsRoot() {
		if s, err := os.UserCacheDir(); err == nil {
			return filepath.Join(s, Arg0Base())
		}
	}
	for _, s := range []string{VarCache, VarRun} {
		if _, err := os.Stat(s); err == nil {
			return filepath.Join(s, Arg0Base())
		}
	}
	return filepath.Join(os.TempDir(), Arg0Base())
})

// Precedence:
//   - $XDG_CONFIG_DIRS
//   - if /opt program, ~/.config/Arg0Base:/etc/opt/Arg0Base:/etc/opt
//   - ~/.config/Arg0Base:/etc/Arg0Base:/etc
func ConfigDirs() string {
	if s := os.Getenv("XDG_CONFIG_DIRS"); len(s) == 0 {
		return s
	}
	etc := Etc
	if ProgramIsOpt() {
		etc = filepath.Join(etc, Opt)
	}
	etc = fmt.Sprint(filepath.Join(etc, Arg0Base()),
		os.PathListSeparator, etc)
	if s, err := os.UserConfigDir(); err == nil {
		etc = fmt.Sprint(filepath.Join(s, Arg0Base()),
			os.PathListSeparator, s,
			os.PathListSeparator, etc)
	}
	return etc
}

// ConfigHome has user-specific configuration files
//
// Precedence:
//   - $XDG_CONFIG_HOME
//   - if not koapp and not super user, ~/Arg0Base
//   - if /opt program, /etc/opt/Arg0Base
//   - /etc/Arg0Base
func ConfigHome() string {
	if s := os.Getenv("XDG_CONFIG_HOME"); len(s) > 0 {
		return s
	}
	if !ProgramIsKoApp() && !UserIsRoot() {
		if s, err := os.UserConfigDir(); err == nil {
			return filepath.Join(s, Arg0Base())
		}
	}
	if ProgramIsOpt() {
		return filepath.Join(EtcOpt, Arg0Base())
	}
	return filepath.Join(Etc, Arg0Base())
}

// Precedence:
//   - $XDG_DATA_DIRS,
//   - ~/.local/share/Arg0Base:/usr/local/share/Arg0Base:/usr/share/Arg0Base
var DataDirs = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_DATA_DIRS"); len(s) > 0 {
		return s
	}
	usrLocalShare := filepath.Join("/usr/local/share", Arg0Base())
	usrShare := filepath.Join("/usr/share", Arg0Base())
	sysShare := fmt.Sprint(usrLocalShare, os.PathListSeparator, usrShare)
	if s, err := os.UserHomeDir(); err == nil {
		homeShare := filepath.Join(s, ".local", "share", Arg0Base())
		return fmt.Sprint(homeShare, os.PathListSeparator, sysShare)
	}
	return sysShare
})

// DataHome has essential, persistent data:
//
// Precedence:
//   - $XDG_DATA_HOME
//   - if koapp or super user
//   - if /usr/local program, /usr/local/share
//   - if /opt ptogram, /opt/share
//   - /usr/share
//   - ~/.local/share/Arg0Base
//   - /tmp
var DataHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_DATA_HOME"); len(s) > 0 {
		return s
	}
	if ProgramIsKoApp() || UserIsRoot() {
		if ProgramIsUsrLocal() {
			return filepath.Join("/usr/local/share", Arg0Base())
		}
		if ProgramIsOpt() {
			return filepath.Join("/opt/share", Arg0Base())
		}
		return filepath.Join("/usr/share", Arg0Base())
	}
	if s, err := os.UserHomeDir(); err == nil {
		return filepath.Join(s, ".local", "share", Arg0Base())
	}
	return os.TempDir()
})

// RunTimeDir is non-essential, ephemeral runtime files and other file objects
// (such as sockets, named pipes, ...)
//
//   - $XDG_RUNTIME_DIR
//   - if not koapp and not super user, ~/.local/run/Arg0Base
//   - if /var/run available, /var/run/Arg0Base
//   - /tmp
var RunTimeDir = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_RUNTIME_DIR"); len(s) > 0 {
		return s
	}
	if !ProgramIsKoApp() && !UserIsRoot() {
		if s, err := os.UserHomeDir(); err == nil {
			return filepath.Join(s, ".local", "run", Arg0Base())
		}
	}
	if fi, err := os.Stat(VarRun); err == nil && fi.IsDir() {
		return filepath.Join(VarRun, Arg0Base())
	}
	return os.TempDir()
})

// Precedence:
//
//   - if not koapp and not super user, ~/.Arg0Base
//   - if /opt program, /etc/opt/.Arg0Base
//   - /etc/.Arg0Base
var SecretHome = sync.OnceValue(func() string {
	dotbase := fmt.Sprint(".", Arg0Base())
	if !ProgramIsKoApp() && !UserIsRoot() {
		if s, err := os.UserHomeDir(); err == nil {
			return filepath.Join(s, dotbase)
		}
	}
	etc := Etc
	if ProgramIsOpt() {
		etc += Opt
	}
	return filepath.Join(etc, dotbase)
})

// StateHome persists between application restarts,
//
//   - actions history (logs, history, recently used files, …)
//
//   - current state of the application that can be reused on a restart
//     (view, layout, open files, undo history, …)
//
// Precedence:
//
//   - $XDG_STATE_HOME
//   - if not koapp and not super user, ~/.local/state/Arg0Base
//   - if /usr/local program, /var/local/Arg0Base
//   - if /opt program, /opt/var/opt/Arg0Base
//   - if /var/lib available, /var/lib/Arg0Base
//   - /tmp
var StateHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_STATE_HOME"); len(s) > 0 {
		return s
	}
	if !ProgramIsKoApp() && !UserIsRoot() {
		if s, err := os.UserHomeDir(); err == nil {
			return filepath.Join(s, ".local", "state", Arg0Base())
		}
	}
	if ProgramIsUsrLocal() {
		return filepath.Join(VarLocal, Arg0Base())
	}
	if ProgramIsOpt() {
		return filepath.Join(VarOpt, Arg0Base())
	}
	if fi, err := os.Stat(VarLib); err == nil && fi.IsDir() {
		return filepath.Join(VarLib, Arg0Base())
	}
	return os.TempDir()
})
