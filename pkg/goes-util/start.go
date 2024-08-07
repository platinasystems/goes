// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"os"
	"os/exec"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xdg"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

// If not a /ko-app, execute feature as a detached process with output piped to
// the system logger; otherwise, it perform within the current process context.
func Start(ctx context.Context, complete bool, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} <feature> [args]
If not a /ko-app, execute feature as a detached process with output piped to
the system logger; otherwise, it perform within the current process context.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		return goes.IntrinsicComplete(ctx, args)
	} else if len(args) == 0 {
		return xerrors.Incomplete("feature")
	}

	var cmd *exec.Cmd
	if xdg.IsKoApp() {
		cmd = exec.Command(xprogram.Path(), args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
	} else {
		cmd = exec.Command(xprogram.Path(),
			append([]string{"log"}, args...)...)
		cmd.Stdin = nil
		cmd.Stdout = nil
		cmd.Stderr = nil
	}

	cmd.Env = DaemonEnv()
	cmd.Dir = xdg.RunTimeDir()
	if _, err = os.Stat(cmd.Dir); err != nil {
		if !os.IsNotExist(err) {
			return err
		}
		cmd.Dir = os.TempDir()
		err = nil
	}
	if !xdg.IsKoApp() {
		cmd.SysProcAttr, err = xexec.DaemonSysProcAttr()
		if err == nil {
			err = cmd.Start()
		}
	} else {
		err = cmd.Run()
	}
	return err
}
