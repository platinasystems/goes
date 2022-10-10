// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/context/rawtty"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()
	tty, err := rawtty.With(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return
	}
	defer tty.Close()
	fmt.Fprint(tty, "Echo input until \"\\r~.\"\n\r")
	_, err = io.Copy(tty, tty)
	if suppress.Errors(err, io.EOF) != nil {
		fmt.Fprint(os.Stderr, "\n\r", err, "\n\r")
	}
}
