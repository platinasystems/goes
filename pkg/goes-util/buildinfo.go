// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"errors"
	"fmt"
	"golang/buildid"
	"os"
	"runtime/debug"
)

var ErrUnavailable = errors.New("unavailable")

func BuildId() (string, error) {
	return buildid.ReadFile(os.Args[0])
}

func BuildInfo() (*debug.BuildInfo, error) {
	var err error
	bi, ok := debug.ReadBuildInfo()
	if !ok {
		err = ErrUnavailable
	}
	return bi, err
}

// Returns [debug.BuildInfo.Main] or its replacement.
func MainModule() (*debug.Module, error) {
	bi, err := BuildInfo()
	if err != nil {
		return nil, err
	}
	m := &bi.Main
	if m.Replace != nil {
		m = m.Replace
	}
	return m, nil
}

// The main module reference in the form of PATH@SYMVER or empty if the
// main module is unavailable, as with GO tests.
func MainReference() (string, error) {
	var ref string
	m, err := MainModule()
	if err != nil {
		return ref, err
	}
	if len(m.Path) > 0 && len(m.Version) > 0 {
		ref = fmt.Sprint(m.Path, "@", m.Version)
	}
	return ref, nil
}

// The main module version if available; empty otherwise.
func MainVersion() (string, error) {
	var ver string
	m, err := MainModule()
	if err != nil {
		return ver, err
	}
	if len(m.Version) > 0 {
		ver = m.Version
	}
	return ver, nil
}
