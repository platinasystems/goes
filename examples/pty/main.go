// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
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
	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

func main() {
	suppressed := []error{io.EOF, context.Canceled}
	ctx, stop := signal.NotifyContext(context.Background(),
		xos.Termination...)
	defer stop()
	tty, err := xcontext.WithRawTTY(ctx)
	if err != nil {
		if err = xerrors.Suppress(err, suppressed...); err != nil {
			log.Print(err)
		}
		return
	}
	defer tty.Close()

	c := exec.Command("sh")

	ptmx, err := pty.Start(c)
	if err != nil {
		if err = xerrors.Suppress(err, suppressed...); err != nil {
			log.Print(err)
		}
		return
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
		_, err := io.Copy(ptmx, tty)
		if err = xerrors.Suppress(err, suppressed...); err != nil {
			log.Print(err)
		}
	}()
	_, err = io.Copy(tty, ptmx)
	if err = xerrors.Suppress(err, suppressed...); err != nil {
		log.Print(err)
	}
}
