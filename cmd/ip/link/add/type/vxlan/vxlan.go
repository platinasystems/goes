// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vxlan

import (
	"fmt"
	"net"
	"strings"
	"syscall"

	"github.com/platinasystems/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "vxlan" }

func (Command) Usage() string {
	return "ip link add type vxlan [ OPTIONS ]..."
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add an vxlan virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
OPTIONS
	{ id | vni } VNI
		VXLAN Network Identifer (or VXLAN Segment Identifier)

	group IPADDR
		IP multicast address ("remote" exclusive)

	remote IPADDR
		IP unicast destination address when the link layer address is
		unknown in the VXLAN device forwarding database. ("group"
		exclusive)

	local ADDR
		IP source address

	dev PHYS_DEV
	       physical device of tunnel endpoint

	ttl  { 1:255 }
		Time-to-live of transimitted packets

	tos { 1:8 }
		Type-of-service of transimitted packets

	flowlabel LABEL
		of transimitted packets

	ageing SECONDS
		FDB lifetime of entries learnt by the kernel.

	maxaddress NUMBER
		maximum number of FDB entries

	dstport PORT
		UDP destination port of remote VXLAN tunnel endpoint

	srcport LOW:HIGH
		range of UDP source port numbers of transimetted packets

	[no-]learning
		specifies whether unknown source link layer addresses and IP
		addresses are entered into the VXLAN device forwarding database
		(FDB)

	[no-]proxy
		proxy ARP

	[no-]rsc
		route short circuit

	[no-]l2miss
		netlink LLADDRESS miss notifications

	[no-]l3miss
		netlink IP ADDR miss notifications

	[no-]udpcsum
		UDP checksum of IPv4 transimitted packets

	[no-]udp6zerocsumtx
		UDP checksum calculation of IPv6 transmitted packets

	[no-]udp6zerocsumrx
		accept IPv6 packets w/o UDP checksum

	[no-]remcsumtx

	[no-]external
		external control plane (e.g. ip route encap);
		otherwise, kernel internal FDB

	gbp
		Group Policy extension (VXLAN-GBP).

		Allows to transport group policy context across VXLAN network
		peers.  If enabled, includes the mark of a packet in the VXLAN
		header for outgoing packets and fills the packet mark based on
		the information found in the VXLAN header for incomming
		packets.

		Format of upper 16 bits of packet mark (flags);

			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			|-|-|-|-|-|-|-|-|-|D|-|-|A|-|-|-|
			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

		D := Don't Learn bit.
		When set, this bit indicates that the egress VTEP MUST NOT
		learn the source address of the encapsulated frame.

		A := Indicates that the group policy has already
		been applied to this packet. Policies MUST NOT be
		applied by devices when the A bit is set.

		Format of lower 16 bits of packet mark (policy ID):

			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+
			|        Group Policy ID        |
			+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+-+

		Example:
			iptables -A OUTPUT [...] -j MARK --set-mark 0x800FF

	gpe
		Generic Protocol extension (VXLAN- GPE)
		Only supported with the external control plane.

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var gaddr, laddr, raddr net.IP
	var s string
	var u8 uint8
	var u16 uint16
	var u32 uint32
	var info nl.Attrs

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"learning", "+learning"},
		[]string{"no-learning", "-learning"},
		[]string{"proxy", "+proxy"},
		[]string{"no-proxy", "-proxy"},
		[]string{"rsc", "+rsc"},
		[]string{"no-rsc", "-rsc"},
		[]string{"l2miss", "+l2miss"},
		[]string{"no-l2miss", "-l2miss"},
		[]string{"l3miss", "+l3miss"},
		[]string{"no-l3miss", "-l3miss"},
		[]string{"udpcsum", "+udpcsum"},
		[]string{"no-udpcsum", "-udpcsum"},
		[]string{"udp6zerocsumtx", "+udp6zerocsumtx"},
		[]string{"no-udp6zerocsumtx", "-udp6zerocsumtx"},
		[]string{"udp6zerocsumrx", "+udp6zerocsumrx"},
		[]string{"no-udp6zerocsumrx", "-udp6zerocsumrx"},
		[]string{"remcsumtx", "+remcsumtx"},
		[]string{"no-remcsumtx", "-remcsumtx"},
		[]string{"remcsumrx", "+remcsumrx"},
		[]string{"no-remcsumrx", "-remcsumrx"},
		[]string{"external", "+external"},
		[]string{"no-external", "-external"},
		[]string{"gbp", "+gbp"},
		[]string{"gpe", "+gpe"},
	)
	args = opt.Parms.More(args,
		[]string{"id", "vni"},
		"group",
		"local",
		"remote",
		"dev",
		[]string{"ttl", "hoplimit"},
		[]string{"tos", "dsfield"},
		"flowlabel",
		"ageing",
		"maxaddress",
		"dstport",
		[]string{"srcport", "port"},
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

	s = opt.Parms.ByName["id"]
	if len(s) == 0 {
		return fmt.Errorf("id: missing")
	}
	if _, err = fmt.Sscan(s, &u32); err != nil {
		return fmt.Errorf("vni: %q %v", s, err)
	} else if u32 >= 1<<24 {
		return fmt.Errorf("vni: %q %v", s, syscall.ERANGE)
	}
	info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_ID,
		Value: nl.Uint32Attr(u32)})
	for _, x := range []struct {
		name string
		p    *net.IP
	}{
		{"group", &gaddr},
		{"local", &laddr},
		{"remote", &raddr},
	} {
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		*x.p = net.ParseIP(s)
		if *x.p == nil {
			return fmt.Errorf("%s: %q invalid", x.name, s)
		}
	}
	for _, addr := range []net.IP{gaddr, raddr} {
		if addr != nil {
			if ip4 := addr.To4(); ip4 != nil {
				info = append(info,
					nl.Attr{Type: rtnl.IFLA_VXLAN_GROUP,
						Value: nl.BytesAttr(ip4)})
			} else {
				info = append(info,
					nl.Attr{Type: rtnl.IFLA_VXLAN_GROUP6,
						Value: nl.BytesAttr(addr.To16())})
			}
			break
		}
	}
	if ip4 := laddr.To4(); ip4 != nil {
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_LOCAL,
			Value: nl.BytesAttr(ip4)})
	} else {
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_LOCAL6,
			Value: nl.BytesAttr(laddr.To16())})
	}
	if s = opt.Parms.ByName["dev"]; len(s) > 0 {
		if dev, found := rtnl.If.IndexByName[s]; !found {
			return fmt.Errorf("dev: %q not found", s)
		} else {
			info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_LINK,
				Value: nl.Uint32Attr(dev)})
		}
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"ttl", rtnl.IFLA_VXLAN_TTL},
		{"tos", rtnl.IFLA_VXLAN_TOS},
	} {
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 || s == "inherit" {
			continue
		}
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		info = append(info, nl.Attr{Type: x.t, Value: nl.Uint8Attr(u8)})
	}
	if s = opt.Parms.ByName["flowlabel"]; len(s) > 0 {
		var u32 uint32
		if _, err = fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("flowlabel: %q %v", s, err)
		}
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_LABEL,
			Value: nl.Be32Attr(u32)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"ageing", rtnl.IFLA_VXLAN_AGEING},
		{"maxaddress", rtnl.IFLA_VXLAN_AGEING},
	} {
		var u32 uint32
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err = fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		info = append(info, nl.Attr{Type: x.t,
			Value: nl.Uint32Attr(u32)})
	}
	if s = opt.Parms.ByName["dstport"]; len(s) > 0 {
		if _, err = fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("dstport: %q %v", s, err)
		}
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_PORT,
			Value: nl.Be16Attr(u16)})
	}
	if s = opt.Parms.ByName["srcport"]; len(s) > 0 {
		var pr rtnl.IflaVxlanPortRange
		colon := strings.Index(s, ":")
		if colon < 1 {
			return fmt.Errorf("srcport: %q invalid", s)
		}
		if _, err = fmt.Sscan(s[:colon], &pr.Low); err != nil {
			return fmt.Errorf("srcport low: %q %v", s, err)
		}
		if _, err = fmt.Sscan(s[colon+1:], &pr.High); err != nil {
			return fmt.Errorf("srcport high: %q %v", s, err)
		}
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_PORT_RANGE,
			Value: pr})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"learning", rtnl.IFLA_VXLAN_LEARNING},
		{"proxy", rtnl.IFLA_VXLAN_PROXY},
		{"rsc", rtnl.IFLA_VXLAN_RSC},
		{"l2miss", rtnl.IFLA_VXLAN_L2MISS},
		{"l3miss", rtnl.IFLA_VXLAN_L2MISS},
		{"udpcsum", rtnl.IFLA_VXLAN_UDP_CSUM},
		{"udp6zerocsumtx", rtnl.IFLA_VXLAN_UDP_ZERO_CSUM6_TX},
		{"udp6zerocsumrx", rtnl.IFLA_VXLAN_UDP_ZERO_CSUM6_RX},
		{"remcsumtx", rtnl.IFLA_VXLAN_REMCSUM_TX},
		{"remcsumrx", rtnl.IFLA_VXLAN_REMCSUM_TX},
	} {
		if opt.Flags.ByName[x.name] {
			info = append(info, nl.Attr{Type: x.t,
				Value: nl.Uint8Attr(1)})
		} else if opt.Flags.ByName["no-"+x.name] {
			info = append(info, nl.Attr{Type: x.t,
				Value: nl.Uint8Attr(0)})
		}
	}
	if opt.Flags.ByName["external"] {
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_COLLECT_METADATA,
			Value: nl.Uint8Attr(1)})
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_LEARNING,
			Value: nl.Uint8Attr(0)})
	} else if opt.Flags.ByName["no-external"] {
		info = append(info, nl.Attr{Type: rtnl.IFLA_VXLAN_COLLECT_METADATA,
			Value: nl.Uint8Attr(0)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"gbp", rtnl.IFLA_VXLAN_GBP},
		{"gpe", rtnl.IFLA_VXLAN_GPE},
	} {
		if opt.Flags.ByName[x.name] {
			info = append(info, nl.Attr{Type: x.t,
				Value: nl.NilAttr{}})
		}
	}

	add.Attrs = append(add.Attrs, nl.Attr{Type: rtnl.IFLA_LINKINFO,
		Value: nl.Attrs{
			nl.Attr{Type: rtnl.IFLA_INFO_KIND,
				Value: nl.KstringAttr("vxlan")},
			nl.Attr{Type: rtnl.IFLA_INFO_DATA, Value: info},
		}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
