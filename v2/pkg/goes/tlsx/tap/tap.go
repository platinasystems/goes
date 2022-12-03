// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build linux

package tap

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

const Key = "tap"

var (
	ErrNoExchange     = errors.New("missing <exchange>")
	ErrUnexpectedArgs = errors.New("unexpected argument(s)")
)

func Daemon(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	uflag := flag.Uint("u", 0, "unit")
	usage := func() error {
		return template.Must(template.New("usage").Parse(`
usage: {{.Command}} [<options>] <exchange>
Exchange tunnel.
{{print .Flags}}`[1:])).Execute(w, struct {
			Command string
			flags.Flags
		}{
			strings.Join(path, " "),
			fs,
		})
	}
	switch path[1] {
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	case "complete":
		complete.Last(w, args, fs.FlagSet,
			certs.Self.DNSNames(),
			certs.Subscriptions.Names())
		return nil
	}
	if err := fs.Parse(args); err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}
	args = fs.Args()
	switch len(args) {
	case 0:
		return ErrNoExchange
	case 1:
	default:
		return fmt.Errorf("%w: %v", ErrUnexpectedArgs, args[1:])
	}
	return tlsx.Tap(ctx, args[0], *uflag)
}
