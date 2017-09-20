// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/platinasystems/go/goes/cmd/ip/internal/group"
	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	nogroup = ^uint32(0)
	Apropos = "link attributes"
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		[]string{"up", "+up"},
		[]string{"down", "no-up", "-up"},
		[]string{"no-master", "-master"},
		[]string{"arp", "+arp"},
		[]string{"no-arp", "-arp"},
		[]string{"dynamic", "+dynamic"},
		[]string{"no-dynamic", "-dynamic"},
		[]string{"multicast", "+multicast"},
		[]string{"no-multicast", "-multicast"},
		[]string{"allmulticast", "+allmulticast"},
		[]string{"no-allmulticast", "-allmulticast"},
		[]string{"promisc", "+promisc"},
		[]string{"no-promisc", "-promisc"},
		[]string{"trailers", "+trailers"},
		[]string{"no-trailers", "-trailers"},
		[]string{"carrier", "+carrier"},
		[]string{"no-carrier", "-carrier"},
		[]string{"protodown", "+protodown"},
		[]string{"no-protodown", "-protodown"},
		[]string{"no-master", "-master"},
		[]string{"no-vrf", "-vrf"},
	}
	Parms = []interface{}{
		"dev",
		"link",
		"index",
		"group",
		"addrgenmode",
		"type",
		"vf",
		"name",
		"alias",
		"qdisc",
		"mtu",
		"address",
		"master",
		"vrf",
		"link-netnsid",
		"netns",
		"mode",
		"state",
		[]string{"broadcast", "brd"},
		"numrxqueues",
		"numtxqueues",
		[]string{"txqueuelen", "qlen", "txqlen"},
	}
)

func New(s string) *Command { return &Command{name: s} }

type Command struct {
	name string
	args []string
	opt  *options.Options

	sr *rtnl.SockReceiver

	hdr   rtnl.Hdr
	msg   rtnl.IfInfoMsg
	attrs rtnl.Attrs

	tinfo rtnl.Attrs

	ifindexByName    map[string]int32
	ifindicesByGroup map[uint32][]int32

	indices []int32

	netns *os.File
}

func (*Command) Apropos() lang.Alt { return apropos }
func (*Command) Man() lang.Alt     { return man }
func (c *Command) String() string  { return c.name }
func (c *Command) Usage() string {
	return fmt.Sprint("ip link ", c.name, ` SUBJECT [ OPTION... ]`)
}

func (c *Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	c.sr = rtnl.NewSockReceiver(sock)

	if err = c.getifindices(); err != nil {
		return err
	}

	c.opt, c.args = options.New(args)
	if err = c.parse(); err != nil {
		return err
	}
	if c.netns != nil {
		defer c.netns.Close()
	}

	c.hdr.Flags = rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK
	c.msg.Family = rtnl.AF_UNSPEC
	switch c.name {
	case "add":
		c.hdr.Type = rtnl.RTM_NEWLINK
		c.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_EXCL
	case "change", "set":
		c.hdr.Type = rtnl.RTM_NEWLINK
	case "replace":
		c.hdr.Type = rtnl.RTM_NEWLINK
		c.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_REPLACE
	case "delete":
		c.hdr.Type = rtnl.RTM_DELLINK
	default:
		return fmt.Errorf("%s: unknown", c)
	}

	for _, ifindex := range c.indices {
		c.msg.Index = ifindex
		if req, err := rtnl.NewMessage(
			c.hdr,
			c.msg,
			c.attrs...,
		); err != nil {
			return err
		} else if err = c.sr.UntilDone(req, func([]byte) {
		}); err != nil {
			return err
		}
	}
	return nil
}

func (c *Command) getifindices() error {
	c.ifindexByName = make(map[string]int32)
	c.ifindicesByGroup = make(map[uint32][]int32)

	req, err := rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETLINK,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
		},
		rtnl.IfInfoMsg{
			Family: rtnl.AF_UNSPEC,
		},
	)
	if err != nil {
		return err
	}
	return c.sr.UntilDone(req, func(b []byte) {
		var ifla rtnl.Ifla
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		msg := rtnl.IfInfoMsgPtr(b)
		ifla.Write(b)
		name := rtnl.Kstring(ifla[rtnl.IFLA_IFNAME])
		c.ifindexByName[name] = msg.Index
		gid := rtnl.Uint32(ifla[rtnl.IFLA_GROUP])
		c.ifindicesByGroup[gid] = append(c.ifindicesByGroup[gid],
			msg.Index)
	})
}

