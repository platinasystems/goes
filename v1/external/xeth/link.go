// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"net"
	"sync"
)

type Maper interface {
	Delete(key interface{})
	Load(key interface{}) (value interface{}, ok bool)
	Store(key, value interface{})
}

type Linker interface {
	Maper
	EthtoolAutoNeg(set ...AutoNeg) AutoNeg
	EthtoolDuplex(set ...Duplex) Duplex
	EthtoolDevPort(set ...DevPort) DevPort
	EthtoolFlags(set ...EthtoolFlagBits) EthtoolFlagBits
	EthtoolSpeed(set ...uint32) uint32
	IfInfoName(set ...string) string
	IfInfoIfIndex(set ...int32) int32
	IfInfoNetNs(set ...NetNs) NetNs
	IfInfoFeatures(set ...uint64) uint64
	IfInfoFlags(set ...net.Flags) net.Flags
	IfInfoDevKind(set ...DevKind) DevKind
	IfInfoHardwareAddr(set ...net.HardwareAddr) net.HardwareAddr
	IPNets(set ...[]*net.IPNet) []*net.IPNet
	IsAdminUp() bool
	IsAutoNeg() bool
	IsBridge() bool
	IsLag() bool
	IsPort() bool
	IsVlan() bool
	NetIfHwL2FwdOffload() bool
	LinkModesSupported(set ...EthtoolLinkModeBits) EthtoolLinkModeBits
	LinkModesAdvertising(set ...EthtoolLinkModeBits) EthtoolLinkModeBits
	LinkModesLPAdvertising(set ...EthtoolLinkModeBits) EthtoolLinkModeBits
	LinkUp(set ...bool) bool
	Lowers(set ...[]Xid) []Xid
	Uppers(set ...[]Xid) []Xid
	Stats(set ...[]uint64) []uint64
	StatNames(set ...[]string) []string
	String() string
	Xid() Xid
}

type LinkAttr uint8

const (
	LinkAttrEthtoolAutoNeg LinkAttr = iota
	LinkAttrEthtoolDevPort
	LinkAttrEthtoolDuplex
	LinkAttrEthtoolFlags
	LinkAttrEthtoolSpeed
	LinkAttrIPNets
	LinkAttrIfInfoName
	LinkAttrIfInfoIfIndex
	LinkAttrIfInfoNetNs
	LinkAttrIfInfoFeatures
	LinkAttrIfInfoFlags
	LinkAttrIfInfoDevKind
	LinkAttrIfInfoHardwareAddr
	LinkAttrLinkModesAdvertising
	LinkAttrLinkModesLPAdvertising
	LinkAttrLinkModesSupported
	LinkAttrLinkUp
	LinkAttrLowers
	LinkAttrStatNames
	LinkAttrStats
	LinkAttrUppers
)

type Link struct {
	sync.Map
	xid Xid
}

var Links sync.Map

func LinkOf(xid Xid) (l *Link) {
	if v, ok := Links.Load(xid); ok {
		l = v.(*Link)
	}
	return
}

func expectLinkOf(xid Xid, requester string) (l *Link) {
	if l = LinkOf(xid); l == nil {
		Unknown.Inc()
	}
	return
}

func LinkRange(f func(xid Xid, l *Link) bool) {
	Links.Range(func(k, v interface{}) bool {
		return f(k.(Xid), v.(*Link))
	})
}

func ListXids() (xids Xids) {
	// scan docker containers to cache their name space attributes
	LinkRange(func(xid Xid, l *Link) bool {
		xids = append(xids, xid)
		return true
	})
	return
}

func RxDelete(xid Xid) (note DevDel) {
	defer Links.Delete(xid)
	note = DevDel(xid)
	l := LinkOf(xid)
	if l == nil {
		return
	}
	for _, entry := range l.IPNets() {
		entry.IP = entry.IP[:cap(entry.IP)]
		entry.Mask = entry.Mask[:cap(entry.Mask)]
		poolIPNet.Put(entry)
	}
	l.Range(func(key, value interface{}) bool {
		defer l.Delete(key)
		if method, found := value.(pooler); found {
			method.Pool()
		}
		return true
	})
	return
}

