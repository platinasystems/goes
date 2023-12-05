// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/context/flagctx"
	"github.com/platinasystems/goes/v2/pkg/context/pathctx"
	"github.com/platinasystems/goes/v2/pkg/context/wctx"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const EchoUsageTemplate = `
usage: {{.Path}} [<options>] [<strings>]
Print string(s) to standard output.
{{.Flag}}`

func EchoUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: pathctx.StringIn(ctx),
		Flag: flagctx.StringIn(ctx),
	}
}

func Echo(ctx context.Context, args ...string) error {
	fs := flag.NewSilentFlagSet("echo")
	ctx = flagctx.Parameter.With(ctx, fs)
	esc := fs.Bool("e", false, "Interpret escapes.")
	nonl := fs.Bool("n", false, "Without trailing newline.")
	if flag.Search[bool]("complete") {
		return complete.Last(args, fs)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return usage.Error(EchoUsageTemplate[1:], EchoUsageData(ctx))
	}
	args = fs.Args()
	cw := write.With(ctx, wctx.Parameter.In(ctx))
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(cw, " ")
		}
		if *esc {
			text := []byte(fmt.Sprintf(`"%s"`, arg))
			err := json.Unmarshal(text, &arg)
			if err != nil {
				return err
			}
		}
		fmt.Fprint(cw, arg)
	}
	if !*nonl {
		fmt.Fprintln(cw)
	}
	return ctx.Err()
}
