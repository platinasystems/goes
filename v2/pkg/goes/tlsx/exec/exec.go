// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"text/template"

	"github.com/creack/pty"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	ex := path[len(path)-1]
	usage := func() error {
		text := map[bool]string{
			true: `
usage: {{.Command}} [<options>] <exchange> <command> [<args>]
Remote execution.
{{print .Flags}}`,
			false: `
usage: {{.Command}} [<options>] <command> [<args>]
Remote execution.
{{print .Flags}}`,
		}[ex == "exec" || ex == ""]
		return template.Must(template.New("usage").Parse(text[1:])).
			Execute(w, struct {
				Command string
				Flags   flags.Flags
			}{
				strings.Join(path, " "),
				fs,
			})
	}
	iflag := fs.String("i", "", "Input FILE or '-' for STDIN.")
	tflag := fs.Bool("t", false, "Allocate a pseudo-TTY.")
	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}
	args = fs.Args()
	if ex == "exec" {
		if len(args) > 0 {
			ex = args[0]
			args = args[1:]
		} else {
			ex = ""
		}
	}
	switch path[1] {
	case "complete":
		if len(ex) == 0 {
			complete.Last(w, args, fs.FlagSet,
				certs.Self.DNSNames(),
				certs.Subscriptions.Names())
			return nil
		}
		args = append([]string{path[1]}, args...)
	case "help":
		if len(ex) == 0 {
			copy(path[1:], path[2:])
			path = path[:len(path)-1]
			return usage()
		}
		args = append([]string{path[1]}, args...)
	}
	if len(ex) == 0 {
		return errors.New("no <exchange>")
	}

	tlsc, err := tlsx.DialAndHandshake(ctx, ex)
	if err != nil {
		return err
	}
	defer tlsc.Close()
	var anyargs []any
	if *tflag {
		if ws, err := pty.GetsizeFull(os.Stdin); err == nil {
			anyargs = append(anyargs,
				"pty", ws.Rows, ws.Cols, ws.X, ws.Y)
		} else {
			return err
		}
	} else if len(*iflag) == 0 {
		r = nil
	} else if *iflag == "-" {
	} else if f, err := os.Open(*iflag); err == nil {
		defer f.Close()
		r = f
	} else {
		return err
	}
	for _, arg := range args {
		anyargs = append(anyargs, arg)
	}
	if err = tlsx.Req(ctx, tlsc, r, w, anyargs...); err != nil {
		if err.Error() == tlsx.ErrExit.Error() {
			err = nil
		} else {
			err = fmt.Errorf("%s: %w", ex, err)
		}
	}
	return err
}
