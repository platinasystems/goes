// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios || freebsd || openbsd

package goes_util

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const daemonCriterion = "all processes whose parents are `goes log ...`"

func Daemons() ([]*os.Process, error) {
	var pids, parents []int

	baseprog := filepath.Base(xprogram.Path())

	cmd := exec.Command("pgrep", "-f", baseprog+" log")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	for {
		var pid int
		if _, err = fmt.Fscan(stdout, &pid); err != nil {
			if errors.Is(err, io.EOF) {
				err = nil
			}
			break
		}
		parents = append(parents, pid)
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}
	for _, parent := range parents {
		cmd = exec.Command("pgrep", "-P", fmt.Sprint(parent))
		if stdout, err = cmd.StdoutPipe(); err != nil {
			return nil, err
		}
		if err = cmd.Start(); err != nil {
			return nil, err
		}
		for {
			var pid int
			_, err = fmt.Fscan(stdout, &pid)
			if err != nil && !errors.Is(err, io.EOF) {
				return nil, err
			}
			pids = append(pids, pid)
		}
		if err = cmd.Wait(); err != nil {
			return nil, err
		}
	}
	procs := make([]*os.Process, len(pids))
	for i, pid := range pids {
		if procs[i], err = os.FindProcess(pid); err != nil {
			break
		}
	}
	return procs, err
}
