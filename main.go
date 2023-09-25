// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	"github.com/platinasystems/goes/v2/pkg/coreutils"
	"github.com/platinasystems/goes/v2/pkg/goes"
	net_tools "github.com/platinasystems/goes/v2/pkg/net/net-tools"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
	"github.com/platinasystems/goes/v2/pkg/net/udpecho"
)

var root = map[string]any{
	"daemon": map[string]any{
		"tlsx": tlsx.Daemons,
		"udp": map[string]any{
			"echo": udpecho.Reply,
		},
	},
	"show": map[string]any{
		"tlsx": tlsx.Show,
	},
	"udp": map[string]any{
		"ping": udpecho.Ping,
	},
}

func main() {
	goes.Merge(root, coreutils.Root)
	goes.Merge(root, net_tools.Root)
	goes.Root = root
	goes.Main()
}
