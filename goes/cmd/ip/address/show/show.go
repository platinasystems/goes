// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show"
	Apropos = "network address"
	Usage   = `
	ip address show [ dev IFNAME ] [ scope SCOPE-ID ] [ to PREFIX ]
		[ FLAG-LIST ] [ label PATTERN ] [ master DEVICE ]
		[ type TYPE ] [ vrf NAME ] [ up ] ]

	SCOPE-ID := [ host | link | global | NUMBER ]

	FLAG-LIST := [ FLAG-LIST ] FLAG

	FLAG := [ permanent | dynamic | secondary | primary | [-]tentative |
		[-]deprecated | [-]dadfailed | temporary | CONFFLAG-LIST ]

	CONFFLAG-LIST := [ CONFFLAG-LIST ] CONFFLAG

	CONFFLAG := [ home | mngtmpaddr | nodad | noprefixroute | autojoin ]

	TYPE := { bridge | bridge_slave | bond | bond_slave | can | dummy |
		hsr | ifb | ipoib | macvlan | macvtap | vcan | veth | vlan |
		vxlan | ip6tnl | ipip | sit | gre | gretap | ip6gre |
		ip6gretap | vti | vrf | nlmon | ipvlan | lowpan | geneve |
		macsec }
	`
	Man = `
SEE ALSO
	ip man address || ip address -man
`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"permanent", "dynamic", "secondary", "primary",
		"-tentative", "tentative",
		"-deprecated", "deprecated",
		"-dadfailed", "dadfailed",
		"temporary",
		"home", "mngtmpaddr", "nodad", "noprefixroute", "autojoin",
		"up",
	}
	Parms = []interface{}{
		"dev", "scope", "to", "label", "master", "type", "vrf",
	}
)

func New(s string) Command { return Command(s) }

type Command string

type show options.Options

func (c Command) Apropos() lang.Alt {
	apropos := Apropos
	if c == "show" {
		apropos += " (default)"
	}
	return lang.Alt{
		lang.EnUS: apropos,
	}
}

func (Command) Man() lang.Alt    { return man }
func (c Command) String() string { return string(c) }
func (Command) Usage() string    { return Usage }

func (Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	show := (*show)(o)
	args = show.Flags.More(args, Flags)
	args = show.Parms.More(args, Parms)

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	show.show()

	return nil
}

func (show *show) show() {
	fmt.Println("FIXME show")
}
