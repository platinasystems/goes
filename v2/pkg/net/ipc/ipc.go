// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/rpc"

	"github.com/platinasystems/goes/v2/pkg/goes/selection"
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
	path selection.Path,
	args ...string,
) error {
	var res string
	name := path[len(path)-1]
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	in := fs.String("i", "", "Input FILE or '-' for STDIN.")
	fs.Usage = func() {
		path.Usage(w, "[-i <FILE|->] COMMAND [OPTION]... [ARG]...\n")
	}
	if path.HasComplete() {
		args = append([]string{"complete"}, args...)
		return nil
	}
	if path.HasHelp() {
		args = append([]string{"help"}, args...)
	}
	err := fs.Parse(args)
	if err != nil {
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
	conn, err := ipc.Dial(ctx)
	if err != nil {
		return err
	}
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
