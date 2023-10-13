// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Rexec(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] <exchange> <command> [<args>]
Remote execution.
{{print .Flags}}
<exchange>
	<name>[@<dns|ip4|\[ip6\]>][:<port>]
`
	fs, h := flag.New()
	iflag := fs.String("i", "", "Input FILE or '-' for STDIN.")
	tflag := fs.Bool("t", false, "Allocate a pseudo-TTY.")
	if complete.Parameter.Value(ctx) {
		if len(args) <= 1 {
			style.Completions(args, fs.FlagSet, Self().DNS0(),
				Subscriptions().Names())
		}
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if help.Parameter.Value(ctx) || *h {
		return style.Usage(usage, struct {
			Path  []string
			Flags fmt.Formatter
		}{path, fs})
	}
	args = fs.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}

	ex := args[0]
	if args = args[1:]; len(args) == 0 {
		return ErrIncomplete
	}

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
