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
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/internal/xeth"
	"github.com/platinasystems/go/vnet"
	"github.com/platinasystems/go/vnet/ethernet"
	"github.com/platinasystems/go/vnet/ip"
	"github.com/platinasystems/go/vnet/ip4"

	"fmt"
	"net"
	"reflect"
	"strconv"
	"sync"
	"syscall"
	"unsafe"
)

var FdbOn bool
var IfinfoDebug bool = true
var IfaddrDebug bool
var FibentryDebug bool
var NeighDebug bool

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
	e.Msgs[e.NumMsgs] = append(e.Msgs[e.NumMsgs], msg...)
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

	// Send events for initial dump of fib entries
	fe := fdbm.GetEvent(vnet.Dynamic)
	vnet.Xeth.DumpFib()
	vnet.Xeth.UntilBreak(func(buf []byte) error {
		ok := fe.EnqueueMsg(buf)
		if !ok {
			// filled event with messages so send event and start a new one
			fe.Signal()
			fe = fdbm.GetEvent(vnet.Dynamic)
			ok := fe.EnqueueMsg(buf)
			if !ok {
				panic("Initial DumpFib: Re-enqueue of msg failed")
			}
		}
		return nil
	}, false)
	fe.Signal()

	// Start go routine that drains XETH socket and shuttles msgs over to vnetd loop
	// for subsequent processing of FIB, IFA and IFINFO updates
	// Can optimize this so multiple messages are sent.
	go func() {
		m := GetMain(v)
		fdbm := &m.FdbMain
		for buf := range vnet.Xeth.RxCh {
			fe := fdbm.GetEvent(vnet.Dynamic)
			msgSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&buf[0])),
				Len:  xeth.PageSize,
				Cap:  xeth.PageSize,
			}))
			ok := fe.EnqueueMsg(msgSlice)
			if !ok {
				panic("Xeth.RxCh reader: Can't enqueue to fresh event!")
			}
			fe.Signal()
		}
	}()
}

