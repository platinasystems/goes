// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
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

var Base = cache.NewReadOnly[string](
	func(p *string) error {
		*p = filepath.Base(Executable())
		return nil
	}).Value

var Executable = cache.NewReadOnly[string](
	func(p *string) error {
		if s, err := os.Executable(); err == nil {
			*p = s
		} else {
			*p = os.Args[0]
		}
		return nil
	}).Value

var Opt = cache.NewReadOnly[bool](
	func(p *bool) (err error) {
		*p = strings.HasPrefix(Executable(),
			filepath.FromSlash("/opt"))
		return
	})

var SuperUser = cache.NewReadOnly[bool](
	func(p *bool) error {
		*p = os.Geteuid() == 0
		return nil
	})

var UsrLocal = cache.NewReadOnly[bool](
	func(p *bool) (err error) {
		*p = strings.HasPrefix(Executable(),
			filepath.FromSlash("/usr/local"))
		return
	})

var Is = struct {
	Opt, SuperUser, UsrLocal func() bool
}{
	Opt:       Opt.Value,
	SuperUser: SuperUser.Value,
	UsrLocal:  UsrLocal.Value,
}

var BuildId = cache.NewReadOnly[string](
	func(p *string) (err error) {
		*p, err = buildid.ReadFile(Executable())
		return
	})

var BuildInfo = cache.NewReadOnly[*debug.BuildInfo](
	func(p **debug.BuildInfo) error {
		if info, ok := debug.ReadBuildInfo(); ok {
			*p = info
			return nil
		}
		return ErrUnavailable
	})

var Build = map[string]any{
	"id":   BuildId,
	"info": BuildInfo,
}

// The main module reference in the form of PATH@SYMVER or empty if the
// main module is unavailable, as with GO tests.
var MainReference = cache.NewReadOnly[string](
	func(p *string) (err error) {
		bi, err := BuildInfo.ValErr()
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
	})

// The main module version.
var MainVersion = cache.NewReadOnly[string](
	func(p *string) (err error) {
		bi, err := BuildInfo.ValErr()
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
	})

var Main = map[string]any{
	"reference": MainReference,
	"version":   MainVersion,
}
