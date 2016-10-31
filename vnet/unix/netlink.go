// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package unix

import (
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/netlink"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/ip6"

	"fmt"
	"sync"
)

type netlinkMain struct {
	loop.Node
	m            *Main
	s            *netlink.Socket
	c            chan netlink.Message
	e            *netlinkEvent
	eventPool    sync.Pool
	add_del_chan chan netlink_add_del
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
		if m.verboseNetlink {
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
	l := nm.m.v.GetLoop()
	l.RegisterNode(nm, "netlink-listener")
	nm.cliInit()
	return
}

type netlinkEvent struct {
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
		e.m.AddEvent(e, e.m)
		e.m.e = nil
	}
}
func (e *netlinkEvent) put() {
	if len(e.msgs) > 0 {
		e.msgs = e.msgs[:0]
	}
	e.m.eventPool.Put(e)
}

func (m *netlinkMain) EventHandler() {}

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
	for _, msg := range e.msgs {
		if e.m.verboseNetlink {
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
				err = e.m.ip4RouteMsg(v)
			case netlink.AF_INET6:
				known = true
				err = e.m.ip6RouteMsg(v)
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
	dst := ip4Address(v.Attrs[netlink.NDA_DST])
	nbr := ethernet.IpNeighbor{
		Si:       intf.si,
		Ethernet: ethernetAddress(v.Attrs[netlink.NDA_LLADDR]),
		Ip:       dst.ToIp(),
	}
	m4 := ip4.GetMain(m.v)
	err = ethernet.GetMain(m.v).AddDelIpNeighbor(&m4.Main, &nbr, isDel)

	// Ignore delete of unknown static Arp entry.
	if err == ethernet.ErrDelUnknownNeighbor && isStatic {
		err = nil
	}
	// not yet
	if false {
		fmt.Printf("nbr if %s, isDel %v, %s -> %s\n", intf, isDel, m4.Main.AddressStringer(&nbr.Ip), &nbr.Ethernet)
	}
	return
}

func (m *Main) ip4RouteMsg(v *netlink.RouteMessage) (err error) {
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
	nh := ip4.NextHop{
		Si:      vnet.SiNil,
		Address: ip4Address(v.Attrs[netlink.RTA_GATEWAY]),
		// FIXME: Not sure how netlink specifies nexthop weight.
		Weight: 1,
	}
	if intf != nil {
		nh.Si = intf.si
	}
	isDel := v.Header.Type == netlink.RTM_DELROUTE
	if false {
		fmt.Printf("route if %s, isDel %v, %s -> %+v %s\n", intf, isDel, &p, &nh, err)
	}
	m4 := ip4.GetMain(m.v)
	err = m4.AddDelRouteNextHop(&p, &nh, isDel)
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
func (m *Main) ip6IfaddrMsg(v *netlink.IfAddrMessage) (err error)     { return }
func (m *Main) ip6NeighborMsg(v *netlink.NeighborMessage) (err error) { return }
func (m *Main) ip6RouteMsg(v *netlink.RouteMessage) (err error)       { return }
