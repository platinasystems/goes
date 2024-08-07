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
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Input(ctx context.Context, complete bool, args []string) error {
	xflag.UsageTemplate(flag.CommandLine, `
usage: {{.Name}} <file> <feature> [args]
Executes feature with file input.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		if len(args) > 1 {
			return goes.IntrinsicComplete(ctx, args[1:])
		}
		// FIXME complete <file>
		return nil
	} else if len(args) == 0 {
		return xerrors.Incomplete("file")
	} else if len(args) == 1 {
		return xerrors.Incomplete("feature")
	}

	r, err := os.Open(args[0])
	if err != nil {
		return err
	}
	defer r.Close()

	cmd := exec.CommandContext(ctx, xprogram.Path(), args[1:]...)
	cmd.Args[0] = xprogram.MainName()
	cmd.Stdin = r
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
