// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// See https://specifications.freedesktop.org/basedir-spec/basedir-spec-latest.html
package xdg

import (
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var (
	Etc    = filepath.FromSlash("/etc")
	EtcOpt = filepath.FromSlash("/etc/opt")

	KoApp = filepath.FromSlash("/ko-app")

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

type Dirs []string

func (dirs Dirs) Search(fn string) (os.FileInfo, error) {
	for _, s := range dirs {
		if len(fn) > 0 && fn != "." {
			s = filepath.Join(s, fn)
		}
		fi, err := os.Stat(s)
		if err == nil {
			return fi, nil
		}
	}
	return nil, os.ErrNotExist
}

func IsKoApp() bool {
	return strings.HasPrefix(xprogram.Path(), KoApp)
}

func IsOpt() bool {
	return strings.HasPrefix(xprogram.Path(), Opt)
}

func IsUsrLocal() bool {
	return strings.HasPrefix(xprogram.Path(), UsrLocal)
}

func IsSuperUser() bool {
	return os.Geteuid() == 0
}

// CacheHome has non-essential, ephemeral data.
var CacheHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		return s
	}
	if IsKoApp() || IsSuperUser() {
		for _, s := range []string{
			VarCache,
			VarRun,
		} {
			fi, err := os.Stat(s)
			if err == nil && fi.IsDir() {
				return s
			}
		}
	} else if s, err := os.UserCacheDir(); err == nil {
		return s
	}
	return os.TempDir()
})

var CacheDirs = sync.OnceValue(func() (dirs Dirs) {
	name := xprogram.MainName()
	if s := os.Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		dirs = append(dirs, filepath.Join(s, name))
		dirs = append(dirs, s)
	} else if IsKoApp() || IsSuperUser() {
		dirs = append(dirs, filepath.Join(VarCache, name))
		dirs = append(dirs, VarCache)
		dirs = append(dirs, filepath.Join(VarRun, name))
		dirs = append(dirs, VarRun)
	} else if s, err := os.UserCacheDir(); err == nil {
		dirs = append(dirs, filepath.Join(s, name))
		dirs = append(dirs, s)
	}
	tmp := os.TempDir()
	dirs = append(dirs, filepath.Join(tmp, name))
	dirs = append(dirs, tmp)
	return
})

// ConfigHome has persistent configuration.
var ConfigHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_CACHE_HOME"); len(s) > 0 {
		return s
	}
	if IsKoApp() || IsSuperUser() {
		if IsOpt() {
			return EtcOpt
		}
	} else if s, err := os.UserConfigDir(); err == nil {
		return s
	}
	return Etc
})

var ConfigDirs = sync.OnceValue(func() (dirs Dirs) {
	name := xprogram.MainName()
	if s := os.Getenv("XDG_CONFIG_DIRS"); len(s) == 0 {
		for _, ss := range strings.Split(s, PathListSeparatorString) {
			dirs = append(dirs, filepath.Join(ss, name))
			dirs = append(dirs, ss)
		}
	} else if IsKoApp() || IsSuperUser() {
		if IsOpt() {
			dirs = append(dirs, filepath.Join(EtcOpt, name))
			dirs = append(dirs, EtcOpt)
		} else {
			dirs = append(dirs, filepath.Join(Etc, name))
			dirs = append(dirs, Etc)
		}
	} else if s, err := os.UserConfigDir(); err == nil {
		dirs = append(dirs, filepath.Join(s, name))
		dirs = append(dirs, s)
	}
	return
})

// DataHome has essential, persistent data.
var DataHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_DATA_HOME"); len(s) > 0 {
		return s
	}
	if IsKoApp() || IsSuperUser() {
		if IsUsrLocal() {
			return UsrLocalShare
		} else if IsOpt() {
			return OptShare
		}
		return UsrShare
	} else if h, err := os.UserHomeDir(); err == nil {
		hls := filepath.Join(h, ".local", "share")
		if _, err = os.Stat(hls); err == nil {
			return hls
		}
		return h
	}
	return os.TempDir()
})

