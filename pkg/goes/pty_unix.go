// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/nbr"
	"github.com/platinasystems/goes/v2/pkg/io/flusher"
	"github.com/platinasystems/goes/v2/pkg/os/host"
	"github.com/platinasystems/goes/v2/pkg/os/utmpx"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
	"golang.org/x/sys/unix"
)

func IntegralPTY(ctx context.Context, args []string) error {
	const usage = `
usage: {{branch .}} [<options>] <command> [<args>]
Run command in an allocated TTY.

Options{{flags .}}`

	var wg sync.WaitGroup
	defer wg.Wait()

	flags := ContextFlags(ctx)
	rFlag := flags.Uint("r", 24, "Rows.")
	cFlag := flags.Uint("c", 80, "Columns.")
	xFlag := flags.Uint("x", 0, "X-pixels.")
	yFlag := flags.Uint("y", 0, "Y-pixels.")

	if ContextComplete(ctx) {
		return complete.Last(args, flags, Root)
	}
	ctx, err := ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if ContextHelp(ctx) {
		if len(args) == 0 {
			return Usage(ctx, usage)
		}
		return Do(ctx, Select, args)
	}
	if args = flags.Args(); len(args) == 0 {
		return ErrIncomplete
	}
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

	r := ContextStdin(ctx)
	w := ContextStdout(ctx)

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
			utx := utmpx.NewUserProcess("root", tty.Name(),
				host.Name(), cmd.Process.Pid)
			defer utx.Died()
		}
		err = cmd.Wait()
	}
	return err
}
