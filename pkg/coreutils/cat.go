// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package coreutils

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/context/read"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Cat(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `{{/*
*/}}usage: {{join . " "}} [<file(s)>|-]
Concatenate file(s) or standard in (-) to output.
`
	if complete.Parameter.Value(ctx) {
		style.Completions(args, "*")
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, path)
	}
	if len(args) == 0 {
		args = append(args, "-")
	}
	cr := read.With(ctx, r)
	cw := write.With(ctx, w)
	for _, fn := range args {
		if fn == "-" {
			io.Copy(cw, cr)
		} else if fi, err := os.Stat(fn); err == nil && fi.IsDir() {
			return fmt.Errorf("%s: is a directory", fn)
		} else if f, err := os.Open(fn); err == nil {
			io.Copy(cw, f)
			f.Close()
		} else {
			return fmt.Errorf("%s: %w", fn, errors.Unwrap(err))
		}
	}
	return ctx.Err()
}
