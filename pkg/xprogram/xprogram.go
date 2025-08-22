// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xprogram

import (
	"golang/buildid"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"runtime/debug"
	"strings"
	"sync"
)

type BuildSetting string

const (
	BuildMode   = BuildSetting("buildmode")
	Compiler    = BuildSetting("compiler")
	CgoEnabled  = BuildSetting("enabled")
	CgoCPPFlags = BuildSetting("cpp")
	CgoCXXFlags = BuildSetting("cxx")
	CgoLDFlags  = BuildSetting("ld")
	GoARCH      = BuildSetting("arch")
	GoMicroARCH = BuildSetting(runtime.GOARCH)
	GoOS        = BuildSetting("os")
	VcsModified = BuildSetting("modified")
	VcsRevision = BuildSetting("revision")
	VcsTime     = BuildSetting("time")
	VcsType     = BuildSetting("type")
)

var Show = map[string]any{
	"show": map[string]any{
		"build": map[string]any{
			"id":   BuildId,
			"info": BuildInfoString,
			"mode": BuildMode,
		},
		"cgo": map[string]any{
			string(CgoCPPFlags): CgoCPPFlags,
			string(CgoCXXFlags): CgoCXXFlags,
			string(CgoEnabled):  CgoEnabled,
			string(CgoLDFlags):  CgoLDFlags,
		},
		string(Compiler): Compiler,
		"go": map[string]any{
			string(GoARCH):      GoARCH,
			string(GoMicroARCH): GoMicroARCH,
			string(GoOS):        GoOS,
		},
		"main": map[string]any{
			"name":      MainName,
			"reference": MainReference,
			"version":   MainVersion,
		},
		"vcs": map[string]any{
			string(VcsModified): VcsModified,
			string(VcsRevision): VcsRevision,
			string(VcsTime):     VcsTime,
			string(VcsType):     VcsType,
		},
	},
}

var BuildSettingKey = map[BuildSetting]string{
	BuildMode:   "-" + string(BuildMode),
	Compiler:    "-" + string(Compiler),
	CgoEnabled:  "CGO_ENABLED",
	CgoCPPFlags: "CGO_CPPFLAGS",
	CgoCXXFlags: "CGO_CXXFLAGS",
	CgoLDFlags:  "CGO_LDFLAGS",
	GoARCH:      "GOARCH",
	GoOS:        "GOOS",
	VcsModified: "vcs." + string(VcsModified),
	VcsRevision: "vcs." + string(VcsRevision),
	VcsTime:     "vcs." + string(VcsTime),
	VcsType:     "vcs",
}

func (bs BuildSetting) String() string {
	bi := BuildInfo()
	if bi == nil {
		return ""
	}
	key, ok := BuildSettingKey[bs]
	if !ok {
		if bs != GoMicroARCH {
			panic(string(bs))
		}
		key = "GO" + strings.ToUpper(string(bs))
	}
	for _, bs := range bi.Settings {
		if bs.Key == key {
			return bs.Value
		}
	}
	return ""
}

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

func BuildId() string {
	s, _ := buildid.ReadFile(Path())
	return s
}

func BuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		bi = nil
	}
	return bi
}

func BuildInfoString() string {
	if bi := BuildInfo(); bi != nil {
		return bi.String()
	}
	return ""
}

func IsOpt() bool {
	return strings.HasPrefix(Path(), filepath.FromSlash("/opt"))
}

func IsUsrLocal() bool {
	return strings.HasPrefix(Path(), filepath.FromSlash("/usr/local"))
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
	const major = "^v[1-9][0-9]*$"
	bi := BuildInfo()
	if bi == nil {
		return filepath.Base(Path())
	}
	name := filepath.Base(bi.Path)
	if t, err := regexp.MatchString(major, name); err != nil {
		name = err.Error()
	} else if t {
		name = filepath.Base(filepath.Dir(bi.Path))
	}
	return name
})

func MainReference() string {
	if mm := MainModule(); mm != nil {
		return mm.Path + "@" + mm.Version
	}
	return ""
}

func MainVersion() string {
	if mm := MainModule(); mm != nil {
		return mm.Version
	}
	return ""
}
