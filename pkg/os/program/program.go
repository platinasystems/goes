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

var SlashExecutable = sync.OnceValue(func() string {
	return filepath.ToSlash(Executable())
})

var IsKoApp = sync.OnceValue(func() bool {
	return strings.HasPrefix(SlashExecutable(), "/ko-app")
})

var IsOpt = sync.OnceValue(func() bool {
	return strings.HasPrefix(SlashExecutable(), "/opt")
})

var IsUsrLocal = sync.OnceValue(func() bool {
	return strings.HasPrefix(SlashExecutable(), "/usr/local")
})

var BuildId = sync.OnceValue(func() string {
	s, err := buildid.ReadFile(Executable())
	if err != nil {
		panic(err)
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
