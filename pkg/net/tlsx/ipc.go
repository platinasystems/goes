// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/goes"
)

const IPCUsage = `
usage: {{branch .}} [-i <file>|-] [<args>]
Daemon IPC.`

func IPC(ctx context.Context, args []string) error {
	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, IPCUsage)
	}
	self := Self()
	if self == nil {
		return ErrNotFound
	}
	branch := goes.ContextBranch(ctx)
	ctx = goes.BranchContext(ctx, append(branch[:1], "ipc"))
	ipc := make([]string, 0, len(branch)+len(args))
	ipc = append(ipc, self.DNS0())
	ipc = append(ipc, branch[1:]...)
	ipc = append(ipc, args...)
	return Rexec(ctx, ipc)
}
