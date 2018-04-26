// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vlan

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command struct{}

func (Command) String() string { return "vlan" }

func (Command) Usage() string {
	return `
ip link add type vlan name IFNAME link DEVICE id ID [ OPTIONS ]...`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a vlan virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
	name IFNAME

	link DEVICE
		physical device associated with new vlan interface

	id { 0..4095 }

OPTIONS
	protocol { 802.1q (default) | 802.1ad }

	ingress-qos-map FROM:TO
		defines a mapping of VLAN header prio field to the Linux
		internal packet priority on incoming frames. The format is
		FROM:TO with multiple mappings separated by spaces.

	egress-qos-map FROM:TO
		defines a mapping of Linux internal packet priority to VLAN
		header prio field but for outgoing frames. The format is the
		same as for ingress-qos-map.

		Linux packet priority can be set by iptables(8):

			iptables -t mangle -A POSTROUTING [...]
				-j CLASSIFY --set-class 0:4

		and this "4" priority can be used in the egress qos
		mapping to set VLAN prio "5":

			ip link set veth0.10 type vlan egress 4:5

	[no-]reorder-hdr
		With reorder-hdr, the VLAN header isn't inserted immediately
		but only before passing to the physical device (if this device
		does not support VLAN offloading), the similar on the RX
		direction - by default the packet will be untagged before being
		received by VLAN device. Reordering allows to accel‐ erate
		tagging on egress and to hide VLAN header on ingress so the
		packet looks like regular Ethernet packet, at the same time it
		might be confusing for packet capture as the VLAN header does
		not exist within the packet.

		VLAN offloading can be checked by ethtool(8):

			ethtool -k <phy_dev> | grep tx-vlan-offload

		Where <phy_dev> is the physical device to which VLAN
		device is bound.

	[no-]gvrp
		GARP VLAN Registration Protocol

	[no-]mvrp
		Multiple VLAN Registration Protocol

	[no-]loose-binding
		bond VLAN to the physical device state

SEE ALSO
	ip link add man type || ip link add type -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var info nl.Attrs
	var iflaVlanFlags rtnl.IflaVlanFlags

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"reorder-hdr", "+reorder-hdr"},
		[]string{"no-reorder-hdr", "-reorder-hdr"},
		[]string{"gvrp", "+gvrp"},
		[]string{"no-gvrp", "-gvrp"},
		[]string{"mvrp", "+mvrp"},
		[]string{"no-mvrp", "-mvrp"},
		[]string{"loose-binding", "+loose-binding"},
		[]string{"no-loose-binding", "-loose-binding"},
	)
	args = opt.Parms.More(args,
		"protocol",
		"id",
		"ingress-qos-map",
		"egress-qos-map",
	)

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt, args)
	if err != nil {
		return err
	}

	for _, x := range []struct {
		set   string
		unset string
		flag  uint32
	}{
		{"reorder-hdr", "no-reorder-hdr",
			rtnl.VLAN_FLAG_REORDER_HDR},
		{"gvrp", "no-gvrp", rtnl.VLAN_FLAG_GVRP},
		{"loose-binding", "no-loose-binding",
			rtnl.VLAN_FLAG_LOOSE_BINDING},
		{"mvrp", "no-mvrp", rtnl.VLAN_FLAG_MVRP},
	} {
		if opt.Flags.ByName[x.set] {
			iflaVlanFlags.Mask |= x.flag
			iflaVlanFlags.Flags |= x.flag
		} else if opt.Flags.ByName[x.unset] {
			iflaVlanFlags.Mask |= x.flag
			iflaVlanFlags.Flags &^= x.flag
		}
	}
	if iflaVlanFlags.Mask != 0 {
		info = append(info, nl.Attr{rtnl.IFLA_VLAN_PROTOCOL,
			iflaVlanFlags})
	}
	if s := opt.Parms.ByName["protocol"]; len(s) > 0 {
		proto, found := map[string]uint16{
			"802.1q":  rtnl.ETH_P_8021Q,
			"802.1ad": rtnl.ETH_P_8021AD,
		}[strings.ToLower(s)]
		if !found {
			return fmt.Errorf("protocol: %q not found", s)
		}
		info = append(info, nl.Attr{rtnl.IFLA_VLAN_PROTOCOL,
			nl.Be16Attr(proto)})
	}
	if s := opt.Parms.ByName["id"]; len(s) > 0 {
		var id uint16
		if _, err := fmt.Sscan(s, &id); err != nil {
			return fmt.Errorf("type vlan id: %q %v",
				s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_VLAN_ID,
			nl.Uint16Attr(id)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"egress-qos-map", rtnl.IFLA_VLAN_EGRESS_QOS},
		{"ingress-qos-map", rtnl.IFLA_VLAN_INGRESS_QOS},
	} {
		var qos rtnl.IflaVlanQosMapping
		s := opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		colon := strings.Index(s, ":")
		if colon < 0 {
			return fmt.Errorf("%s: %q invalid", x.name, s)
		}
		if _, err := fmt.Sscan(s[:colon], &qos.From); err != nil {
			return fmt.Errorf("%s: FROM: %q %v",
				x.name, s[:colon], err)
		}
		if _, err := fmt.Sscan(s[colon+1:], &qos.To); err != nil {
			return fmt.Errorf("%s: TO: %q %v",
				x.name, s[colon+1:], err)
		}
		info = append(info, nl.Attr{x.t, qos})
	}

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr("vlan")},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
