// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package geneve

import (
	"fmt"
	"net"
	"syscall"

	"github.com/platinasystems/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "geneve" }

func (Command) Usage() string {
	return `
ip link add type geneve id ID [ OPTIONS ]...`
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a geneve virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
OPTIONS
	{ id | vni } ID
		Virtual Network Identifer

	remote ADDR
		IP unicast destination of outgoing packets

	ttl TTL
		Time-to-live of outgoing packets

	tos TOS
		Type-of-service of outgoing packets

	flowlabel LABEL
		Flow label of outgoing packets.

	dstport PORT

	[no-]external
	[no-]udpcsum
	[no-]udp6zerocsumtx
	[no-]udp6zerocsumrx

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var info nl.Attrs
	var s string
	var u8 uint8
	var u16 uint16
	var u32 uint32

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"external", "+external"},
		[]string{"no-external", "-external"},
		[]string{"udpcsum", "+udpcsum"},
		[]string{"no-udpcsum", "-udpcsum"},
		[]string{"udp6zerocsumtx", "+udp6zerocsumtx"},
		[]string{"no-udp6zerocsumtx", "-udp6zerocsumtx"},
		[]string{"udp6zerocsumrx", "+udp6zerocsumrx"},
		[]string{"no-udp6zerocsumrx", "-udp6zerocsumrx"},
	)
	args = opt.Parms.More(args,
		[]string{"id", "vni"},
		"remote",
		[]string{"ttl", "hoplimit"},
		[]string{"tos", "dsfield"},
		"flowlabel",
		"dstport",
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
		return fmt.Errorf("missing id")
	}
	if _, err = fmt.Sscan(s, &u32); err != nil {
		return fmt.Errorf("vni: %q %v", s, err)
	} else if u32 >= 1<<24 {
		return fmt.Errorf("vni: %q %v", s, syscall.ERANGE)
	}
	info = append(info, nl.Attr{Type: rtnl.IFLA_GENEVE_ID,
		Value: nl.Uint32Attr(u32)})
	s = opt.Parms.ByName["remote"]
	if len(s) == 0 {
		return fmt.Errorf("missing remote")
	}
	if addr := net.ParseIP(s); addr == nil {
		return fmt.Errorf("remote: %q invalid", s)
	} else if ip4 := addr.To4(); ip4 != nil {
		info = append(info, nl.Attr{Type: rtnl.IFLA_GENEVE_REMOTE,
			Value: nl.BytesAttr(ip4)})
	} else {
		info = append(info, nl.Attr{Type: rtnl.IFLA_GENEVE_REMOTE6,
			Value: nl.BytesAttr(addr.To16())})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"ttl", rtnl.IFLA_GENEVE_TTL},
		{"tos", rtnl.IFLA_GENEVE_TOS},
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
		info = append(info, nl.Attr{Type: rtnl.IFLA_GENEVE_LABEL,
			Value: nl.Be32Attr(u32)})
	}
	if s = opt.Parms.ByName["dstport"]; len(s) > 0 {
		if _, err = fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("dstport: %q %v", s, err)
		}
		info = append(info, nl.Attr{Type: rtnl.IFLA_GENEVE_PORT,
			Value: nl.Be16Attr(u16)})
	}
	if opt.Flags.ByName["external"] {
		info = append(info,
			nl.Attr{Type: rtnl.IFLA_GENEVE_COLLECT_METADATA,
				Value: nl.Uint8Attr(1)})
	} else if opt.Flags.ByName["no-external"] {
		info = append(info,
			nl.Attr{Type: rtnl.IFLA_GENEVE_COLLECT_METADATA,
				Value: nl.Uint8Attr(0)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"udpcsum", rtnl.IFLA_VXLAN_UDP_CSUM},
		{"udp6zerocsumtx", rtnl.IFLA_VXLAN_UDP_ZERO_CSUM6_TX},
		{"udp6zerocsumrx", rtnl.IFLA_VXLAN_UDP_ZERO_CSUM6_RX},
	} {
		if opt.Flags.ByName[x.name] {
			info = append(info, nl.Attr{Type: x.t,
				Value: nl.Uint8Attr(1)})
		} else if opt.Flags.ByName["no-"+x.name] {
			info = append(info, nl.Attr{Type: x.t,
				Value: nl.Uint8Attr(0)})
		}
	}

	add.Attrs = append(add.Attrs, nl.Attr{Type: rtnl.IFLA_LINKINFO,
		Value: nl.Attrs{
			nl.Attr{Type: rtnl.IFLA_INFO_KIND,
				Value: nl.KstringAttr("geneve")},
			nl.Attr{Type: rtnl.IFLA_INFO_DATA, Value: info},
		}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
