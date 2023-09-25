// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

func List() ([]*Netif, error) {
	nl, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	defer nl.Close()
	return list(nl)
}

func list(nl *netlink.Netlink) (nifs []*Netif, err error) {
	h, req := netlink.ExpandNlMsghdr(nil)
	h.Type = netlink.RTM_GETLINK
	h.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	rtgen, req := netlink.ExpandRtGenmsg(req)
	rtgen.Family = netlink.AF_UNSPEC
	err = nl.Request(req, func(rsp []byte) error {
		h := netlink.Pointer[netlink.NlMsghdr](rsp)
		if h.Type == netlink.RTM_NEWLINK {
			nif := new(Netif)
			if err := nif.update(rsp); err != nil {
				return err
			}
			nifs = append(nifs, nif)
		}
		return nil
	})
	if err != nil {
		return
	}
	byIndex := make(map[int]*Netif)
	for _, nif := range nifs {
		byIndex[nif.Index] = nif
	}
	h.Type = netlink.RTM_GETADDR
	h.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	rtgen.Family = netlink.AF_UNSPEC
	err = nl.Request(req, func(rsp []byte) error {
		h, rsp := netlink.ExtractNlMsghdr(rsp)
		if h.Type != netlink.RTM_NEWADDR {
			return nil
		}
		ifaddr, rsp := netlink.ExtractIfAddrmsg(rsp)
		rta, rsp := netlink.ExtractRtAttr(rsp)
		_ = rta
		if nif, ok := byIndex[int(ifaddr.Index)]; ok {
			var ip netip.Addr
			switch ifaddr.Family {
			case netlink.AF_INET:
				ip = netip.AddrFrom4([4]byte(rsp[:4]))
			case netlink.AF_INET6:
				ip = netip.AddrFrom16([16]byte(rsp[:16]))
			}
			prefix := netip.PrefixFrom(ip, int(ifaddr.Prefixlen))
			nif.Prefixes = append(nif.Prefixes, prefix)
		}
		return nil
	})
	return
}

