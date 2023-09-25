// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"io"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/rawtty"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func main() {
	defer style.Recovery(io.EOF, context.Canceled)
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()
	tty, err := rawtty.With(ctx)
	if err != nil {
		panic(err)
	}
	defer tty.Close()

	c := exec.Command("sh")

	ptmx, err := pty.Start(c)
	if err != nil {
		panic(err)
	}
	defer ptmx.Close()

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, syscall.SIGWINCH)
	go func() {
		for range ch {
			if err := pty.InheritSize(os.Stdin, ptmx); err != nil {
				log.Printf("error resizing pty: %s", err)
			}
		}
	}()
	ch <- syscall.SIGWINCH // Initial resize.
	defer func() {
		signal.Stop(ch)
		close(ch)
	}()

	go func() {
		defer style.Recovery(err, io.EOF, context.Canceled)
		if _, err := io.Copy(ptmx, tty); err != nil {
			panic(err)
		}
	}()
	if _, err = io.Copy(tty, ptmx); err != nil {
		panic(err)
	}
}
