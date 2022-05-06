// Copyright © 2016-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"fmt"
	"os"
)

func Cat(ctx context.Context, args ...string) error {
	o := OutputOf(ctx)
	switch Preemption(ctx) {
	case "":
	case "complete":
		for _, s := range CompleteFiles(args) {
			o.Println(s)
		}
		return nil
	case "help":
		Usage(ctx, "[FILE|-]...\n",
			"Concatenate FILE(s) or standard in (-) to output.")
		fallthrough
	default:
		return nil
	}
	if len(args) == 0 {
		args = append(args, "-")
	}
	for _, fn := range args {
		if fn == "-" {
			o.ReadFrom(InputOf(ctx))
		} else if fi, err := os.Stat(fn); err != nil {
			return fmt.Errorf("%s: %w", fn, err)
		} else if fi.IsDir() {
			return fmt.Errorf("%s: Is a directory", fn)
		} else if r, err := os.Open(fn); err != nil {
			return err
		} else {
			o.ReadFrom(r)
			r.Close()
		}
	}
	return ctx.Err()
}
