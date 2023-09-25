// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package tlsx

import "net"

var (
	Dialer net.Dialer
	Dial   = Dialer.DialContext
)
