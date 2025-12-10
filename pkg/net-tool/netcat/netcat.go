// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netcat

import (
	"context"
	"flag"
	"io"
	"net"
	"os"
	"sync"
	"time"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
)

const NetCatUsage = `
usage: {{.Name}} [flags] <host> <port>
Pipe stdin/out with TCP connection to the named or numbered host/port.

{{flags .}}`

var NetCatFlags = xflag.Labels{
	xmain.ConfigFlag,
	xdnsdoh.ConfigFlag,
}

func NetCat(ctx context.Context, args []string) error {
	xflag.TemplateUsage(NetCatUsage)
	err := NetCatFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("host")
	} else if len(args) == 1 {
		return xerrors.Incomplete("port")
	}

	resolver, err := xdnsdoh.Resolver()
	if err != nil {
		return err
	}
	d := &net.Dialer{
		Timeout:   30 * time.Second,
		KeepAlive: 30 * time.Second,
		Resolver:  resolver,
	}

	address := net.JoinHostPort(xdnsdoh.FQDN(args[0]), args[1])
	conn, err := d.DialContext(ctx, "tcp", address)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	defer wg.Wait()
	defer conn.Close()
	wg.Go(func() { io.Copy(os.Stdout, conn) })
	wg.Go(func() { io.Copy(conn, os.Stdin) })
	<-ctx.Done()
	return nil
}
