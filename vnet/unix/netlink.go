// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"fmt"
	"strings"
	"sync"

	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"
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
	e                         *netlinkEvent
	eventPool                 sync.Pool
	add_del_chan              chan netlink_add_del
	unreachable_ip4_next_hops map[ip4.NextHop]unreachable_ip4_next_hop
	current_del_next_hop      ip4.NextHop
	msg_stats                 struct {
		ignored, handled msg_counts
	}
	dummyInterfaceMain
}

// Ignore non-tuntap interfaces (e.g. eth0).
func (m *Main) getInterface(ifindex uint32) (intf *Interface) {
	intf = m.ifByIndex[int(ifindex)]
	return
}

type dummyInterface struct {
	isAdminUp bool
	// Current set of ip4/ip6 addresses for dummy interface collected from netlink IfAddrMessage.
	ip4Addrs map[ip4.Address]ip.FibIndex
	ip6Addrs map[ip6.Address]ip.FibIndex
}

type dummyInterfaceMain struct {
	// Linux ifindex to dummy interface map.
	dummyIfByIndex map[uint32]*dummyInterface
}

// True if given netlink NEWLINK message is for a dummy interface (indicated by ifname starting with "dummy").
func (m *Main) forDummyInterface(msg *netlink.IfInfoMessage) (ok bool) {
	ifname := msg.Attrs[netlink.IFLA_IFNAME].(netlink.StringAttr).String()
	if ok = strings.HasPrefix(ifname, "dummy"); ok {
		if m.dummyIfByIndex == nil {
			m.dummyIfByIndex = make(map[uint32]*dummyInterface)
		}
		if _, known := m.dummyIfByIndex[msg.Index]; !known {
			m.dummyIfByIndex[msg.Index] = &dummyInterface{}
		}
	}
	return
}

func (m *Main) getDummyInterface(ifindex uint32) (i *dummyInterface, ok bool) {
	i, ok = m.dummyIfByIndex[ifindex]
	return
}

func (i *dummyInterface) addDelDummyPuntPrefixes(m *Main, isDel bool) {
	for addr, fi := range i.ip4Addrs {
		m4 := ip4.GetMain(m.v)
		p := ip4.Prefix{Address: addr, Len: 32}
		q := p.ToIpPrefix()
		m4.AddDelRoute(&q, fi, ip.AdjPunt, isDel)
	}
	for addr, fi := range i.ip6Addrs {
		m6 := ip6.GetMain(m.v)
		p := ip6.Prefix{Address: addr, Len: 128}
		q := p.ToIpPrefix()
		m6.AddDelRoute(&q, fi, ip.AdjPunt, isDel)
	}
}
func (m *Main) knownInterface(i uint32) bool { return nil != m.getInterface(i) }

func (m *Main) msgGeneratesEvent(msg netlink.Message) (ok bool) {
	ok = true
	switch v := msg.(type) {
	case *netlink.IfInfoMessage:
		if ok = m.knownInterface(v.Index); !ok {
			ok = m.forDummyInterface(v)
		}
	case *netlink.IfAddrMessage:
		if ok = m.knownInterface(v.Index); !ok {
			_, ok = m.getDummyInterface(v.Index)
		}
	case *netlink.RouteMessage:
		ok = m.knownInterface(uint32(v.Attrs[netlink.RTA_OIF].(netlink.Uint32Attr)))
	case *netlink.NeighborMessage:
		ok = m.knownInterface(v.Index)
	default:
		ok = false // ignore done/error and other messages
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
		msg := <-nm.s.Rx
		m.addMsg(msg)

		// Read any remaining messages without blocking.
	loop:
		for {
			select {
			case msg := <-nm.s.Rx:
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
	mustaddevent := false
	m4 := ip4.GetMain(nm.m.v)
	m4.RegisterFibAddDelHook(nm.m.ip4_fib_add_del)
	err := nm.s.Listen(func(msg netlink.Message) error {
		nm.m.addMsg(msg)
		mustaddevent = true
		return nil
	})
	if err != nil {
		panic(err)
	}
	if mustaddevent {
		nm.e.add()
	}
	go nm.m.listener(l)
}

func (nm *netlinkMain) Init(m *Main) (err error) {
	nm.m = m
	nm.eventPool.New = nm.newEvent
	nm.unreachable_ip4_next_hops = make(map[ip4.NextHop]unreachable_ip4_next_hop)
	l := nm.m.v.GetLoop()
	l.RegisterNode(nm, "netlink-listener")
	nm.cliInit()
	cf := netlink.SocketConfig{
		// Tested and needed to insert/delete 1e6 routes via "netlink route" cli.
		RxBytes: 8 << 20,
	}
	nm.s, err = netlink.NewWithConfig(cf)
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
			// Respect flag admin state changes from unix shell via ifconfig or "ip link" commands.
			known = true
			isUp := v.IfInfomsg.Flags&netlink.IFF_UP != 0
			if di, ok := e.m.getDummyInterface(v.Index); ok {
				// For dummy interfaces add/delete dummy (i.e. loopback) address punts.
				di.isAdminUp = isUp
				di.addDelDummyPuntPrefixes(e.m, !isUp)
			} else {
				intf := e.m.getInterface(v.Index)
				err = intf.si.SetAdminUp(vn, isUp)
			}
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
	isDel := v.Header.Type == netlink.RTM_DELADDR
	if di, ok := m.getDummyInterface(v.Index); ok {
		const fi = 0 // fixme
		q := p.ToIpPrefix()
		if di.isAdminUp || isDel {
			m4.AddDelRoute(&q, fi, ip.AdjPunt, isDel)
		}
		if isDel {
			delete(di.ip4Addrs, p.Address)
		} else {
			if di.ip4Addrs == nil {
				di.ip4Addrs = make(map[ip4.Address]ip.FibIndex)
			}
			di.ip4Addrs[p.Address] = fi
		}
	} else {
		intf := m.getInterface(v.Index)
		err = m4.AddDelInterfaceAddress(intf.si, &p, isDel)
	}
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
		if isDel {
			if u, ok := m.unreachable_ip4_next_hops[nh]; !ok {
				err = ip4.ErrNextHopNotFound
			} else {
				delete(u, p)
			}
		} else {
			m.add_ip4_unreachable_next_hop(p, nh)
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
