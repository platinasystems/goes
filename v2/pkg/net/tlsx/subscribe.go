// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"bytes"
	"context"
	"encoding/pem"
	"errors"

	"github.com/platinasystems/goes/v2/pkg/crypto/keycert"
	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/errors/suppress"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx/state/certs"
)

var ErrInvalidResponse = errors.New("invalid server response")
var Suppressed = []error{context.Canceled}

func Subscribe(ctx context.Context, ex string) error {
	data, err := certs.Self.MarshalPEM()
	if err != nil {
		return egress.Marked(err)
	}

	conn, err := Connect(ctx, ex)
	if err != nil {
		return egress.Marked(err)
	}
	defer conn.Close()

	r := bytes.NewBuffer(data)
	w := new(bytes.Buffer)

	err = Exec(ctx, conn, r, w, "subscribe", "-")
	if err != nil {
		return egress.Marked(suppress.Errors(err, Suppressed...))
	}

	x := new(keycert.X509)
	blk, _ := pem.Decode(w.Bytes())
	if blk == nil {
		return egress.Marked(ErrInvalidResponse)
	}
	if err := x.UnmarshalPEM(blk); err != nil {
		return egress.Marked(err)
	}
	return egress.Marked(certs.Subscriptions.Append(x))
}
