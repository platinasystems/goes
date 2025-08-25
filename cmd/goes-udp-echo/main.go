// Copyright © 2025 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

// This provides a UDP echo server and ping tool.
package main

import (
	"github.com/platinasystems/goes/v2/pkg/goes"
	udp_echo "github.com/platinasystems/goes/v2/pkg/net-tool/udp-echo"
)

func init() { goes.Install(udp_echo.Features) }

func main() { goes.Main() }
