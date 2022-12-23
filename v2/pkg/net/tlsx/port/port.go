// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package port

import (
	"context"
	"net"
	"os/signal"

	"github.com/platinasystems/goes/v2/pkg/os/termination"
	"github.com/platinasystems/goes/v2/pkg/sync/cache"
)

var GoResolve = &net.Resolver{PreferGo: true}

var (
	Exchange = cache.NewReadOnly[int](port{"tlsx-exchange", 8003}.load)
	Registry = cache.NewReadOnly[int](port{"tlsx-registry", 8004}.load)
	RPC      = cache.NewReadOnly[int](port{"tlsx-rpc", 8005}.load)
)

var Map = map[string]any{
	"exchange": Exchange,
	"registry": Registry,
	"rpc":      RPC,
}

type port struct {
	name   string
	number int
}

func (port port) load(p *int) error {
	ctx, stop := signal.NotifyContext(context.Background(),
		termination.Signals...)
	defer stop()

	if v, err := GoResolve.LookupPort(ctx, "tcp", port.name); err == nil {
		*p = v
	} else {
		*p = port.number
	}
	return nil
}
