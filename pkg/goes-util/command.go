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

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"golang.org/x/term"
)

const (
	Command_p_Flag xflag.KeyUsage[bool] = "p Restricted path search."
	Command_v_Flag xflag.KeyUsage[bool] = "v Report path found."
	Command_V_Flag xflag.KeyUsage[bool] = "V More verbose report."
)

func Command(ctx context.Context, complete bool, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <external> [args]
Execute “external” PATH command.

{{flags .}}`)

	pFlag := Command_p_Flag.Define(false)
	vFlag := Command_v_Flag.Define(false)
	vvFlag := Command_V_Flag.Define(false)

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		if nargs := len(args); nargs < 2 {
			var prefix string
			if nargs == 1 {
				prefix = args[0]
			}
			for _, s := range xexec.Match(prefix) {
				fmt.Println(s)
			}
		}
		return nil
	} else if len(args) == 0 {
		return xerrors.Incomplete("external")
	}

	lookpath := exec.LookPath
	if *pFlag {
		lookpath = xexec.RestrictedLookPath
	}
	full, err := lookpath(args[0])
	if err != nil {
		return err
	}
	if *vFlag {
		fmt.Println(full)
		return nil
	}
	if *vvFlag {
		fmt.Println(args[0], "is", full)
		return nil
	}
	cmd := exec.CommandContext(ctx, full, args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if fd := int(os.Stdout.Fd()); term.IsTerminal(fd) {
		cmd.Stderr = os.Stdout
		cmd.SysProcAttr, err = xexec.InteractiveSysProcAttr()
	}
	if err = cmd.Start(); err == nil {
		err = cmd.Wait()
	}
	return err
}
