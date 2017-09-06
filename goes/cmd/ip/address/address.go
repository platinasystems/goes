// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package address

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/address/mod"
	"github.com/platinasystems/go/goes/cmd/ip/address/show"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "address"
	Apropos = "protocol address management"
	Usage   = `
ip address [ {add|change|delete|replace|show(default)}[ OPTION... ]]`
	Man = `
SEE ALSO
	ip address man COMMAND || ip address COMMAND -man
	man ip || ip -man
`
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
	g.Plot(mod.New("add"),
		mod.New("change"),
		mod.New("delete"),
		mod.New("replace"),
		show.New("show"),
		show.New(""),
	)
	return g
}
