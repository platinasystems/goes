// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipc

import (
	"net"
)

// address[:port]
type Tcp string

func NewTcp(ap string) Ipc { return Ipc{Tcp(ap)} }

func (tcp Tcp) Address() (string, error) { return string(tcp), nil }

func (Tcp) Network() string { return "tcp" }

func (tcp Tcp) Listen() (net.Listener, error) {
	return net.Listen("tcp", string(tcp))
}
