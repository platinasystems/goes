// Copyright © 2023-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	"flag"
	"os"
	"os/exec"
	"time"

	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
)

func Cutoff(ctx context.Context, complete bool, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} <duration> <feature> [args]
Perform feature for up to max duration.
`)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); len(args) < 1 {
		return xerrors.Incomplete("duration")
	} else if complete {
		return goes.IntrinsicComplete(ctx, args[1:])
	} else if len(args) < 2 {
		return xerrors.Incomplete("feature")
	}

	timeout, err := time.ParseDuration(args[0])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, xprogram.Path(), args[1:]...)
	cmd.Args[0] = xmain.PackageName()
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
