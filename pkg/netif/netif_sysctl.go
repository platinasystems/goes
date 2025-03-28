// Copyright © 2023-2024 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build darwin || freebsd || netbsd || openbsd

package netif

import (
	"context"
	"net"
	"net/netip"
	"unsafe"

	"github.com/platinasystems/goes/v2/pkg/xnet"
	"golang.org/x/sys/unix"
)

type IfMsgHdr = unix.IfMsghdr
type IfData = unix.IfData
type IfaMsgHdr = unix.IfaMsghdr

func Extract[T Msgs](data []byte) (p *T, body, rem []byte) {
	l := xnet.SysctlMsgLen(data)
	p = Pointer[T](data)
	body = data[Sizeof(p):]
	rem = data[xnet.SysctlAlign(l):]
	return
}

func Pointer[T ~uint16 | Msgs](data []byte) *T {
	return (*T)(unsafe.Pointer(&data[0]))
}

func Sizeof[T Msgs](p *T) int {
	return int(unsafe.Sizeof(*p))
}

func List(ctx context.Context) (nifs NetIfs, err error) {
	rib, err := xnet.SysctlGet(
		unix.CTL_NET,
		xnet.AF_ROUTE,
		0,
		xnet.AF_UNSPEC,
		unix.NET_RT_IFLIST,
		0,
	)
	if err != nil {
		return
	}
	nifByIndex := make(map[int]*NetIf)
	for data := rib; len(data) > xnet.SysctlMsgMin; {
		msglen := xnet.SysctlMsgLen(data)
		if len(data) < msglen {
			break
		}
		if !xnet.SysctlMsgOK(data, unix.RTM_IFINFO) {
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
		if (im.Flags & int32(unix.IFF_UP)) != 0 {
			nif.Flags |= net.FlagUp
		}
		if (im.Flags & int32(unix.IFF_BROADCAST)) != 0 {
			nif.Flags |= net.FlagBroadcast
		}
		if (im.Flags & int32(unix.IFF_LOOPBACK)) != 0 {
			nif.Flags |= net.FlagLoopback
		}
		if (im.Flags & int32(unix.IFF_POINTOPOINT)) != 0 {
			nif.Flags |= net.FlagPointToPoint
		}
		if (im.Flags & int32(unix.IFF_MULTICAST)) != 0 {
			nif.Flags |= net.FlagMulticast
		}
		if (im.Flags & int32(unix.IFF_RUNNING)) != 0 {
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
		dlhdr, body, _ := xnet.SAExtractDataLink(body)
		nif.Name, nif.HardwareAddr, _ = dlhdr.NAS(body)
	}
	for data := rib; len(data) > 0; {
		msglen := xnet.SysctlMsgLen(data)
		if len(data) < msglen {
			break
		}
		if !xnet.SysctlMsgOK(data, unix.RTM_NEWADDR) {
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
		const min = xnet.SAMin
		for i := 0; i < unix.RTAX_MAX && len(body) > min; i++ {
			if (int(ifa.Addrs) & (1 << i)) == 0 {
				continue
			}
			sal := xnet.SALen(body)
			if sal > len(body) {
				break
			}
			switch i {
			case unix.RTAX_DST:
				if a := xnet.SAIP(body); a.IsValid() {
					nif.Extra["dst"] = a
				}
			case unix.RTAX_GATEWAY:
				if a := xnet.SAIP(body); a.IsValid() {
					nif.Extra["gw"] = a
				}
			case unix.RTAX_NETMASK:
				if a := xnet.SAIP(body); a.IsValid() {
					bits, _ = net.IPMask(a.AsSlice()).Size()
				}
			case unix.RTAX_IFA:
				if a := xnet.SAIP(body); a.IsValid() {
					addr = a
				}
			case unix.RTAX_AUTHOR:
			case unix.RTAX_BRD:
				if a := xnet.SAIP(body); a.IsValid() {
					nif.Extra["brd"] = a
				}
			}
			body = body[xnet.SysctlAlign(sal):]
		}
		if addr.IsValid() && bits > 0 {
			prefix := netip.PrefixFrom(addr, bits)
			nif.Prefixes = append(nif.Prefixes, prefix)
		}
	}
	return
}
