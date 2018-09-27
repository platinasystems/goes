// Copyright 2018 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// File to catch message updates from linux kernel (via platina-mk1 driver) that
// signal different networking events (replacement for netlink.go)
//  - prefix/nexthop add/delete/replace
//  - ifaddr add/delete
//  - ifinfo (admin up/down)
//  - neighbor add/delete/replace
//
package unix

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"
	"github.com/platinasystems/go/vnet/unix/internal/dbgfdb"
	"github.com/platinasystems/xeth"
)

var (
	// Function flags
	FdbOn       bool = true
	FdbIfAddrOn bool = true
)

const MAXMSGSPEREVENT = 1000

type fdbEvent struct {
	vnet.Event
	fm      *FdbMain
	evType  vnet.ActionType
	NumMsgs int
	Msgs    [MAXMSGSPEREVENT][]byte
}

type FdbMain struct {
	loop.Node
	m         *Main
	eventPool sync.Pool
}

func (fm *FdbMain) Init(m *Main) {
	fm.m = m
	fm.eventPool.New = fm.newEvent
	l := fm.m.v.GetLoop()
	l.RegisterNode(fm, "fdb-listener")
}

// This needs to be used to initialize the eventpool
func (m *FdbMain) newEvent() interface{} {
	return &fdbEvent{fm: m}
}

func (m *FdbMain) GetEvent(evType vnet.ActionType) *fdbEvent {
	v := m.eventPool.Get().(*fdbEvent)
	*v = fdbEvent{fm: m, evType: evType}
	return v
}

func (e *fdbEvent) Signal() {
	if len(e.Msgs) > 0 {
		e.fm.m.v.SignalEvent(e)
	}
}

func (e *fdbEvent) put() {
	e.NumMsgs = 0
	// Zero out array?
	e.fm.eventPool.Put(e)
}

func (e *fdbEvent) String() (s string) {
	l := e.NumMsgs
	s = fmt.Sprintf("fdb %d:", l)
	return
}

func (e *fdbEvent) EnqueueMsg(msg []byte) bool {
	if e.NumMsgs+1 > MAXMSGSPEREVENT {
		return false
	}
	e.Msgs[e.NumMsgs] = msg
	e.NumMsgs++
	return true
}

func initVnetFromXeth(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain

	// Initiate walk of PortEntry map to send vnetd
	// interface info and namespaces
	ProcessInterfaceInfo(nil, vnet.ReadyVnetd, v, 0)

	// Initiate walk of PortEntry map to send IFAs
	ProcessInterfaceAddr(nil, vnet.ReadyVnetd, v)

	// Initiate walk of PortEntry map to send vnetd ethtool data
	InitInterfaceEthtool(v)

	// Send events for initial dump of fib entries
	fe := fdbm.GetEvent(vnet.Dynamic)
	xeth.DumpFib()
	for msg := range xeth.RxCh {
		if kind := xeth.KindOf(msg); kind == xeth.XETH_MSG_KIND_BREAK {
			xeth.Pool.Put(msg)
			break
		}
		if ok := fe.EnqueueMsg(msg); !ok {
			// filled event with messages so send it and continue
			fe.Signal()
			fe = fdbm.GetEvent(vnet.Dynamic)
			if ok = fe.EnqueueMsg(msg); !ok {
				panic("can't enqueue initial fdb dump")
			}
		}
	}
	fe.Signal()

	// Drain XETH channel into vnet events.
	go gofdb(v)
}

func gofdb(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain
	fe := fdbm.GetEvent(vnet.Dynamic)
	for msg := range xeth.RxCh {
		if ok := fe.EnqueueMsg(msg); !ok {
			fe.Signal()
			fe = fdbm.GetEvent(vnet.Dynamic)
			if ok = fe.EnqueueMsg(msg); !ok {
				panic("Can't enqueue fdb")
			}
		}
		if len(xeth.RxCh) == 0 {
			fe.Signal()
			fe = fdbm.GetEvent(vnet.Dynamic)
		}
	}
}

