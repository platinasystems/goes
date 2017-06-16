// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"fmt"
	"sort"
	"time"

	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
)

type netlink_add_del struct {
	is_del     bool
	is_ip6     bool
	ip4_prefix ip4.Prefix
	count      uint
	ip4_nhs    []ip4.NextHop
	wait       time.Duration
	fib_index  ip.FibIndex
	ns         *net_namespace
}

func (x *netlink_add_del) String() (s string) {
	s = "add"
	if x.is_del {
		s = "del"
	}
	s += " " + fmt.Sprintf("%d", x.count)
	return
}

func (m *netlink_main) netlink_add_del_routes() {
	for {
		x := <-m.add_del_chan
		m.m.v.Logf("start %s\n", &x)
		for i := uint(0); i < x.count; i++ {
			p := x.ip4_prefix.Add(i)

			for i := range x.ip4_nhs {
				nh := &x.ip4_nhs[i]
				intf := m.m.interface_by_si[nh.Si]
				var addrs [2]netlink.Ip4Address
				addrs[0] = netlink.Ip4Address(p.Address)
				addrs[1] = netlink.Ip4Address(nh.Address)
				msg := netlink.NewRouteMessage()
				msg.Type = netlink.RTM_NEWROUTE
				if x.is_del {
					msg.Type = netlink.RTM_DELROUTE
				}
				msg.Flags = netlink.NLM_F_REQUEST |
					netlink.NLM_F_CREATE |
					netlink.NLM_F_REPLACE |
					netlink.NLM_F_ECHO
				msg.Family = netlink.AF_INET
				msg.Table = netlink.RT_TABLE_MAIN
				msg.RouteType = netlink.RTN_UNICAST
				msg.Protocol = netlink.RTPROT_STATIC
				msg.Attrs[netlink.RTA_DST] = &addrs[0]
				msg.DstLen = uint8(p.Len)
				msg.Attrs[netlink.RTA_GATEWAY] = &addrs[1]
				msg.Attrs[netlink.RTA_OIF] = netlink.Int32Attr(intf.ifindex)
				x.ns.NetlinkTx(msg, false)
			}
		}
		m.m.v.Logf("done %s\n", &x)
	}
}

func (m *netlink_main) ip_route(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	var x netlink_add_del

	switch {
	case in.Parse("add"):
		x.is_del = false
	case in.Parse("del"):
		x.is_del = true
	}

	if !in.Parse("%v", &x.ip4_prefix) {
		err = fmt.Errorf("looking for prefix, got `%s'", in)
		return
	}

	x.count = 1
loop:
	for !in.End() {
		var wait float64
		switch {
		case in.Parse("c%*ount %d", &x.count):
		case in.Parse("t%*able %d", &x.fib_index):
		case in.Parse("w%*ait %f", &wait):
			x.wait = time.Duration(wait * float64(time.Second))
		default:
			break loop
		}
	}

	var nh4 ip4.NextHop
	switch {
	case in.Parse("via %v", &nh4, m.m.v):
		x.ip4_nhs = append(x.ip4_nhs, nh4)
	default:
		err = fmt.Errorf("looking for via NEXT-HOP or adjacency, got `%s'", in)
		return
	}

	if m.add_del_chan == nil {
		m.add_del_chan = make(chan netlink_add_del)
		go m.netlink_add_del_routes()
	}
	m.add_del_chan <- x

	return
}

func (m *netlink_main) enable_disable_log(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	v := m.m.verbose_netlink
	for !in.End() {
		switch {
		case in.Parse("e%*nable"):
			v = 1
		case in.Parse("d%*isable"):
			v = 0
		case in.Parse("t%*oggle"):
			if v > 0 {
				v = 0
			} else {
				v = 1
			}
		default:
			err = cli.ParseError
			return
		}
	}
	m.m.verbose_netlink = v
	return
}

type showMsg struct {
	Type    string `format:"%-30s"`
	Ignored uint64 `format:"%16d"`
	Handled uint64 `format:"%16d"`
}
type showMsgs []showMsg

func (ns showMsgs) Less(i, j int) bool { return ns[i].Type < ns[j].Type }
func (ns showMsgs) Swap(i, j int)      { ns[i], ns[j] = ns[j], ns[i] }
func (ns showMsgs) Len() int           { return len(ns) }

func (m *netlink_main) show_summary(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	sm := make(map[netlink.MsgType]showMsg)
	var (
		x  showMsg
		ok bool
	)
	for t, c := range m.msg_stats.handled.by_type {
		if x, ok = sm[t]; ok {
			x.Handled += c
		} else {
			x.Type = t.String()
			x.Handled = c
		}
		sm[t] = x
	}
	for t, c := range m.msg_stats.ignored.by_type {
		if x, ok = sm[t]; ok {
			x.Ignored += c
		} else {
			x.Type = t.String()
			x.Ignored = c
		}
		sm[t] = x
	}

	msgs := showMsgs{}
	for _, v := range sm {
		if v.Ignored+v.Handled != 0 {
			msgs = append(msgs, v)
		}
	}
	sort.Sort(msgs)
	msgs = append(msgs, showMsg{
		Type:    "Total",
		Ignored: m.msg_stats.ignored.total,
		Handled: m.msg_stats.handled.total,
	})

	elib.TabulateWrite(w, msgs)
	return
}

func (m *netlink_main) clear_summary(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
	m.msg_stats.ignored.clear()
	m.msg_stats.handled.clear()
	return
}

func (m *netlink_main) cliInit() (err error) {
	v := m.m.v
	cmds := []cli.Command{
		cli.Command{
			Name:      "netlink route",
			ShortHelp: "add/delete ip4/ip6 routes via netlink",
			Action:    m.ip_route,
		},
		cli.Command{
			Name:      "netlink log",
			ShortHelp: "enable/disable netlink message logging",
			Action:    m.enable_disable_log,
		},
		cli.Command{
			Name:      "show netlink summary",
			ShortHelp: "show summary of netlink messages received",
			Action:    m.show_summary,
		},
		cli.Command{
			Name:      "clear netlink summary",
			ShortHelp: "clear netlink message counters",
			Action:    m.clear_summary,
		},
		cli.Command{
			Name:      "show netlink namespaces",
			ShortHelp: "show netlink namespaces",
			Action:    m.show_net_namespaces,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
	return
}
