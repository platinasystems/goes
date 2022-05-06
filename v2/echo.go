// Copyright © 2015-2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package goes

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
)

func Echo(ctx context.Context, args ...string) error {
	o := OutputOf(ctx)
	fs := flag.NewFlagSet("echo", flag.ContinueOnError)
	esc := fs.Bool("e", false, "interpret escapes")
	nonl := fs.Bool("n", false, "without trailing newline")
	fs.Usage = func() {
		Usage(ctx, "[OPTION]... STRING...\n",
			"Print STRING(s) to standard output.\n",
			"\n",
			fs)
	}
	switch Preemption(ctx) {
	case "":
	case "help":
		fs.Usage()
		fallthrough
	default:
		return nil
	}
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			err = nil
		}
		return err
	}
	args = fs.Args()
	for i, arg := range args {
		if i > 0 {
			o.Print(" ")
		}
		if *esc {
			text := []byte(fmt.Sprintf(`"%s"`, arg))
			err := json.Unmarshal(text, &arg)
			if err != nil {
				return err
			}
		}
		o.Print(arg)
	}
	if !*nonl {
		o.Println()
	}
	return ctx.Err()
}
