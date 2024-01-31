// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"errors"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
)

func Subscribe(ctx context.Context, args []string) error {
	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, `
usage: {{branch .}} <name>[@<address>][:<port>]
Register with exchange.`)
	}
	if len(args) == 0 {
		return ErrIncomplete
	}
	err := InitSelf()
	if err != nil {
		return err
	}
	data, err := Self.MarshalPEM()
	if err != nil {
		return err
	}
	conn, err := Connect(ctx, args[0])
	if err != nil {
		return egress.Mark(err)
	}
	defer conn.Close()

	r := bytes.NewBuffer(data)
	buf := new(bytes.Buffer)

	err = Exec(ctx, conn, r, buf, "subscribe", "-")
	if err != nil {
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		return egress.Mark(err)
	}

	x := new(X509)
	if err = x.UnmarshalText(buf.Bytes()); err != nil {
		return egress.Mark(err)
	}
	Subscriptions.Append(x)
	return nil
}
