// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package netlink

import (
	"bytes"
	"sync"
)

var pool = struct {
	Empty           sync.Pool
	DoneMessage     sync.Pool
	ErrorMessage    sync.Pool
	GenMessage      sync.Pool
	IfAddrMessage   sync.Pool
	IfInfoMessage   sync.Pool
	NeighborMessage sync.Pool
	NetnsMessage    sync.Pool
	NoopMessage     sync.Pool
	RouteMessage    sync.Pool
	AttrArray       sync.Pool
	LinkStats       sync.Pool
	LinkStats64     sync.Pool
	Ip4Address      sync.Pool
	Ip6Address      sync.Pool
	EthernetAddress sync.Pool
	Ip4DevConf      sync.Pool
	Ip6DevConf      sync.Pool
	IfAddrCacheInfo sync.Pool
	RtaCacheInfo    sync.Pool
	RtaMultipath    sync.Pool
	NdaCacheInfo    sync.Pool
	Bytes           sync.Pool
	VlanFlags       sync.Pool
}{
	Empty: sync.Pool{
		New: func() interface{} {
			return new(Empty)
		},
	},
	DoneMessage: sync.Pool{
		New: func() interface{} {
			return new(DoneMessage)
		},
	},
	ErrorMessage: sync.Pool{
		New: func() interface{} {
			return new(ErrorMessage)
		},
	},
	GenMessage: sync.Pool{
		New: func() interface{} {
			return new(GenMessage)
		},
	},
	IfAddrMessage: sync.Pool{
		New: func() interface{} {
			return new(IfAddrMessage)
		},
	},
	IfInfoMessage: sync.Pool{
		New: func() interface{} {
			return new(IfInfoMessage)
		},
	},
	NeighborMessage: sync.Pool{
		New: func() interface{} {
			return new(NeighborMessage)
		},
	},
	NetnsMessage: sync.Pool{
		New: func() interface{} {
			return new(NetnsMessage)
		},
	},
	NoopMessage: sync.Pool{
		New: func() interface{} {
			return new(NoopMessage)
		},
	},
	RouteMessage: sync.Pool{
		New: func() interface{} {
			return new(RouteMessage)
		},
	},
	AttrArray: sync.Pool{
		New: func() interface{} {
			return new(AttrArray)
		},
	},
	LinkStats: sync.Pool{
		New: func() interface{} {
			return new(LinkStats)
		},
	},
	LinkStats64: sync.Pool{
		New: func() interface{} {
			return new(LinkStats64)
		},
	},
	Ip4Address: sync.Pool{
		New: func() interface{} {
			return new(Ip4Address)
		},
	},
	Ip6Address: sync.Pool{
		New: func() interface{} {
			return new(Ip6Address)
		},
	},
	EthernetAddress: sync.Pool{
		New: func() interface{} {
			return new(EthernetAddress)
		},
	},
	Ip4DevConf: sync.Pool{
		New: func() interface{} {
			return new(Ip4DevConf)
		},
	},
	Ip6DevConf: sync.Pool{
		New: func() interface{} {
			return new(Ip6DevConf)
		},
	},
	IfAddrCacheInfo: sync.Pool{
		New: func() interface{} {
			return new(IfAddrCacheInfo)
		},
	},
	RtaCacheInfo: sync.Pool{
		New: func() interface{} {
			return new(RtaCacheInfo)
		},
	},
	RtaMultipath: sync.Pool{
		New: func() interface{} {
			return new(RtaMultipath)
		},
	},
	NdaCacheInfo: sync.Pool{
		New: func() interface{} {
			return new(NdaCacheInfo)
		},
	},
	Bytes: sync.Pool{
		New: func() interface{} {
			return new(bytes.Buffer)
		},
	},
	VlanFlags: sync.Pool{
		New: func() interface{} {
			return new(VlanFlags)
		},
	},
}

func repool(v interface{}) {
	switch t := v.(type) {
	case *Empty:
		pool.Empty.Put(t)
	case *AddressFamilyAttrType:
		pool.Empty.Put((*Empty)(t))
	case *Ip4IfAttrType:
		pool.Empty.Put((*Empty)(t))
	case *Ip6IfAttrType:
		pool.Empty.Put((*Empty)(t))
	case *DoneMessage:
		*t = DoneMessage{}
		pool.DoneMessage.Put(t)
	case *ErrorMessage:
		*t = ErrorMessage{}
		pool.ErrorMessage.Put(t)
	case *GenMessage:
		*t = GenMessage{}
		pool.GenMessage.Put(t)
	case *IfAddrMessage:
		*t = IfAddrMessage{}
		pool.IfAddrMessage.Put(t)
	case *IfInfoMessage:
		*t = IfInfoMessage{}
		pool.IfInfoMessage.Put(t)
	case *NeighborMessage:
		*t = NeighborMessage{}
		pool.NeighborMessage.Put(t)
	case *NetnsMessage:
		*t = NetnsMessage{}
		pool.NetnsMessage.Put(t)
	case *NoopMessage:
		*t = NoopMessage{}
		pool.NoopMessage.Put(t)
	case *RouteMessage:
		*t = RouteMessage{}
		pool.RouteMessage.Put(t)
	case *AttrArray:
		*t = AttrArray{}
		pool.AttrArray.Put(t)
	case *LinkStats:
		*t = LinkStats{}
		pool.LinkStats.Put(t)
	case *LinkStats64:
		*t = LinkStats64{}
		pool.LinkStats64.Put(t)
	case *Ip4Address:
		*t = Ip4Address{}
		pool.Ip4Address.Put(t)
	case *Ip6Address:
		*t = Ip6Address{}
		pool.Ip6Address.Put(t)
	case *EthernetAddress:
		*t = EthernetAddress{}
		pool.EthernetAddress.Put(t)
	case *Ip4DevConf:
		*t = Ip4DevConf{}
		pool.Ip4DevConf.Put(t)
	case *Ip6DevConf:
		*t = Ip6DevConf{}
		pool.Ip6DevConf.Put(t)
	case *IfAddrCacheInfo:
		*t = IfAddrCacheInfo{}
		pool.IfAddrCacheInfo.Put(t)
	case *RtaCacheInfo:
		*t = RtaCacheInfo{}
		pool.RtaCacheInfo.Put(t)
	case *RtaMultipath:
		*t = RtaMultipath{}
		pool.RtaMultipath.Put(t)
	case *NdaCacheInfo:
		*t = NdaCacheInfo{}
		pool.NdaCacheInfo.Put(t)
	case *bytes.Buffer:
		t.Reset()
		pool.Bytes.Put(t)
	case *VlanFlags:
		*t = VlanFlags{}
		pool.VlanFlags.Put(t)
	}
}