func (e *fdbEvent) EventAction() {
	var err error
	m := e.fm
	vn := m.m.v

	for i := 0; i < e.NumMsgs; i++ {
		msg := e.Msgs[i]
		kind := xeth.KindOf(msg)
		dbgfdb.XethMsg.Log("kind:", kind)
		ptr := unsafe.Pointer(&msg[0])
		switch xeth.KindOf(msg) {
		case xeth.XETH_MSG_KIND_NEIGH_UPDATE:
			err = ProcessIpNeighbor((*xeth.MsgNeighUpdate)(ptr), vn)
		case xeth.XETH_MSG_KIND_FIBENTRY:
			err = ProcessFibEntry((*xeth.MsgFibentry)(ptr), vn)
		case xeth.XETH_MSG_KIND_IFA:
			err = ProcessInterfaceAddr((*xeth.MsgIfa)(ptr), e.evType, vn)
		case xeth.XETH_MSG_KIND_IFINFO:
			err = ProcessInterfaceInfo((*xeth.MsgIfinfo)(ptr), e.evType, vn, 0)
		case xeth.XETH_MSG_KIND_ETHTOOL_FLAGS:
			msg := (*xeth.MsgEthtoolFlags)(ptr)
			xethif := xeth.Interface.Indexed(msg.Ifindex)
			ifname := xethif.Ifinfo.Name
			vnet.SetPort(ifname).Flags =
				xeth.EthtoolPrivFlags(msg.Flags)
			fec91 := vnet.PortIsFec91(ifname)
			fec74 := vnet.PortIsFec74(ifname)
			dbgfdb.IfETFlag.Log(ifname, "fec91", fec91, "fec74", fec74)
			var fec ethernet.ErrorCorrectionType
			// if both fec91 and fec74 are on, set fec to fec91
			if fec91 {
				fec = ethernet.ErrorCorrectionCL91
			} else if fec74 {
				fec = ethernet.ErrorCorrectionCL74
			} else {
				fec = ethernet.ErrorCorrectionNone
			}
			media := "fiber"
			if vnet.PortIsCopper(ifname) {
				media = "copper"
			}
			dbgfdb.IfETFlag.Log(ifname, media)
			hi, found := vn.HwIfByName(ifname)
			if found {
				dbgfdb.IfETFlag.Log(ifname, "setting",
					"media", media, "fec", fec)
				hi.SetMedia(vn, media)
				err = ethernet.SetInterfaceErrorCorrection(vn, hi, fec)
				dbgfdb.IfETFlag.Log(err, "on", ifname)
			}
		case xeth.XETH_MSG_KIND_ETHTOOL_SETTINGS:
			msg := (*xeth.MsgEthtoolSettings)(ptr)
			xethif := xeth.Interface.Indexed(msg.Ifindex)
			ifname := xethif.Ifinfo.Name
			vnet.SetPort(ifname).Speed =
				xeth.Mbps(msg.Speed)
			hi, found := vn.HwIfByName(ifname)
			if found {
				var bw float64
				if msg.Autoneg == 0 {
					bw = float64(msg.Speed)
				}
				speedOk := false
				dbgfdb.IfETSetting.Log(ifname, "setting speed", bw)
				switch bw {
				case 0, 1000, 10000, 20000, 25000, 40000, 50000, 100000:
					speedOk = true
				}
				if !speedOk {
					err = fmt.Errorf("unexpected speed: %v",
						bw)
					dbgfdb.IfETSetting.Log(err, "on", ifname)
				} else {
					bw *= 1e6
					err = hi.SetSpeed(vn, vnet.Bandwidth(bw))
					dbgfdb.IfETSetting.Log(err, "on", ifname)
				}
			}

		}
		dbgfdb.XethMsg.Log(err, "with kind", kind)
		xeth.Pool.Put(msg)
	}
	e.put()
}

func ipnetToIP4Prefix(ipnet *net.IPNet) (p ip4.Prefix) {
	for i := range ipnet.IP {
		p.Address[i] = ipnet.IP[i]
	}
	maskOnes, _ := ipnet.Mask.Size()
	p.Len = uint32(maskOnes)
	return
}

func (ns *net_namespace) parseIP4NextHops(msg *xeth.MsgFibentry) (nhs []ip4_next_hop) {
	if ns.ip4_next_hops != nil {
		ns.ip4_next_hops = ns.ip4_next_hops[:0]
	}
	nhs = ns.ip4_next_hops

	if ns == nil {
		dbgfdb.Fib.Log("ns is nil")
	} else {
		dbgfdb.Fib.Log("ns is", ns.name)
	}
	xethNhs := msg.NextHops()
	dbgfdb.Fib.Log(len(xethNhs), "nexthops")
	for i, _ := range xethNhs {
		dbgfdb.Fib.Logf("nexthops[%d]: %#v", i, xethNhs[i])
	}

	// If only 1 nh then assume this is single OIF nexthop
	// otherwise it's multipath
	nh := ip4_next_hop{}
	nh.Weight = 1
	if len(xethNhs) == 1 {
		nh.intf = ns.interface_by_index[uint32(xethNhs[0].Ifindex)]
		if nh.intf == nil {
			dbgfdb.Fib.Log("no ns-intf for ifindex",
				xethNhs[0].Ifindex)
			return
		}
		nh.Si = nh.intf.si
		copy(nh.Address[:], xethNhs[0].IP())
		nhs = append(nhs, nh)
	} else {
		for _, xnh := range xethNhs {
			intf := ns.interface_by_index[uint32(xnh.Ifindex)]
			if intf == nil {
				dbgfdb.Fib.Log("no ns-intf for ifindex",
					xnh.Ifindex)
				continue
			}
			nh.Si = intf.si
			nh.Weight = ip.NextHopWeight(xnh.Weight)
			if nh.Weight == 0 {
				nh.Weight = 1
			}
			copy(nh.Address[:], xnh.IP())
			nhs = append(nhs, nh)
		}
	}
	ns.ip4_next_hops = nhs // save for next call
	return
}

