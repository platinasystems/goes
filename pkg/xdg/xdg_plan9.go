// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdg

import (
	"os"
	"path/filepath"
)

func goosCacheHome() string {
	s := os.Getenv("home")
	if len(s) > 0 {
		s = filepath.Join(s, "lib", "cache")
	}
	return FirstExistingDir(s)
}

func goosConfigHome() string {
	s := os.Getenv("home")
	if len(s) > 0 {
		s = filepath.Join(s, "lib")
	}
	return FirstExistingDir(s)
}

func goosConfigDirs() []string {
	return []string{Lib}
}

func goosDataHome() string {
	s := os.Getenv("home")
	if len(s) > 0 {
		s = filepath.Join(s, "lib")
	}
	return FirstExistingDir(s)
}

func goosDataDirs() []string {
	return []string{Lib}
}

func goosRunTimeDir() string {
	return os.TempDir()
}

func goosStateHome() string {
	s := os.Getenv("home")
	if len(s) > 0 {
		s = filepath.Join(s, "lib", "state")
	}
	return FirstExistingDir(s)
}
