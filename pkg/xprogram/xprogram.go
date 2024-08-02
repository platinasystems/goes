// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xprogram

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sync"
)

func Path() string {
	s, err := os.Executable()
	if err != nil {
		s, err = filepath.EvalSymlinks(os.Args[0])
		if err != nil {
			s = os.Args[0]
		}
	}
	return s
}

func BuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		bi = nil
	}
	return bi
}

func MainModule() *debug.Module {
	bi := BuildInfo()
	if bi == nil {
		return nil
	}
	m := &bi.Main
	if m.Replace != nil {
		m = m.Replace
	}
	return m
}

var MainName = sync.OnceValue(func() string {
	bi := BuildInfo()
	if bi == nil {
		return filepath.Base(Path())
	}
	name := path.Base(bi.Path)
	if t, _ := regexp.MatchString("v[0-9]*", name); t {
		name = path.Base(path.Dir(bi.Path))
	}
	return name
})
