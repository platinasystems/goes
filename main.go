// Copyright © 2022-2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package main

import (
	_ "embed"

	"github.com/platinasystems/goes/v2/pkg/context/selctx"
	"github.com/platinasystems/goes/v2/pkg/coreutils"
	"github.com/platinasystems/goes/v2/pkg/crypto/xcert"
	"github.com/platinasystems/goes/v2/pkg/crypto/xkey"
	"github.com/platinasystems/goes/v2/pkg/goes"
	net_tools "github.com/platinasystems/goes/v2/pkg/net/net-tools"
	"github.com/platinasystems/goes/v2/pkg/net/tlsx"
)

//go:embed LICENSE
var license []byte

//go:embed PATENTS
var patents []byte

var daemons = map[string]any{
	"tlsx": tlsx.Daemons,
}

var root = map[string]any{
	"daemon": daemons,
	"generate": map[string]any{
		"certificate": xcert.Generate,
		"key":         xkey.Generate,
	},
	"show": map[string]any{
		"certificate": xcert.Show,
		"key":         xkey.Show,
		"tlsx":        tlsx.Show,
		"license":     license,
		"patents":     patents,
	},
}

func main() {
	goes.Merge(root, coreutils.Root)
	goes.Merge(root, net_tools.Root)
	goes.Merge(daemons, net_tools.Daemons)
	selctx.Parameter.Default = root
	goes.Main()
}
