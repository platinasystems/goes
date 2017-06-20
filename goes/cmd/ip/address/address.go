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
	ip [ OPTIONS ] address  { COMMAND | help }

	ip address { add | change | replace } IFADDR dev IFNAME [ LIFETIME ]
		[ CONFFLAG-LIST ]

	ip address del IFADDR dev IFNAME [ mngtmpaddr ]

	ip address { save | flush } [ dev IFNAME ] [ scope SCOPE-ID ]
		[ to PREFIX ] [ FLAG-LIST ] [ label PATTERN ] [ up ]

	ip address [ show [ dev IFNAME ] [ scope SCOPE-ID ] [ to PREFIX ]
		[ FLAG-LIST ] [ label PATTERN ] [ master DEVICE ] [ type TYPE ]
		[ vrf NAME ] [ up ] ]

	ip address { showdump | restore }

	IFADDR := PREFIX | ADDR peer PREFIX [ broadcast ADDR ] [ anycast ADDR ]
		[ label LABEL ] [ scope SCOPE-ID ]

	SCOPE-ID := [ host | link | global | NUMBER ]

	FLAG-LIST := [ FLAG-LIST ] FLAG

	FLAG := [ permanent | dynamic | secondary | primary | [-]tentative |
		[-]deprecated | [-]dadfailed | temporary | CONFFLAG-LIST ]

	CONFFLAG-LIST := [ CONFFLAG-LIST ] CONFFLAG

	CONFFLAG := [ home | mngtmpaddr | nodad | noprefixroute | autojoin ]

	LIFETIME := [ valid_lft LFT ] [ preferred_lft LFT ]

	LFT := [ forever | SECONDS ]

	TYPE := [ bridge | bridge_slave | bond | bond_slave | can | dummy | hsr
		| ifb | ipoib | macvlan | macvtap | vcan | veth | vlan | vxlan
		| ip6tnl | ipip | sit | gre | gretap | ip6gre | ip6gretap | vti
		| vrf | nlmon | ipvlan | lowpan | geneve | macsec ]
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
