// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package _type_

import (
	"github.com/platinasystems/go/goes"
	"github.com/platinasystems/go/goes/cmd/helpers"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/basic"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/bridge"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/geneve"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/gre"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/hsr"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/ip6gre"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/ipip"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/ipoib"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/macsec"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/macvlan"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/vlan"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/vrf"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/type/vxlan"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "type"
	Apropos = "specify virtual link type"
	Usage   = `
ip link add type TYPE [[ name ] NAME ] [ OPTION ]... [ ARGS ]...`
	Man = `
TYPES
	bond - Bonding device
	bridge - Ethernet Bridge device
	dummy - Dummy network interface
	gre - Virtual tunnel interface GRE over IPv4
	gretap - Virtual L2 tuunel interface GRE over IPv4
	hsr - High-availability Seamless Redundancy
	ifb - Intermediate Functional Block device
	ip6gre - Virtual tuunel interface GRE over IPv6
	ip6gretap - Virtual L2 tuunel interface GRE over IPv6
	ip6tnl - Virtual tunnel interface IPv4|IPv6 over IPv6
	ipip - Virtual tunnel interface IPv4 over IPv4
	ipoib - IP over Infiniband device
	macsec - 802.1AE MAC-level encryption
	macvlan - Virtual interface base on link layer address (MAC)
	macvtap - Virtual interface based on link layer address (MAC) and TAP
	sit - Virtual tunnel interface IPv6 over IPv4
	vcan - Virtual Controller Area Network interface
	veth - Virtual point-to-point ethernet network interfaces
	vlan - 802.1q tagged virtual LAN interface
	vrf - Virtual Routing and Forwarding device
	vxlan - Virtual eXtended LAN

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
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
	for _, s := range basic.Types {
		g.Plot(basic.New(s))
	}
	g.Plot(bridge.New())
	g.Plot(geneve.New())
	for _, s := range gre.Types {
		g.Plot(gre.New(s))
	}
	g.Plot(hsr.New())
	for _, s := range ip6gre.Types {
		g.Plot(ip6gre.New(s))
	}
	g.Plot(ipip.New())
	g.Plot(ipoib.New())
	g.Plot(macsec.New())
	for _, s := range macvlan.Types {
		g.Plot(macvlan.New(s))
	}
	g.Plot(vlan.New())
	g.Plot(vrf.New())
	g.Plot(vxlan.New())
	return g
}
