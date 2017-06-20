// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package link

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/link/add"
	"github.com/platinasystems/go/goes/cmd/ip/link/del"
	"github.com/platinasystems/go/goes/cmd/ip/link/set"
	"github.com/platinasystems/go/goes/cmd/ip/link/show"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "link"
	Apropos = "network device configuration"
	Usage   = `
	ip link  { COMMAND | help }

	ip link add [ link DEVICE ] [ name ] NAME
		[ txqueuelen PACKETS ]
		[ address LLADDR ] [ broadcast LLADDR ]
		[ mtu MTU ] [ index IDX ]
		[ numtxqueues QUEUE_COUNT ] [ numrxqueues QUEUE_COUNT ]
		type TYPE [ ARGS ]

	ip link delete { DEVICE | group GROUP } type TYPE [ ARGS ]

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

	ip link show [ DEVICE | group GROUP ] [ up ] [ master DEVICE ]
		[ type ETYPE ] [ vrf NAME ]

	ip link help [ TYPE ]

	TYPE := [ bridge | bond | can | dummy | hsr | ifb | ipoib | macvlan |
		macvtap | vcan | veth | vlan | vxlan | ip6tnl | ipip | sit |
		gre | gretap | ip6gre | ip6gretap | vti | nlmon | ipvlan | low‐
		pan | geneve | vrf | macsec ]

	ETYPE := [ TYPE | bridge_slave | bond_slave ]
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
	g.Plot(add.New(),
		del.New(),
		set.New(),
		show.New("show"),
		show.New(""),
	)
	return g
}
