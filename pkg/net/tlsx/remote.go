// Copyright © 2022 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import (
	"net"
	"net/netip"
	"path/filepath"
)

func remoteAddr(conn net.Conn) string {
	na := conn.RemoteAddr()
	if len(na.String()) == 0 {
		na = conn.LocalAddr()
	}
	if na.Network() == "unix" {
		return filepath.Base(na.String())
	} else if ap, err := netip.ParseAddrPort(na.String()); err == nil {
		return ap.Addr().String()
	}
	return ""
}
