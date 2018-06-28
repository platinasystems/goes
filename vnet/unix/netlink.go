// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib/elog"
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/internal/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/gre"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"

	"fmt"
	"sync"
)

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

type netlink_namespace struct {
	netlink_socket_fds [2]int
	netlink_socket_pair
	ip4_next_hops []ip4_next_hop
}

type netlink_socket_pair struct {
	broadcast_socket, unicast_socket *netlink.Socket
}

func (p *netlink_socket_pair) close() {
	if p.broadcast_socket != nil {
		p.broadcast_socket.Close()
	}
	if p.unicast_socket != nil {
		p.unicast_socket.Close()
	}
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

	m            *Main
	eventPool    sync.Pool
	add_del_chan chan netlink_add_del
	msg_stats    struct {
		ignored, handled msg_counts
	}
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
		ok = true
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
	_, ok = ns.si_by_ifindex.get(i)
	return
}

func (ns *net_namespace) route_msg_for_vnet_interface(v *netlink.RouteMessage) (intf *net_namespace_interface, ok bool) {
	if a := v.Attrs[netlink.RTA_OIF]; a != nil {
		intf, ok = ns.interface_by_index[a.(netlink.Uint32Attr).Uint()]
		if !ok {
			if false {
				ns.m.m.v.Logf("%s: route_msg_for_vnet_interface unknown interface: %v\n", ns, v)
			}
			return
		}
		ok = ns.knownInterface(intf.ifindex)
		return
	}
	// Check that all multipath next hops are for known interfaces.
	if a := v.Attrs[netlink.RTA_MULTIPATH]; a != nil {
		mp := a.(*netlink.RtaMultipath)
		for i := range mp.NextHops {
			nh := &mp.NextHops[i]
			if ok = ns.knownInterface(nh.Ifindex); !ok {
				return
			}
		}
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
		var intf *net_namespace_interface
		intf, ok = ns.route_msg_for_vnet_interface(v)

		// Check for tunnel in metadata mode (IFLA_*_COLLECT_METADATA is set).
		// If tunnel destination is reachable via vnet interface, then this route is also reachable.
		if !ok && v.Attrs[netlink.RTA_ENCAP_TYPE] != nil {
			ok = intf.tunnel_metadata_mode
		}
	case *netlink.NeighborMessage:
		ok = ns.knownInterface(v.Index)
	case *netlink.DoneMessage, *netlink.NetnsMessage, *netlink.ErrorMessage:
	default:
		panic(fmt.Errorf("unknown message %s", msg))
	}
	return
}

func (m *netlink_main) listener(ns *net_namespace) {
	// Block until next message.
	for msg := range ns.broadcast_socket.Rx {
		e := ns.getEvent(m.m)
		e.ns = ns
		e.msgs = append(e.msgs, msg)

		// Read any remaining messages without blocking.
	loop:
		for {
			select {
			case msg := <-ns.broadcast_socket.Rx:
				if msg == nil { // channel close
					break loop
				}
				e.msgs = append(e.msgs, msg)
			default:
				break loop
			}
		}

		// Add event to be handled next time through main loop.
		e.signal()
	}
}

type net_namespace_netlink_listen_done_event struct {
	vnet.Event
	ns *net_namespace
	m  *Main
}

func (e *net_namespace_netlink_listen_done_event) String() string {
	return "netlink namespace discovery: " + e.ns.name
}
func (e *net_namespace_netlink_listen_done_event) EventAction() { e.m.namespace_discovery_done() }

func (ns *net_namespace) listen(nm *netlink_main) {
	e := ns.getEvent(nm.m)
	e.ns = ns

	err := ns.broadcast_socket.Listen(func(msg netlink.Message) error {
		e.msgs = append(e.msgs, msg)
		return nil
	})
	if err != nil {
		panic(err)
	}

	e.signal()
	e.m.v.SignalEvent(&net_namespace_netlink_listen_done_event{ns: ns, m: nm.m})

	go nm.listener(ns)
}

func (m *netlink_main) LoopInit(l *loop.Loop) {
	if err := m.m.net_namespace_main.init(); err != nil {
		panic(err)
	}
	m.m.net_namespace_main.watch_for_new_net_namespaces()
}

func (nm *netlink_main) Init(m *Main) {
	nm.m = m
	nm.eventPool.New = nm.newEvent
	l := nm.m.v.GetLoop()
	l.RegisterNode(nm, "netlink-listener")
	nm.cliInit()
	nm.namespace_register_nodes()
}

type netlinkEvent struct {
	vnet.Event
	m    *Main
	ns   *net_namespace
	msgs []netlink.Message
}

