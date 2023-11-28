// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
)

func Echo(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] [<strings>]
Print string(s) to standard output.
{{SprintDefault .Flags}}`
	fs := flag.NewSilentFlagSet("echo")
	esc := fs.Bool("e", false, "Interpret escapes.")
	nonl := fs.Bool("n", false, "Without trailing newline.")
	if flag.Search[bool]("complete") {
		style.Completions(args, fs)
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return style.Usage(usage, struct {
			Path  []string
			Flags *flag.FlagSet
		}{path, fs})
	}
	args = fs.Args()
	cw := write.With(ctx, w)
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
