// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"errors"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/context/ctxparm"
	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/errors/usage"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

const SubscribeUsageTemplate = `
usage: {{.}} <name>[@<address>][:<port>]
Register with exchange.`

func SubscribeUsageData(ctx context.Context) any {
	return strings.Join(ctxparm.Strings.In(ctx), " ")
}

func Subscribe(ctx context.Context, args ...string) error {
	if *complete.Help {
		return nil
	}
	if *usage.Help {
		return usage.Error(SubscribeUsageTemplate[1:],
			SubscribeUsageData(ctx))
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
