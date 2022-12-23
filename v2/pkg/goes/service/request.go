// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package service

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/rpc"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
)

const RequestUsage = `
usage: {{.Command}} [<options>] <host> <command> [<args>]
Run command on <host>.
{{print .Flags}}`

type Connecter interface {
	Connect(context.Context, string) (net.Conn, error)
}

type Request struct {
	Connecter
}

func (req Request) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	in := fs.String("i", "", "Input FILE or '-' for STDIN.")
	usage := func() error {
		return template.Must(template.New("usage").
			Parse(RequestUsage[1:])).Execute(w, struct {
			Command string
			Flags   flags.Flags
		}{
			strings.Join(path, " "),
			fs,
		})
	}

	err := fs.Parse(args)
	if err == flags.ErrHelp {
		return usage()
	} else if err != nil {
		return err
	}

	if args = fs.Args(); len(args) == 0 {
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

	if path[1] == "complete" || path[1] == "help" {
		args = append([]string{path[1]}, args...)
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
		fmt.Fprint(w, res)
	}
	return err
}