func (m *netlink_main) newEvent() interface{} {
	return &netlinkEvent{m: m.m}
}
func (ns *net_namespace) getEvent(m *Main) *netlinkEvent {
	v := m.eventPool.Get().(*netlinkEvent)
	*v = netlinkEvent{m: v.m}
	return v
}
func (e *netlinkEvent) signal() {
	if len(e.msgs) > 0 {
		e.m.v.SignalEvent(e)
	}
}
func (e *netlinkEvent) put() {
	if len(e.msgs) > 0 {
		e.msgs = e.msgs[:0]
	}
	e.m.eventPool.Put(e)
}

type msgKindCount struct {
	kind  netlink.MsgType
	count uint16
}

func (a *msgKindCount) update(msg netlink.Message, sʹ string) (s string) {
	var t netlink.MsgType
	s = sʹ
	if msg != nil {
		t = msg.MsgType()
		if a.count > 0 && t == a.kind {
			a.count++
			return
		}
	}
	if a.count > 0 {
		s += " "
		if a.count > 1 {
			s += fmt.Sprintf("%d ", a.count)
		}
		s += a.kind.String() + "\n"
	}
	a.kind = t
	a.count = 1
	return
}

func (e *netlinkEvent) String() (s string) {
	l := len(e.msgs)
	s = fmt.Sprintf("netlink %d:", l)
	var st msgKindCount
	for _, msg := range e.msgs {
		s = st.update(msg, s)
	}
	s = st.update(nil, s)
	return
}

type netlinkElogEvent struct {
	nsName       elog.StringRef
	numMsg       uint16
	numKindCount uint16
	kindCounts   [(elog.EventDataBytes - 1*4 - 2*2) / 4]msgKindCount
}

func (e *netlinkEvent) ElogData() elog.Logger {
	var le netlinkElogEvent
	le.nsName = e.ns.elog_name
	le.numMsg = uint16(len(e.msgs))
	var (
		k msgKindCount
		i int
	)
	for _, msg := range e.msgs {
		t := msg.MsgType()
		if t != k.kind && k.count != 0 {
			le.kindCounts[i] = k
			i++
			if i >= len(le.kindCounts) {
				break
			}
		}
		if k.kind == t {
			k.count++
		} else {
			k.kind = t
			k.count = 1
		}
	}
	if k.count != 0 && i+1 < len(le.kindCounts) {
		le.kindCounts[i] = k
		i++
	}
	le.numKindCount = uint16(i)
	return &le
}

func (e *netlinkElogEvent) Elog(l *elog.Log) {
	if e.numKindCount > 1 {
		l.Logf("netlink %s %d msg", e.nsName, e.numMsg)
		for i := uint16(0); i < e.numKindCount; i++ {
			l.Logf("%d %s", e.kindCounts[i].count, e.kindCounts[i].kind)
		}
	} else {
		l.Logf("netlink %s %d %s", e.nsName, e.kindCounts[0].count, e.kindCounts[0].kind)
	}
}

func (ns *net_namespace) siForIfIndex(ifIndex uint32) (si vnet.Si, ok bool) {
	si, ok = ns.si_by_ifindex.get(ifIndex)
	return
}

func (ns *net_namespace) fibIndexForNamespace() ip.FibIndex { return ip.FibIndex(ns.index) }
func (ns *net_namespace) fibInit(is_del bool) {
	m4 := ip4.GetMain(ns.m.m.v)
	var name string
	if !is_del {
		name = ns.name
	}
	fi := ns.fibIndexForNamespace()
	m4.SetFibNameForIndex(name, fi)
	if is_del {
		m4.FibReset(fi)
	}
}
func (ns *net_namespace) validateFibIndexForSi(si vnet.Si) {
	m4 := ip4.GetMain(ns.m.m.v)
	fi := ns.fibIndexForNamespace()

	m4.SetFibIndexForSi(si, fi)
	return
}

func (e *netlinkEvent) EventAction() {
	var err error
	m := e.m
	vn := m.v
	known := false

	for imsg, msg := range e.msgs {
		if e.ns.is_deleted() {
			continue
		}

		if v, ok := msg.(*netlink.IfInfoMessage); ok {
			if err := e.ns.add_del_interface(m, v); err != nil {
				m.v.Logf("namespace %s, add/del interface %s: %v\n", e.ns, v.Attrs[netlink.IFLA_IFNAME].String(), err)
				continue
			}
		}

		if !e.ns.msg_for_vnet_interface(msg) {
			m.msg_stats.ignored.count(msg)
			if m.verbose_netlink {
				m.v.Logf("%s: netlink ignore %s\n", e.ns, msg)
			}
			continue
		}

		isLastInEvent := imsg+1 == len(e.msgs)
		if m.verbose_netlink {
			m.v.Logf("%s: netlink %s\n", e.ns, msg)
		}
		if false {
			fmt.Printf("********* %s: netlink %s\n", e.ns, msg)
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
			} else if si, ok := e.ns.siForIfIndex(v.Index); ok {
				e.ns.validateFibIndexForSi(si)
				err = si.SetAdminUp(vn, isUp)
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
		}
		if !known {
			err = fmt.Errorf("unkown")
		}
		if err != nil {
			m.v.Logf("%s: netlink %s: %s\n", e.ns, err, msg.String())
		}
		m.msg_stats.handled.count(msg)
	}

	// Return all messages to pools.
	for _, msg := range e.msgs {
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
	} else if si, ok := e.ns.siForIfIndex(v.Index); ok {
		e.ns.validateFibIndexForSi(si)
		err = m4.AddDelInterfaceAddress(si, &p, isDel)
	}
	return
}

