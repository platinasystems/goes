// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xprogram

import (
	"os"
	"path"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"sync"
)

type BuildSettingKey string

const (
	BuileMode   BuildSettingKey = "-buildmode"
	Compiler    BuildSettingKey = "-compiler"
	CgoEnabled  BuildSettingKey = "CGO_ENABLED"
	CgoCPPFlags BuildSettingKey = "CGO_CPPFLAGS"
	CgoCXXFlags BuildSettingKey = "CGO_CXXFLAGS"
	CgoLDFlags  BuildSettingKey = "CGO_LDFLAGS"
	GoARCH      BuildSettingKey = "GOARCH"
	GoOS        BuildSettingKey = "GOOS"
	Vcs         BuildSettingKey = "vcs"
	VcsModified BuildSettingKey = "vcs.modified"
	VcsRevision BuildSettingKey = "vcs.revision"
	VcsTime     BuildSettingKey = "vcs.time"
)

var BuildSettings = map[string]any{
	string(BuileMode):   BuileMode,
	string(Compiler):    Compiler,
	string(CgoEnabled):  CgoEnabled,
	string(CgoCPPFlags): CgoCPPFlags,
	string(CgoCXXFlags): CgoCXXFlags,
	string(CgoLDFlags):  CgoLDFlags,
	string(GoARCH):      GoARCH,
	string(GoOS):        GoOS,
	string(Vcs):         Vcs,
	string(VcsModified): VcsModified,
	string(VcsRevision): VcsRevision,
	string(VcsTime):     VcsTime,
}

func (bsk BuildSettingKey) String() string {
	if bi := BuildInfo(); bi != nil {
		for _, bs := range bi.Settings {
			if bs.Key == string(bsk) {
				return bs.Value
			}
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

func BuildInfo() *debug.BuildInfo {
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		bi = nil
	}
	return bi
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
