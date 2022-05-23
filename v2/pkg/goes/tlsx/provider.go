// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"sync"

	"github.com/platinasystems/goes/v2/pkg/goes/cat"
	"github.com/platinasystems/goes/v2/pkg/goes/command"
	"github.com/platinasystems/goes/v2/pkg/goes/echo"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/authorized"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var Service = selection.Map{
	"authorize": AuthorizeOrRevoke,
	"cat":       cat.Func,
	"command":   command.Func,
	"echo":      echo.Func,
	"revoke":    AuthorizeOrRevoke,
}

func Host(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	const synopsis = `
Service consumer requests through subscribed exchange(s).`
	fs := flag.NewFlagSet("host", flag.ContinueOnError)
	fs.Usage = func() {
		path.Usage(w, synopsis)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	if args = fs.Args(); len(args) == 0 {
		args = append(args, "")
	}
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}
	var wg sync.WaitGroup
	for _, ski := range certs.Exchanges.SKIs() {
		wg.Add(1)
		go tlsx.Accept(ctx, &wg, ski, Service.Select)
	}
	wg.Wait()
	return nil
}

func AuthorizeOrRevoke(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	op := path[len(path)-1]
	val := op == "authorize"
	fs := flag.NewFlagSet(op, flag.ContinueOnError)
	fs.Usage = func() {
		syn := "Authorize host consumer."
		if !val {
			syn = "Revoke consumer authorization."
		}
		path.Usage(w, "[<subject-key-id(s)>]\n", syn)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	args = fs.Args()
	if path.HasComplete() {
		return nil
	}
	if path.HasHelp() {
		fs.Usage()
		return nil
	}
	if len(args) == 0 {
		authorized.Range(func(ski string) bool {
			fmt.Fprintln(w, ski)
			return true
		})
		return nil
	}
	if _, err = hex.DecodeString(args[0]); err != nil {
		return err
	}
	return authorized.Store(args[0], val)
}
