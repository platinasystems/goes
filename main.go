// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	_ "embed"
	"maps"

	"github.com/platinasystems/goes/v2/pkg/coreutils"
	"github.com/platinasystems/goes/v2/pkg/goes"
	net_tools "github.com/platinasystems/goes/v2/pkg/net/net-tools"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
)

//go:embed LICENSE
var license []byte

//go:embed PATENTS
var patents []byte

func main() {
	maps.Copy(goes.Root, coreutils.Root)
	maps.Copy(goes.Root, net_tools.Root)
	goes.Root["tlsx"] = tlsx.Command
	goes.Show["license"] = license
	goes.Show["patents"] = patents
	goes.Show["tlsx"] = tlsx.Show
	maps.Copy(goes.Daemons, net_tools.Daemons)
	goes.Daemons["tlsx"] = tlsx.Daemon
	goes.Main()
}
