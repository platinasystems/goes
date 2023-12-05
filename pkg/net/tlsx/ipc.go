// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"

	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
)

const IPCUsageTemplate = `
usage: {{.}} [-i <file>|-] [<args>]
Daemon IPC.`

func IPCUsageData(ctx context.Context) any {
	return pathctx.StringIn(ctx)
}

func IPC(ctx context.Context, args ...string) error {
	if flag.Search[bool]("complete") {
		return nil
	}
	path := pathctx.Parameter.In(ctx)
	if flag.Search[bool]("help") {
		return usage.Error(IPCUsageTemplate[1:], IPCUsageData(ctx))
	}
	ipc := struct{ path, args []string }{
		path: append(path[:1], "exec"),
		args: make([]string, 0, len(path)+len(args)),
	}
	ipc.args = append(ipc.args, Self().DNS0())
	ipc.args = append(ipc.args, path[1:]...)
	ipc.args = append(ipc.args, args...)
	return Rexec(ctx, ipc.args...)
}
