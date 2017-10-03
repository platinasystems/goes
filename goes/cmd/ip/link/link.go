// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package link

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/link/counters"
	"github.com/platinasystems/go/goes/cmd/ip/link/mod"
	"github.com/platinasystems/go/goes/cmd/ip/link/show"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "link"
	Apropos = "network device configuration"
	Usage   = `
ip link [ COMMAND[ OPTION... ]]

COMMAND := {add | change | counters | delete | replace | set | show(default)}`
	Man = `
SEE ALSO
	ip link man COMMAND || ip link COMMAND -man
	man ip || ip -man`
)

func New() *goes.Goes {
	g := goes.New(Name, Usage,
		lang.Alt{
			lang.EnUS: Apropos,
		},
		lang.Alt{
			lang.EnUS: Man,
		})
	g.Plot(helpers.New()...)
	g.Plot(counters.New(),
		mod.New("add"),
		mod.New("change"),
		mod.New("delete"),
		mod.New("replace"),
		mod.New("set"),
		show.New("show"),
		show.New(""),
	)
	return g
}
