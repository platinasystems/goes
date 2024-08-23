// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// See https://specifications.freedesktop.org/basedir-spec/latest/
package xdg

import (
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const PathListSeparatorString = string(os.PathListSeparator)

// CacheHome has non-essential, ephemeral data.
func CacheHome() string {
	s := os.Getenv("XDG_CACHE_HOME")
	if len(s) > 0 {
		return s
	}
	switch runtime.GOOS {
	case "darwin", "ios":
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, "Library/Caches")
		}
	case "plan9":
		if s = os.Getenv("home"); len(s) > 0 {
			s = filepath.Join(s, "lib", "cache")
		}
	case "windows":
		if s = WindowsAppDataLocal(); len(s) > 0 {
			s = filepath.Join(s, "cache")
		}
	default:
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, ".cache")
		}
	}
	return s
}

func ConfigDirs() (dirs []string) {
	s := os.Getenv("XDG_CONFIG_DIRS")
	if len(s) > 0 {
		dirs = strings.Split(s, PathListSeparatorString)
	} else {
		switch runtime.GOOS {
		case "darwin", "ios":
			dirs = append(dirs,
				"/Library/Application Support",
				"/Library/Preferences",
			)
		case "plan9":
			dirs = append(dirs, "/lib")
		case "windows":
			var dirs []string
			if s = WindowsProgramData(); len(s) > 0 {
				dirs = append(dirs, s)
			}
			if s = WindowsAppDataRoaming(); len(s) > 0 {
				dirs = append(dirs, s)
			}
			if len(dirs) > 0 {
				return dirs
			}
		default:
			if xprogram.IsOpt() {
				dirs = append(dirs, "/etc/opt")
			} else {
				dirs = append(dirs, "/etc")
			}
		}
	}
	return
}

// ConfigHome has persistent configuration.
func ConfigHome() string {
	s := os.Getenv("XDG_CONFIG_HOME")
	if len(s) > 0 {
		return s
	}
	switch runtime.GOOS {
	case "darwin", "ios":
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, "Library/Application Support")
		}
	case "plan9":
		if s = os.Getenv("home"); len(s) > 0 {
			s = filepath.Join(s, "lib")
		}
	case "windows":
		s = WindowsAppDataLocal()
	default:
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, ".config")
		}
	}
	return s
}

func DataDirs() (dirs []string) {
	s := os.Getenv("XDG_DATA_DIRS")
	if len(s) > 0 {
		dirs = strings.Split(s, PathListSeparatorString)
	} else {
		switch runtime.GOOS {
		case "darwin", "ios":
			dirs = append(dirs, "/Library/Application Support")
		case "plan9":
			dirs = append(dirs, "/lib")
		case "windows":
			if s := WindowsAppDataRoaming(); len(s) > 0 {
				dirs = append(dirs, s)
			}
			if s := WindowsProgramData(); len(s) > 0 {
				dirs = append(dirs, s)
			}
		default:
			if xprogram.IsUsrLocal() {
				dirs = append(dirs, "/usr/local")
			} else if xprogram.IsOpt() {
				dirs = append(dirs, "/opt/share")
			} else {
				dirs = append(dirs, "/usr/share")
			}
		}
	}
	return
}

// DataHome has essential, persistent data.
func DataHome() string {
	s := os.Getenv("XDG_DATA_HOME")
	if len(s) > 0 {
		return s
	}
	switch runtime.GOOS {
	case "darwin", "ios":
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, "Library/Application Support")
		}
	case "plan9":
		if s = os.Getenv("home"); len(s) > 0 {
			s = filepath.Join(s, "lib")
		}
	case "windows":
		s = WindowsAppDataLocal()
	default:
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, ".local/share")
		}
	}
	return s
}

// RunTimeDir has non-essential, ephemeral runtime files and other objects,
// such as sockets and named pipes.
func RunTimeDir() string {
	s := os.Getenv("XDG_RUNTIME_DIR")
	if len(s) > 0 {
		return s
	}
	switch runtime.GOOS {
	case "darwin", "ios":
		if s = SudoUserHome(); len(s) > 0 {
			return filepath.Join(s, "Library/Application Support")
		}
	case "plan9":
	case "windows":
		s = WindowsAppDataLocal()
	default:
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, ".local/run")
		}
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
	if len(s) > 0 {
		return s
	}
	switch runtime.GOOS {
	case "darwin", "ios":
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, "Library/Application Support")
		}
	case "plan9":
		if s = os.Getenv("home"); len(s) > 0 {
			s = filepath.Join(s, "lib", "state")
		}
	case "windows":
		s = WindowsAppDataLocal()
	default:
		if s = SudoUserHome(); len(s) > 0 {
			s = filepath.Join(s, ".local/state")
		}
	}
	return s
}

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

func WindowsAppDataLocal() string {
	s := os.Getenv("LocalAppData")
	if len(s) > 0 {
		return s
	}
	if u, err := user.Current(); err == nil {
		s = filepath.Join(u.HomeDir, "AppData", "Local")
	}
	return s
}

func WindowsAppDataRoaming() string {
	s := os.Getenv("AppData")
	if len(s) > 0 {
		return s
	}
	if u, err := user.Current(); err == nil {
		s = filepath.Join(u.HomeDir, "AppData", "Roaming")
	}
	return s
}

func WindowsProgramData() string {
	s := os.Getenv("ProgramData")
	if len(s) > 0 {
		return s
	}
	if s = os.Getenv("ALLUSERSPROFILE"); len(s) > 0 {
		return s
	}
	if s = os.Getenv("SystemDrive"); len(s) > 0 {
		s = filepath.Join(s, "ProgramData")
	}
	return s
}
