// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pty

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/nbr"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/io/flusher"
	"github.com/platinasystems/goes/v2/pkg/os/utmpx"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
	"golang.org/x/sys/unix"
)

const ExecUsage = `
usage: {{branch .}} <rows> <cols> <x-pixels> <y-pixels> <command> [<args>]
Run command in an allocated TTY.`

func Exec(ctx context.Context, args []string) error {
	branch := goes.ContextBranch(ctx)
	r := goes.ContextStdin(ctx)
	w := goes.ContextStdout(ctx)

	if *complete.Help {
		return nil
	}
	if *usage.Help {
		return goes.Usage(ctx, ExecUsage)
	}
	if n := len(args); n < 5 {
		return fmt.Errorf("missing %s", []string{
			"<rows>",
			"<cols>",
			"<x-pixels>",
			"<y-pixels>",
			"<command>",
		}[n])
	}

	var ws pty.Winsize
	for i, pu := range []*uint16{
		&ws.Rows,
		&ws.Cols,
		&ws.X,
		&ws.Y,
	} {
		if _, err := fmt.Sscan(args[i], pu); err != nil {
			return fmt.Errorf("%q: %w",
				args[i], err)
		}
	}
	args = args[4:]

	ptmx, tty, err := pty.Open()
	if err != nil {
		return fmt.Errorf("pty: %w", err)
	}
	if err = pty.Setsize(tty, &ws); err != nil {
		ptmx.Close()
		return fmt.Errorf("resize: %w", err)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go func() {
		in, flush := flusher.New(r)
		io.Copy(ptmx, in)
		ptmx.Close()
		flush()
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		io.Copy(w, nbr.With(ctx, ptmx))
		wg.Done()
	}()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	cmd.Stdin = tty
	cmd.Stdout = tty
	cmd.Stderr = tty
	cmd.SysProcAttr = &unix.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err = cmd.Start(); err == nil {
		if os.Geteuid() == 0 {
			line := tty.Name()
			ra := path[1]
			pid := cmd.Process.Pid
			utx := utmpx.NewUserProcess("root", line, ra, pid)
			defer utx.Died()
		}
		err = cmd.Wait()
	}
	return err
}
