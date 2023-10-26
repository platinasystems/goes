// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"net"
	"net/netip"

	"github.com/platinasystems/goes/v2/pkg/net/netlink"
)

const (
	IFA_UNSPEC  = netlink.IFA_UNSPEC
	IFLA_UNSPEC = netlink.IFLA_UNSPEC
)

var rta = netlink.ExtractRtAttr

func List() ([]*Netif, error) {
	nl, err := netlink.Open()
	if err != nil {
		return nil, err
	}
	defer nl.Close()
	return list(nl)
}

func list(nl *netlink.Netlink) ([]*Netif, error) {
	nifs, err := ifinfos(nl)
	if err == nil {
		err = ifaddrs(nl, nifs)
	}
	return nifs, err
}

func ifinfos(nl *netlink.Netlink) ([]*Netif, error) {
	var nifs []*Netif
	req, msg := netlink.ExpandNlMsghdr(nil)
	req.Type = netlink.RTM_GETLINK
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	rtgen, msg := netlink.ExpandRtGenmsg(msg)
	rtgen.Family = netlink.AF_UNSPEC
	if err := nl.Request(msg); err != nil {
		return nifs, err
	}
	for {
		rsp, data, err := nl.Next()
		if err != nil {
			return nifs, err
		} else if rsp.Seq != req.Seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			return nifs, netlink.ExtractError(data)
		} else if rsp.Type != netlink.RTM_NEWLINK {
			continue
		}
		nif := new(Netif)
		if err = nif.ifinfo(data); err != nil {
			return nifs, err
		}
		nifs = append(nifs, nif)
	}
	return nifs, nil
}

func ifaddrs(nl *netlink.Netlink, nifs []*Netif) error {
	byIndex := make(map[int]*Netif)
	for _, nif := range nifs {
		byIndex[nif.Index] = nif
	}
	req, msg := netlink.ExpandNlMsghdr(nil)
	req.Type = netlink.RTM_GETADDR
	req.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_DUMP
	rtgen, msg := netlink.ExpandRtGenmsg(msg)
	rtgen.Family = netlink.AF_UNSPEC
	if err := nl.Request(msg); err != nil {
		return err
	}
	for {
		rsp, data, err := nl.Next()
		if err != nil {
			return err
		} else if rsp.Seq != req.Seq {
			continue
		} else if rsp.Type == netlink.NLMSG_DONE {
			break
		} else if rsp.Type == netlink.NLMSG_ERROR {
			return netlink.ExtractError(data)
		} else if rsp.Type != netlink.RTM_NEWADDR {
			continue
		}
		ifaddr, data := netlink.ExtractIfAddrmsg(data)
		nif, ok := byIndex[int(ifaddr.Index)]
		if !ok {
			continue
		}
		var addr, local netip.Addr
		for t, v, m := rta(data); t != IFA_UNSPEC; t, v, m = rta(m) {
			switch t {
			case netlink.IFA_ADDRESS,
				netlink.IFA_LOCAL,
				netlink.IFA_BROADCAST,
				netlink.IFA_ANYCAST,
				netlink.IFA_MULTICAST:
				switch ifaddr.Family {
				case netlink.AF_INET:
					v = v[:4]
				case netlink.AF_INET6:
					v = v[:16]
				}
				a, aok := netip.AddrFromSlice(v)
				if !aok {
					continue
				}
				switch t {
				case netlink.IFA_ADDRESS:
					addr = a
				case netlink.IFA_LOCAL:
					local = a
				case netlink.IFA_BROADCAST:
					nif.Extra["l3broadcast"] = a
				case netlink.IFA_ANYCAST:
					nif.Extra["anycast"] = a
				case netlink.IFA_MULTICAST:
					nif.Multicasts =
						append(nif.Multicasts, a)
				}
			case netlink.IFA_LABEL:
				nif.Extra["label"] = netlink.CloneString(v)
			}
		}
		if addr.IsValid() {
			if local.IsValid() && local.Compare(addr) != 0 {
				nif.Extra["peer"] = addr
				addr = local
			}
			prefix := netip.PrefixFrom(addr, int(ifaddr.Prefixlen))
			nif.Prefixes = append(nif.Prefixes, prefix)
		}
	}
	return nil
}

