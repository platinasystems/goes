// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"

	"fmt"
	"sync"
)

type unreachable_ip4_next_hop map[ip4.Prefix]struct{}

type msg_counts struct {
	total   uint64
	by_type map[netlink.MsgType]uint64
}

func (c *msg_counts) count(m netlink.Message) {
	if c.by_type == nil {
		c.by_type = make(map[netlink.MsgType]uint64)
	}
	c.by_type[m.MsgType()]++
	c.total++
}

type netlinkMain struct {
	loop.Node
	m                         *Main
	s                         *netlink.Socket
	c                         chan netlink.Message
	e                         *netlinkEvent
	eventPool                 sync.Pool
	add_del_chan              chan netlink_add_del
	unreachable_ip4_next_hops map[ip4.NextHop]unreachable_ip4_next_hop
	current_del_next_hop      ip4.NextHop
	msg_stats                 struct {
		ignored, handled msg_counts
	}
}

// Ignore non-tuntap interfaces (e.g. eth0).
func (m *Main) getInterface(ifindex uint32) (intf *Interface) {
	intf = m.ifByIndex[int(ifindex)]
	return
}
func (m *Main) knownInterface(i uint32) bool { return nil != m.getInterface(i) }

func (m *Main) msgGeneratesEvent(msg netlink.Message) (ok bool) {
	ok = true
	switch v := msg.(type) {
	case *netlink.IfInfoMessage:
		ok = m.knownInterface(v.Index)
	case *netlink.IfAddrMessage:
		ok = m.knownInterface(v.Index)
	case *netlink.RouteMessage:
		ok = m.knownInterface(uint32(v.Attrs[netlink.RTA_OIF].(netlink.Uint32Attr)))
	case *netlink.NeighborMessage:
		ok = m.knownInterface(v.Index)
	case *netlink.DoneMessage, *netlink.ErrorMessage:
		ok = false // ignore done/error messages
	default:
		panic("unknown netlink message")
	}
	return
}

func (m *Main) addMsg(msg netlink.Message) {
	e := m.getEvent()
	if m.msgGeneratesEvent(msg) {
		e.msgs = append(e.msgs, msg)
	} else {
		m.msg_stats.ignored.count(msg)
		if m.verboseNetlink > 1 {
			m.v.Logf("netlink ignore %s\n", msg)
		}
		// Done with message.
		msg.Close()
	}
}

func (m *Main) listener(l *loop.Loop) {
	nm := &m.netlinkMain
	for {
		// Block until next message.
		msg := <-nm.c
		m.addMsg(msg)

		// Read any remaining messages without blocking.
	loop:
		for {
			select {
			case msg := <-nm.c:
				m.addMsg(msg)
			default:
				break loop
			}
		}

		// Add event to be handled next time through main loop.
		nm.e.add()
	}
}

func (nm *netlinkMain) LoopInit(l *loop.Loop) {
	m4 := ip4.GetMain(nm.m.v)
	m4.RegisterFibAddDelHook(nm.m.ip4_fib_add_del)

	var err error
	nm.c = make(chan netlink.Message, 64)
	cf := netlink.SocketConfig{
		Rx: nm.c,
		// Tested and needed to insert/delete 1e6 routes via "netlink route" cli.
		RcvbufBytes: 8 << 20,
	}
	nm.s, err = netlink.NewWithConfig(cf)
	if err != nil {
		panic(err)
	}
	go nm.s.Listen()
	go nm.m.listener(l)
}

func (nm *netlinkMain) Init(m *Main) (err error) {
	nm.m = m
	nm.eventPool.New = nm.newEvent
	nm.unreachable_ip4_next_hops = make(map[ip4.NextHop]unreachable_ip4_next_hop)
	l := nm.m.v.GetLoop()
	l.RegisterNode(nm, "netlink-listener")
	nm.cliInit()
	return
}

type netlinkEvent struct {
	vnet.Event
	m    *Main
	msgs []netlink.Message
}

func (m *netlinkMain) newEvent() interface{} {
	return &netlinkEvent{m: m.m}
}

func (m *netlinkMain) getEvent() *netlinkEvent {
	if m.e == nil {
		m.e = m.eventPool.Get().(*netlinkEvent)
	}
	return m.e
}
func (e *netlinkEvent) add() {
	if len(e.msgs) > 0 {
		e.m.v.SignalEvent(e)
		e.m.e = nil
	}
}
func (e *netlinkEvent) put() {
	if len(e.msgs) > 0 {
		e.msgs = e.msgs[:0]
	}
	e.m.eventPool.Put(e)
}

