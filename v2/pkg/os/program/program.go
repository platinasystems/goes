// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This package provides the program full name, base name, and main module
// reference.
package program

import (
	"errors"
	"fmt"
	"golang/buildid"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var ErrUnavailable = errors.New("unavailable")

var Base = cache.New[string](func(p *string) error {
	*p = filepath.Base(Executable())
	return nil
}).Value

var Executable = cache.New[string](func(p *string) error {
	if s, err := os.Executable(); err == nil {
		*p = s
	} else {
		*p = os.Args[0]
	}
	return nil
}).Value

var Cache = struct{ Opt, UsrLocal, SuperUser *cache.Cache[bool] }{
	Opt: cache.New[bool](func(p *bool) (err error) {
		*p = strings.HasPrefix(Executable(), filepath.FromSlash("/opt"))
		return
	}),
	UsrLocal: cache.New[bool](func(p *bool) (err error) {
		*p = strings.HasPrefix(Executable(),
			filepath.FromSlash("/usr/local"))
		return
	}),
	SuperUser: cache.New[bool](func(p *bool) error {
		*p = os.Geteuid() == 0
		return nil
	}),
}

var Is = struct {
	Opt, UsrLocal, SuperUser func() bool
}{
	Opt:       Cache.Opt.Value,
	UsrLocal:  Cache.UsrLocal.Value,
	SuperUser: Cache.SuperUser.Value,
}

var Build = struct {
	Id   *cache.Cache[string]
	Info *cache.Cache[*debug.BuildInfo]
}{
	Id: cache.New[string](func(p *string) (err error) {
		*p, err = buildid.ReadFile(Executable())
		return
	}),
	Info: cache.New[*debug.BuildInfo](func(p **debug.BuildInfo) error {
		if info, ok := debug.ReadBuildInfo(); ok {
			*p = info
			return nil
		}
		return ErrUnavailable
	}),
}

var Main = struct {
	// The main module reference in the form of PATH@SYMVER or empty if the
	// main module is unavailable, as with GO tests.
	Reference *cache.Cache[string]
	// Returns the main module version.
	Version *cache.Cache[string]
}{
	Reference: cache.New[string](func(p *string) (err error) {
		bi, err := Build.Info.ValErr()
		if err == nil {
			m := &bi.Main
			if m.Replace != nil {
				m = m.Replace
			}
			if len(m.Path) > 0 && len(m.Version) > 0 {
				*p = fmt.Sprint(m.Path, "@", m.Version)
			} else {
				err = ErrUnavailable
			}
		}
		return
	}),
	Version: cache.New[string](func(p *string) (err error) {
		bi, err := Build.Info.ValErr()
		if err == nil {
			m := &bi.Main
			if m.Replace != nil {
				m = m.Replace
			}
			if len(m.Version) > 0 {
				*p = m.Version
			} else {
				err = ErrUnavailable
			}
		}
		return
	}),
}
