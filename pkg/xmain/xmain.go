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

var (
	CacheDir  string
	CacheFlag = xflag.Label{"cache",
		"Directory of non-essential, ephemeral data.",
		func() any {
			var ok bool
			if CacheDir, ok = LookupEnv("CACHE"); ok {
			} else if os.Geteuid() == 0 {
				CacheDir = BestDir(fhs.Cache(),
					xdg.CacheHome())
			} else {
				CacheDir = BestDir(xdg.CacheHome(),
					fhs.Cache())
			}
			return &CacheDir
		}}
	CacheFile = func(s string) string { return File(CacheDir, s) }
)

var (
	ConfigDir  string
	ConfigFlag = xflag.Label{"config",
		"Directory of configuration files.",
		func() any {
			var ok bool
			if ConfigDir, ok = LookupEnv("CONFIG"); ok {
			} else if xprogram.IsKoApp() {
				ConfigDir = SubDir(fhs.Config())
			} else if os.Geteuid() == 0 {
				ConfigDir = BestDir(fhs.Config(),
					xdg.ConfigHome())
			} else {
				ConfigDir = BestDir(xdg.ConfigHome(),
					fhs.Config())
			}
			return &ConfigDir
		}}
	ConfigFile = func(s string) string { return File(ConfigDir, s) }
)

var (
	DataDir  string
	DataFlag = xflag.Label{"data",
		"Directory of essential, persistent data.",
		func() any {
			var ok bool
			DataDir, ok = os.LookupEnv("KO_DATA_PATH")
			if ok {
			} else if DataDir, ok = LookupEnv("DATA"); ok {
			} else if os.Geteuid() == 0 {
				DataDir = BestDir(fhs.Data(),
					xdg.DataHome())
			} else {
				DataDir = BestDir(xdg.DataHome(),
					fhs.Data())
			}
			return &DataDir
		}}
	DataFile = func(s string) string { return File(DataDir, s) }
)

var (
	RunDir  string
	RunFlag = xflag.Label{"run",
		"Directory of ephemeral files and sockets.",
		func() any {
			var ok bool
			if RunDir, ok = LookupEnv("RUN"); ok {
			} else if os.Geteuid() == 0 {
				RunDir = BestDir(fhs.RunTime(),
					xdg.RunTimeDir())
			} else {
				RunDir = BestDir(xdg.RunTimeDir(),
					fhs.RunTime())
			}
			return &RunDir
		}}
	RunFile = func(s string) string { return File(RunDir, s) }
)

var (
	StateDir  string
	StateFlag = xflag.Label{"state",
		"Directory that persists through restart.",
		func() any {
			var ok bool
			if StateDir, ok = LookupEnv("STATE"); ok {
			} else if xprogram.IsKoApp() {
				StateDir = SubDir(fhs.State())
			} else if os.Geteuid() == 0 {
				StateDir = BestDir(fhs.State(),
					xdg.StateHome())
			} else {
				StateDir = BestDir(xdg.StateHome(),
					fhs.State())
			}
			return &StateDir
		}}
	StateFile = func(s string) string { return File(StateDir, s) }
)

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

// If “name” doesn't equal "-" and doesn't have a [filepath.Separator],
// [filepath.Join] it to “dir”;
// otherwise, return unchanged.
func File(dir, name string) string {
	if name != "-" &&
		strings.IndexRune(name, filepath.Separator) < 0 &&
		len(dir) > 0 {
		name = filepath.Join(dir, name)
	}
	return name
}

// [os.LookupEnv] of keyword with [EnvPrefix] and given “suffix”.
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
	return "(unavailable)"
}

// [filepath.Join] “dir” with [PackageName]
func SubDir(dir string) string {
	return filepath.Join(dir, PackageName())
}
