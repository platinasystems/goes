// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package address

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/ip/address/mod"
	"github.com/platinasystems/goes/cmd/ip/address/show"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "address",
	USAGE: `
ip address [ {add|change|delete|replace|show(default)}[ OPTION... ]]`,
	APROPOS: lang.Alt{
		lang.EnUS: "protocol address management",
	},
	MAN: lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip address man COMMAND || ip address COMMAND -man
	man ip || ip -man`,
	},
	ByName: map[string]cmd.Cmd{
		"add":     mod.Command("add"),
		"delete":  mod.Command("delete"),
		"replace": mod.Command("replace"),
		"show":    show.Command("show"),
		"":        show.Command(""),
	},
}
