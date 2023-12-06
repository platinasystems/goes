// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const EchoUsageTemplate = `
usage: {{.Path}} [<options>] [<strings>]
Print string(s) to standard output.
{{.Flag}}`

func EchoUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		Path: strings.Join(ctxparm.Strings.In(ctx), " "),
		Flag: ctxparm.SprintFlagsIn(ctx),
	}
}

func Echo(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	flags := usage.NewFlags("echo")
	ctx = ctxparm.Flags.With(ctx, flags)
	esc := flags.Bool("e", false, "Interpret escapes.")
	nonl := flags.Bool("n", false, "Without trailing newline.")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(EchoUsageTemplate[1:], EchoUsageData(ctx))
	}
	args = flags.Args()
	cw := write.With(ctx, ctxparm.Writer.In(ctx))
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
