// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package pty

import (
	"context"
	"fmt"
	"io"
	"os/exec"
	"strings"
	"sync"
	"syscall"
	"text/template"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/nbr"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/input"
	"github.com/platinasystems/goes/v2/pkg/os/program"
	"github.com/platinasystems/goes/v2/pkg/os/utmpx"
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return template.Must(template.New("usage").Parse(`
usage: {{.}} <rows> <cols> <x-pixels> <y-pixels> <command> [<args>]
run command in an allocated TTY.
`[1:])).Execute(w, strings.Join(path, " "))
	}
	var wg sync.WaitGroup
	defer wg.Wait()

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

	wg.Add(1)
	go func() {
		ir, flush := input.New(r)
		io.Copy(ptmx, ir)
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
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid:  true,
		Setctty: true,
	}
	if err = cmd.Start(); err == nil {
		if program.IsSuperUser.Value() {
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
