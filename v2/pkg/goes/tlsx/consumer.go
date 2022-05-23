// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"context"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/goes/complete"
	"github.com/platinasystems/goes/v2/pkg/goes/selection"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/alias"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var ErrMissingHostArg = errors.New("missing host <dns> or <subject-key-id>")

func Jump(
	ctx context.Context,
	r io.Reader,
	w io.Writer,
	path selection.Path,
	args ...string,
) error {
	fs := flag.NewFlagSet("jump", flag.ContinueOnError)
	ex := fs.String("x", certs.Exchanges.FirstDNS(),
		"Exchange <dns> or <subject-key-id>.")
	in := fs.String("i", "", "Input FILE or '-' for STDIN.")
	fs.Usage = func() {
		path.Usage(w, "[<options>] <host> [<request> [<args>]]\n",
			"Run request on host connected through exchange.\n",
			fs,
		)
	}
	err := fs.Parse(args)
	if err != nil {
		return err
	}
	var host, ski string
	args = fs.Args()
	if path.HasComplete() {
		if len(args) < 2 {
			complete.Last(w, args, alias.Keys())
			return nil
		}
		host = args[0]
		args[0] = "complete"
	} else if path.HasHelp() {
		if len(args) < 1 {
			fs.Usage()
			return nil
		}
		host = args[0]
		args[0] = "help"
	} else if len(args) > 0 {
		host = args[0]
		args = args[1:]
	} else {
		return ErrMissingHostArg
	}
	if _, err = hex.DecodeString(host); err == nil {
		ski = host
	} else if val, ok := alias.Load(host); ok {
		ski = val
	} else {
		return fmt.Errorf("%s: %w", host, alias.ErrNotFound)
	}
	if len(*in) == 0 {
		r = io.LimitReader(nil, 0)
	} else if *in == "-" {
		// Req with Stdin
	} else if f, err := os.Open(*in); err == nil {
		defer f.Close()
		r = f
	} else {
		return err
	}
	cn, err := tlsx.DialAndHandshake(ctx, *ex)
	if err != nil {
		return err
	}
	defer cn.Close()
	svc := new(strings.Builder)
	err = tlsx.Req(ctx, cn, nil, svc, "connect", ski)
	if err != nil {
		return err
	}
	return tlsx.Req(ctx, cn, r, w, args)
}
