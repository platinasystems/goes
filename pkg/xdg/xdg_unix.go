// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build unix && !darwin && !ios

package xdg

import "path/filepath"

func goosCacheHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, ".cache")
	}
	return FirstExistingDir(s, VarCache, VarRun)
}

func goosConfigDirs() (dirs []string) {
	if IsOpt() {
		dirs = []string{filepath.Join(EtcOpt, "xdg")}
	} else {
		dirs = []string{filepath.Join(Etc, "xdg")}
	}
	return
}

func goosConfigHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, ".config")
	}
	etc := Etc
	if IsOpt() {
		etc = EtcOpt
	}
	return FirstExistingDir(s, etc)
}

func goosDataDirs() []string {
	return []string{UsrLocalShare, UsrShare}
}

func goosDataHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, ".local", "share")
	}
	share := UsrShare
	if IsUsrLocal() {
		share = UsrLocalShare
	} else if IsOpt() {
		share = OptShare
	}
	return FirstExistingDir(s, share)
}

func goosRunTimeDir() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, ".local", "run")
	}
	return FirstExistingDir(s, VarRun)
}

func goosStateHome() string {
	s := SudoUserHome()
	if len(s) > 0 {
		s = filepath.Join(s, ".local", "state")
	}
	state := VarLib
	if IsUsrLocal() {
		state = VarLocal
	} else if IsOpt() {
		state = VarOpt
	}
	return FirstExistingDir(s, state)
}
