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

const (
	Name    = "vlan"
	Apropos = "add a vlan virtual link"
	Usage   = `
ip link add type vlan link IFNAME id ID [ OPTIONS ]...`
	Man = `
OPTIONS
	link IFNAME
	id ID
	protocol PROTOCOL

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
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

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
	err := opt.OnlyName(args)
	if err != nil {
		return err
	}

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt)
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
			"802.1q":  0x8100,
			"802.1ad": 0x88a8,
		}[s]
		if !found {
			return fmt.Errorf("protocol: %q not found", s)
		}
		info = append(info, nl.Attr{rtnl.IFLA_VLAN_PROTOCOL,
			nl.Uint16Attr(proto)})
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
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(Name)},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
