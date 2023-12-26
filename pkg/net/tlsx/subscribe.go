// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"errors"

	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/goes"
)

const SubscribeUsage = `
usage: {{branch .}} <name>[@<address>][:<port>]
Register with exchange.`

func Subscribe(ctx context.Context, args []string) error {
	if goes.ContextComplete(ctx) {
		return nil
	}
	if goes.ContextHelp(ctx) {
		return goes.Usage(ctx, SubscribeUsage)
	}
	if len(args) == 0 {
		return ErrIncomplete
	}
	data, err := Selfie.MarshalPEM()
	if err != nil {
		return egress.Mark(err)
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

	x := new(xcert.X509)
	if err = x.UnmarshalText(buf.Bytes()); err != nil {
		return egress.Mark(err)
	}
	Subscriptions().Append(x)
	return nil
}
