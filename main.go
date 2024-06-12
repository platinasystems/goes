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
	"github.com/platinasystems/goes/v2/pkg/net/vpn"
)

//go:generate go run ./tool/gendoctxt

//go:embed LICENSE
var license []byte

//go:embed PATENTS
var patents []byte

func main() {
	maps.Copy(goes.RootShows(), map[string]any{
		"license": license,
		"patents": patents,
	})
	maps.Copy(goes.Root, coreutils.Root)
	maps.Copy(goes.Root, net_tools.Root)
	maps.Copy(goes.RootDaemons(), net_tools.Daemons)
	goes.Root["vpn"] = vpn.Commands
	goes.RootDaemons()["vpn"] = vpn.Daemons
	goes.RootShows()["vpn"] = vpn.Shows
	goes.Main()
}
