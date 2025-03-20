// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios || freebsd || openbsd

package goes_util

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

const daemonCriterion = "all processes whose parents are `goes log ...`"

func Daemons(ctx context.Context) (procs []*os.Process, err error) {
	baseprog := filepath.Base(xprogram.Path())

	parents, err := pgrep(ctx, "-f", baseprog+" log")
	if err != nil {
		return
	}
	if len(parents) == 0 {
		return nil, nil
	}
	for _, parent := range parents {
		var pids []int
		pids, err = pgrep(ctx, "-P", fmt.Sprint(parent))
		if err != nil {
			break
		}
		for _, pid := range pids {
			if proc, e := os.FindProcess(pid); e == nil {
				procs = append(procs, proc)
			}
		}
	}
	return
}

func pgrep(ctx context.Context, args ...string) (pids []int, err error) {
	output, err := exec.CommandContext(ctx, "pgrep", args...).Output()
	if err != nil {
		return
	}
	scanner := bufio.NewScanner(bytes.NewReader(output))
	for scanner.Scan() {
		var pid int
		line := scanner.Text()
		if _, err = fmt.Sscan(line, &pid); err != nil {
			return
		}
		pids = append(pids, pid)
	}
	return
}
