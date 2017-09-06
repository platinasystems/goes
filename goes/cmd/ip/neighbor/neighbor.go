// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package neighbor

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/neighbor/mod"
	"github.com/platinasystems/go/goes/cmd/ip/neighbor/show"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "neighbor"
	Apropos = "neighbor/arp tables management"
	Usage   = `ip neighbor { show (default) | flush } [ proxy ]
	[ to PREFIX ] [ dev DEV ] [ nud STATE ] [ vrf NAME ]

ip neighbor { add | del | change | replace }
	{ ADDR [ lladdr LLADDR ] [ nud STATE ] | proxy ADDR } [ dev DEV ]

STATE := { permanent | noarp | stale | reachable | none | incomplete |
	delay | probe | failed }`
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
		mod.New("del"),
		mod.New("delete"),
		mod.New("replace"),
		show.New("show"),
		show.New("flush"),
		show.New(""),
	)
	return g
}
