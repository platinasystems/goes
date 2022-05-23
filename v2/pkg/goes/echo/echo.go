// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"

	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	fs := flag.NewFlagSet("echo", flag.ContinueOnError)
	esc := fs.Bool("e", false, "interpret escapes")
	nonl := fs.Bool("n", false, "without trailing newline")
	fs.Usage = func() {
		path.Usage(w, "[<options>] [<strings>]\n",
			"Print string(s) to standard output.\n",
			fs,
		)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	args = fs.Args()
	if path.HasComplete() {
		complete.Last(w, args, fs)
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}
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
