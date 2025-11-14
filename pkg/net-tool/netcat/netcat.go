// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netcat

import (
	"context"
	"flag"
	"io"
	"net"
	"net/netip"
	"os"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
	"github.com/platinasystems/goes/v2/pkg/xmain"
	"github.com/platinasystems/goes/v2/pkg/xnet/xdns/xdnsdoh"
)

func NetCat(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <host> <port>
Pipe stdin/out with TCP connection to the named or numbered host/port.

{{flags .}}`)

	err := xflag.Labels{
		xmain.ConfigFlag,
	}.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("host")
	} else if len(args) == 1 {
		return xerrors.Incomplete("port")
	}

	addrs, err := xdnsdoh.LookupNetIP(ctx, args[0])
	if err != nil {
		return err
	}

	nw := "tcp"
	if addrs[0].Is4() {
		nw = "tcp4"
	} else if addrs[0].Is6() {
		nw = "tcp6"
	}

	port, err := net.DefaultResolver.LookupPort(ctx, nw, args[1])
	if err != nil {
		return err
	}

	ap := netip.AddrPortFrom(addrs[0], uint16(port))

	var d net.Dialer
	conn, err := d.DialContext(ctx, nw, ap.String())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Go(func() { io.Copy(os.Stdout, conn) })
	wg.Go(func() { io.Copy(conn, os.Stdin) })
	<-ctx.Done()
	conn.Close()
	wg.Wait()
	return nil
}
