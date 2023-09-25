// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This provides an approximation of the Unix net-tools command.
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes"
	net_tools "github.com/platinasystems/goes/v2/pkg/net/net-tools"
)

func main() {
	goes.Exec(net_tools.Netstat)
}
