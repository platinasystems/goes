// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package neighbor

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/ip/neighbor/mod"
	"github.com/platinasystems/goes/cmd/ip/neighbor/show"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "neighbor",
	USAGE: `ip neighbor { show (default) | flush } [ proxy ]
	[ to PREFIX ] [ dev DEV ] [ nud STATE ] [ vrf NAME ]

ip neighbor { add | del | change | replace }
	{ ADDR [ lladdr LLADDR ] [ nud STATE ] | proxy ADDR } [ dev DEV ]

STATE := { permanent | noarp | stale | reachable | none | incomplete |
	delay | probe | failed }`,
	APROPOS: lang.Alt{
		lang.EnUS: "neighbor/arp tables management",
	},
	MAN: lang.Alt{
		lang.EnUS: Man,
	},
	ByName: map[string]cmd.Cmd{
		"add":     mod.Command("add"),
		"change":  mod.Command("change"),
		"delete":  mod.Command("delete"),
		"replace": mod.Command("replace"),
		"show":    show.Command("show"),
		"flush":   show.Command("flush"),
		"":        show.Command(""),
	},
}
