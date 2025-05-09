// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || ios || freebsd || openbsd

package xos

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func CmdLine(ctx context.Context, proc *os.Process) (string, error) {
	spid := fmt.Sprint(proc.Pid)
	data, err := exec.CommandContext(ctx, "ps", "-o", "command=", spid).
		Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(data)), nil
}
