// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemon

import (
	"context"
	"flag"
	"io"
	"os/exec"
	"sync"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/exchange"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/registry"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/service"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/tap"
)

const DaemonUsage = `
usage: {{.}} start <routines> [<command> [<args>]]
Ordered start of one or more of these daemon go-routines,

  exchange [bridge [leasing <network> <base>]]
  registry
  rpc
  tap [-name <name>] [-prefix <prefix>] [-unix #] <exchange>

e.g. an exchange VPN,

  {{.}} start registry exchange bridge leasing <network> <base>

a named service w/in the exchange,

  {{.}} start tap -name <name> -prefix <prefix> <exchange> <command> [<args>]

an exchange node providing non-blocking (exchange) and blocking (rpc) service,

  {{.}} start tap <exchange> exchange rpc
`

var (
	Exchange exchange.T
	Tap      tap.T
)

func Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) (err error) {
	var wg sync.WaitGroup
	defer wg.Wait()

	usage := func() error {
		return template.Must(template.New("usage").
			Parse(DaemonUsage[1:])).
			Execute(w, path[0])
	}

	switch path[1] {
	case "complete":
		complete.Last(w, args, []string{
			"bridge",
			"exchange",
			"leasing",
			"registry",
			"rpc",
			"tap",
			"-unit",
			"-prefix",
		})
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}

	cctx, cancel := context.WithCancel(ctx)
	for ctx.Err() == nil && len(args) > 0 {
		switch args[0] {
		case "exchange":
			args, err = Exchange.Configure(args[1:])
			if err != nil {
				cancel()
			} else {
				wg.Add(1)
				go Exchange.Routine(cctx, &wg)
			}
		case "registry":
			args = args[1:]
			wg.Add(1)
			go registry.Routine(cctx, &wg)
		case "rpc":
			args = args[1:]
			wg.Add(1)
			go service.Routine(ctx, &wg)
		case "tap":
			args, err = Tap.Configure(ctx, args[1:])
			if err != nil {
				cancel()
			} else {
				wg.Add(1)
				go Tap.Routine(cctx, &wg)
			}
		default:
			err = exec.CommandContext(cctx, args[0], args[1:]...).
				Run()
			cancel()
		}
	}
	if err == flag.ErrHelp {
		err = nil
		usage()
	}
	return
}
