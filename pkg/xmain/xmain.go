// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xmain

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/fhs"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

var Features = map[string]any{
	"show": map[string]any{
		"main": map[string]any{
			"name":      PackageName,
			"reference": Reference,
			"version":   Version,
		},
	},
}

var Cache = xflag.NewDir("cache", `
Directory of non-essential, ephemeral data.
`[1:], func() string {
	s, ok := LookupEnv("CACHE")
	if !ok {
		if os.Geteuid() == 0 {
			s = BestDir(fhs.Cache(), xdg.CacheHome())
		} else {
			s = BestDir(xdg.CacheHome(), fhs.Cache())
		}
	}
	return s
})

var Config = xflag.NewDir("config", `
Directory of configuration files.
`[1:], func() string {
	s, ok := LookupEnv("CONFIG")
	if !ok {
		if xprogram.IsKoApp() {
			s = SubDir(fhs.Config())
		} else if os.Geteuid() == 0 {
			s = BestDir(fhs.Config(), xdg.ConfigHome())
		} else {
			s = BestDir(xdg.ConfigHome(), fhs.Config())
		}
	}
	return s
})

var Data = xflag.NewDir("data", `
Directory of essential, persistent data.
`[1:], func() string {
	s, ok := os.LookupEnv("KO_DATA_PATH")
	if !ok {
		if s, ok = LookupEnv("DATA"); !ok {
			if os.Geteuid() == 0 {
				s = BestDir(fhs.Data(), xdg.DataHome())
			} else {
				s = BestDir(xdg.DataHome(), fhs.Data())
			}
		}
	}
	return s
})

var Run = xflag.NewDir("run", `
Directory of non-essential, ephemeral files and sockets.
`[1:], func() string {
	s, ok := LookupEnv("RUN")
	if !ok {
		if os.Geteuid() == 0 {
			s = BestDir(fhs.RunTime(), xdg.RunTimeDir())
		} else {
			s = BestDir(xdg.RunTimeDir(), fhs.RunTime())
		}
	}
	return s
})

var State = xflag.NewDir("state", `
Directory of files that persist between restarts.
`[1:], func() string {
	s, ok := LookupEnv("STATE")
	if !ok {
		if xprogram.IsKoApp() {
			s = SubDir(fhs.State())
		} else if os.Geteuid() == 0 {
			s = BestDir(fhs.State(), xdg.StateHome())
		} else {
			s = BestDir(xdg.StateHome(), fhs.State())
		}
	}
	return s
})

// BestDir returns the primary [SubDir] if it exists,
// or the first existing alternate;
// otherwise, return the primary if none exist.
func BestDir(primary string, alternates ...string) string {
	primaryMain := SubDir(primary)
	if fi, err := os.Stat(primaryMain); err == nil && fi.IsDir() {
		return primaryMain
	}
	for _, alt := range alternates {
		altMain := SubDir(alt)
		if fi, err := os.Stat(altMain); err == nil && fi.IsDir() {
			return altMain
		}
	}
	return primaryMain
}

// EnvPrefix returns the [PackageName] as an underscored, uppercase,
// environment keyword prefix. (e.g. GOES_, or GOES_VPN_)
var EnvPrefix = sync.OnceValue(func() string {
	return ToUnderscoredUpper(PackageName()) + "_"
})

// [os.LookupEnv] of keyword with [EnvPreifx] and given “suffix”.
func LookupEnv(suffix string) (string, bool) {
	suffix = ToUnderscoredUpper(suffix)
	return os.LookupEnv(fmt.Sprint(EnvPrefix(), suffix))
}

// Main module wi/in [xprogram.BuildInfo]
func Module() *debug.Module {
	bi := xprogram.BuildInfo()
	if bi == nil {
		return nil
	}
	m := &bi.Main
	if m.Replace != nil {
		m = m.Replace
	}
	return m
}

// Main package name w/in [xprogram.BuidInfo]
var PackageName = sync.OnceValue(func() string {
	const major = "^v[1-9][0-9]*$"
	bi := xprogram.BuildInfo()
	if bi == nil {
		return filepath.Base(xprogram.Path())
	}
	name := filepath.Base(bi.Path)
	if t, err := regexp.MatchString(major, name); err != nil {
		name = err.Error()
	} else if t {
		name = filepath.Base(filepath.Dir(bi.Path))
	}
	return name
})

// Main [Module] Path + "@" + Version.
func Reference() string {
	if mm := Module(); mm != nil {
		return mm.Path + "@" + mm.Version
	}
	return ""
}

func ToUnderscoredUpper(s string) string {
	return strings.Replace(strings.ToUpper(s), "-", "_", -1)
}

// Main [Module] Version.
func Version() string {
	if mm := Module(); mm != nil {
		return mm.Version
	}
	return ""
}

// [filepath.Join] “dir” with [PackageName]
func SubDir(dir string) string {
	return filepath.Join(dir, PackageName())
}