type eventSumState struct {
	lastType  netlink.MsgType
	lastCount uint
}

func (a *eventSumState) update(msg netlink.Message, sʹ string) (s string) {
	var t netlink.MsgType
	s = sʹ
	if msg != nil {
		t = msg.MsgType()
		if a.lastCount > 0 && t == a.lastType {
			a.lastCount++
			return
		}
	}
	if a.lastCount > 0 {
		s += " "
		if a.lastCount > 1 {
			s += fmt.Sprintf("%d ", a.lastCount)
		}
		s += a.lastType.String()
	}
	a.lastType = t
	a.lastCount = 1
	return
}

func (e *netlinkEvent) String() (s string) {
	l := len(e.msgs)
	s = fmt.Sprintf("netlink %d:", l)
	var st eventSumState
	for _, msg := range e.msgs {
		s = st.update(msg, s)
	}
	s = st.update(nil, s)
	return
}

func (e *netlinkEvent) EventAction() {
	var err error
	vn := e.m.v
	known := false
	for imsg, msg := range e.msgs {
		isLastInEvent := imsg+1 == len(e.msgs)
		if e.m.verboseNetlink > 0 {
			e.m.v.Logf("netlink %s\n", msg)
		}
		switch v := msg.(type) {
		case *netlink.IfInfoMessage:
			known = true
			intf := e.m.getInterface(v.Index)
			// Respect flag admin state changes from unix shell via ifconfig or "ip link" commands.
			err = intf.si.SetAdminUp(vn, v.IfInfomsg.Flags&netlink.IFF_UP != 0)
		case *netlink.IfAddrMessage:
			switch v.Family {
			case netlink.AF_INET:
				known = true
				err = e.m.ip4IfaddrMsg(v)
			case netlink.AF_INET6:
				known = true
				err = e.m.ip6IfaddrMsg(v)
			}
		case *netlink.RouteMessage:
			switch v.Family {
			case netlink.AF_INET:
				known = true
				err = e.m.ip4RouteMsg(v, isLastInEvent)
			case netlink.AF_INET6:
				known = true
				err = e.m.ip6RouteMsg(v, isLastInEvent)
			}
		case *netlink.NeighborMessage:
			switch v.Family {
			case netlink.AF_INET:
				known = true
				err = e.m.ip4NeighborMsg(v)
			case netlink.AF_INET6:
				known = true
				err = e.m.ip6NeighborMsg(v)
			}
		}
		if !known {
			err = fmt.Errorf("unkown")
		}
		if err != nil {
			e.m.v.Logf("netlink %s: %s\n", err, msg.String())
		}
		e.m.msg_stats.handled.count(msg)
		// Return message to pools.
		msg.Close()
	}
	e.put()
}

func ip4Prefix(t netlink.Attr, l uint8) (p ip4.Prefix) {
	p.Len = uint32(l)
	if t != nil {
		a := t.(*netlink.Ip4Address)
		for i := range a {
			p.Address[i] = a[i]
		}
	}
	return
}

func ip4Address(t netlink.Attr) (a ip4.Address) {
	if t != nil {
		b := t.(*netlink.Ip4Address)
		for i := range b {
			a[i] = b[i]
		}
	}
	return
}

func ip4NextHop(t netlink.Attr, w ip.NextHopWeight, intf *Interface) (n ip4.NextHop) {
	if t != nil {
		b := t.(*netlink.Ip4Address)
		for i := range b {
			n.Address[i] = b[i]
		}
		n.Si = vnet.SiNil
		if intf != nil {
			n.Si = intf.si
		}
		n.Weight = w
	}
	return
}

func ethernetAddress(t netlink.Attr) (a ethernet.Address) {
	if t != nil {
		b := t.(*netlink.EthernetAddress)
		for i := range b {
			a[i] = b[i]
		}
	}
	return
}

func (m *Main) ifAttr(t netlink.Attr) (intf *Interface) {
	if t != nil {
		intf = m.getInterface(t.(netlink.Uint32Attr).Uint())
	}
	return
}

func (m *Main) ip4IfaddrMsg(v *netlink.IfAddrMessage) (err error) {
	p := ip4Prefix(v.Attrs[netlink.IFA_ADDRESS], v.Prefixlen)
	m4 := ip4.GetMain(m.v)
	intf := m.getInterface(v.Index)
	isDel := v.Header.Type == netlink.RTM_DELADDR
	err = m4.AddDelInterfaceAddress(intf.si, &p, isDel)
	return
}

