// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/ifaddr"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/ifarp"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

func List() ([]*NetIf, map[int]*NetIf, map[string]*NetIf, error) {
	nl, err := netlink.Open()
	if err != nil {
		return nil, nil, nil, err
	}
	defer nl.Close()
	nifs, err := ifinfos(nl)
	nifByIndex := make(map[int]*NetIf)
	nifByName := make(map[string]*NetIf)
	if err == nil {
		for _, nif := range nifs {
			nifByIndex[nif.Index] = nif
			nifByName[nif.Name] = nif
		}
		err = ifaddrs(nl, nifByIndex)
	}
	return nifs, nifByIndex, nifByName, err
}

func ifinfos(nl *netlink.NL) ([]*NetIf, error) {
	var nifs []*NetIf
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETLINK
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	gen, req := netlink.ExpandRtGenMsg(req)
	gen.Family = af.UNSPEC
	if err := nl.Request(req); err != nil {
		return nifs, err
	}
	seq := hdr.SEQ
	for {
		rsp, data, err := nl.Next()
		if err != nil {
			return nifs, err
		} else if rsp.SEQ != seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			e, _ := netlink.ExtractMsgErr(data)
			return nifs, e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWLINK {
			continue
		}
		nif := new(NetIf)
		if err = nif.ifinfo(data); err != nil {
			return nifs, err
		}
		nifs = append(nifs, nif)
	}
	return nifs, nil
}

func ifaddrs(nl *netlink.NL, nifByIndex map[int]*NetIf) error {
	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_GETADDR
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	gen, req := netlink.ExpandRtGenMsg(req)
	gen.Family = af.UNSPEC
	if err := nl.Request(req); err != nil {
		return err
	}
	seq := hdr.SEQ
	for {
		rsp, data, err := nl.Next()
		if err != nil {
			return err
		} else if rsp.SEQ != seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			e, _ := netlink.ExtractMsgErr(data)
			return e.Err()
		} else if rsp.Type != rtnetlink.RTM_NEWADDR {
			continue
		}
		m, data := netlink.ExtractIfAddrMsg(data)
		nif, ok := nifByIndex[int(m.Index)]
		if !ok {
			continue
		}
		var addr, local netip.Addr
		for netlink.HasAttr(data) {
			t, v, datá := netlink.ExtractAttr(data)
			data = datá
			switch ifaddr.Ifa(t) {
			case ifaddr.IFA_UNSPEC:
			case
				ifaddr.IFA_ADDRESS,
				ifaddr.IFA_LOCAL,
				ifaddr.IFA_BROADCAST,
				ifaddr.IFA_ANYCAST,
				ifaddr.IFA_MULTICAST:
				a, aok := netlink.IP(m.Family, v)
				if !aok {
					continue
				}
				switch ifaddr.Ifa(t) {
				case ifaddr.IFA_ADDRESS:
					addr = a
				case ifaddr.IFA_LOCAL:
					local = a
				case ifaddr.IFA_BROADCAST:
					nif.Extra["l3broadcast"] = a
				case ifaddr.IFA_ANYCAST:
					nif.Extra["anycast"] = a
				case ifaddr.IFA_MULTICAST:
					nif.Multicasts =
						append(nif.Multicasts, a)
				}
			case ifaddr.IFA_LABEL:
				nif.Extra["label"] = netlink.CloneString(v)
			}
		}
		if addr.IsValid() {
			if local.IsValid() && local.Compare(addr) != 0 {
				nif.Extra["peer"] = addr
				addr = local
			}
			prefix := netip.PrefixFrom(addr, int(m.PrefixLen))
			nif.Prefixes = append(nif.Prefixes, prefix)
		}
	}
	return nil
}