func ProcessIpNeighbor(msg *xeth.MsgNeighUpdate, v *vnet.Vnet) (err error) {

	// For now only doing IPv4
	if msg.Family != syscall.AF_INET {
		return
	}

	kind := xeth.Kind(msg.Kind)
	dbgfdb.Neigh.Log(kind)
	var macIsZero bool = true
	for _, i := range msg.Lladdr {
		if i != 0 {
			macIsZero = false
			break
		}
	}
	isDel := macIsZero
	m := GetMain(v)
	ns := getNsByInode(m, msg.Net)
	netns := xeth.Netns(msg.Net)
	if ns == nil {
		dbgfdb.Ns.Log("namespace", netns, "not found")
		return
	}
	si, ok := ns.siForIfIndex(uint32(msg.Ifindex))
	if !ok {
		dbgfdb.Neigh.Log("no si for", msg.Ifindex, "in", ns.name)
		return
	}
	nbr := ethernet.IpNeighbor{
		Si:       si,
		Ethernet: ethernet.Address(msg.Lladdr),
		Ip:       ip.Address(msg.Dst),
	}
	m4 := ip4.GetMain(v)
	em := ethernet.GetMain(v)
	dbgfdb.Neigh.Log(addDel(isDel), "nbr", nbr)
	_, err = em.AddDelIpNeighbor(&m4.Main, &nbr, isDel)

	// Ignore delete of unknown neighbor.
	if err == ethernet.ErrDelUnknownNeighbor {
		err = nil
	}
	return
}

// Zero Gw processing covers 2 major sub-cases:
// 1. Interface-address setting
//    If local table entry and is a known interface of vnet i.e. front-panel then
//    install an interface address
// 2. Dummy interface setting
//    If not a known interface of vnet, we assume it's a dummy and install as a punt
//    adjacency (FIXME - need to filter routes through eth0 and others)
func ProcessZeroGw(msg *xeth.MsgFibentry, v *vnet.Vnet, ns *net_namespace, isDel, isLocal, isMainUc bool) (err error) {
	xethNhs := msg.NextHops()
	pe := vnet.GetPortByIndex(xethNhs[0].Ifindex)
	if pe != nil {
		// Adds (local comes first followed by main-uc):
		// If local-local route then stash /32 prefix into Port[] table
		// If main-unicast route then lookup port in Port[] table and marry
		// local prefix and main-unicast prefix-len to install interface-address
		// Dels (main-uc comes first followed by local):
		//
		if FdbIfAddrOn {
			return
		}
		m := GetMain(v)
		ns := getNsByInode(m, pe.Net)
		if ns == nil {
			dbgfdb.Ns.Log("namespace", pe.Net, "not found")
			return
		}
		dbgfdb.Ns.Log("namespace", pe.Net, "found")
		if isLocal {
			dbgfdb.Fib.Log(addDel(isDel), "local", msg.Prefix())
		} else if isMainUc {
			dbgfdb.Fib.Log(addDel(isDel), "main", msg.Prefix())
			//m4 := ip4.GetMain(v)
			//ns.Ip4IfaddrMsg(m4, msg.Prefix(), uint32(xethNhs[0].Ifindex), isDel)
		} else {
			dbgfdb.Fib.Log(addDel(isDel),
				"neither local nor main", msg.Prefix())
		}
	} else {
		// dummy processing
		if isLocal {
			dbgfdb.Fib.Log("dummy install punt for", msg.Prefix())
			m4 := ip4.GetMain(v)
			in := msg.Prefix()
			var addr ip4.Address
			for i := range in.IP {
				addr[i] = in.IP[i]
			}
			// Filter 127.*.*.* routes
			if addr[0] == 127 {
				return
			}
			p := ip4.Prefix{Address: addr, Len: 32}
			q := p.ToIpPrefix()
			m4.AddDelRoute(&q, ns.fibIndexForNamespace(), ip.AdjPunt, isDel)
		}
	}
	return
}

