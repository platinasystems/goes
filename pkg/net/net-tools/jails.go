// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build freebsd

package net_tools

import (
	"context"
	"fmt"
)

const HaveJails = true

func EnterJail(ctx context.Context, name string) error {
	// syscall.ForkExec self with ProcAttr.Sys.Jail
	return fmt.Errorf("jail %w", FIXME)
}