func (nif *Netif) update(data []byte) error {
	var stats64 *netlink.RtnlLinkStats[uint64]
	h, rsp := netlink.ExtractNlMsghdr(data)
	if h.Type != netlink.RTM_NEWLINK {
		return nil
	}
	ifinfo, rsp := netlink.ExtractIfInfomsg(rsp)
	if nif.Index != 0 && int(ifinfo.Index) != nif.Index {
		return nil
	}
	nif.Index = int(ifinfo.Index)
	nif.Type = ARPHRD(ifinfo.Type)
	if nif.Extra == nil {
		nif.Extra = make(map[string]any)
	}
	iff := IFF(ifinfo.Flags)
	nif.Extra["iff"] = iff
	nif.parseIFF(iff)
	for len(rsp) > netlink.SizeofRtAttr {
		rta := netlink.Pointer[netlink.RtAttr](rsp)
		val := rsp[netlink.SizeofRtAttr:int(rta.Len)]
		rsp = rsp[align.RTA.Roundup(int(rta.Len)):]
		var t byte
		for _, b := range val {
			t |= b
		}
		switch rta.Type {
		case netlink.IFLA_UNSPEC:
		case netlink.IFLA_ADDRESS:
			if t != 0 {
				nif.HardwareAddr = make(net.HardwareAddr,
					len(val))
				copy(nif.HardwareAddr, val)
			}
		case netlink.IFLA_BROADCAST:
			if t != 0 {
				ha := make(net.HardwareAddr, len(val))
				copy(ha, val)
				nif.Extra["broadcast"] = ha
			}
		case netlink.IFLA_IFNAME:
			nif.Name = netlink.CloneString(val)
		case netlink.IFLA_MTU:
			nif.MTU = int(*(netlink.Pointer[uint32](val)))
		case netlink.IFLA_LINK:
			nif.Extra["link"] = *(netlink.Pointer[uint32](val))
		case netlink.IFLA_QDISC:
			nif.Extra["qdisc"] = netlink.CloneString(val)
		case netlink.IFLA_STATS:
			if stats64 == nil {
				delete(nif.Extra, "stats")
				stats := netlink.Pointer[netlink.
					RtnlLinkStats[uint32]](val)
				nif.Extra["stats"] = *stats
				nif.Rx.Packets = uint64(stats64.RxPackets)
				nif.Rx.Bytes = uint64(stats64.RxBytes)
				nif.Rx.Drops = uint64(stats64.RxDropped)
				nif.Rx.Errors = uint64(stats64.RxErrors)
				nif.Tx.Packets = uint64(stats64.TxPackets)
				nif.Tx.Bytes = uint64(stats64.TxBytes)
				nif.Tx.Drops = uint64(stats64.TxDropped)
				nif.Tx.Errors = uint64(stats64.TxErrors)
				nif.Collisions = uint64(stats64.Collisions)
			}
		case netlink.IFLA_COST:
		case netlink.IFLA_PRIORITY:
		case netlink.IFLA_MASTER:
			nif.Extra["master"] = *(netlink.Pointer[uint32](val))
		case netlink.IFLA_WIRELESS:
		case netlink.IFLA_PROTINFO:
		case netlink.IFLA_TXQLEN:
			nif.Extra["qlen"] = *(netlink.Pointer[uint32](val))
		case netlink.IFLA_MAP:
			if t != 0 {
				nif.Extra["ifmap"] = *(netlink.
					Pointer[netlink.RtnlLinkIfmap](val))
			}
		case netlink.IFLA_WEIGHT:
			nif.Extra["weight"] = *(netlink.Pointer[uint32](val))
		case netlink.IFLA_OPERSTATE:
			delete(nif.Extra, "state")
			if s, ok := netlink.IfOperName[val[0]]; ok {
				nif.Extra["state"] = s
			}
		case netlink.IFLA_LINKMODE:
			delete(nif.Extra, "mode")
			if s, ok := netlink.IfLinkModeName[val[0]]; ok {
				nif.Extra["mode"] = s
			}
		case netlink.IFLA_LINKINFO:
			// nested
		case netlink.IFLA_NET_NS_PID:
			nif.Extra["ns-pid"] = *(netlink.Pointer[int32](val))
		case netlink.IFLA_IFALIAS:
			nif.Extra["alias"] = netlink.CloneString(val)
		case netlink.IFLA_NUM_VF:
			nif.Extra["num-vf"] = *(netlink.Pointer[int32](val))
		case netlink.IFLA_VFINFO_LIST:
		case netlink.IFLA_STATS64:
			delete(nif.Extra, "stats")
			stats64 = netlink.Pointer[netlink.
				RtnlLinkStats[uint64]](val)
			nif.Extra["stats"] = *stats64
			nif.Rx.Packets = stats64.RxPackets
			nif.Rx.Bytes = stats64.RxBytes
			nif.Rx.Drops = stats64.RxDropped
			nif.Rx.Errors = stats64.RxErrors
			nif.Tx.Packets = stats64.TxPackets
			nif.Tx.Bytes = stats64.TxBytes
			nif.Tx.Drops = stats64.TxDropped
			nif.Tx.Errors = stats64.TxErrors
			nif.Collisions = stats64.Collisions
		case netlink.IFLA_VF_PORTS:
		case netlink.IFLA_PORT_SELF:
		case netlink.IFLA_AF_SPEC:
			// nested
		case netlink.IFLA_GROUP:
			nif.Extra["group"] = *(netlink.Pointer[int32](val))
		case netlink.IFLA_NET_NS_FD:
			nif.Extra["ns-pid"] = *(netlink.Pointer[int32](val))
		case netlink.IFLA_EXT_MASK:
		case netlink.IFLA_PROMISCUITY:
			nif.Extra["promiscuity"] = *(netlink.Pointer[int32](val))
		case netlink.IFLA_NUM_TX_QUEUES:
			nif.Extra["tx-queues"] = *(netlink.Pointer[uint32](val))
		case netlink.IFLA_NUM_RX_QUEUES:
			nif.Extra["rx-queues"] = *(netlink.Pointer[uint32](val))
		case netlink.IFLA_CARRIER:
			if val[0] == 0 {
				nif.Extra["carrier"] = "down"
			} else {
				nif.Extra["carrier"] = "up"
			}
		case netlink.IFLA_PHYS_PORT_ID:
			// unspecified
		case netlink.IFLA_CARRIER_CHANGES:
		case netlink.IFLA_PHYS_SWITCH_ID:
		case netlink.IFLA_LINK_NETNSID:
		case netlink.IFLA_PHYS_PORT_NAME:
		case netlink.IFLA_PROTO_DOWN:
		case netlink.IFLA_GSO_MAX_SEGS:
		case netlink.IFLA_GSO_MAX_SIZE:
		case netlink.IFLA_PAD:
		case netlink.IFLA_XDP:
		case netlink.IFLA_EVENT:
		}
	}
	return nil
}
