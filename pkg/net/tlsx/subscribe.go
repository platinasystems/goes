// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"errors"
	"io"

	"github.com/platinasystems/goes/v2/pkg/context/help"
	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/text/complete"
)

func Subscribe(
	ctx context.Context,
	w io.Writer,
	path []string,
	args ...string,
) error {
	const usage = `usage: {{join . " "}} <name>[@<address>][:<port>]
Register with exchange.
`
	if complete.Parameter.Value(ctx) {
		return nil
	}
	if help.Parameter.Value(ctx) {
		return style.Usage(usage, path)
	}
	if len(args) == 0 {
		return ErrIncomplete
	}
	data, err := Selfie.MarshalPEM()
	if err != nil {
		return egress.Marked(err)
	}
	conn, err := Connect(ctx, args[0])
	if err != nil {
		return egress.Marked(err)
	}
	defer conn.Close()

	r := bytes.NewBuffer(data)
	buf := new(bytes.Buffer)

	err = Exec(ctx, conn, r, buf, "subscribe", "-")
	if err != nil {
		if errors.Is(err, context.Canceled) {
			err = nil
		}
		return egress.Marked(err)
	}

	x := new(xcert.X509)
	if err = x.UnmarshalText(buf.Bytes()); err != nil {
		return egress.Marked(err)
	}
	Subscriptions().Append(x)
	return nil
}
