// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netlink"
)

const (
	Name    = "show"
	Apropos = "link attributes"
	Usage   = `
	ip link show [ DEVICE | group GROUP ] [ up ] [ master DEVICE ]
		[ type ETYPE ] [ vrf NAME ]
`

	Man = `
SEE ALSO
	ip man link || ip link -man
`
)

const (
	IF_LINK_MODE_DEFAULT linkmode = iota
	IF_LINK_MODE_DORMANT
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"up",
	}
	Parms = []interface{}{
		"group",
		"master",
		"type",
		"vrf",
	}
)

func New(name string) Command { return Command(name) }

type Command string

type dev struct {
	name  string
	index uint
	flags netlink.IfInfoFlags
	addrs map[string]string
	attrs map[string]string
	stats netlink.LinkStats64
}

type show options.Options

type newline struct{}
type linebreak struct{}
type linkmode uint8
type group uint32

func (c Command) Apropos() lang.Alt {
	apropos := Apropos
	if c == "show" {
		apropos += " (default)"
	}
	return lang.Alt{
		lang.EnUS: apropos,
	}
}

func (Command) Man() lang.Alt    { return man }
func (c Command) String() string { return string(c) }
func (Command) Usage() string    { return Usage }

func (Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	show := (*show)(o)
	args = show.Flags.More(args, Flags)
	args = show.Parms.More(args, Parms)

	msgs, err := load()
	if err != nil {
		return err
	}

	for _, info := range msgs {
		show.show(info)
	}

	return nil
}

/*
1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
     link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
or with -details...
     link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00 promiscuity 0 addrgenmode eui64 numtxqueues 1 numrxqueues 1
with -stats...
    RX: bytes  packets  errors  dropped overrun mcast
    0          0        0       0       0       0
    TX: bytes  packets  errors  dropped carrier collsns
    0          0        0       0       0       0
*/
func (show *show) show(info *netlink.IfInfoMessage) {
	u8 := func(attr netlink.Attr) uint8 {
		return attr.(netlink.Uint8er).Uint()
	}
	u32 := func(attr netlink.Attr) uint32 {
		return attr.(netlink.Uint32er).Uint()
	}
	show.print(info.Index)
	show.print(": ", info.Attrs[netlink.IFLA_IFNAME])
	show.print(": <", info.IfInfomsg.Flags, ">")
	show.print(" mtu ", info.Attrs[netlink.IFLA_MTU])
	show.print(" qdisc ", info.Attrs[netlink.IFLA_QDISC])
	show.print(" state ", info.Attrs[netlink.IFLA_OPERSTATE])
	show.print(" mode ", linkmode(u8(info.Attrs[netlink.IFLA_LINKMODE])))
	show.print(" group ", group(u32(info.Attrs[netlink.IFLA_GROUP])))
	show.print(" qlen ", info.Attrs[netlink.IFLA_TXQLEN])
	show.print(linebreak{})
	show.print("    link/", netlink.L2IfType(info.L2IfType))
	show.print(" ", info.Attrs[netlink.IFLA_ADDRESS])
	show.print(" brd ", info.Attrs[netlink.IFLA_BROADCAST])
	if show.Flags.ByName["-d"] {
		show.print(" promiscuity ",
			info.Attrs[netlink.IFLA_PROMISCUITY])
		attr := info.Attrs[netlink.IFLA_INET6_ADDR_GEN_MODE]
		if attr != nil {
			show.print(" addrgenmode  ", attr)
		}
		show.print(" numtxqueues ",
			info.Attrs[netlink.IFLA_NUM_TX_QUEUES])
		show.print(" numrxqueues ",
			info.Attrs[netlink.IFLA_NUM_RX_QUEUES])
	}
	if show.Flags.ByName["-s"] {
		show.print(linebreak{})
		show.print("    FIXME stats")
	}
	show.print(newline{})
}

func (show *show) print(args ...interface{}) {
	for _, v := range args {
		switch t := v.(type) {
		case newline:
			os.Stdout.Write([]byte{'\n'})
		case linebreak:
			if show.Flags.ByName["-o"] {
				os.Stdout.Write([]byte{'\\'})
			} else {
				os.Stdout.Write([]byte{'\n'})
			}
		case []byte:
			os.Stdout.Write(t)
		case string:
			os.Stdout.WriteString(t)
		case group:
			// FIXME lookup in /etc/iproute2/group
			if t == 0 {
				os.Stdout.WriteString("default")
			} else {
				fmt.Print(t)
			}
		case netlink.IfInfoFlags:
			s := strings.Replace(t.String(), " ", "", -1)
			os.Stdout.WriteString(strings.ToUpper(s))
		case netlink.IfOperState:
			os.Stdout.WriteString(strings.ToUpper(t.String()))
		case linkmode:
			switch t {
			case IF_LINK_MODE_DEFAULT:
				os.Stdout.WriteString("DEFAULT")
			case IF_LINK_MODE_DORMANT:
				os.Stdout.WriteString("DORMANT")
			default:
				fmt.Print(t)
			}
		case netlink.L2IfType:
			os.Stdout.WriteString(strings.ToLower(t.String()))
		default:
			if method, found := v.(io.WriterTo); found {
				method.WriteTo(os.Stdout)
			} else {
				fmt.Print(v)
			}
		}
	}
}

func load() (info []*netlink.IfInfoMessage, err error) {
	nl, err := netlink.New(netlink.RTNLGRP_LINK)
	if err != nil {
		return
	}
	h := func(msg netlink.Message) error {
		switch msg.MsgType() {
		case netlink.RTM_NEWLINK, netlink.RTM_GETLINK:
			// FIXME what about RTM_DELLINK and RTM_SETLINK ?
			info = append(info, msg.(*netlink.IfInfoMessage))
		default:
			msg.Close()
		}
		return nil
	}
	req := netlink.ListenReq{netlink.RTM_GETLINK, netlink.AF_PACKET}
	err = nl.Listen(h, req)
	if err == nil {
		sort.Slice(info, func(i, j int) bool {
			return info[i].Index < info[j].Index
		})
	}
	return
}
