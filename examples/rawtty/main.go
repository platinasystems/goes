// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/context/rawtty"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

func main() {
	defer style.Recovery(io.EOF)
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()
	tty, err := rawtty.With(ctx)
	if err != nil {
		panic(err)
	}
	defer tty.Close()
	fmt.Fprint(tty, "Echo input until \"\\r~.\"\n\r")
	if _, err = io.Copy(tty, tty); err != nil {
		panic(err)
	}
}
