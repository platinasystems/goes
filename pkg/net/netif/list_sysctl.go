// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build !netlink && !linux

package netif

import (
	"net"
	"net/netip"
	"time"

	"github.com/platinasystems/goes/v2/pkg/log/style"
	"github.com/platinasystems/goes/v2/pkg/net/sysctl"
	"github.com/platinasystems/goes/v2/pkg/syscall/align"
)

func List() (nifs []*Netif, err error) {
	byIndex := make(map[int]*Netif)
	rib, err := sysctl.Get(sysctl.CTL_NET, sysctl.AF_ROUTE, 0,
		sysctl.AF_UNSPEC, sysctl.NET_RT_IFLIST, 0)
	if err != nil {
		return
	}
	var h *sysctl.MsgSubHdr
	for rem := rib; len(rem) > sysctl.Sizeof(h); {
		var (
			im   *sysctl.IfMsghdr
			body []byte
		)
		h = sysctl.Pointer[sysctl.MsgSubHdr](rem)
		if h.Type != sysctl.RTM_IFINFO {
			if h.Type != sysctl.RTM_NEWADDR {
				style.Note("type ", h.Type, "?")
			}
			rem = rem[h.Msglen:]
			continue
		}
		im, body, rem = sysctl.ExtractIfMsghdr(rem)
		nif, ok := byIndex[int(im.Index)]
		if !ok {
			nif = new(Netif)
			nif.Index = int(im.Index)
			nif.Extra = make(map[string]any)
			nifs = append(nifs, nif)
			byIndex[nif.Index] = nif
		}
		iff := IFF(im.Flags)
		nif.Extra["iff"] = iff
		nif.parseIFF(iff)
		nif.Type = IFT(im.Data.Type)
		nif.MTU = int(im.Data.Mtu)
		nif.Rx.Packets = uint64(im.Data.Ipackets)
		nif.Rx.Bytes = uint64(im.Data.Ibytes)
		nif.Rx.Errors = uint64(im.Data.Ierrors)
		nif.Tx.Packets = uint64(im.Data.Opackets)
		nif.Tx.Bytes = uint64(im.Data.Obytes)
		nif.Tx.Errors = uint64(im.Data.Oerrors)
		nif.Collisions = uint64(im.Data.Collisions)
		nif.Rx.Drops = uint64(im.Data.Iqdrops)
		if im.Data.Recvquota != 0 {
			nif.Extra["recvquota"] = im.Data.Recvquota
		}
		if im.Data.Xmitquota != 0 {
			nif.Extra["xmitquota"] = im.Data.Xmitquota
		}
		if im.Data.Metric != 0 {
			nif.Extra["metric"] = im.Data.Metric
		}
		if im.Data.Baudrate != 0 {
			nif.Extra["baudrate"] = im.Data.Baudrate
		}
		if im.Data.Imcasts != 0 {
			nif.Extra["imcasts"] = im.Data.Imcasts
		}
		if im.Data.Omcasts != 0 {
			nif.Extra["omcasts"] = im.Data.Omcasts
		}
		if im.Data.Noproto != 0 {
			nif.Extra["noproto"] = im.Data.Noproto
		}
		if im.Data.Recvtiming != 0 {
			nif.Extra["recvtiming"] = im.Data.Recvtiming
		}
		if im.Data.Xmittiming != 0 {
			nif.Extra["xmittiming"] = im.Data.Xmittiming
		}
		if im.Data.Lastchange.Sec != 0 {
			nif.Extra["lastchange"] = time.Unix(
				int64(im.Data.Lastchange.Sec),
				int64(im.Data.Lastchange.Usec)*1000)
		}
		if im.Data.Hwassist != 0 {
			nif.Extra["hwassist"] = im.Data.Hwassist
		}
		nif.parseSockaddrDatalink(body)
	}
	for rem := rib; len(rem) > sysctl.Sizeof(h); {
		var (
			ifa  *sysctl.IfaMsghdr
			body []byte
		)
		h = sysctl.Pointer[sysctl.MsgSubHdr](rem)
		if h.Type != sysctl.RTM_NEWADDR {
			rem = rem[h.Msglen:]
			continue
		}
		ifa, body, rem = sysctl.ExtractIfaMsghdr(rem)
		nif, ok := byIndex[int(ifa.Index)]
		if !ok {
			continue
		}
		nif.parseAddrs(ifa.Addrs, body)
	}
	return
}

