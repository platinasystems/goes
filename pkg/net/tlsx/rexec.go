// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/goes"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const RexecUsage = `
usage: {{branch .}} [<options>] <exchange> <command> [<args>]
Remote execution.

<exchange>
	<name>[@<dns|ip4|\[ip6\]>][:<port>]
{{flags .}}`

func Rexec(ctx context.Context, args []string) error {
	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	if err := InitRootCAs(); err != nil {
		return err
	}
	if goes.ContextComplete(ctx) {
		if len(args) <= 1 {
			return complete.Last(args, Self.DNS0(),
				Subscriptions.Names())
		}
		return nil
	}
	iflag := flags.String("i", "", "Input FILE or '-' for STDIN.")
	tflag := flags.Bool("t", false, "Allocate a pseudo-TTY.")
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, RexecUsage)
	}
	args = flags.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}

	ex := args[0]
	if args = args[1:]; len(args) == 0 {
		return ErrIncomplete
	}

	r := goes.ContextStdin(ctx)
	w := goes.ContextStdout(ctx)

	var anyargs []any
	if *tflag {
		if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
			anyargs = append(anyargs,
				"pty", ws.Rows, ws.Cols, ws.X, ws.Y)
		} else {
			return err
		}
	} else if len(*iflag) == 0 {
		r = nil
	} else if *iflag == "-" {
	} else if f, err := os.Open(*iflag); err == nil {
		defer f.Close()
		r = f
	} else {
		return err
	}
	for _, arg := range args {
		anyargs = append(anyargs, arg)
	}

	conn, err := Connect(ctx, ex)
	if err != nil {
		return err
	}

	tlsc, err := greetServer(ctx, ex, conn)
	if err != nil {
		return err
	}

	if err = Exec(ctx, tlsc, r, w, anyargs...); err != nil {
		err = fmt.Errorf("%s: %w", ex, err)
	}
	return err
}
