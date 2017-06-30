// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "set"
	Apropos = "link attributes"
	Usage   = `
	ip link set { DEVICE | group GROUP }
		[ { up | down } ]
		[ type ETYPE TYPE_ARGS ]
		[ arp { on | off } ]
		[ dynamic { on | off } ]
		[ multicast { on | off } ]
		[ allmulticast { on | off } ]
		[ promisc { on | off } ]
		[ protodown { on | off } ]
		[ trailers { on | off } ]
		[ txqueuelen PACKETS ]
		[ name NEWNAME ]
		[ address LLADDR ]
		[ broadcast LLADDR ]
		[ mtu MTU ]
		[ netns { PID | NETNSNAME } ]
		[ link-netnsid ID ]
		[ alias NAME ]
		[ vf NUM [ mac LLADDR ]
			[ vlan VLANID [ qos VLAN-QOS ] ]
			[ rate TXRATE ]
			[ max_tx_rate TXRATE ]
			[ min_tx_rate TXRATE ]
			[ spoofchk { on | off } ]
			[ query_rss { on | off } ]
			[ state { auto | enable | disable } ]
			[ trust { on | off } ]
			[ node_guid eui64 ]
			[ port_guid eui64 ] ]
		[ master DEVICE ]
		[ nomaster ]
		[ vrf NAME ]
		[ addrgenmode { eui64 | none | stable_secret | random } ]
	`
	Man = `
SEE ALSO
	ip man link || ip link -man
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"up", "down", "nomaster",
	}
	Parms = []interface{}{
		"dev",
		"group",
		"type",
		"arp",
		"dynamic",
		"multicast",
		"allmulticast",
		"promisc",
		"protodown",
		"trailers",
		"txqueuelen",
		"name",
		"address",
		"broadcast",
		"mtu",
		"netns",
		"link-netnsid",
		"alias",
		"vf",
		"master",
		"vrf",
		"addrgenmode",
	}
	VfParms = []string{
		"mac",
		"vlan",
		"qos",
		"rate",
		"max_tx_rate",
		"min_tx_rate",
		"spoofchk",
		"query_rss",
		"state",
		"trust",
		"node_guid",
		"port_guid",
	}
)

func New() Command { return Command{} }

type Command struct{}

type set options.Options

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	set := (*set)(o)
	args = set.Flags.More(args, Flags)
	args = set.Parms.More(args, Parms)

	if s := set.Parms.ByName["group"]; len(s) > 0 {
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
		fmt.Println("FIXME set group", s)
	} else {
		switch len(args) {
		case 0:
			return fmt.Errorf("DEVICE: missing")
		case 1:
			fmt.Println("FIXME set", args[0])
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	}

	return nil
}
