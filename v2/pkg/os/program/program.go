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

var BuildId = cache.New[string](func(p *string) (err error) {
	*p, err = buildid.ReadFile(Executable())
	return
})

var BuildInfo = cache.New[*debug.BuildInfo](func(p **debug.BuildInfo) error {
	if info, ok := debug.ReadBuildInfo(); ok {
		*p = info
		return nil
	}
	return ErrUnavailable
})

var Executable = cache.New[string](func(p *string) error {
	if s, err := os.Executable(); err == nil {
		*p = s
	} else {
		*p = os.Args[0]
	}
	return nil
}).Value

var IsOpt = cache.New[bool](func(p *bool) (err error) {
	*p = strings.HasPrefix(Executable(), filepath.FromSlash("/opt"))
	return
})

var IsUsrLocal = cache.New[bool](func(p *bool) (err error) {
	*p = strings.HasPrefix(Executable(),
		filepath.FromSlash("/usr/local"))
	return
})

var IsSuperUser = cache.New[bool](func(p *bool) error {
	*p = os.Geteuid() == 0
	return nil
})

var Is = struct {
	Opt, UsrLocal, SuperUser func() bool
}{
	Opt:       IsOpt.Value,
	UsrLocal:  IsUsrLocal.Value,
	SuperUser: IsSuperUser.Value,
}

// Returns the main module reference in the form of PATH@SYMVER or empty
// if the main module is unavailable, as with GO tests.
var MainReference = cache.New[string](func(p *string) (err error) {
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

// Returns the main module version.
var MainVersion = cache.New[string](func(p *string) (err error) {
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