var DataDirs = sync.OnceValue(func() (dirs Dirs) {
	name := xprogram.MainName()
	if s := os.Getenv("XDG_DATA_DIRS"); len(s) > 0 {
		for _, ss := range strings.Split(s, PathListSeparatorString) {
			dirs = append(dirs, filepath.Join(ss, name))
			dirs = append(dirs, ss)
		}
	} else if IsKoApp() || IsSuperUser() {
		if IsUsrLocal() {
			dirs = append(dirs, filepath.Join(UsrLocalShare, name))
			dirs = append(dirs, UsrLocalShare)
		} else if IsOpt() {
			dirs = append(dirs, filepath.Join(OptShare, name))
			dirs = append(dirs, OptShare)
		} else {
			dirs = append(dirs, filepath.Join(UsrShare, name))
			dirs = append(dirs, UsrShare)
		}
	} else if h, err := os.UserHomeDir(); err == nil {
		hls := filepath.Join(h, ".local", "share")
		dirs = append(dirs, filepath.Join(hls, name))
		dirs = append(dirs, hls)
	}
	tmp := os.TempDir()
	dirs = append(dirs, filepath.Join(tmp, name))
	dirs = append(dirs, tmp)
	return
})

// RunTimeDir has non-essential, ephemeral runtime files and other objects,
// such as sockets and named pipes.
var RunTimeDir = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_RUNTIME_DIR"); len(s) > 0 {
		return s
	}
	if IsKoApp() || IsSuperUser() {
		for _, s := range []string{
			VarRun,
		} {
			fi, err := os.Stat(s)
			if err == nil && fi.IsDir() {
				return s
			}
		}
	} else if h, err := os.UserHomeDir(); err == nil {
		for _, s := range []string{
			filepath.Join(h, ".local", "run"),
			filepath.Join(h, ".local"),
			h,
		} {
			fi, err := os.Stat(s)
			if err == nil && fi.IsDir() {
				return s
			}
		}
	}
	return os.TempDir()
})

var RunTimeDirs = sync.OnceValue(func() (dirs Dirs) {
	name := xprogram.MainName()
	if s := os.Getenv("XDG_RUNTIME_DIR"); len(s) > 0 {
		dirs = append(dirs, filepath.Join(s, name))
		dirs = append(dirs, s)
	} else if IsKoApp() || IsSuperUser() {
		dirs = append(dirs, filepath.Join(VarRun, name))
		dirs = append(dirs, VarRun)
	} else if h, err := os.UserHomeDir(); err == nil {
		dirs = append(dirs, filepath.Join(h, ".local", "run", name))
		dirs = append(dirs, filepath.Join(h, ".local", name))
		dirs = append(dirs, filepath.Join(h, ".local", "run"))
		dirs = append(dirs, filepath.Join(h, ".local"))
	}
	tmp := os.TempDir()
	dirs = append(dirs, filepath.Join(tmp, name))
	dirs = append(dirs, tmp)
	return
})

// StateHome persists between application restarts,
//
//   - actions history (logs, history, recently used files, …)
//
//   - current state of the application that can be reused on a restart
//     (view, layout, open files, undo history, …)
var StateHome = sync.OnceValue(func() string {
	if s := os.Getenv("XDG_STATE_HOME"); len(s) > 0 {
		return s
	}
	if IsKoApp() {
		return VarLib
	}
	if IsSuperUser() {
		if IsUsrLocal() {
			return VarLocal
		} else if IsOpt() {
			return VarOpt
		}
		return VarLib
	} else if h, err := os.UserHomeDir(); err == nil {
		return filepath.Join(h, ".local", "state")
	}
	return os.TempDir()
})

var StateDirs = sync.OnceValue(func() (dirs Dirs) {
	name := xprogram.MainName()
	if s := os.Getenv("XDG_STATE_HOME"); len(s) > 0 {
		dirs = append(dirs, filepath.Join(s, name))
		dirs = append(dirs, s)
	} else if IsKoApp() {
		dirs = append(dirs, filepath.Join(VarLib, name))
		dirs = append(dirs, VarLib)
	} else if IsSuperUser() {
		if IsUsrLocal() {
			dirs = append(dirs, filepath.Join(VarLocal, name))
			dirs = append(dirs, VarLocal)
		} else if IsOpt() {
			dirs = append(dirs, filepath.Join(VarOpt, name))
			dirs = append(dirs, VarOpt)
		} else {
			dirs = append(dirs, filepath.Join(VarLib, name))
			dirs = append(dirs, VarLib)
		}
	} else if h, err := os.UserHomeDir(); err == nil {
		hls := filepath.Join(h, ".local", "state")
		dirs = append(dirs, filepath.Join(hls, name))
		dirs = append(dirs, hls)
	}
	tmp := os.TempDir()
	dirs = append(dirs, filepath.Join(tmp, name))
	dirs = append(dirs, tmp)
	return
})
