// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"context"
	"fmt"
	"net"
)

type Ipcer interface {
	Address() (string, error)
	Listen() (net.Listener, error)
	Network() string
}

type Ipc struct{ Ipcer }

func (ipc Ipc) Dial(ctx context.Context) (net.Conn, error) {
	address, err := ipc.Address()
	if err != nil {
		return nil, err
	}
	return new(net.Dialer).DialContext(ctx, ipc.Network(), address)
}

func (ipc Ipc) String() string {
	address, err := ipc.Address()
	if err != nil {
		return fmt.Sprint(ipc.Network(), "://", err)
	}
	return fmt.Sprint(ipc.Network(), "://", address)
}
