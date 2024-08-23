// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios || freebsd || openbsd

package goes_util

import (
	"errors"
	"fmt"
	"io"
	"os/exec"
	"path/filepath"

	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Daemons() (daemons []int, err error) {
	var parents []int

	baseprog := filepath.Base(xprogram.Path())

	cmd := exec.Command("pgrep", "-f", baseprog+" log")
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return
	}
	if err = cmd.Start(); err != nil {
		return
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
		return
	}
	for _, parent := range parents {
		cmd = exec.Command("pgrep", "-P", fmt.Sprint(parent))
		if stdout, err = cmd.StdoutPipe(); err != nil {
			break
		}
		if err = cmd.Start(); err != nil {
			break
		}
		for {
			var pid int
			if _, err = fmt.Fscan(stdout, &pid); err != nil {
				if errors.Is(err, io.EOF) {
					err = nil
				}
				break
			}
			daemons = append(daemons, pid)
		}
		if err = cmd.Wait(); err != nil {
			break
		}
	}
	return
}
