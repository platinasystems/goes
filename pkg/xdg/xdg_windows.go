// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdg

import (
	"os"
	"os/user"
	"path/filepath"
)

func windowsLocalAppData() string {
	s := os.Getenv("LocalAppData")
	if len(s) > 0 {
		return s
	}
	if u, err := user.Current(); err == nil {
		s = filepath.Join(u.HomeDir, "AppData", "Local")
	}
	return s
}

func windowsRoamingAppData() string {
	s := os.Getenv("AppData")
	if len(s) > 0 {
		return s
	}
	if u, err := user.Current(); err == nil {
		s = filepath.Join(u.HomeDir, "AppData", "Roaming")
	}
	return s
}

func windowsProgramData() string {
	s := os.Getenv("ProgramData")
	if len(s) > 0 {
		return s
	}
	if s = os.Getenv("ALLUSERSPROFILE"); len(s) > 0 {
		return s
	}
	if s = os.Getenv("SystemDrive"); len(s) > 0 {
		s = filepath.Join(s, "ProgramData")
	}
	return s
}

func goosCacheHome() string {
	s := windowsLocalAppData()
	if len(s) > 0 {
		s = filepath.Join(s, "cache")
	}
	return FirstExistingDir(s)
}

func goosConfigDirs() (dirs []string) {
	s := windowsProgramData()
	if len(s) > 0 {
		dirs = append(dirs, s)
	}
	s = windowsRoamingAppData()
	if len(s) > 0 {
		dirs = append(dirs, s)
	}
	return
}

func goosConfigHome() string {
	return FirstExistingDir(windowsLocalAppData())
}

func goosDataDirs() (dirs []string) {
	s := windowsRoamingAppData()
	if len(s) > 0 {
		dirs = append(dirs, s)
	}
	s = windowsProgramData()
	if len(s) > 0 {
		dirs = append(dirs, s)
	}
	return
}

func goosDataHome() string {
	return FirstExistingDir(windowsLocalAppData())
}

func goosRunTimeDir() string {
	return FirstExistingDir(windowsLocalAppData())
}

func goosStateHome() string {
	return FirstExistingDir(windowsLocalAppData())
}
