// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/internal/netlink"
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
func (c *msg_counts) clear() {
	c.total = 0
	for i := range c.by_type {
		c.by_type[i] = 0
	}
}

type netlink_socket_pair struct {
	broadcast_socket, unicast_socket *netlink.Socket
}

func (p *netlink_socket_pair) close() {
	p.broadcast_socket.Close()
	p.unicast_socket.Close()
}

func (p *netlink_socket_pair) configure(broadcast_fd, unicast_fd int) (err error) {
	cf := netlink.SocketConfig{
		// Tested and needed to insert/delete 1e6 routes via "netlink route" cli.
		RxBytes:           8 << 20,
		DontListenAllNsid: true,
	}
	cf.Groups = []netlink.MulticastGroup{
		netlink.RTNLGRP_LINK,
		netlink.RTNLGRP_NEIGH,
		netlink.RTNLGRP_IPV4_IFADDR,
		netlink.RTNLGRP_IPV4_ROUTE,
		netlink.RTNLGRP_IPV4_MROUTE,
		netlink.RTNLGRP_IPV6_IFADDR,
		netlink.RTNLGRP_IPV6_ROUTE,
		netlink.RTNLGRP_IPV6_MROUTE,
		netlink.RTNLGRP_NSID,
	}
	if p.broadcast_socket, err = netlink.NewWithConfigAndFile(cf, broadcast_fd); err != nil {
		return
	}
	cf.Groups = []netlink.MulticastGroup{netlink.NOOP_RTNLGRP}
	if p.unicast_socket, err = netlink.NewWithConfigAndFile(cf, unicast_fd); err != nil {
		p.broadcast_socket.Close()
		return
	}
	return
}

func (p *netlink_socket_pair) NetlinkTx(request netlink.Message, wait bool) (reply netlink.Message) {
	p.unicast_socket.Tx <- request
	if wait {
		reply = <-p.unicast_socket.Rx
	}
	return
}

type netlink_main struct {
	loop.Node
	net_namespace_main

	m                         *Main
	eventPool                 sync.Pool
	add_del_chan              chan netlink_add_del
	unreachable_ip4_next_hops map[ip4.NextHop]unreachable_ip4_next_hop
	current_del_next_hop      ip4.NextHop
	msg_stats                 struct {
		ignored, handled msg_counts
	}
}

// Ignore non-tuntap interfaces (e.g. eth0).
func (ns *net_namespace) getTuntapInterface(ifindex uint32) (intf *tuntap_interface, ok bool) {
	ns.mu.Lock()
	intf, ok = ns.vnet_tuntap_interface_by_ifindex[ifindex]
	ns.mu.Unlock()
	return
}

type dummy_interface struct {
	isAdminUp bool
	// Current set of ip4/ip6 addresses for dummy interface collected from netlink IfAddrMessage.
	ip4Addrs map[ip4.Address]ip.FibIndex
	ip6Addrs map[ip6.Address]ip.FibIndex
}

// True if given netlink NEWLINK message is for a dummy interface as indicated by IFLA_INFO_KIND.
func (ns *net_namespace) forDummyInterface(msg *netlink.IfInfoMessage) (ok bool) {
	if k := msg.InterfaceKind(); k == netlink.InterfaceKindDummy {
		if ns.dummy_interface_by_ifindex == nil {
			ns.dummy_interface_by_ifindex = make(map[uint32]*dummy_interface)
		}
		if _, known := ns.dummy_interface_by_ifindex[msg.Index]; !known {
			ns.dummy_interface_by_ifindex[msg.Index] = &dummy_interface{}
		}
	}
	return
}

func (ns *net_namespace) getDummyInterface(ifindex uint32) (i *dummy_interface, ok bool) {
	i, ok = ns.dummy_interface_by_ifindex[ifindex]
	return
}

