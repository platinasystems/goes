// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"fmt"
	"os/exec"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func ShowDaemons(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}}
List PIDs with this same executable and /dev/null stdin.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	pids, err := Daemons()
	if err != nil {
		return err
	}
	for _, pid := range pids {
		fmt.Println(pid)
	}
	return nil
}

// If not a /ko-app, execute feature as a detached process with output piped to
// the system logger; otherwise, [goes.Reselect] [goes.Features] within the
// current process context.
func StartDaemon(ctx context.Context, complete bool, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} <feature> [args]
If not a /ko-app, execute feature as a detached process with output piped to
the system logger; otherwise, perform within the current process context.
`)
	if len(args) > 0 && args[0] == "-h" {
		flag.CommandLine.Usage()
		return nil
	}
	if complete {
		return goes.IntrinsicComplete(ctx, args)
	} else if len(args) == 0 {
		return xerrors.Incomplete("feature")
	}

	if xprogram.IsKoApp() {
		return goes.Reselect(ctx, args)
	}

	cmd := exec.Command(xprogram.Path(),
		append([]string{"log"}, args...)...)
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Args[0] = xprogram.MainName()
	cmd.Env = DaemonEnv()
	cmd.Dir = DaemonWorkingDirectory
	attr, err := xexec.DaemonSysProcAttr()
	if err == nil {
		cmd.SysProcAttr = attr
		err = cmd.Start()
	}
	return err
}

func StopDaemons(ctx context.Context, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}}
Terminate all processes with this same executable and /dev/null stdin.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	pids, err := Daemons()
	if err != nil {
		return err
	}
	for _, pid := range pids {
		terr := Terminate(pid)
		if err == nil {
			err = terr
		}
	}
	return err
}
