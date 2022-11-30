// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/rpc"
	"strings"
	"text/template"

	"github.com/platinasystems/goes/v2/pkg/flag/flags"
	"github.com/platinasystems/goes/v2/pkg/os/program"
)

type Ipcer interface {
	Address() (string, error)
	Listen() (net.Listener, error)
	Network() string
	String() string
}

type Ipc struct{ Ipcer }

func (ipc Ipc) Func(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path []string,
	args ...string,
) error {
	fs := flags.New()
	in := fs.String("i", "", "Input FILE or '-' for STDIN.")
	usage := func() error {
		return template.Must(template.New("usage").Parse(`
usage: {{.Command}} [<options>] <command> [<args>]
Run command through IPC server.
{{print .Flags}}`[1:])).Execute(w, struct {
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
	args = fs.Args()
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
	conn, err := ipc.Dial(ctx)
	if err != nil {
		return err
	}
	var res string
	c := rpc.NewClient(conn)
	call := c.Go("Service.Select", args, &res, nil)
	select {
	case <-call.Done:
		err = call.Error
		c.Close()
	case <-ctx.Done():
		err = ctx.Err()
		c.Close()
		<-call.Done
	}
	if len(res) > 0 {
		fmt.Fprint(w, res)
	}
	return err
}

// Dial the address returned from Ipc.Adress().
func (ipc Ipc) Dial(ctx context.Context) (net.Conn, error) {
	address, err := ipc.Address()
	if err != nil {
		return nil, err
	}
	return new(net.Dialer).DialContext(ctx, ipc.Network(), address)
}

func (ipc Ipc) String() string {
	address, err := ipc.Address()
	if err != nil {
		return fmt.Sprint(ipc.Network(), "://", err)
	}
	return fmt.Sprint(ipc.Network(), "://", address)
}

func join(prefix string, args []any) string {
	return fmt.Sprint(prefix, program.Base(), fmt.Sprint(args...))
}
