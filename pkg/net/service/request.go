// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
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

	"github.com/platinasystems/goes/v2/pkg/flag"
	"github.com/platinasystems/goes/v2/pkg/log/style"
)

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
	const usage = `{{$path := join .Path " "}}{{/*
*/}}usage: {{$path}} [<options>] <host> <command> [<args>]
Run command on <host>.
{{SprintDefault .Flags}}`
	cmd := path[len(path)-1]
	fs := flag.NewSilentFlagSet(cmd)
	in := fs.String("i", "", "Input FILE or '-' for STDIN.")
	if flag.Search[bool]("complete") {
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if flag.Search[bool]("help", fs) {
		return style.Usage(usage, struct {
			Path  []string
			Flags *flag.FlagSet
		}{path, fs})
	}
	args = fs.Args()
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
		fmt.Fprint(w, res)
	}
	return err
}