func (e *fdbEvent) EventAction() {
	var err error
	m := e.fm
	vn := m.m.v

	fmt.Println("fdb EventAction:", e.NumMsgs)

	for i := 0; i < e.NumMsgs; i++ {

		msg := e.Msgs[i]
		if true {
			fmt.Printf("********* fdb message type: %d\n", xeth.KindOf(msg))
		}

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
			ifname := xeth.Ifname(msg.Ifname)
			vnet.SetPort(ifname.String()).Flags =
				xeth.EthtoolFlagBits(msg.Flags)
			fec91 := vnet.PortIsFec91(ifname.String())
			fec74 := vnet.PortIsFec74(ifname.String())
			if false {
				fmt.Printf("XETH_MSG_KIND_ETHTOOL_FLAGS: %s Fec91 %v\n",
					ifname.String(), fec91)
				fmt.Printf("XETH_MSG_KIND_ETHTOOL_FLAGS: %s Fec74 %v\n",
					ifname.String(), fec74)
			}
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
			if vnet.PortIsCopper(ifname.String()) {
				media = "copper"
			}
			hi, found := vn.HwIfByName(ifname.String())
			if found {
				hi.SetMedia(vn, media)
				err = ethernet.SetInterfaceErrorCorrection(vn, hi, fec)
				if err != nil {
					fmt.Printf("XETH_MSG_KIND_ETHTOOL_FLAGS: %s Error setting fec: %v\n",
						ifname.String(), err)
				}
			}
		case xeth.XETH_MSG_KIND_ETHTOOL_SETTINGS:
			msg := (*xeth.MsgEthtoolSettings)(ptr)
			ifname := xeth.Ifname(msg.Ifname)
			vnet.SetPort(ifname.String()).Speed =
				xeth.Mbps(msg.Speed)
			hi, found := vn.HwIfByName(ifname.String())
			if found {
				bw := float64(msg.Speed)
				speedOk := false
				if false {
					fmt.Printf("xeth.XETH_MSG_KIND_ETHTOOL_SETTINGS processing speed %v\n", bw)
				}
				switch bw {
				case 0, 1000, 10000, 20000, 25000, 40000, 50000, 100000:
					speedOk = true
				}
				if !speedOk {
					err = fmt.Errorf("speed expected 0|1000|10000|20000|25000|40000|50000|100000 got %v",
						bw)
					fmt.Println("SetSpeed error1:", err)
				} else {
					bw *= 1e6
					err = hi.SetSpeed(vn, vnet.Bandwidth(bw))
					if err != nil {
						fmt.Println("SetSpeed error2:", err)
					}

				}
			}

		}
		if err != nil {
			fmt.Println("Error processing msg from driver:", err)
		}
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

	if FibentryDebug {
		if ns == nil {
			fmt.Println("parseIP4NextHops: ns nil!")
		} else {
			fmt.Println("parseIP4NextHops: ns", ns.name)
		}
	}
	xethNhs := msg.NextHops()
	if FibentryDebug {
		fmt.Println("parseIP4NextHops: length nhs:", len(xethNhs))
		for i, _ := range xethNhs {
			fmt.Printf("parseIP4NextHops: [%d] %s\n", i, xethNhs[i].String())
		}
	}

	// If only 1 nh then assume this is single OIF nexthop
	// otherwise it's multipath
	nh := ip4_next_hop{}
	nh.Weight = 1
	if len(xethNhs) == 1 {
		nh.intf = ns.interface_by_index[uint32(xethNhs[0].Ifindex)]
		if nh.intf == nil {
			fmt.Println("parseIP4NextHops: Can't find ns-intf for ifindex", xethNhs[0].Ifindex)
			return
		}
		nh.Si = nh.intf.si
		copy(nh.Address[:], xethNhs[0].IP())
		nhs = append(nhs, nh)
	} else {
		for _, xnh := range xethNhs {
			intf := ns.interface_by_index[uint32(xnh.Ifindex)]
			if intf == nil {
				fmt.Println("parseIP4NextHops: Can't find ns-intf for ifindex", xnh.Ifindex)
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

	if NeighDebug {
		fmt.Println(msg)
	}
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
	if ns != nil {
		si, ok := ns.siForIfIndex(uint32(msg.Ifindex))
		if !ok {
			fmt.Println("ProcessIpNeighbor: Can't find si in namespace", msg.Ifindex, ns.name)
			return
		}
		nbr := ethernet.IpNeighbor{
			Si:       si,
			Ethernet: ethernet.Address(msg.Lladdr),
			Ip:       ip.Address(msg.Dst),
		}
		m4 := ip4.GetMain(v)
		em := ethernet.GetMain(v)
		if NeighDebug {
			fmt.Printf("ProcessIpNeighbor: nbr (%v) isDel %v\n", nbr, isDel)
		}
		_, err = em.AddDelIpNeighbor(&m4.Main, &nbr, isDel)

		// Ignore delete of unknown neighbor.
		if err == ethernet.ErrDelUnknownNeighbor {
			err = nil
		}
	} else {
		fmt.Println("ProcessIpNeighbor: Can't find namespace", msg.Net)
	}
	return
}

// NB:
// Using these tests you could replace interface-address message and just use
// fibentry - use this test for interface address routes
// 	if (msg.Id == xeth.RT_TABLE_LOCAL && msg.Type == xeth.RTN_LOCAL) ||
//		(msg.Id == xeth.RT_TABLE_MAIN && msg.Type == xeth.RTN_UNICAST) {
func ProcessFibEntry(msg *xeth.MsgFibentry, v *vnet.Vnet) (err error) {

	if FibentryDebug {
		fmt.Printf("MsgFibentry: ns net %d\n", msg.Net)
	}
	// fwiw netlink handling also filters RTPROT_KERNEL and RTPROT_REDIRECT
	if msg.Type != xeth.RTN_UNICAST || msg.Id != xeth.RT_TABLE_MAIN {
		fmt.Println("ProcessFibEntry: Msg type not unicast and table main", msg.Type, msg.Id)
		return
	}

	isDel := msg.Event == xeth.FIB_EVENT_ENTRY_DEL
	isReplace := msg.Event == xeth.FIB_EVENT_ENTRY_REPLACE

	p := ipnetToIP4Prefix(msg.Prefix())

	m := GetMain(v)
	ns := getNsByInode(m, msg.Net)
	if ns != nil {
		nhs := ns.parseIP4NextHops(msg)
		m4 := ip4.GetMain(v)

		if FibentryDebug {
			fmt.Println("ProcessFibEntry: parsed nexthops for ns", ns.name)
		}

		for _, nh := range nhs {
			// For now process normal nexthops -
			// do something special for tunnels later.
			// Ignore routes with zero nexthops - these represent connected local
			// routes that are taken care of with interface-address message processing.
			gw := nh.Address
			var gwIsZero bool = true
			for _, i := range gw {
				if i != 0 {
					gwIsZero = false
					break
				}
			}
			if gwIsZero {
				return
			}
			if FibentryDebug {
				fmt.Printf("RouteMsg for Prefix %v adding nexthop %v isDel %v isReplace %v\n",
					p, gw, isDel, isReplace)
			}
			if err = m4.AddDelRouteNextHop(&p, &nh.NextHop, isDel, isReplace); err != nil {
				fmt.Println("AddDelRouteNextHop: returns err:", err)
				return
			}
			//This flag should only be set once on first nh because it deletes any previously set nh
			isReplace = false
		}
	} else {
		if FibentryDebug {
			fmt.Println("ProcessFibEntry: namespace not found for Net", msg.Net)
			fmt.Println("ProcessFibEntry: Message contents", msg.String())
		}
	}
	return
}

func (ns *net_namespace) Ip4IfaddrMsg(m4 *ip4.Main, ipnet *net.IPNet, ifindex uint32, isDel bool) (err error) {
	// FIXME - need to add dummy interface processing here.
	p := ipnetToIP4Prefix(ipnet)
	if IfaddrDebug {
		fmt.Printf("Ip4IfaddrMsg: %v --> %v\n", ipnet, p)
	}
	if si, ok := ns.siForIfIndex(ifindex); ok {
		if IfaddrDebug {
			fmt.Printf("Ip4IfaddrMsg: si %d isDel %v\n", si, isDel)
		}
		ns.validateFibIndexForSi(si)
		err = m4.AddDelInterfaceAddress(si, &p, isDel)
	} else {
		if IfaddrDebug {
			fmt.Println("Ip4IfaddrMsg: si not found for ifindex:", ifindex)
		}
	}
	return
}

func ProcessInterfaceAddr(msg *xeth.MsgIfa, action vnet.ActionType, v *vnet.Vnet) (err error) {

	var ifname xeth.Ifname
	switch action {
	case vnet.PreVnetd:
		// stash addresses for later use
		ifname = xeth.Ifname(msg.Ifname)
		pe := vnet.SetPort(ifname.String())
		if IfaddrDebug {
			fmt.Println("XETH_MSG_KIND_IFA: PreVnetd:", ifname.String())
		}
		if msg.IsAdd() {
			if IfaddrDebug {
				fmt.Println("XETH_MSG_KIND_IFA: PreVnetd Add", ifname.String(), msg.Event, msg.Prefix())
			}
			pe.AddIPNet(msg.Prefix())
		} else if msg.IsDel() {
			if IfaddrDebug {
				fmt.Println("XETH_MSG_KIND_IFA: PreVnetd Del", ifname.String(), msg.Event, msg.Prefix())
			}
			pe.DelIPNet(msg.Prefix())
		}
	case vnet.ReadyVnetd:
		// Walk Port map and flush any IFAs we gathered at prevnetd time
		if IfaddrDebug {
			fmt.Println("XETH_MSG_KIND_IFA: ReadyVnetd Add")
		}
		sendFdbEventIfAddr(v)

	case vnet.PostReadyVnetd:
		if IfaddrDebug {
			fmt.Println("XETH_MSG_KIND_IFA: PostReadyVnetd Add", msg.IsAdd())
		}
		fallthrough
	case vnet.Dynamic:
		if IfaddrDebug {
			fmt.Println("XETH_MSG_KIND_IFA: Dynamic Add", msg.IsAdd())
		}
		// vnetd is up and running and received an event, so call into vnet api
		ifname = xeth.Ifname(msg.Ifname)
		pe, found := vnet.Ports[ifname.String()]
		if !found {
			if IfaddrDebug {
				fmt.Println("XETH_MSG_KIND_IFA :", ifname.String(), "not found!")
			}
			err = fmt.Errorf("Dynamic IFA - Ifname unknown:", ifname.String())
			return
		}
		if FdbOn {
			if msg.IsAdd() {
				if IfaddrDebug {
					fmt.Println("XETH_MSG_KIND_IFA: Add", ifname.String(),
						msg.Event, msg.Prefix())
				}
				pe.AddIPNet(msg.Prefix())
			} else if msg.IsDel() {
				if IfaddrDebug {
					fmt.Println("XETH_MSG_KIND_IFA: Del", ifname.String(),
						msg.Event, msg.Prefix())
				}
				pe.DelIPNet(msg.Prefix())
			}

			m := GetMain(v)
			ns := getNsByInode(m, pe.Net)
			if ns != nil {
				if IfaddrDebug {
					fmt.Println("XETH_MSG_KIND_IFA- : Namespace found", pe.Net)
				}
				m4 := ip4.GetMain(v)
				ns.Ip4IfaddrMsg(m4, msg.Prefix(), uint32(pe.Ifindex), msg.IsDel())
			} else {
				if IfaddrDebug {
					fmt.Println("XETH_MSG_KIND_IFA- : Namespace not found", pe.Net)
				}
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
		var ifname [16]uint8
		copy(ifname[:], pe.Ifname)
		if IfaddrDebug {
			fmt.Println("sendFdbEventIfAddr:", pe.Ifname)
		}

		for _, peipnet := range pe.IPNets {
			msg := xeth.MsgIfa{
				Kind:    xeth.XETH_MSG_KIND_IFA,
				Ifname:  ifname,
				Event:   xeth.IFA_ADD,
				Address: ipnetToUint(peipnet, true),
				Mask:    ipnetToUint(peipnet, false),
			}
			if IfaddrDebug {
				ifn := xeth.Ifname(msg.Ifname)
				fmt.Println("sendFdbEventIfAddr:", ifn.String(), msg.Address, msg.Mask)
			}
			msgSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
				Data: uintptr(unsafe.Pointer(&msg)),
				Len:  xeth.PageSize,
				Cap:  xeth.PageSize,
			}))

			ok := fe.EnqueueMsg(msgSlice)
			if !ok {
				// filled event with messages so send event and start a new one
				fe.Signal()
				fe = fdbm.GetEvent(vnet.PostReadyVnetd)
				ok := fe.EnqueueMsg(msgSlice)
				if !ok {
					panic("sendFdbEventIfAddr: Re-enqueue of msg failed")
				}
			}
		}
	}
	if IfaddrDebug {
		fmt.Println("sendFdbEventIfAddr sending nummsgs:", fe.NumMsgs)
	}
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
		fmt.Println("Found namespace for net", net, nsFound.name)
	} else {
		name := strconv.FormatUint(net, 10)
		fmt.Println("pleaseDoAddNamepace: Trying to add namespace", name)
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
		if IfinfoDebug {
			fmt.Println("XETH_MSG_KIND_IFINFO: maybeAddNamespaces Add single ns:", net)
		}
		pleaseDoAddNamepace(v, net)
	} else {
		// March through all port-entries.
		// If we haven't seen a Net before we need to create a net_namespace
		for _, pe := range vnet.Ports {
			if IfinfoDebug {
				fmt.Println("XETH_MSG_KIND_IFINFO: maybeAddNamespaces ReadyVnetd Add", pe.Ifname)
			}
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

func ProcessInterfaceInfo(msg *xeth.MsgIfinfo, action vnet.ActionType, v *vnet.Vnet, puntIndex uint8) (err error) {

	switch action {
	case vnet.PreVnetd:
		ifname := xeth.Ifname(msg.Ifname)
		pe := vnet.SetPort(ifname.String())
		pe.Net = msg.Net
		pe.Ifindex = msg.Ifindex
		pe.Iflinkindex = msg.Iflinkindex
		pe.Ifname = ifname.String()
		vnet.SetPortByIndex(msg.Ifindex, pe.Ifname)
		pe.Iff = xeth.Iff(msg.Flags)
		pe.Vid = msg.Id
		copy(pe.Addr[:], msg.Addr[:])
		pe.Portindex = msg.Portindex
		pe.Subportindex = msg.Subportindex
		pe.PuntIndex = puntIndex
		if IfinfoDebug {
			fmt.Println("XETH_MSG_KIND_IFINFO: Prevnetd:", ifname.String(), msg.Ifindex, msg.Portindex, msg.Subportindex, msg.Devtype, msg.Id)
		}

	case vnet.ReadyVnetd:
		// Walk Port map and flush into vnet/fe layers the interface info we gathered
		// at prevnetd time. Both namespace and interface creation messages sent during this processing.
		if IfinfoDebug {
			fmt.Println("XETH_MSG_KIND_IFINFO: ReadyVnetd Add")
		}
		maybeAddNamespaces(v, 0)
		// Signal that all namespaces are now initialized??
		sendFdbEventIfInfo(v)

	case vnet.PostReadyVnetd:
		fallthrough
	case vnet.Dynamic:
		ifname := xeth.Ifname(msg.Ifname)
		if IfinfoDebug {
			fmt.Println("XETH_MSG_KIND_IFINFO:", action, msg)
		}
		// FIXME Add support for Dynamic cases known:
		// - Addition of vlans or dummy interfaces (DO NOT SUPPORT dynamic addition of fp ports)
		// - Need to cleanup old namespaces when no ports reference them
		//
		// FIXME - no way to distinguish initial port creations with dynamic attempts (ie when vnetd is up)
		//
		m := GetMain(v)
		maybeAddNamespaces(v, msg.Net)
		ns := getNsByInode(m, msg.Net)
		if ns != nil {
			if IfinfoDebug {
				fmt.Println("XETH_MSG_KIND_IFINFO-: Found namespace:", action, ns.name, msg.Ifindex)
			}
			pe := vnet.GetPortByIndex(msg.Ifindex)
			if pe == nil {
				fmt.Println("XETH_MSG_KIND_IFINFO: pe is nil")
				return
			}
			if msg.Net != pe.Net {
				// This ifindex has been set into a new namespace so
				// 1. Remove ifindex from previous namespace
				// 2. Add ifindex to new namespace
				nsOld := getNsByInode(m, pe.Net)
				if nsOld == nil {
					fmt.Println("XETH_MSG_KIND_IFINFO-: Couldn't find old ns:", pe.Net)
					return
				}
				nsOld.addDelMk1Interface(m, true, ifname.String(),
					uint32(msg.Ifindex), msg.Addr)

				ns.addDelMk1Interface(m, false, ifname.String(),
					uint32(msg.Ifindex), msg.Addr)

				pe.Net = msg.Net
				if IfinfoDebug {
					fmt.Println("XETH_MSG_KIND_IFINFO-Moving", ifname.String(), pe.Net, msg.Net)
				}
			} else if msg.Net == pe.Net && action == vnet.PostReadyVnetd {
				// Goes has restarted with interfaces already in existent namespaces,
				// so create vnet representation of interface in this ns.
				if IfinfoDebug {
					fmt.Println("XETH_MSG_KIND_IFINFO - PostReadyVnetd", ifname.String(), msg.Net)
				}
				ns.addDelMk1Interface(m, false, ifname.String(),
					uint32(msg.Ifindex), msg.Addr)
			}
			ifindex := uint32(msg.Ifindex)
			if ns.interface_by_index[ifindex] != nil {
				// Process admin-state flags
				if si, ok := ns.siForIfIndex(ifindex); ok {
					ns.validateFibIndexForSi(si)
					isUp := msg.Flags&xeth.IFF_UP == xeth.IFF_UP
					err = si.SetAdminUp(v, isUp)
					if err != nil {
						fmt.Println("ProcessInterfaceInfo: admin up/down", err)
					} else {
						ifname := xeth.Ifname(msg.Ifname)
						if IfinfoDebug {
							fmt.Printf("ProcessInterfaceInfo: %s isUp %v flags (%s)",
								ifname.String(), isUp, xeth.Iff(msg.Flags))
						}
					}
				} else {
					fmt.Println("ProcessInterfaceInfo: admin up/down: Can't get si!",
						ifindex, ns.name)
				}
			} else {
				if action == vnet.Dynamic {
					ifname := xeth.Ifname(msg.Ifname)
					_, found := vnet.Ports[ifname.String()]
					if !found {
						pe := vnet.SetPort(ifname.String())
						if IfinfoDebug {
							fmt.Println("ProcessInterfaceInfo: Setting into", ifname.String(),
								pe.Net, msg.Net)
						}
						pe.Net = msg.Net
						pe.Ifindex = msg.Ifindex
						pe.Iflinkindex = msg.Iflinkindex
						pe.Ifname = ifname.String()
						vnet.SetPortByIndex(msg.Ifindex, pe.Ifname)
						pe.Iff = xeth.Iff(msg.Flags)
						pe.Vid = msg.Id
						copy(pe.Addr[:], msg.Addr[:])
						pe.Portindex = msg.Portindex
						pe.Subportindex = msg.Subportindex
						pe.PuntIndex = puntIndex
					}
				}
				ns.addDelMk1Interface(m, false, ifname.String(),
					uint32(msg.Ifindex), msg.Addr)
			}
		} else {
			if IfinfoDebug {
				fmt.Printf("XETH_MSG_KIND_IFINFO- : Namespace not found %d\n", msg.Net)
			}
		}
	}
	return nil
}

func sendFdbEventIfInfo(v *vnet.Vnet) {
	m := GetMain(v)
	fdbm := &m.FdbMain
	fe := fdbm.GetEvent(vnet.PostReadyVnetd)
	for _, pe := range vnet.Ports {
		var ifname [16]uint8
		copy(ifname[:], pe.Ifname)
		if IfinfoDebug {
			fmt.Println("sendFdbEventIfInfo:", pe.Ifname, pe)
		}
		msg := xeth.MsgIfinfo{
			Kind:         xeth.XETH_MSG_KIND_IFINFO,
			Ifname:       ifname,
			Ifindex:      pe.Ifindex,
			Iflinkindex:  pe.Iflinkindex,
			Addr:         pe.Addr,
			Net:          pe.Net,
			Portindex:    pe.Portindex,
			Subportindex: pe.Subportindex,
			Flags:        uint32(pe.Iff),
			Id:           pe.Vid,
		}
		msgSlice := *(*[]byte)(unsafe.Pointer(&reflect.SliceHeader{
			Data: uintptr(unsafe.Pointer(&msg)),
			Len:  xeth.PageSize,
			Cap:  xeth.PageSize,
		}))
		ok := fe.EnqueueMsg(msgSlice)
		if !ok {
			// filled event with messages so send event and start a new one
			fe.Signal()
			fe = fdbm.GetEvent(vnet.PostReadyVnetd)
			ok := fe.EnqueueMsg(msgSlice)
			if !ok {
				panic("sendFdbEventIfInfo: Re-enqueue of msg failed")
			}
		}
	}
	fe.Signal()
}

func ipnetToUint(ipnet *net.IPNet, ipNotMask bool) uint32 {
	if ipNotMask {
		return *(*uint32)(unsafe.Pointer(&ipnet.IP[0]))
	} else {
		return *(*uint32)(unsafe.Pointer(&ipnet.Mask[0]))
	}
}