func (nif *Netif) parseAddrs(addrs int32, body []byte) {
	var (
		addr netip.Addr
		bits int
		h    *sysctl.SockaddrSubHdr
	)
	sahdrsz := sysctl.Sizeof(h)
	for i := 0; i < sysctl.RTAX_MAX && len(body) >= sahdrsz; i++ {
		if (int(addrs) & (1 << i)) == 0 {
			continue
		}
		h = sysctl.Pointer[sysctl.SockaddrSubHdr](body)
		switch h.Family {
		case sysctl.AF_INET:
			sa, _, _ := sysctl.ExtractSockaddrIn(body)
			switch i {
			case sysctl.RTAX_DST:
				nif.Extra["dst"] = netip.AddrFrom4(sa.Addr)
			case sysctl.RTAX_GATEWAY:
				nif.Extra["gw"] = netip.AddrFrom4(sa.Addr)
			case sysctl.RTAX_GENMASK:
			case sysctl.RTAX_NETMASK:
				bits, _ = net.IPMask(sa.Addr[:]).Size()
			case sysctl.RTAX_IFP:
			case sysctl.RTAX_IFA:
				addr = netip.AddrFrom4(sa.Addr)
			case sysctl.RTAX_AUTHOR:
			case sysctl.RTAX_BRD:
				nif.Extra["brd"] = netip.AddrFrom4(sa.Addr)
			}
		case sysctl.AF_INET6:
			sa, _, _ := sysctl.ExtractSockaddrIn6(body)
			switch i {
			case sysctl.RTAX_DST:
				nif.Extra["dst"] = netip.AddrFrom16(sa.Addr)
			case sysctl.RTAX_GATEWAY:
				nif.Extra["gw"] = netip.AddrFrom16(sa.Addr)
			case sysctl.RTAX_GENMASK:
			case sysctl.RTAX_NETMASK:
				bits, _ = net.IPMask(sa.Addr[:]).Size()
			case sysctl.RTAX_IFP:
			case sysctl.RTAX_IFA:
				addr = netip.AddrFrom16(sa.Addr)
			case sysctl.RTAX_AUTHOR:
			case sysctl.RTAX_BRD:
				nif.Extra["brd"] = netip.AddrFrom16(sa.Addr)
			}
		}
		body = body[align.Sysctl.Roundup(int(h.Len)):]
	}
	if bits > 0 && !addr.IsUnspecified() {
		prefix := netip.PrefixFrom(addr, bits)
		nif.Prefixes = append(nif.Prefixes, prefix)
	}
}

func (nif *Netif) parseSockaddrDatalink(body []byte) {
	var sa *sysctl.SockaddrDatalink
	if len(body) < sysctl.Sizeof(sa) {
		return
	}
	sa, body, _ = sysctl.ExtractSockaddrDatalink(body)
	if sa.Family != sysctl.AF_LINK {
		return
	}
	nlen, alen, slen := int(sa.Nlen), int(sa.Alen), int(sa.Slen)
	if nlen == 0xff {
		nlen = 0
	}
	if alen == 0xff {
		alen = 0
	}
	if slen == 0xff {
		slen = 0
	}
	if len(body) < nlen+alen+slen {
		panic("insufficient data")
	}
	if nlen > 0 && nlen < len(body) {
		nif.Name = string(body[:nlen])
		body = body[nlen:]
	}
	if alen > 0 && alen < len(body) {
		nif.HardwareAddr = make(net.HardwareAddr, alen)
		copy(nif.HardwareAddr, body[:alen])
		body = body[alen:]
	}
	return
}
