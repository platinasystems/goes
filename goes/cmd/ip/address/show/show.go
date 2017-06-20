// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
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
	theseFlags = []string{
		"permanent", "dynamic", "secondary", "primary",
		"-tentative", "tentative",
		"-deprecated", "deprecated",
		"-dadfailed", "dadfailed",
		"temporary",
		"home", "mngtmpaddr", "nodad", "noprefixroute", "autojoin",
		"up",
	}
	theseParms = []string{
		"dev", "scope", "to", "label", "master", "type", "vrf",
	}
)

func New(s string) Command { return Command(s) }

type Command string

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

func (c Command) Main(args ...string) error {
	command := c
	if len(command) == 0 {
		command = "show"
	}

	ipFlag, ipParm, args := options.New(args)
	flag, args := flags.New(args, theseFlags...)
	parm, args := parms.New(args, theseParms...)

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	fmt.Println("FIXME", command)

	_ = ipFlag
	_ = ipParm
	_ = flag
	_ = parm
	return nil
}
