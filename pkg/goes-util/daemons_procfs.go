// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build android || linux || netbsd

package goes_util

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const daemonCriterion = "all matching process executables and /dev/null stdin"

func Daemons() ([]*os.Process, error) {
	var pids []int

	thisPid := os.Getpid()
	thisProg := xprogram.Path()
	isKoApp := xprogram.IsKoApp()

	procexes, err := filepath.Glob("/proc/*/exe")
	if err != nil {
		return nil, xerrors.Label(err, "glob")
	}

	for _, procexe := range procexes {
		var pid int
		var s string
		procdir := filepath.Dir(procexe)
		_, err = fmt.Sscan(filepath.Base(procdir), &pid)
		if err != nil || pid == thisPid {
			continue
		}
		if s, err = filepath.
			EvalSymlinks(procexe); err != nil || s != thisProg {
			continue
		}
		procfd0 := filepath.Join(procdir, "fd", "0")
		if s, err = filepath.
			EvalSymlinks(procfd0); err != nil || s != os.DevNull {
			continue
		}
		if isKoApp {
			pids = append(pids, pid)
			continue
		}
		procfd1 := filepath.Join(procdir, "fd", "1")
		if s, err = filepath.
			EvalSymlinks(procfd1); err != nil || s != os.DevNull {
			pids = append(pids, pid)
		}
	}
	sort.Reverse(sort.IntSlice(pids))
	procs := make([]*os.Process, len(pids))
	for i, pid := range pids {
		if procs[i], err = os.FindProcess(pid); err != nil {
			break
		}
	}
	return procs, err
}
