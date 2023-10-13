// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"io"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func IPC(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `usage: {{join . " "}} [-i <file>|-] [<args>]
Daemon IPC.
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, path)
	}
	ipc := struct{ path, args []string }{
		path: []string{path[0], "exec"},
		args: make([]string, 0, len(path)+len(args)),
	}
	ipc.args = append(ipc.args, Self().DNS0())
	ipc.args = append(ipc.args, path[1:]...)
	ipc.args = append(ipc.args, args...)
	return Rexec(ctx, r, w, ipc.path, ipc.args...)
}
