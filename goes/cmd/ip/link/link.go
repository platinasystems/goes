// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package link

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/link/add"
	"github.com/platinasystems/go/goes/cmd/ip/link/counters"
	"github.com/platinasystems/go/goes/cmd/ip/link/delete"
	"github.com/platinasystems/go/goes/cmd/ip/link/mod"
	"github.com/platinasystems/go/goes/cmd/ip/link/show"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "link"
	Apropos = "network device configuration"
	Usage   = `
ip link [ COMMAND[ OPTION... ]]

COMMAND := {add | change | counters | delete | replace | set | show(default)}`
	Man = `
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
	man ip || ip -man`
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
		counters.New(),
		mod.New("change"),
		delete.New(),
		mod.New("replace"),
		mod.New("set"),
		show.New("show"),
		show.New(""),
	)
	return g
}
