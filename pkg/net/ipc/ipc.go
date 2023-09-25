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
	Err() error
	Listen() (net.Listener, error)
	net.Addr
}

type Ipc struct{ Ipcer }

func (ipc Ipc) Connect(ctx context.Context, _ string) (net.Conn, error) {
	address := ipc.String()
	if err := ipc.Err(); err != nil {
		return nil, err
	}
	return new(net.Dialer).DialContext(ctx, ipc.Network(), address)
}

func (ipc Ipc) Format(w fmt.State, verb rune) {
	address := ipc.String()
	if err := ipc.Err(); err != nil {
		fmt.Fprint(w, ipc.Network(), "://", err)
	} else {
		fmt.Fprint(w, ipc.Network(), "://", address)
	}
}
