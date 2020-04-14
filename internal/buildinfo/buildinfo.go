// Copyright © 2019-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// +build go1.12

// This package provides a runtime/debug.BuildInfo Formatter.
// Usage,
//
//	if bi := buildinfo.New(); bi.Version() != buildinfo.Unavailable {
//		fmt.Println(bi)
//	} else {
//		fmt.Println(buildinfo.Unavailable)
//	}
package buildinfo

import (
	"fmt"
	"io"
	"runtime/debug"
)

type BuildInfo struct {
	*debug.BuildInfo
}

func New() BuildInfo {
	if bi, ok := debug.ReadBuildInfo(); ok {
		return BuildInfo{bi}
	}
	return BuildInfo{}
}

func (bi BuildInfo) Format(f fmt.State, c rune) {
	if bi.BuildInfo == nil {
		f.Write(unavailable)
	} else {
		modinfo(f, &bi.Main)
		for _, dep := range bi.Deps {
			f.Write([]byte("\n\t"))
			modinfo(f, dep)
		}
	}
}

func (bi BuildInfo) Version() string {
	s := Unavailable
	if bi.BuildInfo != nil {
		s = bi.Main.Version
	}
	return s
}

func modinfo(w io.Writer, m *debug.Module) {
	w.Write([]byte(m.Path))
	if m.Replace != nil {
		w.Write([]byte("="))
		w.Write([]byte(m.Replace.Path))
		if len(m.Replace.Version) > 0 {
			w.Write([]byte("@"))
			w.Write([]byte(m.Replace.Version))
		}
	} else {
		w.Write([]byte("@"))
		w.Write([]byte(m.Version))
	}
}
