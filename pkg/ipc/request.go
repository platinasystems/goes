// Copyright © 2022-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"net/rpc"
	"os"

	"github.com/platinasystems/goes/v2/pkg/xerrors"
	"github.com/platinasystems/goes/v2/pkg/xflag"
)

type Connecter interface {
	Connect(context.Context, string) (net.Conn, error)
}

type Request struct {
	Connecter
}

func (req Request) Func(ctx context.Context, args []string) error {
	flags := flag.NewFlagSet("ioc", flag.ContinueOnError)
	xflag.UsageTemplate(flags, `
usage: {{.Name}} [flags] <host> <command> [<args>]
Run command on <host>.

{{flags .}}`)
	in := flags.String("i", "", "Input FILE or '-' for STDIN.")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	args = flags.Args()
	if len(args) == 0 {
		return xerrors.Incomplete("host")
	}
	host := args[0]
	if args = args[1:]; len(args) == 0 {
		return xerrors.Incomplete("command")
	}
	if len(*in) > 0 {
		var b []byte
		if *in == "-" {
			b, err = io.ReadAll(os.Stdin)
		} else {
			b, err = os.ReadFile(*in)
		}
		if err != nil {
			return err
		}
		args = append([]string{"-i", string(b)}, args...)
	}

	conn, err := req.Connect(ctx, host)
	if err != nil {
		return err
	}

	var res string
	cl := rpc.NewClient(conn)
	call := cl.Go("Service.Select", args, &res, nil)
	select {
	case <-call.Done:
		err = call.Error
		cl.Close()
	case <-ctx.Done():
		err = ctx.Err()
		cl.Close()
		<-call.Done
	}

	if len(res) > 0 {
		fmt.Print(res)
	}
	return err
}