// Valid() if xid has mapped attributes
func Valid(xid Xid) bool {
	_, ok := Links.Load(xid)
	return ok
}

func (l *Link) EthtoolAutoNeg(set ...AutoNeg) (an AutoNeg) {
	if len(set) > 0 {
		an = set[0]
		l.Store(LinkAttrEthtoolAutoNeg, an)
	} else if v, ok := l.Load(LinkAttrEthtoolAutoNeg); ok {
		an = v.(AutoNeg)
	}
	return
}

func (l *Link) EthtoolDuplex(set ...Duplex) (duplex Duplex) {
	if len(set) > 0 {
		duplex = set[0]
		l.Store(LinkAttrEthtoolDuplex, duplex)
	} else if v, ok := l.Load(LinkAttrEthtoolDuplex); ok {
		duplex = v.(Duplex)
	}
	return
}

func (l *Link) EthtoolDevPort(set ...DevPort) (devport DevPort) {
	if len(set) > 0 {
		devport = set[0]
		l.Store(LinkAttrEthtoolDevPort, devport)
	} else if v, ok := l.Load(LinkAttrEthtoolDevPort); ok {
		devport = v.(DevPort)
	}
	return
}

func (l *Link) EthtoolFlags(set ...EthtoolFlagBits) (bits EthtoolFlagBits) {
	if len(set) > 0 {
		bits = set[0]
		l.Store(LinkAttrEthtoolFlags, bits)
	} else if v, ok := l.Load(LinkAttrEthtoolFlags); ok {
		bits = v.(EthtoolFlagBits)
	}
	return
}

func (l *Link) EthtoolSpeed(set ...uint32) (mbps uint32) {
	if len(set) > 0 {
		mbps = set[0]
		l.Store(LinkAttrEthtoolSpeed, mbps)
	} else if v, ok := l.Load(LinkAttrEthtoolSpeed); ok {
		mbps = v.(uint32)
	}
	return
}

func (l *Link) IfInfoName(set ...string) (name string) {
	if len(set) > 0 {
		name = set[0]
		l.Store(LinkAttrIfInfoName, name)
	} else if v, ok := l.Load(LinkAttrIfInfoName); ok {
		name = v.(string)
	}
	return
}

func (l *Link) IfInfoIfIndex(set ...int32) (ifindex int32) {
	if len(set) > 0 {
		ifindex = set[0]
		l.Store(LinkAttrIfInfoIfIndex, ifindex)
	} else if v, ok := l.Load(LinkAttrIfInfoIfIndex); ok {
		ifindex = v.(int32)
	}
	return
}

func (l *Link) IfInfoNetNs(set ...NetNs) (netns NetNs) {
	if len(set) > 0 {
		netns = set[0]
		l.Store(LinkAttrIfInfoNetNs, netns)
	} else if v, ok := l.Load(LinkAttrIfInfoNetNs); ok {
		netns = v.(NetNs)
	}
	return
}

func (l *Link) IfInfoFeatures(set ...uint64) (features uint64) {
	if len(set) > 0 {
		features = set[0]
		l.Store(LinkAttrIfInfoFeatures, features)
	} else if v, ok := l.Load(LinkAttrIfInfoFeatures); ok {
		features = v.(uint64)
	}
	return
}

func (l *Link) IfInfoFlags(set ...net.Flags) (flags net.Flags) {
	if len(set) > 0 {
		flags = set[0]
		l.Store(LinkAttrIfInfoFlags, flags)
	} else if v, ok := l.Load(LinkAttrIfInfoFlags); ok {
		flags = v.(net.Flags)
	}
	return
}

func (l *Link) IfInfoDevKind(set ...DevKind) (devkind DevKind) {
	if len(set) > 0 {
		devkind = set[0]
		l.Store(LinkAttrIfInfoDevKind, devkind)
	} else if v, ok := l.Load(LinkAttrIfInfoDevKind); ok {
		devkind = v.(DevKind)
	}
	return
}

