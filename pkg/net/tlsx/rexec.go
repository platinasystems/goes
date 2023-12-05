// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"fmt"
	"os"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/rctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const RexecUsageTemplate = `
usage: {{.Path}} [<options>] <exchange> <command> [<args>]
Remote execution.

<exchange>
	<name>[@<dns|ip4|\[ip6\]>][:<port>]
{{.Flag}}`

func RexecUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		pathctx.StringIn(ctx),
		flagctx.StringIn(ctx),
	}
}

func Rexec(ctx context.Context, args ...string) error {
	fs := flag.NewSilentFlagSet("rexec")
	ctx = flagctx.Parameter.With(ctx, fs)
	iflag := fs.String("i", "", "Input FILE or '-' for STDIN.")
	tflag := fs.Bool("t", false, "Allocate a pseudo-TTY.")
	if flag.Search[bool]("complete") {
		if len(args) <= 1 {
			return complete.Last(args, fs, Self().DNS0(),
				Subscriptions().Names())
		}
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(RexecUsageTemplate[1:], RexecUsageData(ctx))
	}
	args = fs.Args()
	if len(args) == 0 {
		return ErrIncomplete
	}

	ex := args[0]
	if args = args[1:]; len(args) == 0 {
		return ErrIncomplete
	}

	r := rctx.Parameter.In(ctx)
	w := wctx.Parameter.In(ctx)

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
