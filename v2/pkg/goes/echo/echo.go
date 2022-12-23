// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package echo

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/context/write"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
)

func Func(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	esc := fs.Bool("e", false, "Interpret escapes.")
	nonl := fs.Bool("n", false, "Without trailing newline.")
	usage := func() error {
		return template.Must(template.New("usage").Parse(`
usage: {{.Command}} [<options>] [<strings>]
Print string(s) to standard output.
{{print .Flags}}`[1:])).Execute(w, struct {
			Command string
			Flags   flags.Flags
		}{
			strings.Join(path, " "),
			fs,
		})
	}
	switch path[1] {
	case "complete":
		complete.Last(w, args, fs.FlagSet)
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}
	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
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
