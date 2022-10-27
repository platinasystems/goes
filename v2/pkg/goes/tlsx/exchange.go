// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"flag"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/ipc"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/alias"
)

func ApproveOrDeny(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) (err error) {
	op := path[len(path)-1]
	fs := flag.NewFlagSet(op, flag.ContinueOnError)
	xflag := fs.String("x", "", "Exchange <dns>:<port> (default IPC).")
	fs.Usage = func() {
		op_ := "Approve"
		if op != "approve" {
			op_ = "Deny"
		}
		path.Usage(w, "[<options>] [<subject-key-id(s)>]\n",
			op_, " client(s) subscription.\n",
			fs,
		)
	}
	if err = fs.Parse(args); err != nil {
		return
	}
	if path.HasComplete() {
		complete.Last(w, args, alias.Keys())
		return
	}
	if path.HasHelp() {
		fs.Usage()
		return
	}
	cn, err := tlsx.DialAndHandshake(ctx, *xflag)
	if err != nil {
		return err
	}
	defer cn.Close()
	return tlsx.Req(ctx, cn, nil, w, op, fs.Args())
}

func Clients(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) (err error) {
	fs := flag.NewFlagSet("certs", flag.ContinueOnError)
	xflag := fs.String("x", "", "Exchange <dns>:<port> (default IPC).")
	fs.Usage = func() {
		path.Usage(w, "[<options>]\n",
			"List certificates of exchange clients\n",
			fs,
		)
	}
	if err = fs.Parse(args); err != nil {
		return
	}
	if path.HasComplete() {
		return
	}
	if path.HasHelp() {
		fs.Usage()
		return
	}
	cn, err := tlsx.DialAndHandshake(ctx, *xflag)
	if err != nil {
		fmt.Fprintln(w, "greet failed:", err)
		return err
	}
	defer cn.Close()
	return tlsx.Req(ctx, cn, nil, w, "clients")
}

func Exchange(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	fs := flag.NewFlagSet("exchange", flag.ContinueOnError)
	rflag := fs.Uint("r", 0, "Registry port. (default IPC)")
	xflag := fs.Uint("x", 0, "Exchange port. (default IPC)")
	fs.Usage = func() {
		path.Usage(w, "[<options>]\n",
			"Start TLS exchange service.\n",
			fs,
		)
	}
	if path.HasComplete() {
		complete.Last(w, args, fs)
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	var xln, rln net.Listener
	if *xflag == 0 {
		xln, err = ipc.Exchange().Listen()
		if err == nil {
			rln, err = ipc.Registry().Listen()
		}
	} else {
		xln, err = net.Listen("tcp", fmt.Sprint(":", *xflag))
		if err == nil {
			rln, err = net.Listen("tcp", fmt.Sprint(":", *rflag))
		}
	}
	if err != nil {
		return err
	}
	var wg sync.WaitGroup
	wg.Add(1)
	go tlsx.Exchange(ctx, &wg, xln)
	wg.Add(1)
	go tlsx.Registry(ctx, &wg, rln, *xflag)
	wg.Wait()
	return nil
}