func addrIsZero(addr ip4.Address) bool {
	var aiz bool = true
	for _, i := range addr {
		if i != 0 {
			aiz = false
			break
		}
	}
	return aiz
}

// NB:
// Using these tests you could replace interface-address message and just use
// fibentry - use this test for interface address routes
// 	if (msg.Id == xeth.RT_TABLE_LOCAL && msg.Type == xeth.RTN_LOCAL) ||
//		(msg.Id == xeth.RT_TABLE_MAIN && msg.Type == xeth.RTN_UNICAST) {
func ProcessFibEntry(msg *xeth.MsgFibentry, v *vnet.Vnet) (err error) {

	var isLocal bool = msg.Id == xeth.RT_TABLE_LOCAL && msg.Type == xeth.RTN_LOCAL
	var isMainUc bool = msg.Id == xeth.RT_TABLE_MAIN && msg.Type == xeth.RTN_UNICAST

	netns := xeth.Netns(msg.Net)
	rtn := xeth.Rtn(msg.Id)
	rtt := xeth.RtTable(msg.Type)
	// fwiw netlink handling also filters RTPROT_KERNEL and RTPROT_REDIRECT
	if msg.Type != xeth.RTN_UNICAST || msg.Id != xeth.RT_TABLE_MAIN {
		if isLocal {
			dbgfdb.Fib.Log(rtn, "table", rtt, msg.Prefix(),
				"in", netns)
		} else {
			dbgfdb.Fib.Log(nil, "ignore", rtn, "table", rtt,
				"in", netns)
			return
		}
	} else {
		dbgfdb.Fib.Log(rtn, "table", rtt, msg.Prefix(), "in", netns)
	}

	isDel := msg.Event == xeth.FIB_EVENT_ENTRY_DEL
	isReplace := msg.Event == xeth.FIB_EVENT_ENTRY_REPLACE

	p := ipnetToIP4Prefix(msg.Prefix())

	m := GetMain(v)
	ns := getNsByInode(m, msg.Net)
	if ns == nil {
		dbgfdb.Ns.Log("namespace", netns, "not found")
		return
	}
	nhs := ns.parseIP4NextHops(msg)
	m4 := ip4.GetMain(v)

	dbgfdb.Fib.Log(len(nhs), "nexthops for", netns)

	// Check for dummy processing
	xethNhs := msg.NextHops()
	if len(xethNhs) == 1 {
		var nhAddr ip4.Address
		copy(nhAddr[:], xethNhs[0].IP())
		if addrIsZero(nhAddr) {
			ProcessZeroGw(msg, v, ns, isDel, isLocal, isMainUc)
			return
		}
	}

	// Regular nexthop processing
	for _, nh := range nhs {
		if addrIsZero(nh.Address) {
			ProcessZeroGw(msg, v, nil, isDel, isLocal, isMainUc)
			return
		}

		dbgfdb.Fib.Log(addDelReplace(isDel, isReplace), "nexthop", nh.Address,
			"for", msg.Prefix, "in", netns)
		err = m4.AddDelRouteNextHop(&p, &nh.NextHop, isDel, isReplace)
		if err != nil {
			dbgfdb.Fib.Log(err)
			return
		}
		//This flag should only be set once on first nh because it deletes any previously set nh
		isReplace = false
	}
	return
}

func (ns *net_namespace) Ip4IfaddrMsg(m4 *ip4.Main, ipnet *net.IPNet, ifindex uint32, isDel bool) (err error) {
	p := ipnetToIP4Prefix(ipnet)
	dbgfdb.Ifa.Log(ipnet, "-->", p)
	if si, ok := ns.siForIfIndex(ifindex); ok {
		dbgfdb.Ifa.Log(addDel(isDel), "si", si)
		ns.validateFibIndexForSi(si)
		err = m4.AddDelInterfaceAddress(si, &p, isDel)
		dbgfdb.Ifa.Log(err)
	} else {
		dbgfdb.Ifa.Log("no si for ifindex:", ifindex)
	}
	return
}