func (m *Main) ip4NeighborMsg(v *netlink.NeighborMessage) (err error) {
	if v.Ndmsg.Type != netlink.RTN_UNICAST {
		return
	}
	isDel := v.Header.Type == netlink.RTM_DELNEIGH
	isStatic := false
	switch v.State {
	case netlink.NUD_NOARP, netlink.NUD_NONE:
		// ignore these
		return
	case netlink.NUD_FAILED:
		isDel = true
	case netlink.NUD_PERMANENT:
		isStatic = true
	}
	intf := m.getInterface(v.Index)
	nh := ip4NextHop(v.Attrs[netlink.NDA_DST], next_hop_weight, intf)
	nbr := ethernet.IpNeighbor{
		Si:       intf.si,
		Ethernet: ethernetAddress(v.Attrs[netlink.NDA_LLADDR]),
		Ip:       nh.Address.ToIp(),
	}
	m4 := ip4.GetMain(m.v)
	em := ethernet.GetMain(m.v)
	// Save away currently deleted next hop for use in ip4_fib_add_del callback.
	if isDel {
		m.current_del_next_hop = nh
	}
	err = em.AddDelIpNeighbor(&m4.Main, &nbr, isDel)

	// Ignore delete of unknown static Arp entry.
	if err == ethernet.ErrDelUnknownNeighbor && isStatic {
		err = nil
	}
	// Add previously unreachable prefixes when next-hop becomes reachable.
	if err == nil && !isDel {
		if u, ok := m.unreachable_ip4_next_hops[nh]; ok {
			delete(m.unreachable_ip4_next_hops, nh)
			for p := range u {
				err = m4.AddDelRouteNextHop(&p, &nh, isDel)
				if err != nil {
					return
				}
			}
		}
	}
	return
}

func (m *Main) add_ip4_unreachable_next_hop(p ip4.Prefix, nh ip4.NextHop) {
	u := m.unreachable_ip4_next_hops[nh]
	if u == nil {
		u = unreachable_ip4_next_hop(make(map[ip4.Prefix]struct{}))
	}
	u[p] = struct{}{}
	m.unreachable_ip4_next_hops[nh] = u
}

func (m *Main) ip4_fib_add_del(fib_index ip.FibIndex, p *ip4.Prefix, adj ip.Adj, isDel bool, isRemap bool) {
	if isDel && isRemap && adj == ip.AdjNil {
		m.add_ip4_unreachable_next_hop(*p, m.current_del_next_hop)
	}
}

// FIXME: Not sure how netlink specifies nexthop weight, so we just set all weights to equal.
const next_hop_weight = 1

func (m *Main) ip4RouteMsg(v *netlink.RouteMessage, isLastInEvent bool) (err error) {
	switch v.Protocol {
	case netlink.RTPROT_KERNEL, netlink.RTPROT_REDIRECT:
		// Ignore all except routes that are static (RTPROT_BOOT) or originating from routing-protocols.
		return
	}
	if v.RouteType != netlink.RTN_UNICAST {
		return
	}
	p := ip4Prefix(v.Attrs[netlink.RTA_DST], v.DstLen)
	intf := m.ifAttr(v.Attrs[netlink.RTA_OIF])

	nh := ip4NextHop(v.Attrs[netlink.RTA_GATEWAY], next_hop_weight, intf)
	isDel := v.Header.Type == netlink.RTM_DELROUTE
	m4 := ip4.GetMain(m.v)
	err = m4.AddDelRouteNextHop(&p, &nh, isDel)
	if err == ip4.ErrNextHopNotFound {
		err = nil
		if u, ok := m.unreachable_ip4_next_hops[nh]; !ok {
			if !isDel {
				m.add_ip4_unreachable_next_hop(p, nh)
			} else {
				err = ip4.ErrNextHopNotFound
			}
		} else if isDel {
			delete(u, p)
		}
	}
	return
}

func ip6Prefix(t netlink.Attr, l uint8) (p ip6.Prefix) {
	p.Len = uint32(l)
	if t != nil {
		a := t.(*netlink.Ip6Address)
		for i := range a {
			p.Address[i] = a[i]
		}
	}
	return
}

// not yet
func (m *Main) ip6IfaddrMsg(v *netlink.IfAddrMessage) (err error)                   { return }
func (m *Main) ip6NeighborMsg(v *netlink.NeighborMessage) (err error)               { return }
func (m *Main) ip6RouteMsg(v *netlink.RouteMessage, isLastInEvent bool) (err error) { return }
