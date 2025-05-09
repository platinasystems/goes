// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// darwin doesn't permit sysctl access to sysctl so fall back to ps
// haven't tried other, non-procfs based unix

//go:build ignore

package xos

import (
	"bytes"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xos"
)

func CmdLine(proc *os.Process) (string, error) {
	buf, err := xos.SysctlGet(
		xos.CTL_KERN,
		xos.KERN_PROCARGS2,
		int32(proc.Pid),
	)
	if err != nil {
		return "", err
	}
	for i, c := range buf {
		if c == 0 {
			return string(bytes.Clone(buf[:i])), nil
		}
	}
	return string(buf), nil
}
