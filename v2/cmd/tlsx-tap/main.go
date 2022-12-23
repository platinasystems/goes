// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/net/tlsx/tap"
	"github.com/platinasystems/goes/v2/pkg/os/termination"
)

const Usage = `
usage: tlsx-tap [-unit #] [-prefix <prefix>] <command> [<args>]
`

func main() {
	var t tap.T
	var wg sync.WaitGroup
	defer wg.Wait()

	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()
	cctx, cancel := context.WithCancel(ctx)
	defer cancel()

	args, err := t.Configure(cctx, os.Args[1:])
	if err == flag.ErrHelp {
		fmt.Print(Usage[1:])
		return
	} else {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	wg.Add(1)
	go t.Routine(cctx, &wg)

	if len(args) > 0 {
		err = exec.CommandContext(cctx, args[0], args[1:]...).Run()
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}
}
