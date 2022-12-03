// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"io"
	"net"
	"strings"
	"sync"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/address"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

func Daemon(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	rflag := fs.String("r", ":8004", "Registry [<address>]:<port>.")
	xflag := fs.String("x", ":8003", "Exchange [<address>]:<port>.")
	usage := func() error {
		return template.Must(template.New("usage").Parse(`
usage: {{.Command}} [<options>]
Start TLS exchange service.
{{print .Flags}}`[1:])).Execute(w, struct {
			Command string
			flags.Flags
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
	rln, err := net.Listen("tcp", *rflag)
	if err != nil {
		return err
	}
	xln, err := net.Listen("tcp", *xflag)
	if err != nil {
		rln.Close()
		return err
	}
	address.Store(certs.Self.SKI(), xln.Addr().String())
	var wg sync.WaitGroup
	wg.Add(1)
	go tlsx.Registry(ctx, &wg, rln)
	wg.Add(1)
	go tlsx.Exchange(ctx, &wg, xln, Service)
	wg.Wait()
	return nil
}
