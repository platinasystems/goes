// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package ip6

import (
	"net"
)

func (a *Address) String() string {
	return (net.IP)(a[:]).String()
}
