// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package link

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/ip/link/add"
	"github.com/platinasystems/goes/cmd/ip/link/counters"
	"github.com/platinasystems/goes/cmd/ip/link/delete"
	"github.com/platinasystems/goes/cmd/ip/link/mod"
	"github.com/platinasystems/goes/cmd/ip/link/show"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "link",
	USAGE: `
ip link [ COMMAND[ OPTION... ]]

COMMAND := {add | change | counters | delete | replace | set | show(default)}`,
	APROPOS: lang.Alt{
		lang.EnUS: "network device configuration",
	},
	MAN: lang.Alt{
		lang.EnUS: `
EXAMPLES
	ip link set dev ppp0 mtu 1400
		Change the MTU the ppp0 device.

	ip link add link eth0 name eth0.10 type vlan id 10
		Creates a new vlan device eth0.10 on device eth0.

	ip link delete dev eth0.10
		 Removes vlan device.

	ip link help gre
		Display help for the gre link type.

	ip link add name tun1 type ipip remote 192.168.1.1 local 192.168.1.2 \
		ttl 225 encap gue encap-sport auto encap-dport 5555 \
		encap-csum encap-remcsum

		Creates an IPIP that is encapsulated with Generic UDP
		Encapsula‐ tion, and the outer UDP checksum and remote checksum
		offload are enabled.

	ip link add link wpan0 lowpan0 type lowpan
		Creates a 6LoWPAN interface named lowpan0 on the underlying
		IEEE 802.15.4 device wpan0.

SEE ALSO
	ip link man COMMAND || ip link COMMAND -man
	man ip || ip -man`,
	},
	ByName: map[string]cmd.Cmd{
		"add":      add.Goes,
		"change":   mod.Command("change"),
		"counters": counters.Command{},
		"delete":   delete.Command{},
		"replace":  mod.Command("replace"),
		"set":      mod.Command("set"),
		"show":     show.Command("show"),
		"":         show.Command(""),
	},
}
