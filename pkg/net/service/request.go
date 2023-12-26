// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package service

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"

	"github.com/platinasystems/goes/v2/pkg/goes"
)

type Connecter interface {
	Connect(context.Context, string) (net.Conn, error)
}

type Request struct {
	Connecter
}

const RequestUsage = `
usage: {{branch .}} [<options>] <host> <command> [<args>]
Run command on <host>.
{{flags .}}`

func (req Request) Func(ctx context.Context, args []string) error {
	var flag flag.FlagSet
	ctx = goes.FlagsContext(ctx, &flags)
	branch := goes.ContextBranch(ctx)
	cmd := branch[len(branch)-1]
	if goes.ContextComplete(ctx) {
		return nil
	}
	in := flags.String("i", "", "Input FILE or '-' for STDIN.")
	ctx, err := goes.ParseFlagsContext(ctx, args)
	if err != nil {
		return err
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, RequestUsage)
	}
	args = flags.Args()
	if len(args) == 0 {
		return errors.New("missing <host>")
	}
	host := args[0]
	if args = args[1:]; len(args) == 0 {
		return errors.New("missing <command>")
	}
	if len(*in) > 0 {
		var b []byte
		if *in == "-" {
			b, err = ioutil.ReadAll(r)
		} else {
			b, err = ioutil.ReadFile(*in)
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
		fmt.Fprint(goes.ContextStdout(ctx), res)
	}
	return err
}
