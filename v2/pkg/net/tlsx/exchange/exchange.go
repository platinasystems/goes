// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exchange

import (
	"context"
	"errors"
	"html/template"
	"net/netip"
	"strings"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/container/slice"
	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/bridge"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/lease"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/service"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

const Usage = `
usage: {{.}} [-p <port>] <prefix>
Start exchange at <port> (default 8003).
`

var ErrIncomplete = errors.New("incomplete")

func Daemon(
	ctx context.Context,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	port := fs.Uint("p", 8003, "service port")
	fs.BoolVar(&certs.Restricted, "r", false, "restrict clients to self")

	usage := func() error {
		return template.Must(template.New("usage").Parse(Usage[1:])).
			Execute(style.Plain.Notice.Writer(),
				strings.Join(path, " "))
	}

	switch path[1] {
	case "complete":
		return nil
	case "help":
		copy(path[1:], path[2:])
		path = slice.Cut[string](path, 1, 1)
		path[1] = "start" // replace "daemon"
		return usage()
	default:
		if !program.IsKoApp() {
			style.System()
		}
	}

	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}
	args = fs.Args()

	if len(args) == 0 {
		return ErrIncomplete
	}

	if prefix, err := netip.ParsePrefix(args[0]); err != nil {
		return err
	} else {
		lease.Enable(prefix)
	}

	var wg sync.WaitGroup
	defer wg.Wait()

	wg.Add(1)
	go bridge.Routine(ctx, &wg)

	wg.Add(1)
	go service.Routine(ctx, &wg, *port)

	return nil
}