func ProcessInterfaceAddr(msg *xeth.MsgIfa, action vnet.ActionType, v *vnet.Vnet) (err error) {
	if !FdbIfAddrOn {
		return
	}
	if msg == nil {
		sendFdbEventIfAddr(v)
		return
	}
	xethif := xeth.Interface.Indexed(msg.Ifindex)
	if xethif == nil {
		err = fmt.Errorf("can't find %d", msg.Ifindex)
		return
	}
	ifname := xethif.Name
	if len(ifname) == 0 {
		err = fmt.Errorf("interface %d has no name", msg.Ifindex)
		return
	}
	ifaevent := xeth.IfaEvent(msg.Event)
	switch action {
	case vnet.PreVnetd:
		// stash addresses for later use
		pe := vnet.SetPort(ifname)
		dbgfdb.Ifa.Log("PreVnetd", ifaevent, msg.IPNet(), "to", ifname)
		if msg.IsAdd() {
			pe.AddIPNet(msg.IPNet())
		} else if msg.IsDel() {
			pe.DelIPNet(msg.IPNet())
		}
	case vnet.ReadyVnetd:
		// Walk Port map and flush any IFAs we gathered at prevnetd time
		dbgfdb.Ifa.Log("ReadyVnetd", ifaevent)
		sendFdbEventIfAddr(v)

		if false {
			m := GetMain(v)
			for _, pe := range vnet.Ports {
				ns := getNsByInode(m, pe.Net)
				if ns != nil {
					dbgfdb.Ifa.Log("ReadyVnetd namespace",
						pe.Net, pe.Ifname)
					m4 := ip4.GetMain(v)
					for _, peipnet := range pe.IPNets {
						ns.Ip4IfaddrMsg(m4, peipnet, uint32(pe.Ifindex), false)
					}
				} else {
					dbgfdb.Ns.Log("ReadyVnetd namespace",
						pe.Net, "not found")
				}
			}
		}

	case vnet.PostReadyVnetd:
		dbgfdb.Ifa.Log("PostReadyVnetd", ifaevent)
		fallthrough
	case vnet.Dynamic:
		dbgfdb.Ifa.Log("Dynamic", ifaevent)
		// vnetd is up and running and received an event, so call into vnet api
		pe, found := vnet.Ports[ifname]
		if !found {
			err = fmt.Errorf("Dynamic IFA - %q unknown", ifname)
			dbgfdb.Ifa.Log(err)
			return
		}
		if FdbOn {
			if action == vnet.Dynamic {
				dbgfdb.Ifa.Log(ifname, ifaevent, msg.IPNet())
				if msg.IsAdd() {
					pe.AddIPNet(msg.IPNet())
				} else if msg.IsDel() {
					pe.DelIPNet(msg.IPNet())
				}
			}

			m := GetMain(v)
			ns := getNsByInode(m, pe.Net)
			if ns != nil {
				dbgfdb.Ns.Log("namespace", pe.Net, "found")
				m4 := ip4.GetMain(v)
				ns.Ip4IfaddrMsg(m4, msg.IPNet(), uint32(pe.Ifindex), msg.IsDel())
			} else {
				dbgfdb.Ns.Log("namespace", pe.Net, "not found")
			}
		}
	}
	return
}

func sendFdbEventIfAddr(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain
	fe := fdbm.GetEvent(vnet.PostReadyVnetd)

	for _, pe := range vnet.Ports {
		xethif := xeth.Interface.Indexed(pe.Ifindex)
		ifname := xethif.Name
		dbgfdb.Ifa.Log(ifname)

		for _, peipnet := range pe.IPNets {
			buf := xeth.Pool.Get(xeth.SizeofMsgIfa)
			msg := (*xeth.MsgIfa)(unsafe.Pointer(&buf[0]))
			msg.Kind = xeth.XETH_MSG_KIND_IFA
			msg.Ifindex = xethif.Index
			msg.Event = xeth.IFA_ADD
			msg.Address = ipnetToUint(peipnet, true)
			msg.Mask = ipnetToUint(peipnet, false)
			dbgfdb.Ifa.Log(ifname, msg.IPNet())
			ok := fe.EnqueueMsg(buf)
			if !ok {
				// filled event with messages so send event and start a new one
				fe.Signal()
				fe = fdbm.GetEvent(vnet.PostReadyVnetd)
				ok := fe.EnqueueMsg(buf)
				if !ok {
					panic("sendFdbEventIfAddr: Re-enqueue of msg failed")
				}
			}
		}
	}
	dbgfdb.Ifa.Log("sending", fe.NumMsgs, "messages")
	fe.Signal()
}

