// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cli"
	"github.com/platinasystems/go/netlink"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"

	"fmt"
	"sort"
	"time"
)

type netlink_add_del struct {
	is_del     bool
	is_ip6     bool
	ip4_prefix ip4.Prefix
	count      uint
	ip4_nhs    []ip4.NextHop
	wait       time.Duration
	fib_index  ip.FibIndex
}

func (x *netlink_add_del) String() (s string) {
	s = "add"
	if x.is_del {
		s = "del"
	}
	s += " " + fmt.Sprintf("%d", x.count)
	return
}

func (m *netlinkMain) add_del() {
	for {
		x := <-m.add_del_chan
		m.m.v.Logf("start %s\n", &x)
		n_tx := 0
		for i := uint(0); i < x.count; i++ {
			p := x.ip4_prefix.Add(i)

			for i := range x.ip4_nhs {
				nh := &x.ip4_nhs[i]
				intf := m.m.ifBySi[nh.Si]
				var addrs [2]netlink.Ip4Address
				addrs[0] = netlink.Ip4Address(p.Address)
				addrs[1] = netlink.Ip4Address(nh.Address)
				msg := &netlink.RouteMessage{}
				msg.Type = netlink.RTM_NEWROUTE
				if x.is_del {
					msg.Type = netlink.RTM_DELROUTE
				}
				msg.Flags = netlink.NLM_F_CREATE | netlink.NLM_F_REPLACE | netlink.NLM_F_ECHO
				msg.Family = netlink.AF_INET
				msg.Table = netlink.RT_TABLE_MAIN
				msg.RouteType = netlink.RTN_UNICAST
				msg.Protocol = netlink.RTPROT_STATIC
				msg.Attrs[netlink.RTA_DST] = &addrs[0]
				msg.DstLen = uint8(p.Len)
				msg.Attrs[netlink.RTA_GATEWAY] = &addrs[1]
				msg.Attrs[netlink.RTA_OIF] = netlink.Int32Attr(intf.ifindex)
				m.s.TxAdd(msg)
				n_tx++
				if n_tx > 256 {
					m.s.TxFlush()
					if x.wait != 0 {
						time.Sleep(x.wait)
					}
					n_tx = 0
				}
			}
		}
		if n_tx > 0 {
			m.s.TxFlush()
			if x.wait != 0 {
				time.Sleep(x.wait)
			}
		}
		m.m.v.Logf("done %s\n", &x)
	}
}

func (m *netlinkMain) ip_route(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
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
		go m.add_del()
	}
	m.add_del_chan <- x

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

func (m *netlinkMain) show_summary(c cli.Commander, w cli.Writer, in *cli.Input) (err error) {
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
		msgs = append(msgs, v)
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

func (m *netlinkMain) cliInit() (err error) {
	v := m.m.v
	cmds := []cli.Command{
		cli.Command{
			Name:      "netlink route",
			ShortHelp: "add/delete ip4/ip6 routes via netlink",
			Action:    m.ip_route,
		},
		cli.Command{
			Name:      "show netlink summary",
			ShortHelp: "summary netlink messages received",
			Action:    m.show_summary,
		},
	}
	for i := range cmds {
		v.CliAdd(&cmds[i])
	}
	return
}
