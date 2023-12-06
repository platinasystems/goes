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
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/context/read"
	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const CatUsageTemplate = `
usage: {{.}} [<file(s)>|-]
Concatenate file(s) or standard in (-) to output.`

func CatUsageData(ctx context.Context) any {
	return strings.Join(ctxparm.Strings.In(ctx), " ")
}

func Cat(ctx context.Context, args ...string) error {
	if *complete.Help {
		return complete.Last(args, "*")
	}
	if *usage.Help {
		return usage.Error(CatUsageTemplate[1:], CatUsageData(ctx))
	}
	if len(args) == 0 {
		args = append(args, "-")
	}
	cr := read.With(ctx, ctxparm.Reader.In(ctx))
	cw := write.With(ctx, ctxparm.Writer.In(ctx))
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
