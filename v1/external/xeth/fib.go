// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"bytes"
	"net"
	"sync"
	"unsafe"

	"github.com/platinasystems/goes/external/xeth/internal"
)

type FibEntryEvent uint8

type RtScope uint8
type RtTable uint32
type Rtn uint8
type RtnhFlags uint32

type FibEntry struct {
	net.IPNet
	NHs []*NH
	NetNs
	RtTable
	FibEntryEvent
	Rtn
	Tos uint8
	Ref
}

type NH struct {
	net.IP
	Xid
	Ifindex int32
	Weight  int32
	RtnhFlags
	RtScope
}

const (
	CompatRtTable  = RtTable(RT_TABLE_COMPAT)
	DefaultRtTable = RtTable(RT_TABLE_DEFAULT)
	MainRtTable    = RtTable(RT_TABLE_MAIN)
	LocalRtTable   = RtTable(RT_TABLE_LOCAL)
)

var poolFibEntry = sync.Pool{
	New: func() interface{} {
		return &FibEntry{
			IPNet: net.IPNet{
				IP:   make([]byte, net.IPv6len, net.IPv6len),
				Mask: make([]byte, net.IPv6len, net.IPv6len),
			},
		}
	},
}

var poolNH = sync.Pool{
	New: func() interface{} {
		return &NH{
			IP: make([]byte, net.IPv6len, net.IPv6len),
		}
	},
}

func newFibEntry() *FibEntry {
	fe := poolFibEntry.Get().(*FibEntry)
	fe.Hold()
	return fe
}

func newNH() *NH {
	return poolNH.Get().(*NH)
}

func (fe *FibEntry) Pool() {
	if fe.Release() > 0 {
		return
	}
	for _, nh := range fe.NHs {
		nh.Pool()
	}
	fe.NHs = fe.NHs[:0]
	fe.IPNet.IP = fe.IPNet.IP[:net.IPv6len]
	fe.IPNet.Mask = fe.IPNet.Mask[:net.IPv6len]
	poolFibEntry.Put(fe)
}

func (nh *NH) Pool() {
	nh.IP = nh.IP[:net.IPv6len]
	poolNH.Put(nh)
}

// to sort a list of fib entries,
//	sort.Slice(fib, func(i, j int) bool {
//		return fib[i].Less(fib[j])
//	})
func (feI *FibEntry) Less(feJ *FibEntry) bool {
	if feI.NetNs != feJ.NetNs {
		return feI.NetNs < feJ.NetNs
	}
	if feI.RtTable != feJ.RtTable {
		return feI.RtTable < feJ.RtTable
	}
	switch bytes.Compare(feI.IPNet.IP, feJ.IPNet.IP) {
	case 0:
		onesI, _ := feI.IPNet.Mask.Size()
		onesJ, _ := feJ.IPNet.Mask.Size()
		return onesI < onesJ
	case -1:
		return true
	case 1:
		return false
	}
	return false
}

func fib4(msg *internal.MsgFibEntry) *FibEntry {
	fe := newFibEntry()
	fe.NetNs = NetNs(msg.Net)
	*(*uint32)(unsafe.Pointer(&fe.IPNet.IP[0])) = msg.Address
	*(*uint32)(unsafe.Pointer(&fe.IPNet.Mask[0])) = msg.Mask
	fe.IPNet.IP = fe.IPNet.IP[:net.IPv4len]
	fe.IPNet.Mask = fe.IPNet.Mask[:net.IPv4len]
	fe.FibEntryEvent = FibEntryEvent(msg.Event)
	fe.Rtn = Rtn(msg.Type)
	fe.RtTable = RtTable(msg.Table)
	fe.Tos = msg.Tos
	for _, nh := range msg.NextHops() {
		xid := fe.NetNs.Xid(nh.Ifindex)
		fenh := newNH()
		*(*uint32)(unsafe.Pointer(&fenh.IP[0])) = nh.Gw
		fenh.IP = fenh.IP[:net.IPv4len]
		fenh.Xid = xid
		fenh.Ifindex = nh.Ifindex
		fenh.Weight = nh.Weight
		fenh.RtnhFlags = RtnhFlags(nh.Flags)
		fenh.RtScope = RtScope(nh.Scope)
		fe.NHs = append(fe.NHs, fenh)
	}
	fe.NetNs.fibentry(fe)
	return fe
}

func fib6(msg *internal.MsgFib6Entry) *FibEntry {
	netns := NetNs(msg.Net)
	fe := newFibEntry()
	fe.NetNs = netns
	copy(fe.IPNet.IP, msg.Address[:])
	fe.IPNet.Mask = net.CIDRMask(int(msg.Length), net.IPv6len*8)
	fe.FibEntryEvent = FibEntryEvent(msg.Event)
	fe.Rtn = Rtn(msg.Type)
	fe.RtTable = RtTable(msg.Table)
	nhxid := netns.Xid(msg.Nh.Ifindex)
	nh := newNH()
	copy(nh.IP, msg.Nh.Gw[:])
	nh.Xid = nhxid
	nh.Ifindex = msg.Nh.Ifindex
	nh.Weight = msg.Nh.Weight
	nh.RtnhFlags = RtnhFlags(msg.Nh.Flags)
	fe.NHs = append(fe.NHs, nh)
	for _, sibling := range msg.Siblings() {
		sibxid := netns.Xid(sibling.Ifindex)
		nh = newNH()
		copy(nh.IP, sibling.Gw[:])
		nh.Xid = sibxid
		nh.Ifindex = sibling.Ifindex
		nh.Weight = sibling.Weight
		nh.RtnhFlags = RtnhFlags(sibling.Flags)
		fe.NHs = append(fe.NHs, nh)
	}
	fe.NetNs.fibentry(fe)
	return fe
}
