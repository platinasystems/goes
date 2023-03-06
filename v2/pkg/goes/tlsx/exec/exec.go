// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
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
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/greet"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

const Usage = `
usage: {{.Command}} [<options>] <exchange> <command> [<args>]
Remote execution.
{{print .Flags}}
<exchange>
	<name>[@<dns|ip4|\[ip6\]>][:<port>]
`

var ErrIncomplete = errors.New("incomplete")

var cut = slice.Cut[string]

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	iflag := fs.String("i", "", "Input FILE or '-' for STDIN.")
	tflag := fs.Bool("t", false, "Allocate a pseudo-TTY.")

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
			if name, err := certs.Self.Name(); err == nil {
				complete.Last(w, args, fs.FlagSet, name,
					certs.Subscriptions.Names())
			}
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
		return ErrIncomplete
	}

	ex := args[0]
	if args = args[1:]; len(args) == 0 {
		return ErrIncomplete
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

	conn, err := tlsx.Connect(ctx, ex)
	if err != nil {
		return err
	}

	tlsc, err := greet.Server(ctx, ex, conn)
	if err != nil {
		return err
	}

	if err = tlsx.Exec(ctx, tlsc, r, w, anyargs...); err != nil {
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
usage: {{.}} [-i <file>|-] [<args>]
Daemon IPC.
`
	switch path[1] {
	case "complete":
		return nil
	case "help":
		return template.Must(template.New("usage").
			Parse(usage[1:])).
			Execute(w, strings.Join(cut(path, 1, 1), " "))
	}
	ipc := struct{ path, args []string }{
		path: []string{path[0], "exec"},
		args: make([]string, 0, len(path)+len(args)),
	}
	name, err := certs.Self.Name()
	if err != nil {
		return egress.Marked(err)
	}
	ipc.args = append(ipc.args, name)
	ipc.args = append(ipc.args, path[1:]...)
	ipc.args = append(ipc.args, args...)
	return Func(ctx, r, w, ipc.path, ipc.args...)
}
