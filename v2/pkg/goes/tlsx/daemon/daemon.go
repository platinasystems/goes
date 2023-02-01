// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package daemon

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os/exec"
	"sync"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/exchange"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/registry"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/service"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/tap"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const DaemonUsage = `
usage: {{.}} start <routines> [<command> [<args>]]
Ordered start of one or more of these daemon go-routines,

  exchange [-b [-l <prefix>]]
  registry
  rpc
  tap [-p <prefix>] [-u <unit>] [-x <exchange>]

e.g. an exchange VPN,

  {{.}} start registry exchange -b -l <prefix>

an exchange service,

  {{.}} start tap -p <prefix> -x <exchange> <command> [<args>]

an exchange node providing non-blocking (exchange) and blocking (rpc) service,

  {{.}} start tap -x <exchange> exchange rpc
`

var (
	Exchange exchange.T
	Tap      tap.T
)

func Func(
	ctx context.Context,
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
			"exchange",
			"registry",
			"rpc",
			"tap",
			"-b", // bridge
			"-l", // leasing
			"-u", // unit
			"-p", // prefix
		})
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = path[:len(path)-1]
		return usage()
	}

	if !program.IsKoApp() {
		style.System()
	}

	cctx, cancel := context.WithCancel(ctx)
	for ctx.Err() == nil && len(args) > 0 {
		switch args[0] {
		case "exchange":
			args, err = Exchange.Configure(args[1:])
			if err != nil {
				err = fmt.Errorf("exchange: %w", err)
				cancel()
			} else {
				style.Note("go exchange...")
				wg.Add(1)
				go Exchange.Routine(cctx, &wg)
			}
		case "registry":
			args = args[1:]
			style.Note("go registry...")
			wg.Add(1)
			go registry.Routine(cctx, &wg)
		case "rpc":
			style.Note("rpc...")
			args = args[1:]
			wg.Add(1)
			go service.Routine(ctx, &wg, Exchange.Selection)
		case "tap":
			args, err = Tap.Configure(ctx, args[1:])
			if err != nil {
				err = fmt.Errorf("tap: %w", err)
				cancel()
			} else {
				style.Note("go tap...")
				wg.Add(1)
				go Tap.Routine(cctx, &wg)
			}
		default:
			err = exec.CommandContext(cctx, args[0], args[1:]...).
				Run()
			if err != nil {
				err = fmt.Errorf("%v %w", args, err)
			}
			cancel()
		}
	}
	if err == flag.ErrHelp {
		err = nil
		usage()
	}
	return
}
