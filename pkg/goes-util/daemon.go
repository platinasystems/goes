// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xos"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsignal"
)

func AlarmDaemons(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [pid]...
Send alarm to identified or ` + daemonCriterion + ".\n")
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	return DoPIDsOrDaemons(ctx, args, func(proc *os.Process) error {
		return proc.Signal(xsignal.Alarm)
	})
}

func DoDaemons(ctx context.Context, f func(*os.Process) error) error {
	procs, err := Daemons(ctx)
	if err != nil {
		return err
	}
	return DoProcs(ctx, procs, f)
}

func DoPIDs(
	ctx context.Context, args []string, f func(*os.Process) error,
) error {
	var procs []*os.Process
	for _, s := range args {
		if pid, err := strconv.ParseInt(s, 10, 0); err != nil {
			return err
		} else if proc, err := os.FindProcess(int(pid)); err != nil {
			return err
		} else {
			procs = append(procs, proc)
		}
	}
	return DoProcs(ctx, procs, f)
}

func DoPIDsOrDaemons(
	ctx context.Context, args []string, f func(*os.Process) error,
) error {
	if len(args) > 0 {
		return DoPIDs(ctx, args, f)
	}
	return DoDaemons(ctx, f)
}

func DoProcs(
	ctx context.Context, procs []*os.Process, f func(*os.Process) error,
) (err error) {
	for _, proc := range procs {
		cl, clerr := xos.CmdLine(ctx, proc)
		if t := f(proc); t != nil && err == nil {
			err = fmt.Errorf("%d: %w", proc.Pid, t)
		} else if clerr == nil {
			fmt.Print(proc.Pid, "\t", cl, "\n")
		} else {
			fmt.Print(proc.Pid, "\t", clerr, "\n")
		}
	}
	return
}

func ShowDaemons(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [pid]...
Print identified or ` + daemonCriterion + ".\n")
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	return DoPIDsOrDaemons(ctx, args, func(proc *os.Process) error {
		return nil
	})
}

// If not a /ko-app, execute feature as a detached process with output piped to
// the system logger; otherwise, [goes.Reselect] [goes.Features] within the
// current process context.
func StartDaemon(ctx context.Context, complete bool, args []string) error {
	xflag.TemplateUsage(`
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

	for _, s := range args {
		if s == "-h" || s == "-help" || s == "--help" {
			return goes.Reselect(ctx, args)
		}
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
	xflag.TemplateUsage(`
usage: {{.Name}} [pid]...
Terminate identified or ` + daemonCriterion + ".\n")
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	}
	return DoPIDsOrDaemons(ctx, args, func(proc *os.Process) error {
		return proc.Signal(xsignal.Terminate)
	})
}