func (i *dummy_interface) addDelDummyPuntPrefixes(m *Main, isDel bool) {
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
func (ns *net_namespace) knownInterface(i uint32) (ok bool) {
	_, ok = ns.getTuntapInterface(i)
	if !ok {
		_, ok = ns.si_by_ifindex[i]
	}
	return
}

func (ns *net_namespace) msg_for_vnet_interface(msg netlink.Message) (ok bool) {
	ok = true
	switch v := msg.(type) {
	case *netlink.IfInfoMessage:
		if ok = ns.knownInterface(v.Index); !ok {
			ok = ns.forDummyInterface(v)
		}
	case *netlink.IfAddrMessage:
		if ok = ns.knownInterface(v.Index); !ok {
			_, ok = ns.getDummyInterface(v.Index)
		}
	case *netlink.RouteMessage:
		ok = ns.knownInterface(uint32(v.Attrs[netlink.RTA_OIF].(netlink.Uint32Attr)))
	case *netlink.NeighborMessage:
		ok = ns.knownInterface(v.Index)
	case *netlink.DoneMessage, *netlink.NetnsMessage, *netlink.ErrorMessage:
	default:
		panic(fmt.Errorf("unknown message %s", msg))
	}
	return
}

func (m *Main) addMsg(ns *net_namespace, msg netlink.Message) {
	if msg == nil {
		// Can happen when reading message from closed channel.
		return
	}
	e := ns.getEvent(m)
	e.ns = ns
	e.msgs = append(e.msgs, msg)
}

func (nm *netlink_main) listener(ns *net_namespace) {
	// Block until next message.
	for msg := range ns.broadcast_socket.Rx {
		nm.m.addMsg(ns, msg)

		// Read any remaining messages without blocking.
	loop:
		for {
			select {
			case msg := <-ns.broadcast_socket.Rx:
				nm.m.addMsg(ns, msg)
			default:
				break loop
			}
		}

		// Add event to be handled next time through main loop.
		ns.current_event.add()
	}
}

func (ns *net_namespace) listen(nm *netlink_main) {
	err := ns.broadcast_socket.Listen(func(msg netlink.Message) error {
		nm.m.addMsg(ns, msg)
		return nil
	})
	if err != nil {
		panic(err)
	}

	// Add artificial done message to mark end of initial dump for this namespace.
	{
		msg := netlink.NewDoneMessage()
		msg.Type = netlink.NLMSG_DONE
		nm.m.addMsg(ns, msg)
	}

	ns.current_event.add()
	go nm.listener(ns)
}

func (nm *netlink_main) LoopInit(l *loop.Loop) {
	m4 := ip4.GetMain(nm.m.v)
	m4.RegisterFibAddDelHook(nm.m.ip4_fib_add_del)
	if err := nm.namespace_init(); err != nil {
		panic(err)
	}
	nm.watch_for_new_net_namespaces()
}

func (nm *netlink_main) Init(m *Main) {
	nm.m = m
	nm.eventPool.New = nm.newEvent
	nm.unreachable_ip4_next_hops = make(map[ip4.NextHop]unreachable_ip4_next_hop)
	l := nm.m.v.GetLoop()
	l.RegisterNode(nm, "netlink-listener")
	nm.cliInit()
	nm.namespace_register_nodes()
}

type netlinkEvent struct {
	vnet.Event
	m    *Main
	msgs []netlink.Message
	ns   *net_namespace
}

func (m *netlink_main) newEvent() interface{} {
	return &netlinkEvent{m: m.m}
}

func (ns *net_namespace) getEvent(m *Main) *netlinkEvent {
	if ns.current_event == nil {
		ns.current_event = m.eventPool.Get().(*netlinkEvent)
	}
	return ns.current_event
}
func (e *netlinkEvent) add() {
	if len(e.msgs) > 0 {
		e.m.v.SignalEvent(e)
		e.ns.current_event = nil
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

func (ns *net_namespace) siForIfIndex(ifIndex uint32) (si vnet.Si, i *tuntap_interface, ok bool) {
	i, ok = ns.getTuntapInterface(ifIndex)
	if ok {
		si = i.si
	} else {
		si, ok = ns.si_by_ifindex[ifIndex]
	}
	if !ok {
		si = vnet.SiNil
	}
	return
}

func (ns *net_namespace) fibIndexForNamespace() ip.FibIndex { return ip.FibIndex(ns.index) }
func (ns *net_namespace) validateFibIndexForNamespace(si vnet.Si) (err error) {
	m4 := ip4.GetMain(ns.m.m.v)
	err = m4.SetFibIndexForSi(si, ns.fibIndexForNamespace())
	return
}

func (e *netlinkEvent) EventAction() {
	var err error
	m := e.m
	vn := m.v
	known := false

	for imsg, msg := range e.msgs {
		if v, ok := msg.(*netlink.IfInfoMessage); ok {
			e.ns.add_del_interface(m, v)
		}

		if !e.ns.msg_for_vnet_interface(msg) {
			m.msg_stats.ignored.count(msg)
			if m.verbose_netlink > 1 {
				m.v.Logf("%s: netlink ignore %s\n", e.ns, msg)
			}
			// Done with message.
			msg.Close()
			continue
		}

		isLastInEvent := imsg+1 == len(e.msgs)
		if m.verbose_netlink > 0 {
			m.v.Logf("%s: netlink %s\n", e.ns, msg)
		}
		switch v := msg.(type) {
		case *netlink.IfInfoMessage:
			// Respect flag admin state changes from unix shell via ifconfig or "ip link" commands.
			known = true
			isUp := v.IfInfomsg.Flags&netlink.IFF_UP != 0
			if di, ok := e.ns.getDummyInterface(v.Index); ok {
				// For dummy interfaces add/delete dummy (i.e. loopback) address punts.
				di.isAdminUp = isUp
				di.addDelDummyPuntPrefixes(m, !isUp)
			} else if si, intf, ok := e.ns.siForIfIndex(v.Index); ok && intf != nil {
				if intf.flags_synced() {
					if isUp {
						err = e.ns.validateFibIndexForNamespace(si)
					}
					if err == nil {
						err = si.SetAdminUp(vn, isUp)
					}
					if !isUp && err == nil {
						err = e.ns.validateFibIndexForNamespace(si)
					}
				} else if intf.flag_sync_in_progress {
					intf.check_flag_sync_done(v)
				}
			}
		case *netlink.IfAddrMessage:
			switch v.Family {
			case netlink.AF_INET:
				known = true
				err = e.ip4IfaddrMsg(v)
			case netlink.AF_INET6:
				known = true
				err = e.ip6IfaddrMsg(v)
			}
		case *netlink.RouteMessage:
			switch v.Family {
			case netlink.AF_INET:
				known = true
				err = e.ip4RouteMsg(v, isLastInEvent)
			case netlink.AF_INET6:
				known = true
				err = e.ip6RouteMsg(v, isLastInEvent)
			}
		case *netlink.NeighborMessage:
			switch v.Family {
			case netlink.AF_INET:
				known = true
				err = e.ip4NeighborMsg(v)
			case netlink.AF_INET6:
				known = true
				err = e.ip6NeighborMsg(v)
			}
		case *netlink.NetnsMessage:
			known = true
			err = e.netnsMessage(v)
		case *netlink.DoneMessage:
			known = true
			err = e.ns.netlink_dump_done(m)
		}
		if !known {
			err = fmt.Errorf("unkown")
		}
		if err != nil {
			m.v.Logf("%s: netlink %s: %s\n", e.ns, err, msg.String())
		}
		m.msg_stats.handled.count(msg)
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

func ip4NextHop(t netlink.Attr, w ip.NextHopWeight, si vnet.Si) (n ip4.NextHop) {
	if t != nil {
		b := t.(*netlink.Ip4Address)
		for i := range b {
			n.Address[i] = b[i]
		}
		n.Si = si
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

func (ns *net_namespace) ifAttr(t netlink.Attr) (si vnet.Si, ok bool) {
	si = vnet.SiNil
	if t != nil {
		si, _, ok = ns.siForIfIndex(t.(netlink.Uint32Attr).Uint())
	}
	return
}

func (e *netlinkEvent) ip4IfaddrMsg(v *netlink.IfAddrMessage) (err error) {
	p := ip4Prefix(v.Attrs[netlink.IFA_ADDRESS], v.Prefixlen)
	m4 := ip4.GetMain(e.m.v)
	isDel := v.Header.Type == netlink.RTM_DELADDR
	if di, ok := e.ns.getDummyInterface(v.Index); ok {
		fi := e.ns.fibIndexForNamespace()
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
	} else if si, _, ok := e.ns.siForIfIndex(v.Index); ok {
		err = m4.AddDelInterfaceAddress(si, &p, isDel)
	}
	return
}

func (e *netlinkEvent) ip4NeighborMsg(v *netlink.NeighborMessage) (err error) {
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
	si, _, ok := e.ns.siForIfIndex(v.Index)
	if !ok {
		// Ignore neighbors for non vnet interfaces.
		return
	}
	nh := ip4NextHop(v.Attrs[netlink.NDA_DST], next_hop_weight, si)
	nbr := ethernet.IpNeighbor{
		Si:       si,
		Ethernet: ethernetAddress(v.Attrs[netlink.NDA_LLADDR]),
		Ip:       nh.Address.ToIp(),
	}
	m4 := ip4.GetMain(e.m.v)
	em := ethernet.GetMain(e.m.v)
	// Save away currently deleted next hop for use in ip4_fib_add_del callback.
	if isDel {
		e.m.current_del_next_hop = nh
	}
	err = em.AddDelIpNeighbor(&m4.Main, &nbr, isDel)

	// Ignore delete of unknown static Arp entry.
	if err == ethernet.ErrDelUnknownNeighbor && isStatic {
		err = nil
	}
	// Add previously unreachable prefixes when next-hop becomes reachable.
	if err == nil && !isDel {
		if u, ok := e.m.unreachable_ip4_next_hops[nh]; ok {
			delete(e.m.unreachable_ip4_next_hops, nh)
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

func (e *netlinkEvent) ip4RouteMsg(v *netlink.RouteMessage, isLastInEvent bool) (err error) {
	switch v.Protocol {
	case netlink.RTPROT_KERNEL, netlink.RTPROT_REDIRECT:
		// Ignore all except routes that are static (RTPROT_BOOT) or originating from routing-protocols.
		return
	}
	if v.RouteType != netlink.RTN_UNICAST {
		return
	}
	// No linux VRF support.  Only main table is meaningful.
	if v.Table != netlink.RT_TABLE_MAIN {
		e.m.v.Logf("netlink ignore route with table not main: %s\n", v)
		return
	}
	si, ok := e.ns.ifAttr(v.Attrs[netlink.RTA_OIF])
	if !ok {
		// Ignore routes for non vnet interfaces.
		return
	}
	p := ip4Prefix(v.Attrs[netlink.RTA_DST], v.DstLen)
	nh := ip4NextHop(v.Attrs[netlink.RTA_GATEWAY], next_hop_weight, si)
	isDel := v.Header.Type == netlink.RTM_DELROUTE
	m4 := ip4.GetMain(e.m.v)
	err = m4.AddDelRouteNextHop(&p, &nh, isDel)
	if err == ip4.ErrNextHopNotFound {
		err = nil
		if isDel {
			if u, ok := e.m.unreachable_ip4_next_hops[nh]; !ok {
				err = ip4.ErrNextHopNotFound
			} else {
				delete(u, p)
			}
		} else {
			e.m.add_ip4_unreachable_next_hop(p, nh)
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
func (e *netlinkEvent) ip6IfaddrMsg(v *netlink.IfAddrMessage) (err error)                   { return }
func (e *netlinkEvent) ip6NeighborMsg(v *netlink.NeighborMessage) (err error)               { return }
func (e *netlinkEvent) ip6RouteMsg(v *netlink.RouteMessage, isLastInEvent bool) (err error) { return }
