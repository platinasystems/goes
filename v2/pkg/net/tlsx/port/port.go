// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package port

import (
	"context"
	"net"

	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var GoResolve = &net.Resolver{PreferGo: true}

var (
	Exchange = cache.NewContext[int](port{"tlsx-exchange", 8003}.load)
	Registry = cache.NewContext[int](port{"tlsx-registry", 8004}.load)
	RPC      = cache.NewContext[int](port{"tlsx-rpc", 8005}.load)
)

type port struct {
	name   string
	number int
}

func (port port) load(ctx context.Context, p *int) error {
	if v, err := GoResolve.LookupPort(ctx, "tcp", port.name); err == nil {
		*p = v
	} else {
		*p = port.number
	}
	return nil
}
