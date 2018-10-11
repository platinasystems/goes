// Copyright 2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package vnet

import (
	"net"

	"github.com/platinasystems/go/elib"
	"github.com/platinasystems/go/elib/cpu"
	"github.com/platinasystems/go/elib/dep"
	"github.com/platinasystems/go/elib/loop"
	"github.com/platinasystems/go/elib/parse"
	"github.com/platinasystems/xeth"
)

//Debug Flags
var LogSvi bool
var AdjDebug bool

// drivers/net/ethernet/xeth/platina_mk1.c: xeth.MsgIfinfo
// PortEntry go/main/goes-platina-mk1/vnetd.go:vnetdInit() xeth.XETH_MSG_KIND_IFINFO
//   also BridgeEntry, BridgeMemberEntry
// PortProvision go/main/goes-platina-mk1/vnetd.go:parsePortConfig() from entry Ports
// BridgeProvision parseBridgeConfig() from entry Bridges,BridgeMembers
// PortConfig fe1/platform.go:parsePortConfig() PortProvision
// BridgeConfig parseBridgeConfig() BridgeProvision
// Port fe1/internal/fe1a/port_init.go:PortInit()
type PortEntry struct {
	Net          uint64
	Ifindex      int32
	Iflinkindex  int32 // system side eth# ifindex
	Ifname       string
	Flags        xeth.EthtoolPrivFlags
	Iff          net.Flags
	Speed        xeth.Mbps
	Vid          uint16 // port_vid
	Portindex    int16
	Subportindex int8
	PuntIndex    uint8 // 0-based meth#, derived from Iflinkindex
	Devtype      uint8
	Addr         [xeth.ETH_ALEN]uint8
	IPNets       []*net.IPNet
}

var Ports map[string]*PortEntry
var PortsByIndex map[int32]*PortEntry // FIXME - driver only sends platina-mk1 type

type BridgeEntry struct {
	Net         uint64
	Ifindex     int32
	Iflinkindex int32 // system side eth# ifindex
	PuntIndex   uint8
	Addr        [xeth.ETH_ALEN]uint8
	IPNets      []*net.IPNet
}

// indexed by customer vid
var Bridges map[uint16]*BridgeEntry

type BridgeMemberEntry struct {
	Vid      uint16
	IsTagged bool
	PortVid  uint16
}

var BridgeMembers map[string]*BridgeMemberEntry

func SetBridge(vid uint16) *BridgeEntry {
	if Bridges == nil {
		Bridges = make(map[uint16]*BridgeEntry)
	}
	entry, found := Bridges[vid]
	if !found {
		entry = new(BridgeEntry)
		Bridges[vid] = entry
	}
	return entry
}

func SetBridgeMember(ifname string) *BridgeMemberEntry {
	if BridgeMembers == nil {
		BridgeMembers = make(map[string]*BridgeMemberEntry)
	}
	entry, found := BridgeMembers[ifname]
	if !found {
		entry = new(BridgeMemberEntry)
		BridgeMembers[ifname] = entry
	}
	return entry
}

func SetPort(ifname string) *PortEntry {
	if Ports == nil {
		Ports = make(map[string]*PortEntry)
	}
	entry, found := Ports[ifname]
	if !found {
		entry = new(PortEntry)
		Ports[ifname] = entry
	}
	entry.Ifname = ifname
	return entry
}

func SetPortByIndex(ifindex int32, ifname string) *PortEntry {
	if PortsByIndex == nil {
		PortsByIndex = make(map[int32]*PortEntry)
	}
	PortsByIndex[ifindex] = Ports[ifname]
	return PortsByIndex[ifindex]
}

func GetPortByIndex(ifindex int32) (entry *PortEntry) {
	if PortsByIndex == nil {
		return nil
	}
	entry, _ = PortsByIndex[ifindex]
	return entry
}

func GetNumSubports(ifname string) (numSubports uint) {
	numSubports = 0
	entry, found := Ports[ifname]
	if !found {
		return
	}
	portindex := entry.Portindex
	for _, pe := range Ports {
		if pe.Portindex == int16(portindex) {
			numSubports++
		}
	}
	return
}

func IfName(portindex, subportindex int) (name string) {
	name = ""
	for _, pe := range Ports {
		if int(pe.Portindex) == portindex && int(pe.Subportindex) == subportindex {
			name = pe.Ifname
		}
	}
	return
}

var (
	PortIsCopper = func(ifname string) bool { return false }
	PortIsFec74  = func(ifname string) bool { return false }
	PortIsFec91  = func(ifname string) bool { return false }
)

type RxTx int

const (
	Rx RxTx = iota
	Tx
	NRxTx
)

var rxTxStrings = [...]string{
	Rx: "rx",
	Tx: "tx",
}

func (x RxTx) String() (s string) {
	return elib.Stringer(rxTxStrings[:], int(x))
}

//go:generate gentemplate -id initHook -d Package=vnet -d DepsType=initHookVec -d Type=initHook -d Data=hooks github.com/platinasystems/go/elib/dep/dep.tmpl
type initHook func(v *Vnet)

var initHooks initHookVec

func AddInit(f initHook, deps ...*dep.Dep) { initHooks.Add(f, deps...) }

func (v *Vnet) configure(in *parse.Input) (err error) {
	if err = v.ConfigurePackages(in); err != nil {
		return
	}
	if err = v.InitPackages(); err != nil {
		return
	}
	return
}
func (v *Vnet) TimeDiff(t0, t1 cpu.Time) float64 { return v.loop.TimeDiff(t1, t0) }

func (v *Vnet) Run(in *parse.Input) (err error) {
	loop.AddInit(func(l *loop.Loop) {
		v.interfaceMain.init()
		v.CliInit()
		v.eventInit()
		for i := range initHooks.hooks {
			initHooks.Get(i)(v)
		}
		if err := v.configure(in); err != nil {
			panic(err)
		}
	})
	v.loop.Run()
	err = v.ExitPackages()
	return
}

func (v *Vnet) Quit() { v.loop.Quit() }

func (pe *PortEntry) AddIPNet(ipnet *net.IPNet) {
	pe.IPNets = append(pe.IPNets, ipnet)
}

func (pe *PortEntry) DelIPNet(ipnet *net.IPNet) {
	for i, peipnet := range pe.IPNets {
		if peipnet.IP.Equal(ipnet.IP) {
			n := len(pe.IPNets) - 1
			copy(pe.IPNets[i:], pe.IPNets[i+1:])
			pe.IPNets = pe.IPNets[:n]
			break
		}
	}
}

func (pe *PortEntry) HardwareAddr() net.HardwareAddr {
	return net.HardwareAddr(pe.Addr[:])
}

type ActionType int

const (
	PreVnetd       ActionType = iota // before vnetd is started
	ReadyVnetd                       // vnetd has declared it's ready
	PostReadyVnetd                   // vnetd processing something initated from previous state
	Dynamic                          // free-run case
)
