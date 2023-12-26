// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"

	"github.com/platinasystems/goes/v2/pkg/goes"
)

const EchoUsage = `
usage: {{branch .}} [<options>] [<strings>]
Print string(s) to standard output.
{{flags .}}`

func Echo(ctx context.Context, args []string) error {
	var flags flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	esc := flags.Bool("e", false, "Interpret escapes.")
	nonl := flags.Bool("n", false, "Without trailing newline.")
	if goes.ContextComplete(ctx) {
		return nil
	}
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, EchoUsage)
	}
	args = flags.Args()
	w := goes.ContextStdout(ctx)
	for i, arg := range args {
		if i > 0 {
			fmt.Fprint(w, " ")
		}
		if *esc {
			text := []byte(fmt.Sprintf(`"%s"`, arg))
			err := json.Unmarshal(text, &arg)
			if err != nil {
				return err
			}
		}
		fmt.Fprint(w, arg)
	}
	if !*nonl {
		fmt.Fprintln(w)
	}
	return ctx.Err()
}
