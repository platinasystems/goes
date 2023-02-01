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
	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var cut = slice.Cut[string]

const Usage = `
usage: {{.Command}} [<options>] <exchange> <command> [<args>]
Remote execution.
{{print .Flags}}`

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	usage := func() error {
		return template.Must(template.New("usage").
			Parse(Usage[1:])).Execute(w, struct {
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
	switch path[1] {
	case "complete":
		if len(args) <= 1 {
			complete.Last(w, args, fs.FlagSet,
				certs.Self.Name(),
				certs.Subscriptions.Names())
			return nil
		}
		args = append([]string{path[1]}, args...)
	case "help":
		if len(args) == 0 {
			path = cut(path, 1, 1)
			return usage()
		}
		args = append([]string{path[1]}, args...)
	}

	if len(args) == 0 {
		return errors.New("no <exchange>")
	}

	ex := args[0]
	if args = args[1:]; len(args) == 0 {
		return errors.New("no <command>")
	}

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

	conn, err := tlsx.Exchange.Connect(ctx, ex)
	if err != nil {
		return err
	}

	if err = tlsx.Exec(ctx, conn, r, w, anyargs...); err != nil {
		err = fmt.Errorf("%s: %w", ex, err)
	}
	return err
}

func IPC(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `
usage: {{.}} <command> [<args>]
Daemon IPC.`
	switch path[1] {
	case "complete":
		return nil
	case "help":
		return template.Must(template.New("usage").
			Parse(usage[1:])).
			Execute(w, strings.Join(
				cut(path, 1, 1), " ",
			))
	}
	last := len(path) - 1
	cmd := path[last]
	path[last] = "exec"
	args = append([]string{certs.Self.Value().X509.Name, cmd},
		args...)
	return Func(ctx, r, w, path, args...)
}