func pleaseDoAddNamepace(v *vnet.Vnet, net uint64) {
	// Ignore 1 which is default ns and created at init time
	if net == 1 {
		return
	}
	// First try and see if an existing namespace has this net number.
	// If so, just grab it. Otherwise, create a new one.
	m := GetMain(v)
	nm := &m.net_namespace_main
	if nsFound, ok := nm.namespace_by_inode[net]; ok {
		dbgfdb.Ns.Log(nsFound.name, "found for net", net)
	} else {
		name := strconv.FormatUint(net, 10)
		dbgfdb.Ns.Log("trying to add namespace", name)
		nm.addDelNamespace(name, false)
	}
}

// FIXME - need to add logic to handle a namespace that has been orphaned and needs
// to be cleaned out.
func maybeAddNamespaces(v *vnet.Vnet, net uint64) {
	// If specified find or create the namespace with inode-num "net".
	// Otherwise, we walk the PortEntry table and create namespaces
	// that we don't know about
	if net > 0 {
		dbgfdb.Ns.Log("add single ns for", net)
		pleaseDoAddNamepace(v, net)
	} else {
		// March through all port-entries.
		// If we haven't seen a Net before we need to create a net_namespace
		for _, pe := range vnet.Ports {
			dbgfdb.Ns.Log("ReadyVnetd add", pe.Net, "for", pe.Ifname)
			pleaseDoAddNamepace(v, pe.Net)
		}
	}
}

func getNsByInode(m *Main, netNum uint64) *net_namespace {
	if netNum == 1 {
		return &m.default_namespace
	} else {
		return m.namespace_by_inode[netNum]
	}
}

func makePortEntry(msg *xeth.MsgIfinfo, puntIndex uint8) (pe *vnet.PortEntry) {
	ifname := xeth.Ifname(msg.Ifname)
	pe = vnet.SetPort(ifname.String())
	pe.Net = msg.Net
	pe.Ifindex = msg.Ifindex
	pe.Iflinkindex = msg.Iflinkindex
	pe.Ifname = ifname.String()
	vnet.SetPortByIndex(msg.Ifindex, pe.Ifname)
	pe.Iff = net.Flags(msg.Flags)
	pe.Vid = msg.Id
	copy(pe.Addr[:], msg.Addr[:])
	pe.Portindex = msg.Portindex
	// -1 is unspecified - from driver
	if msg.Subportindex >= 0 {
		pe.Subportindex = msg.Subportindex
	}
	pe.PuntIndex = puntIndex
	pe.Devtype = msg.Devtype
	return
}

func isSpecialDevtype(msg *xeth.MsgIfinfo) bool {
	if msg.Devtype == xeth.XETH_DEVTYPE_LINUX_VLAN ||
		msg.Devtype == xeth.XETH_DEVTYPE_XETH_BRIDGE {
		return true
	}
	// Hack until driver sends correct devtype - REMOVE
	ifname := xeth.Ifname(msg.Ifname)
	if strings.Contains(ifname.String(), "xethbr") {
		return true
	}
	return false
}

