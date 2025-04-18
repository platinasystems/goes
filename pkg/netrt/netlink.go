// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netrt

import "github.com/platinasystems/goes/v2/pkg/netlink"

type nl struct {
	seq  uint32
	sock *netlink.NL
}

func (nl nl) Close() error {
	return nl.sock.Close()
}
