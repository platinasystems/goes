// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd || openbsd

package netif

import (
	"context"
	"net"
	"net/netip"
	"syscall"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/net/sockaddr"
	"github.com/platinasystems/goes/v2/pkg/net/sysctl"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

type IfMsgHdr = syscall.IfMsghdr
type IfData = syscall.IfData
type IfaMsgHdr = syscall.IfaMsghdr

func Extract[T Msgs](data []byte) (p *T, body, rem []byte) {
	l := sysctl.MsgLen(data)
	p = Pointer[T](data)
	body = data[Sizeof(p):]
	rem = data[sysctl.Align(l):]
	return
}

func Pointer[T ~uint16 | Msgs](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

func Sizeof[T Msgs](p *T) int {
	return int(unsafe.Sizeof(*p))
}

func List(ctx context.Context) (nifs []*NetIf, err error) {
	rib, err := sysctl.Get(
		syscall.CTL_NET,
		af.ROUTE,
		0,
		af.UNSPEC,
		syscall.NET_RT_IFLIST,
		0,
	)
	if err != nil {
		return nifs, err
	}
	nifByIndex := make(map[int]*NetIf)
	for data := rib; len(data) > sysctl.MsgMin; {
		msglen := sysctl.MsgLen(data)
		if len(data) < msglen {
			break
		}
		if !sysctl.MsgOK(data, syscall.RTM_IFINFO) {
			data = data[msglen:]
			continue
		}
		im, body, datá := Extract[IfMsgHdr](data)
		data = datá
		nif, ok := nifByIndex[int(im.Index)]
		if !ok {
			nif = &NetIf{
				Extra: make(map[string]any),
			}
			nif.Index = int(im.Index)
			nifs = append(nifs, nif)
			nifByIndex[nif.Index] = nif
		}
		if (im.Flags & int32(syscall.IFF_UP)) != 0 {
			nif.Flags |= net.FlagUp
		}
		if (im.Flags & int32(syscall.IFF_BROADCAST)) != 0 {
			nif.Flags |= net.FlagBroadcast
		}
		if (im.Flags & int32(syscall.IFF_LOOPBACK)) != 0 {
			nif.Flags |= net.FlagLoopback
		}
		if (im.Flags & int32(syscall.IFF_POINTOPOINT)) != 0 {
			nif.Flags |= net.FlagPointToPoint
		}
		if (im.Flags & int32(syscall.IFF_MULTICAST)) != 0 {
			nif.Flags |= net.FlagMulticast
		}
		if (im.Flags & int32(syscall.IFF_RUNNING)) != 0 {
			nif.Flags |= net.FlagRunning
		}
		nif.Type = int(im.Data.Type)
		nif.MTU = int(im.Data.Mtu)
		nif.Rx.Packets = uint64(im.Data.Ipackets)
		nif.Rx.Bytes = uint64(im.Data.Ibytes)
		nif.Rx.Errors = uint64(im.Data.Ierrors)
		nif.Tx.Packets = uint64(im.Data.Opackets)
		nif.Tx.Bytes = uint64(im.Data.Obytes)
		nif.Tx.Errors = uint64(im.Data.Oerrors)
		nif.Collisions = uint64(im.Data.Collisions)
		nif.Rx.Drops = uint64(im.Data.Iqdrops)
		nif.ExtraIfData(&im.Data)
		dlhdr, body, _ := sockaddr.ExtractDlHdr(body)
		nif.Name, nif.HardwareAddr, _ = dlhdr.NAS(body)
	}
	for data := rib; len(data) > 0; {
		msglen := sysctl.MsgLen(data)
		if len(data) < msglen {
			break
		}
		if !sysctl.MsgOK(data, syscall.RTM_NEWADDR) {
			data = data[msglen:]
			continue
		}
		ifa, body, datá := Extract[IfaMsgHdr](data)
		data = datá
		nif, ok := nifByIndex[int(ifa.Index)]
		if !ok {
			continue
		}
		var (
			addr netip.Addr
			bits int
		)
		const min = sockaddr.Min
		for i := 0; i < syscall.RTAX_MAX && len(body) > min; i++ {
			if (int(ifa.Addrs) & (1 << i)) == 0 {
				continue
			}
			sal := sockaddr.Len(body)
			if sal > len(body) {
				break
			}
			switch i {
			case syscall.RTAX_DST:
				nif.Extra["dst"] = sockaddr.IP(body)
			case syscall.RTAX_GATEWAY:
				nif.Extra["gw"] = sockaddr.IP(body)
			case syscall.RTAX_NETMASK:
				ip := sockaddr.IP(body)
				bits, _ = net.IPMask(ip.AsSlice()).Size()
			case syscall.RTAX_IFA:
				addr = sockaddr.IP(body)
			case syscall.RTAX_AUTHOR:
			case syscall.RTAX_BRD:
				nif.Extra["brd"] = sockaddr.IP(body)
			}
			body = body[sysctl.Align(sal):]
		}
		if addr.IsValid() && bits > 0 {
			prefix := netip.PrefixFrom(addr, bits)
			nif.Prefixes = append(nif.Prefixes, prefix)
		}
	}
	return nifs, nil
}