func ProcessInterfaceInfo(msg *xeth.MsgIfinfo, action vnet.ActionType, v *vnet.Vnet, puntIndex uint8) (err error) {
	if msg == nil {
		sendFdbEventIfInfo(v)
		return
	}
	kind := xeth.Kind(msg.Kind)
	ifname := (*xeth.Ifname)(&msg.Ifname).String()
	ifindex := uint32(msg.Ifindex)
	reason := xeth.IfinfoReason(msg.Reason)
	netns := xeth.Netns(msg.Net)
	switch action {
	case vnet.PreVnetd:
		makePortEntry(msg, puntIndex)
		dbgfdb.Ifinfo.Log("Prevnetd", kind)

	case vnet.ReadyVnetd:
		// Walk Port map and flush into vnet/fe layers the interface info we gathered
		// at prevnetd time. Both namespace and interface creation messages sent during this processing.
		dbgfdb.Ifinfo.Log("ReadyVnetd add", ifname)
		// Signal that all namespaces are now initialized??
		sendFdbEventIfInfo(v)

	case vnet.PostReadyVnetd:
		fallthrough
	case vnet.Dynamic:
		dbgfdb.Ifinfo.Log(action, kind, ifname, netns)
		m := GetMain(v)
		ns := getNsByInode(m, msg.Net)
		if ns == nil {
			dbgfdb.Ns.Log("namespace", netns, "not found")
			return
		}
		dbgfdb.Ns.Log(action, "namespace", ns.name, "found for", ifname)
		pe := vnet.GetPortByIndex(msg.Ifindex)
		if pe == nil {
			// If a vlan interface we allow dynamic creation so create a cached entry
			if msg.Devtype >= xeth.XETH_DEVTYPE_LINUX_UNKNOWN {
				dbgfdb.Ifinfo.Log("Creating Ports entry for",
					ifname, msg.Ifindex, msg.Net)
				pe = makePortEntry(msg, puntIndex)
			} else {
				dbgfdb.Ifinfo.Log("pe is nil - returning")
				return
			}
		}
		if msg.Net != pe.Net {
			// This ifindex has been set into a new namespace so
			// 1. Remove ifindex from previous namespace
			// 2. Add ifindex to new namespace
			nsOld := getNsByInode(m, pe.Net)
			if nsOld == nil {
				// old namespace already removed
				dbgfdb.Ns.Log("Couldn't find old ns:", pe.Net)
			} else {
				nsOld.addDelMk1Interface(m, true, ifname,
					uint32(msg.Ifindex), msg.Addr, msg.Devtype, msg.Iflinkindex, msg.Id)
			}

			ns.addDelMk1Interface(m, false, ifname,
				uint32(msg.Ifindex), msg.Addr, msg.Devtype, msg.Iflinkindex, msg.Id)

			dbgfdb.Ifinfo.Log("moving", ifname, pe.Net, netns)
			pe.Net = msg.Net
		} else if action == vnet.PostReadyVnetd {
			// Goes has restarted with interfaces already in existent namespaces,
			// so create vnet representation of interface in this ns.
			// Or this is a dynamically created vlan interface.
			dbgfdb.Ifinfo.Log(action, ifname, netns)
			ns.addDelMk1Interface(m, false, ifname,
				uint32(msg.Ifindex), msg.Addr, msg.Devtype,
				msg.Iflinkindex, msg.Id)
		} else if msg.Devtype >= xeth.XETH_DEVTYPE_LINUX_UNKNOWN {
			// create or delete interfaces based on reg/unreg reason
			dbgfdb.Ifinfo.Log(ifname, reason, "vlan", netns)
			isUnreg := reason == xeth.XETH_IFINFO_REASON_UNREG
			ns.addDelMk1Interface(m, isUnreg, ifname,
				uint32(msg.Ifindex), msg.Addr, msg.Devtype,
				msg.Iflinkindex, msg.Id)
			if reason == xeth.XETH_IFINFO_REASON_UNREG {
				return
			}
		}
		if ns.interface_by_index[ifindex] != nil {
			// Process admin-state flags
			if si, ok := ns.siForIfIndex(ifindex); ok {
				ns.validateFibIndexForSi(si)
				flags := net.Flags(msg.Flags)
				isUp := flags&net.FlagUp == net.FlagUp
				err = dbgfdb.Ifinfo.Log(si.SetAdminUp(v, isUp))
			} else {
				dbgfdb.Si.Log("can't get si of", ifname)
			}
		} else {
			// NB: This is the dynamic front-panel-port-creation case which our lower layers
			// don't support yet. Driver does not send us these but here as a placeholder.
			dbgfdb.Ifinfo.Log("Attempting dynamic port-creation of", ifname)
			if false {
				if action == vnet.Dynamic {
					_, found := vnet.Ports[ifname]
					if !found {
						pe := vnet.SetPort(ifname)
						dbgfdb.Ifinfo.Log("setting",
							ifname, "in", netns)
						pe.Net = msg.Net
						pe.Ifindex = msg.Ifindex
						pe.Iflinkindex = msg.Iflinkindex
						pe.Ifname = ifname
						vnet.SetPortByIndex(msg.Ifindex, pe.Ifname)
						pe.Iff = net.Flags(msg.Flags)
						pe.Vid = msg.Id
						copy(pe.Addr[:], msg.Addr[:])
						pe.Portindex = msg.Portindex
						pe.Subportindex = msg.Subportindex
						pe.PuntIndex = puntIndex
					}
				}
				ns.addDelMk1Interface(m, false, ifname,
					uint32(msg.Ifindex), msg.Addr,
					msg.Devtype, msg.Iflinkindex,
					msg.Id)
			}
		}
	}
	return nil
}

