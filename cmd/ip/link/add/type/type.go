// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package _type_

import (
	"github.com/platinasystems/goes"
	"github.com/platinasystems/goes/cmd"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/basic"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/bridge"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/geneve"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/gre"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/hsr"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/ip6gre"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/ipip"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/ipoib"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/macsec"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/macvlan"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/vlan"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/vrf"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/vxlan"
	"github.com/platinasystems/goes/cmd/ip/link/add/type/xeth"
	"github.com/platinasystems/goes/lang"
)

var Goes = &goes.Goes{
	NAME: "type",
	USAGE: `
ip link add type TYPE [[ name ] NAME ] [ OPTION ]... [ ARGS ]...`,
	APROPOS: lang.Alt{
		lang.EnUS: "specify virtual link type",
	},
	MAN: lang.Alt{
		lang.EnUS: `
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
	xeth - ethernet multiplexor

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	},
	ByName: map[string]cmd.Cmd{
		"bridge":    bridge.Command{},
		"dummy":     basic.Command("dummy"),
		"geneve":    geneve.Command{},
		"gre":       gre.Command("gre"),
		"gretap":    gre.Command("gretap"),
		"hsr":       hsr.Command{},
		"ifb":       basic.Command("ifb"),
		"ip6gre":    ip6gre.Command("ip6gre"),
		"ip6gretap": ip6gre.Command("ip6gretap"),
		"ipip":      ipip.Command{},
		"ipoib":     ipoib.Command{},
		"macsec":    macsec.Command{},
		"macvlan":   macvlan.Command("macvlan"),
		"macvtap":   macvlan.Command("macvtap"),
		"vcan":      basic.Command("vcan"),
		"vlan":      vlan.Command{},
		"vrf":       vrf.Command{},
		"vxlan":     vxlan.Command{},
		"xeth":      xeth.Command{},
	},
}
