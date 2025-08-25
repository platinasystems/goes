// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"fmt"
	"io"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/xcontext"
	"github.com/platinasystems/goes/v2/pkg/xos"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(),
		xos.Termination...)
	defer stop()
	tty, err := xcontext.WithRawTTY(ctx)
	if err != nil {
		if err != io.EOF {
			panic(err)
		}
		return
	}
	defer tty.Close()
	fmt.Fprint(tty, "Echo input until \"\\r~.\"\n\r")
	_, err = io.Copy(tty, tty)
	if err != nil && err != io.EOF {
		panic(err)
	}
}
