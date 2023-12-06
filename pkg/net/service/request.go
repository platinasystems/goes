// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package service

import (
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/rpc"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

type Connecter interface {
	Connect(context.Context, string) (net.Conn, error)
}

type Request struct {
	Connecter
}

const RequestUsageTemplate = `
usage: {{.Path}} [<options>] <host> <command> [<args>]
Run command on <host>.
{{.Flag}}`

func RequestUsageData(ctx context.Context) any {
	return struct{ Path, Flag string }{
		strings.Join(ctxparm.Strings.In(ctx), " "),
		ctxparm.SprintFlagsIn(ctx),
	}
}

func (req Request) Func(ctx context.Context, args ...string) error {
	path := ctxparm.Strings.In(ctx)
	cmd := path[len(path)-1]
	if *complete.Help {
		return nil
	}
	ctx, flags := ctxparm.NewFlagsIn(ctx, cmd)
	in := flags.String("i", "", "Input FILE or '-' for STDIN.")
	err := flags.Parse(args)
	if err != nil {
		return err
	}
	if *usage.Help {
		return usage.Error(RequestUsageTemplate[1:],
			RequestUsageData(ctx))
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
		fmt.Fprint(ctxparm.Writer.In(ctx), res)
	}
	return err
}
