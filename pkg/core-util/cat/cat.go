// Copyright © 2022-2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package cat

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xflag"
)

const CatUsage = `
usage: {{.Name}} [file(s)|-]
Concatenate file(s) or stdin (-) to stdout.
`

func Cat(ctx context.Context, complete bool, args []string) error {
	xflag.TemplateUsage(CatUsage)
	err := flag.CommandLine.Parse(args)
	if err != nil {
		return err
	} else if args = flag.Args(); complete {
		// FIXME complete file(s)
		return nil
	} else if len(args) == 0 {
		args = append(args, "-")
	}

	for _, fn := range args {
		if fn == "-" {
			if _, err := io.Copy(os.Stdout, os.Stdin); err != nil {
				return err
			}
		} else if fi, err := os.Stat(fn); err == nil && fi.IsDir() {
			return fmt.Errorf("%s: is a directory", fn)
		} else if f, err := os.Open(fn); err == nil {
			_, err = io.Copy(os.Stdout, f)
			f.Close()
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("%s: %w", fn, errors.Unwrap(err))
		}
	}
	return ctx.Err()
}