func (c *Command) parse() error {
	var err error
	var dev, link int32
	var gid uint32

	c.args = c.opt.Flags.More(c.args, Flags...)
	c.args = c.opt.Parms.More(c.args, Parms...)

	for _, x := range []struct {
		set   string
		unset string
		flag  uint32
	}{
		{"up", "down", rtnl.IFF_UP},
		{"allmulticast", "no-allmulticast", rtnl.IFF_ALLMULTI},
		{"no-arp", "arp", rtnl.IFF_NOARP},
		{"dynamic", "no-dynamic", rtnl.IFF_DYNAMIC},
		{"multicast", "no-multicast", rtnl.IFF_MULTICAST},
		{"promisc", "no-promisc", rtnl.IFF_PROMISC},
		{"no-trailers", "trailers", rtnl.IFF_NOTRAILERS},
	} {
		if c.opt.Flags.ByName[x.set] {
			c.msg.Change |= x.flag
			c.msg.Flags |= x.flag
		} else if c.opt.Flags.ByName[x.unset] {
			c.msg.Change |= x.flag
			c.msg.Flags &^= x.flag
		}
	}
	for _, x := range []struct {
		set   string
		unset string
		t     uint16
	}{
		{"carrier", "no-carrier", rtnl.IFLA_CARRIER},
		{"protodown", "no-protodown", rtnl.IFLA_PROTO_DOWN},
	} {
		if c.opt.Flags.ByName[x.set] {
			c.attrs = append(c.attrs,
				rtnl.Attr{x.t, rtnl.Uint8Attr(1)})
		} else if c.opt.Flags.ByName[x.unset] {
			c.attrs = append(c.attrs,
				rtnl.Attr{x.t, rtnl.Uint8Attr(0)})
		}
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"no-master", rtnl.IFLA_MASTER},
		{"no-vrf", rtnl.IFLA_MASTER},
	} {
		if c.opt.Flags.ByName[x.name] {
			c.attrs = append(c.attrs,
				rtnl.Attr{x.t, rtnl.Int32Attr(0)})
		}
	}
	for _, x := range []struct {
		name string
		p    *int32
	}{
		{"dev", &dev},
		{"link", &link},
	} {
		if s := c.opt.Parms.ByName[x.name]; len(s) > 0 {
			var found bool
			*x.p, found = c.ifindexByName[s]
			if !found {
				return fmt.Errorf("%s: %q not found",
					x.name, s)
			}
		}
	}
	if s := c.opt.Parms.ByName["index"]; c.name == "add" && len(s) > 0 {
		var ifindex int32
		if _, err = fmt.Sscan(s, &ifindex); err != nil {
			return fmt.Errorf("index: %q %v", s, err)
		}
		c.indices = []int32{ifindex}
	}
	if s := c.opt.Parms.ByName["group"]; len(s) > 0 {
		if gid = group.Id(s); gid == nogroup {
			return fmt.Errorf("group: %q not found", s)
		}
	} else {
		gid = nogroup
	}
	for _, x := range []struct {
		name  string
		parse func(string) error
	}{
		{"addrgenmode", c.parseAddrGenMode},
		{"type", c.parseType},
		{"vf", c.parseVf},
	} {
		if s := c.opt.Parms.ByName[x.name]; len(s) > 0 {
			if err = x.parse(s); err != nil {
				return err
			}
		}
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"name", rtnl.IFLA_IFNAME},
		{"alias", rtnl.IFLA_IFALIAS},
		{"qdisc", rtnl.IFLA_QDISC},
	} {
		s := c.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		c.attrs = append(c.attrs, rtnl.Attr{x.t, rtnl.KstringAttr(s)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"link-netnsid", rtnl.IFLA_LINK_NETNSID},
		{"mtu", rtnl.IFLA_MTU},
		{"numtxqueues", rtnl.IFLA_NUM_TX_QUEUES},
		{"numrxqueues", rtnl.IFLA_NUM_RX_QUEUES},
		{"txqueuelen", rtnl.IFLA_TXQLEN},
	} {
		var u32 uint32
		s := c.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		_, err := fmt.Sscan(s, &u32)
		if err != nil {
			return fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		c.attrs = append(c.attrs, rtnl.Attr{x.t, rtnl.Uint32Attr(u32)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"address", rtnl.IFLA_ADDRESS},
		{"broadcast", rtnl.IFLA_BROADCAST},
	} {
		s := c.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		mac, err := net.ParseMAC(s)
		if err != nil {
			return fmt.Errorf("%s: %q %v", x.name, s, err)
		}
		c.attrs = append(c.attrs, rtnl.Attr{x.t, rtnl.BytesAttr(mac)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"master", rtnl.IFLA_MASTER},
		{"vrf", rtnl.IFLA_MASTER},
	} {
		s := c.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		ifindex, found := c.ifindexByName[s]
		if !found {
			return fmt.Errorf("%s: %q not found", x.name, s)
		}
		c.attrs = append(c.attrs,
			rtnl.Attr{x.t, rtnl.Int32Attr(ifindex)})
	}
	if s := c.opt.Parms.ByName["netns"]; len(s) > 0 {
		var id int32
		var t uint16
		c.netns, err = os.Open(filepath.Join("/var/run/netns", s))
		if err == nil {
			t = rtnl.IFLA_NET_NS_FD
			id = int32(c.netns.Fd())
		} else if _, err := fmt.Sscan(s, &id); err != nil {
			return fmt.Errorf("netns: %q %v", s, err)
		} else {
			t = rtnl.IFLA_NET_NS_PID
		}
		c.attrs = append(c.attrs, rtnl.Attr{t, rtnl.Int32Attr(id)})
	}
	if s := c.opt.Parms.ByName["mode"]; len(s) > 0 {
		mode, found := rtnl.IfLinkModeByName[s]
		if !found {
			return fmt.Errorf("mode: %q unknown", s)
		}
		c.attrs = append(c.attrs,
			rtnl.Attr{rtnl.IFLA_LINKMODE, rtnl.Uint8Attr(mode)})
	}
	if s := c.opt.Parms.ByName["state"]; len(s) > 0 {
		u8, found := rtnl.IfOperByName[s]
		if !found {
			return fmt.Errorf("state: %q unknown", s)
		}
		c.attrs = append(c.attrs,
			rtnl.Attr{rtnl.IFLA_OPERSTATE, rtnl.Uint8Attr(u8)})
	}
	if c.name == "add" {
		switch len(c.args) {
		case 0:
		case 1:
			c.attrs = append(c.attrs, rtnl.Attr{rtnl.IFLA_IFNAME,
				rtnl.KstringAttr(c.args[0])})
		default:
			return fmt.Errorf("%v: unexpected", c.args[1:])
		}
		// FIXME what about link?
		if len(c.indices) == 0 {
			c.indices = []int32{0}
		}
		return nil
	}
	switch len(c.args) {
	case 0:
		if dev != 0 {
			c.indices = []int32{dev}
		} else if gid != nogroup {
			c.indices = c.ifindicesByGroup[gid]
			if len(c.indices) == 0 {
				return fmt.Errorf("group %d empty", gid)
			}
		} else {
			return fmt.Errorf("no dev | group")
		}
	case 1:
		if ifindex, found := c.ifindexByName[c.args[0]]; !found {
			return fmt.Errorf("dev: %q not found", c.args[0])
		} else {
			c.indices = []int32{ifindex}
		}
		if gid != nogroup {
			c.attrs = append(c.attrs, rtnl.Attr{rtnl.IFLA_GROUP,
				rtnl.Uint32Attr(gid)})
		}
	default:
		return fmt.Errorf("%v: unexpected", c.args[1:])
	}
	return nil
}

func (c *Command) parseAddrGenMode(s string) error {
	m, found := rtnl.In6AddrGenModeByName[s]
	if !found {
		return fmt.Errorf("addrgenmode: %q unknown", s)
	}
	c.attrs = append(c.attrs,
		rtnl.Attr{rtnl.IFLA_AF_SPEC,
			rtnl.Attr{uint16(rtnl.AF_INET6),
				rtnl.Attr{rtnl.IFLA_INET6_ADDR_GEN_MODE,
					rtnl.Uint8Attr(m),
				},
			},
		},
	)
	return nil
}
