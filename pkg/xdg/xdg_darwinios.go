// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios

package xdg

import "path/filepath"

func goosCacheHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, "Library", "Caches")
	}
	return FirstExistingDir(s, VarRun)
}

func goosConfigDirs() (dirs []string) {
	s := SudoUserHome()
	if len(s) > 0 {
		dirs = append(dirs, filepath.Join(s, "Library", "Preferences"))
	}
	dirs = append(dirs,
		"/Library/Application Support",
		"/Library/Preferences")
	return
}

func goosConfigHome() string {
	var dirs []string
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, "Library", "Application Support")
		dirs = append(dirs, s)
	}
	if IsOpt() {
		dirs = append(dirs, EtcOpt)
	} else {
		dirs = append(dirs, Etc)
	}
	return FirstExistingDir(dirs...)
}

func goosDataDirs() []string {
	return []string{filepath.Join("Library", "Application Support")}
}

func goosDataHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, "Library", "Application Support")
	}
	return FirstExistingDir(s)
}

func goosRunTimeDir() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, "Library", "Application Support")
	}
	return FirstExistingDir(s)
}

func goosStateHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, "Library", "Application Support")
	}
	return FirstExistingDir(s)
}
