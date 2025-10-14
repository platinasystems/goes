// Copyright © 2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xdns

import (
	"context"
	"net"
)

var Resolver = &net.Resolver{
	PreferGo:     true,
	StrictErrors: true,
}

var Dialer = &net.Dialer{
	Resolver: Resolver,
}

var DialContext = Dialer.DialContext

// An Asker sends a buffered query through an associated connection.
// If sucessful, it returns the raw binary response within the same,
// probably expanded buffer.
type Asker interface {
	Ask(context.Context, []byte) ([]byte, error)
}