func (nif *Netif) ifinfo(data []byte) error {
	var stats64 *netlink.RtnlLinkStats[uint64]
	ifinfo, data := netlink.ExtractIfInfomsg(data)
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
	for t, v, data := rta(data); t != IFLA_UNSPEC; t, v, data = rta(data) {
		var isnt0 bool
		for _, b := range v {
			if b != 0 {
				isnt0 = true
			}
		}
		switch t {
		case netlink.IFLA_UNSPEC:
		case netlink.IFLA_ADDRESS:
			if isnt0 {
				nif.HardwareAddr =
					net.HardwareAddr(netlink.Clone(v))
			}
		case netlink.IFLA_BROADCAST:
			if isnt0 {
				nif.Extra["broadcast"] =
					net.HardwareAddr(netlink.Clone(v))
			}
		case netlink.IFLA_IFNAME:
			nif.Name = netlink.CloneString(v)
		case netlink.IFLA_MTU:
			nif.MTU = int(*(netlink.Pointer[uint32](v)))
		case netlink.IFLA_LINK:
			nif.Extra["link"] = *(netlink.Pointer[uint32](v))
		case netlink.IFLA_QDISC:
			nif.Extra["qdisc"] = netlink.CloneString(v)
		case netlink.IFLA_STATS:
			if stats64 == nil {
				stats := netlink.Pointer[netlink.
					RtnlLinkStats[uint32]](v)
				nif.Rx.Packets = uint64(stats.RxPackets)
				nif.Rx.Bytes = uint64(stats.RxBytes)
				nif.Rx.Drops = uint64(stats.RxDropped)
				nif.Rx.Errors = uint64(stats.RxErrors)
				nif.Tx.Packets = uint64(stats.TxPackets)
				nif.Tx.Bytes = uint64(stats.TxBytes)
				nif.Tx.Drops = uint64(stats.TxDropped)
				nif.Tx.Errors = uint64(stats.TxErrors)
				nif.Collisions = uint64(stats.Collisions)
			}
		case netlink.IFLA_COST:
		case netlink.IFLA_PRIORITY:
		case netlink.IFLA_MASTER:
			nif.Extra["master"] = *(netlink.Pointer[uint32](v))
		case netlink.IFLA_WIRELESS:
		case netlink.IFLA_PROTINFO:
		case netlink.IFLA_TXQLEN:
			nif.Extra["qlen"] = *(netlink.Pointer[uint32](v))
		case netlink.IFLA_MAP:
			if isnt0 {
				nif.Extra["ifmap"] = *(netlink.
					Pointer[netlink.RtnlLinkIfmap](v))
			}
		case netlink.IFLA_WEIGHT:
			nif.Extra["weight"] = *(netlink.Pointer[uint32](v))
		case netlink.IFLA_OPERSTATE:
			delete(nif.Extra, "state")
			if s, ok := netlink.IfOperName[v[0]]; ok {
				nif.Extra["state"] = s
			}
		case netlink.IFLA_LINKMODE:
			delete(nif.Extra, "mode")
			if s, ok := netlink.IfLinkModeName[v[0]]; ok {
				nif.Extra["mode"] = s
			}
		case netlink.IFLA_LINKINFO:
			// nested
		case netlink.IFLA_NET_NS_PID:
			nif.Extra["ns-pid"], _ = netlink.Extract[int32](v)
		case netlink.IFLA_IFALIAS:
			nif.Extra["alias"] = netlink.CloneString(v)
		case netlink.IFLA_NUM_VF:
			nif.Extra["num-vf"], _ = netlink.Extract[int32](v)
		case netlink.IFLA_VFINFO_LIST:
		case netlink.IFLA_STATS64:
			stats64 = netlink.Pointer[netlink.
				RtnlLinkStats[uint64]](v)
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
			nif.Extra["group"] = *(netlink.Pointer[int32](v))
		case netlink.IFLA_NET_NS_FD:
			nif.Extra["ns-pid"] = *(netlink.Pointer[int32](v))
		case netlink.IFLA_EXT_MASK:
		case netlink.IFLA_PROMISCUITY:
			nif.Extra["promiscuity"] = *(netlink.Pointer[int32](v))
		case netlink.IFLA_NUM_TX_QUEUES:
			nif.Extra["tx-queues"] = *(netlink.Pointer[uint32](v))
		case netlink.IFLA_NUM_RX_QUEUES:
			nif.Extra["rx-queues"] = *(netlink.Pointer[uint32](v))
		case netlink.IFLA_CARRIER:
			if v[0] == 0 {
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