func (e *netlinkEvent) ip4NeighborMsg(v *netlink.NeighborMessage) (err error) {
	if v.Ndmsg.Type != netlink.RTN_UNICAST {
		return
	}
	isDel := v.Header.Type == netlink.RTM_DELNEIGH
	si, ok := e.ns.siForIfIndex(v.Index)
	if false { // debug print
		fmt.Printf("netlink NEIGH isDel=%v NDA_LLADDR=%v NDA_DST=%-16v si=%-10v state=%v\n", isDel, ethernetAddress(v.Attrs[netlink.NDA_LLADDR]), v.Attrs[netlink.NDA_DST], si.Name(e.m.v), v.State)
	}
	if !isDel {
		switch v.State {
		case netlink.NUD_NOARP, netlink.NUD_NONE:
			// ignore these
			return
		case netlink.NUD_INCOMPLETE, netlink.NUD_STALE, netlink.NUD_PROBE, netlink.NUD_DELAY:
			// ignore these, too; transient state, don't add yet
			return
		case netlink.NUD_FAILED:
			//do not delete neighbor on FAIL; matches Linux behavior
			//isDel = true
			return
		}
		// Only states that'll add neighbor are NUD_REACHABLE and NUD_PERMANENT
	}
	if !ok {
		// Ignore neighbors for non vnet interfaces.
		return
	}
	const next_hop_weight = 1
	nh := ip4NextHop(v.Attrs[netlink.NDA_DST], next_hop_weight, si)
	nbr := ethernet.IpNeighbor{
		Si:       si,
		Ethernet: ethernetAddress(v.Attrs[netlink.NDA_LLADDR]),
		Ip:       nh.Address.ToIp(),
	}
	m4 := ip4.GetMain(e.m.v)
	em := ethernet.GetMain(e.m.v)
	_, err = em.AddDelIpNeighbor(&m4.Main, &nbr, isDel)

	// Ignore delete of unknown neighbor.
	if err == ethernet.ErrDelUnknownNeighbor {
		err = nil
	}
	return
}

func set_ip4_next_hop_address(a netlink.Attr, nh *ip4.NextHop) {
	if a != nil {
		copy(nh.Address[:], a.(*netlink.Ip4Address)[:])
	}
}

type ip4_next_hop struct {
	ip4.NextHop
	intf  *net_namespace_interface
	attrs []netlink.Attr
}

