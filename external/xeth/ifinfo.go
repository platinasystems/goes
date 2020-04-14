// Copyright © 2018-2020 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package xeth

import (
	"net"

	"github.com/platinasystems/goes/external/xeth/internal"
)

type DevKind uint8
type DevNew Xid
type DevDel Xid
type DevUp Xid
type DevDown Xid
type DevDump Xid
type DevUnreg Xid
type DevReg Xid

func (xid Xid) RxIfInfo(msg *internal.MsgIfInfo) interface{} {
	var note interface{} = DevDump(xid)
	l := mayMakeLinkOf(xid)
	if len(l.IfInfoName()) == 0 {
		note = DevNew(xid)
		name := make([]byte, internal.SizeofIfName)
		for i, c := range msg.Ifname[:] {
			if c == 0 {
				name = name[:i]
				break
			} else {
				name[i] = byte(c)
			}
		}
		l.IfInfoName(string(name))
		l.IfInfoDevKind(DevKind(msg.Kind))
		ha := make(net.HardwareAddr, internal.SizeofEthAddr)
		copy(ha, msg.Addr[:])
		l.IfInfoHardwareAddr(ha)
	}
	l.IfInfoIfIndex(msg.Ifindex)
	l.IfInfoNetNs(NetNs(msg.Net))
	l.IfInfoFlags(net.Flags(msg.Flags))
	return note
}

func (xid Xid) RxUp() DevUp {
	up := DevUp(xid)
	if l := expectLinkOf(xid, "admin-up"); l != nil {
		flags := l.IfInfoFlags()
		flags |= net.FlagUp
		l.IfInfoFlags(flags)
	}
	return up
}

func (xid Xid) RxDown() DevDown {
	down := DevDown(xid)
	if l := expectLinkOf(xid, "admin-down"); l != nil {
		flags := l.IfInfoFlags()
		flags &^= net.FlagUp
		l.IfInfoFlags(flags)
	}
	return down
}

func (xid Xid) RxReg(netns NetNs) DevReg {
	reg := DevReg(xid)
	if l := LinkOf(xid); l != nil {
		ifindex := l.IfInfoIfIndex()
		if netns != DefaultNetNs {
			DefaultNetNs.Xid(ifindex, 0)
			netns.Xid(ifindex, xid)
			l.IfInfoNetNs(netns)
		} else {
			DefaultNetNs.Xid(ifindex, xid)
			l.IfInfoNetNs(DefaultNetNs)
		}
	}
	return reg
}

func (xid Xid) RxUnreg() (unreg DevUnreg) {
	unreg = DevUnreg(xid)
	if l := expectLinkOf(xid, "netns-reg"); l != nil {
		ifindex := l.IfInfoIfIndex()
		oldns := l.IfInfoNetNs()
		oldns.Xid(ifindex, 0)
		DefaultNetNs.Xid(ifindex, xid)
		l.IfInfoNetNs(DefaultNetNs)
	}
	return unreg
}
