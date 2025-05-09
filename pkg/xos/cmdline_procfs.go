// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build android || linux || netbsd

package xos

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CmdLine(ctx context.Context, proc *os.Process) (string, error) {
	fn := filepath.Join("/proc", fmt.Sprint(proc.Pid), "cmdline")
	data, err := os.ReadFile(fn)
	if err != nil {
		return "", err
	}
	for i, c := range data {
		if c == 0 {
			data[i] = ' '
		}
	}
	return strings.TrimSpace(string(data)), nil
}