func (ns *net_namespace) parse_ip4_next_hops(v *netlink.RouteMessage) (nhs []ip4_next_hop) {
	if ns.ip4_next_hops != nil {
		ns.ip4_next_hops = ns.ip4_next_hops[:0]
	}
	nhs = ns.ip4_next_hops

	nh := ip4_next_hop{}
	nh.Weight = 1
	nh.attrs = v.Attrs[:]
	nh_ok := false
	if a := v.Attrs[netlink.RTA_OIF]; a != nil {
		nh.intf = ns.interface_by_index[a.(netlink.Uint32Attr).Uint()]
		nh.Si = nh.intf.si
		nh_ok = true
	}
	set_ip4_next_hop_address(v.Attrs[netlink.RTA_GATEWAY], &nh.NextHop)
	if nh_ok {
		nhs = append(nhs, nh)
	} else if a := v.Attrs[netlink.RTA_MULTIPATH]; a != nil {
		mp := a.(*netlink.RtaMultipath)
		for i := range mp.NextHops {
			mnh := &mp.NextHops[i]
			intf := ns.interface_by_index[mnh.Ifindex]
			nh.attrs = mnh.Attrs[:]
			nh.Si = intf.si
			nh.Weight = ip.NextHopWeight(mnh.Hops)
			if nh.Weight == 0 {
				nh.Weight = 1
			}
			if gw := nh.attrs[netlink.RTA_GATEWAY]; gw != nil {
				set_ip4_next_hop_address(gw, &nh.NextHop)
			} else {
				panic("RTA_MULTIPATH next-hop without RTA_GATEWAY")
			}
			nhs = append(nhs, nh)
		}
	}

	ns.ip4_next_hops = nhs // save for next call
	return
}

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

	isReplace := false
	switch v.Flags {
	case netlink.NLM_F_REPLACE:
		//Override existing
		isReplace = true
	case netlink.NLM_F_EXCL:
		//Do not touch, if it exists
	case netlink.NLM_F_CREATE:
		//Create, if it does not exist
	case netlink.NLM_F_APPEND:
		//Add to end of list
	}

	p := ip4Prefix(v.Attrs[netlink.RTA_DST], v.DstLen)
	isDel := v.Header.Type == netlink.RTM_DELROUTE

	nhs := e.ns.parse_ip4_next_hops(v)
	if isReplace && false { //debug print
		fmt.Printf("netlink NEWROUTE Flags=Replace GATEWAY=%v DST=%v\n", v.Attrs[netlink.RTA_GATEWAY], v.Attrs[netlink.RTA_DST])
	}
	m4 := ip4.GetMain(e.m.v)

	for i := range nhs {
		nh := &nhs[i]
		intf := nh.intf

		// Check for tunnel via RTA_ENCAP_TYPE/RTA_ENCAP attributes.
		if encap_type, ok := nh.attrs[netlink.RTA_ENCAP_TYPE].(netlink.LwtunnelEncapType); ok {
			as := nh.attrs[netlink.RTA_ENCAP].(*netlink.AttrArray)
			switch encap_type {
			case netlink.LWTUNNEL_ENCAP_IP:
				err = e.ip4_in_ip4_route(&p, as, intf, isDel)
			case netlink.LWTUNNEL_ENCAP_IP6:
				err = e.ip4_in_ip6_route(&p, as, intf, isDel)
			}
			if err != nil {
				return
			}
			continue
		}

		// Otherwise its a normal next hop.
		gw := nh.attrs[netlink.RTA_GATEWAY]
		if gw != nil {
			if err = m4.AddDelRouteNextHop(&p, &nh.NextHop, isDel, isReplace); err != nil {
				return
			}
		}
		//This flag should only be set once on first nh because it deletes any previously set nh
		isReplace = false
	}
	return
}

func (e *netlinkEvent) ip4_in_ip4_route(p *ip4.Prefix, as *netlink.AttrArray, intf *net_namespace_interface, isDel bool) (err error) {
	switch intf.kind {
	case netlink.InterfaceKindIp4GRE, netlink.InterfaceKindIpip:
	default:
		err = fmt.Errorf("unsupported ip4 tunnel type: %v", intf.kind)
		return
	}

	var (
		nbr   ip4.Neighbor
		flags uint16
	)
	h := &nbr.Header
	h.Ip_version_and_header_length = 0x45 // v4 no options.
	for k, a := range as.X {
		if a == nil {
			continue
		}
		kind := netlink.LwtunnelIp4AttrKind(k)
		switch kind {
		case netlink.LWTUNNEL_IP_SRC:
			h.Src = ip4.Address(*a.(*netlink.Ip4Address))
		case netlink.LWTUNNEL_IP_DST:
			h.Dst = ip4.Address(*a.(*netlink.Ip4Address))
		case netlink.LWTUNNEL_IP_TTL:
			h.Ttl = a.(netlink.Uint8Attr).Uint()
		case netlink.LWTUNNEL_IP_TOS:
			h.Tos = a.(netlink.Uint8Attr).Uint()
		case netlink.LWTUNNEL_IP_ID:
		case netlink.LWTUNNEL_IP_FLAGS:
			flags = a.(netlink.Uint16Attr).Uint()
		}
	}

	h.Protocol = ip.IP_IN_IP
	if intf.kind == netlink.InterfaceKindIp4GRE {
		h.Protocol = ip.GRE
	}

	if h.Protocol == ip.GRE {
		g := gre.Header{
			Type: ethernet.TYPE_IP4.FromHost(),
		}
		nbr.Payload = make([]byte, gre.SizeofHeader)
		g.Write(nbr.Payload[:])
	}

	// not yet used
	_ = flags

	m4 := ip4.GetMain(e.m.v)

	// By default lookup neighbor in FIB for namespace.
	nbr.FibIndex = e.ns.fibIndexForNamespace()
	nbr.Weight = 1
	err = m4.AddDelRouteNeighbor(p, &nbr, e.ns.fibIndexForNamespace(), isDel)
	return
}

func (e *netlinkEvent) ip4_in_ip6_route(p *ip4.Prefix, as *netlink.AttrArray, intf *net_namespace_interface, isDel bool) (err error) {
	panic("not yet")
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
