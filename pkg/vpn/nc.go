// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vpn

import (
	"context"
	"flag"
	"io"
	"net"
	"os"
	"strconv"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

func NetCat(ctx context.Context, args []string) error {
	xflag.TemplateUsage(`
usage: {{.Name}} [flags] <guest> <port>
Pipe stdin/out with TCP connection to numbered port of named guest.

{{flags .}}`)

	err := RestFlags.Define()
	if err != nil {
		return err
	} else if err = flag.CommandLine.Parse(args); err != nil {
		return err
	} else if args = flag.Args(); len(args) == 0 {
		return xerrors.Incomplete("guest")
	} else if len(args) == 1 {
		return xerrors.Incomplete("port")
	} else if err = restInit(); err != nil {
		return err
	}

	sub, err := RestWhoisReq(ctx, args[0])
	if err != nil {
		return xerrors.Label(err, "guest")
	}
	port64, err := strconv.ParseInt(args[1], 10, 16)
	if err != nil {
		return xerrors.Label(err, "port")
	}

	rest.CloseIdleConnections()

	conn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   net.IP(sub.Addr.AsSlice()),
		Port: int(port64),
	})
	if err != nil {
		return err
	}

	wg.Go(func() { io.Copy(os.Stdout, conn) })
	wg.Go(func() { io.Copy(conn, os.Stdin) })
	<-ctx.Done()
	conn.Close()
	wg.Wait()
	return nil
}
