// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes_util

import (
	"context"
	_ "embed"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xexec"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xprogram"
	"github.com/platinasystems/goes/v2/pkg/xsync"
	"golang.org/x/sys/unix"
)

func Pty(ctx context.Context, complete bool, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <feature< [args]
Execute feature in an allocated TTY.

{{flags .}}`)

	rFlag := flag.Uint("r", 24, "Rows.")
	cFlag := flag.Uint("c", 80, "Columns.")
	xFlag := flag.Uint("x", 0, "X-pixels.")
	yFlag := flag.Uint("y", 0, "Y-pixels.")

	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		return goes.IntrinsicComplete(ctx, args)
	} else if len(args) == 0 {
		return xerrors.Incomplete("feature")
	}

	var wg xsync.WaitGroup
	defer wg.Wait()

	ws := pty.Winsize{
		Rows: uint16(*rFlag),
		Cols: uint16(*cFlag),
		X:    uint16(*xFlag),
		Y:    uint16(*yFlag),
	}
	ptmx, tty, err := pty.Open()
	if err != nil {
		return err
	}
	if err = pty.Setsize(tty, &ws); err != nil {
		ptmx.Close()
		return fmt.Errorf("resize: %w", err)
	}

	wg.Go(func() {
		io.Copy(ptmx, os.Stdin)
		ptmx.Close()
		io.ReadAll(os.Stdin)
	})

	wg.Go(func() {
		io.Copy(os.Stdout, xcontext.WithNBR(ctx, ptmx))
	})

	cmd := exec.CommandContext(ctx, xprogram.Path(), args...)
	cmd.Args[0] = xprogram.MainName()
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.SysProcAttr = &unix.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err = cmd.Start(); err == nil {
		if os.Geteuid() == 0 {
			hostname, _ := os.Hostname()
			utx := xexec.NewUserProcess("root", tty.Name(),
				hostname, cmd.Process.Pid)
			defer utx.Died()
		}
		err = cmd.Wait()
	}
	return err
}
