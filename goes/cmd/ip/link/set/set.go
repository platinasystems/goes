// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package set

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
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
	theseFlags = []string{
		"up", "down", "nomaster",
	}
	theseParms = []string{
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
	theseVfParms = []string{
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

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	ipFlag, ipParm, args := options.New(args)
	flag, args := flags.New(args, theseFlags...)
	parm, args := parms.New(args, theseParms...)

	subject := parm["group"]
	switch len(subject) {
	case 0:
		switch len(args) {
		case 0:
			return fmt.Errorf("DEVICE: missing")
		case 1:
			subject = args[0]
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	default:
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
	}

	fmt.Println("FIXME", Name, subject)

	_ = ipFlag
	_ = ipParm
	_ = flag

	return nil
}