func (nif *NetIf) ifinfo(data []byte) error {
	var stats64 *iflink.Stats[uint64]
	m, data := netlink.ExtractIfInfoMsg(data)
	if nif.Index != 0 && int(m.Index) != nif.Index {
		return nil
	}
	nif.Index = int(m.Index)
	nif.Type = ifarp.ARPHRD(m.Type)
	if nif.Extra == nil {
		nif.Extra = make(map[string]any)
	}
	if (m.Flags & uint32(iflink.IFF_UP)) != 0 {
		nif.Flags |= net.FlagUp
	}
	if (m.Flags & uint32(iflink.IFF_BROADCAST)) != 0 {
		nif.Flags |= net.FlagBroadcast
	}
	if (m.Flags & uint32(iflink.IFF_LOOPBACK)) != 0 {
		nif.Flags |= net.FlagLoopback
	}
	if (m.Flags & uint32(iflink.IFF_POINTOPOINT)) != 0 {
		nif.Flags |= net.FlagPointToPoint
	}
	if (m.Flags & uint32(iflink.IFF_MULTICAST)) != 0 {
		nif.Flags |= net.FlagMulticast
	}
	if (m.Flags & uint32(iflink.IFF_RUNNING)) != 0 {
		nif.Flags |= net.FlagRunning
	}
	for netlink.HasAttr(data) {
		t, v, datá := netlink.ExtractAttr(data)
		data = datá
		var isnt0 bool
		for _, b := range v {
			if b != 0 {
				isnt0 = true
			}
		}
		switch iflink.Ifla(t) {
		case iflink.IFLA_UNSPEC:
		case iflink.IFLA_ADDRESS:
			if isnt0 {
				nif.HardwareAddr =
					net.HardwareAddr(netlink.Clone(v))
			}
		case iflink.IFLA_BROADCAST:
			if isnt0 {
				nif.Extra["broadcast"] =
					net.HardwareAddr(netlink.Clone(v))
			}
		case iflink.IFLA_IFNAME:
			nif.Name = netlink.CloneString(v)
		case iflink.IFLA_MTU:
			nif.MTU = int(*(netlink.Pointer[uint32](v)))
		case iflink.IFLA_LINK:
			nif.Extra["link"] = *(netlink.Pointer[uint32](v))
		case iflink.IFLA_QDISC:
			nif.Extra["qdisc"] = netlink.CloneString(v)
		case iflink.IFLA_STATS:
			if stats64 == nil {
				s32 := netlink.Pointer[iflink.Stats[uint32]](v)
				nif.Rx.Packets = uint64(s32.RxPackets)
				nif.Rx.Bytes = uint64(s32.RxBytes)
				nif.Rx.Drops = uint64(s32.RxDropped)
				nif.Rx.Errors = uint64(s32.RxErrors)
				nif.Tx.Packets = uint64(s32.TxPackets)
				nif.Tx.Bytes = uint64(s32.TxBytes)
				nif.Tx.Drops = uint64(s32.TxDropped)
				nif.Tx.Errors = uint64(s32.TxErrors)
				nif.Collisions = uint64(s32.Collisions)
			}
		case iflink.IFLA_COST:
		case iflink.IFLA_PRIORITY:
		case iflink.IFLA_MASTER:
			nif.Extra["master"] = *(netlink.Pointer[uint32](v))
		case iflink.IFLA_WIRELESS:
		case iflink.IFLA_PROTINFO:
		case iflink.IFLA_TXQLEN:
			nif.Extra["qlen"] = *(netlink.Pointer[uint32](v))
		case iflink.IFLA_MAP:
			if isnt0 {
				nif.Extra["ifmap"] = *(netlink.
					Pointer[iflink.IfMap](v))
			}
		case iflink.IFLA_WEIGHT:
			nif.Extra["weight"] = *(netlink.Pointer[uint32](v))
		case iflink.IFLA_OPERSTATE:
			delete(nif.Extra, "state")
			nif.Extra["state"] = iflink.IfOper(v[0]).String()
		case iflink.IFLA_LINKMODE:
			delete(nif.Extra, "mode")
			nif.Extra["mode"] = iflink.IfLinkMode(v[0]).String()
		case iflink.IFLA_LINKINFO:
			// nested
		case iflink.IFLA_NET_NS_PID:
			nif.Extra["ns-pid"] = *(netlink.Pointer[int32](v))
		case iflink.IFLA_IFALIAS:
			nif.Extra["alias"] = netlink.CloneString(v)
		case iflink.IFLA_NUM_VF:
			nif.Extra["num-vf"] = *(netlink.Pointer[int32](v))
		case iflink.IFLA_VFINFO_LIST:
		case iflink.IFLA_STATS64:
			stats64 = netlink.Pointer[iflink.Stats[uint64]](v)
			nif.Rx.Packets = stats64.RxPackets
			nif.Rx.Bytes = stats64.RxBytes
			nif.Rx.Drops = stats64.RxDropped
			nif.Rx.Errors = stats64.RxErrors
			nif.Tx.Packets = stats64.TxPackets
			nif.Tx.Bytes = stats64.TxBytes
			nif.Tx.Drops = stats64.TxDropped
			nif.Tx.Errors = stats64.TxErrors
			nif.Collisions = stats64.Collisions
		case iflink.IFLA_VF_PORTS:
		case iflink.IFLA_PORT_SELF:
		case iflink.IFLA_AF_SPEC:
			// nested
		case iflink.IFLA_GROUP:
			nif.Extra["group"] = *(netlink.Pointer[int32](v))
		case iflink.IFLA_NET_NS_FD:
			nif.Extra["ns-pid"] = *(netlink.Pointer[int32](v))
		case iflink.IFLA_EXT_MASK:
		case iflink.IFLA_PROMISCUITY:
			nif.Extra["promiscuity"] = *(netlink.Pointer[int32](v))
		case iflink.IFLA_NUM_TX_QUEUES:
			nif.Extra["tx-queues"] = *(netlink.Pointer[uint32](v))
		case iflink.IFLA_NUM_RX_QUEUES:
			nif.Extra["rx-queues"] = *(netlink.Pointer[uint32](v))
		case iflink.IFLA_CARRIER:
			if v[0] == 0 {
				nif.Extra["carrier"] = "down"
			} else {
				nif.Extra["carrier"] = "up"
			}
		case iflink.IFLA_PHYS_PORT_ID:
			// unspecified
		case iflink.IFLA_CARRIER_CHANGES:
		case iflink.IFLA_PHYS_SWITCH_ID:
		case iflink.IFLA_LINK_NETNSID:
		case iflink.IFLA_PHYS_PORT_NAME:
		case iflink.IFLA_PROTO_DOWN:
		case iflink.IFLA_GSO_MAX_SEGS:
		case iflink.IFLA_GSO_MAX_SIZE:
		case iflink.IFLA_PAD:
		case iflink.IFLA_XDP:
		case iflink.IFLA_EVENT:
		}
	}
	return nil
}