func sendFdbEventIfInfo(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain
	fe := fdbm.GetEvent(vnet.PostReadyVnetd)
	xeth.Interface.Iterate(func(entry *xeth.InterfaceEntry) error {
		dbgfdb.Ifinfo.Log(entry.Name)
		buf := xeth.Pool.Get(xeth.SizeofMsgIfinfo)
		msg := (*xeth.MsgIfinfo)(unsafe.Pointer(&buf[0]))
		msg.Kind = xeth.XETH_MSG_KIND_IFINFO
		copy(msg.Ifname[:], entry.Name)
		msg.Ifindex = entry.Index
		msg.Iflinkindex = entry.Link
		copy(msg.Addr[:], entry.HardwareAddr())
		msg.Net = uint64(entry.Netns)
		msg.Id = entry.Id
		msg.Portindex = entry.Port
		msg.Subportindex = entry.Subport
		msg.Flags = uint32(entry.Flags)
		msg.Devtype = uint8(entry.DevType)
		ok := fe.EnqueueMsg(buf)
		if !ok {
			// filled event with messages so send event and start a new one
			fe.Signal()
			fe = fdbm.GetEvent(vnet.PostReadyVnetd)
			ok := fe.EnqueueMsg(buf)
			if !ok {
				panic("sendFdbEventIfInfo: Re-enqueue of msg failed")
			}
		}
		return nil
	})
	fe.Signal()
}

func ipnetToUint(ipnet *net.IPNet, ipNotMask bool) uint32 {
	if ipNotMask {
		return *(*uint32)(unsafe.Pointer(&ipnet.IP[0]))
	} else {
		return *(*uint32)(unsafe.Pointer(&ipnet.Mask[0]))
	}
}

func InitInterfaceEthtool(v *vnet.Vnet) {
	sendFdbEventEthtoolSettings(v)
	sendFdbEventEthtoolFlags(v)
}

func sendFdbEventEthtoolSettings(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain
	fe := fdbm.GetEvent(vnet.PostReadyVnetd)
	for _, pe := range vnet.Ports {
		xethif := xeth.Interface.Indexed(pe.Ifindex)
		ifindex := xethif.Ifinfo.Index
		ifname := xethif.Ifinfo.Name
		if xethif.Ifinfo.DevType != xeth.XETH_DEVTYPE_XETH_PORT {
			continue
		}
		dbgfdb.Ifinfo.Log(ifname, pe)
		buf := xeth.Pool.Get(xeth.SizeofMsgEthtoolSettings)
		msg := (*xeth.MsgEthtoolSettings)(unsafe.Pointer(&buf[0]))
		msg.Kind = xeth.XETH_MSG_KIND_ETHTOOL_SETTINGS
		msg.Ifindex = ifindex
		msg.Speed = uint32(pe.Speed)
		msg.Autoneg = pe.Autoneg
		// xeth layer is cacheing the rest of this message
		// in future can just reference that and send it along here
		ok := fe.EnqueueMsg(buf)
		if !ok {
			// filled event with messages so send event and start a new one
			fe.Signal()
			fe = fdbm.GetEvent(vnet.PostReadyVnetd)
			ok := fe.EnqueueMsg(buf)
			if !ok {
				panic("sendFdbEventEthtoolSettings: Re-enqueue of msg failed")
			}
		}
	}
	fe.Signal()
}

func sendFdbEventEthtoolFlags(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain
	fe := fdbm.GetEvent(vnet.PostReadyVnetd)
	for _, pe := range vnet.Ports {
		xethif := xeth.Interface.Indexed(pe.Ifindex)
		ifindex := xethif.Ifinfo.Index
		ifname := xethif.Ifinfo.Name
		if xethif.Ifinfo.DevType != xeth.XETH_DEVTYPE_XETH_PORT {
			continue
		}
		dbgfdb.Ifinfo.Log(ifname, pe)
		buf := xeth.Pool.Get(xeth.SizeofMsgEthtoolFlags)
		msg := (*xeth.MsgEthtoolFlags)(unsafe.Pointer(&buf[0]))
		msg.Kind = xeth.XETH_MSG_KIND_ETHTOOL_FLAGS
		msg.Ifindex = ifindex
		msg.Flags = uint32(pe.Flags)
		// xeth layer is cacheing the rest of this message
		// in future can just reference that and send it along here
		ok := fe.EnqueueMsg(buf)
		if !ok {
			// filled event with messages so send event and start a new one
			fe.Signal()
			fe = fdbm.GetEvent(vnet.PostReadyVnetd)
			ok := fe.EnqueueMsg(buf)
			if !ok {
				panic("sendFdbEventEthtoolFlags: Re-enqueue of msg failed")
			}
		}
	}
	fe.Signal()
}

func addDel(isDel bool) string {
	if isDel {
		return "del"
	}
	return "add"
}

func addDelReplace(isDel, isReplace bool) string {
	if isReplace {
		return "replace"
	} else if isDel {
		return "del"
	}
	return "add"
}