func (l *Link) IfInfoHardwareAddr(set ...net.HardwareAddr) (ha net.HardwareAddr) {
	if len(set) > 0 {
		ha = set[0]
		l.Store(LinkAttrIfInfoHardwareAddr, ha)
	} else if v, ok := l.Load(LinkAttrIfInfoHardwareAddr); ok {
		ha = v.(net.HardwareAddr)
	}
	return
}

func (l *Link) IPNets(set ...[]*net.IPNet) (nets []*net.IPNet) {
	if len(set) > 0 {
		nets = set[0]
		l.Store(LinkAttrIPNets, nets)
	} else if v, ok := l.Load(LinkAttrIPNets); ok {
		nets = v.([]*net.IPNet)
	}
	return
}

func (l *Link) IsAdminUp() bool {
	return l.IfInfoFlags()&net.FlagUp == net.FlagUp
}

func (l *Link) IsAutoNeg() bool {
	return l.EthtoolAutoNeg() == AUTONEG_ENABLE
}

func (l *Link) IsBridge() bool {
	return l.IfInfoDevKind() == DevKindBridge
}

func (l *Link) IsLag() bool {
	return l.IfInfoDevKind() == DevKindLag
}

func (l *Link) IsPort() bool {
	return l.IfInfoDevKind() == DevKindPort
}

func (l *Link) IsVlan() bool {
	return l.IfInfoDevKind() == DevKindVlan
}

func (l *Link) NetIfHwL2FwdOffload() bool {
	return (l.IfInfoFeatures() & NetIfHwL2FwdOffload) == NetIfHwL2FwdOffload
}

func (l *Link) LinkModesSupported(set ...EthtoolLinkModeBits) EthtoolLinkModeBits {
	return l.linkmodes(LinkAttrLinkModesSupported, set...)
}

func (l *Link) LinkModesAdvertising(set ...EthtoolLinkModeBits) EthtoolLinkModeBits {
	return l.linkmodes(LinkAttrLinkModesAdvertising, set...)
}

func (l *Link) LinkModesLPAdvertising(set ...EthtoolLinkModeBits) EthtoolLinkModeBits {
	return l.linkmodes(LinkAttrLinkModesLPAdvertising, set...)
}

func (l *Link) linkmodes(attr LinkAttr, set ...EthtoolLinkModeBits) (modes EthtoolLinkModeBits) {
	if len(set) > 0 {
		modes = set[0]
		l.Store(attr, modes)
	} else if v, ok := l.Load(attr); ok {
		modes = v.(EthtoolLinkModeBits)
	}
	return
}

func (l *Link) LinkUp(set ...bool) (up bool) {
	if len(set) > 0 {
		up = set[0]
		l.Store(LinkAttrLinkUp, up)
	} else if v, ok := l.Load(LinkAttrLinkUp); ok {
		up = v.(bool)
	}
	return
}

func (l *Link) Lowers(set ...[]Xid) (xids []Xid) {
	if len(set) > 0 {
		xids = set[0]
		l.Store(LinkAttrLowers, xids)
	} else if v, ok := l.Load(LinkAttrLowers); ok {
		xids = v.([]Xid)
	}
	return
}

func (l *Link) Uppers(set ...[]Xid) (xids []Xid) {
	if len(set) > 0 {
		xids = set[0]
		l.Store(LinkAttrUppers, xids)
	} else if v, ok := l.Load(LinkAttrUppers); ok {
		xids = v.([]Xid)
	}
	return
}

func (l *Link) Stats(set ...[]uint64) (stats []uint64) {
	if len(set) > 0 {
		stats = set[0]
		l.Store(LinkAttrStats, stats)
	} else if v, ok := l.Load(LinkAttrStats); ok {
		stats = v.([]uint64)
	}
	return
}

func (l *Link) StatNames(set ...[]string) (names []string) {
	if len(set) > 0 {
		names = set[0]
		l.Store(LinkAttrStatNames, names)
	} else if v, ok := l.Load(LinkAttrStatNames); ok {
		names = v.([]string)
	}
	return
}

func (l *Link) String() string {
	return l.IfInfoName()
}

func (l *Link) Xid() Xid {
	return l.xid
}
