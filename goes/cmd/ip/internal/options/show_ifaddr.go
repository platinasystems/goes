// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (opt *Options) ShowIfAddr(b []byte, withCacheInfo bool) {
	var ifa rtnl.Ifa
	var ifaf uint32
	ifa.Write(b)
	msg := rtnl.IfAddrMsgPtr(b)

	if val := ifa[rtnl.IFA_FLAGS]; len(val) > 0 {
		ifaf = rtnl.Uint32(val)
	} else {
		ifaf = uint32(msg.Flags)
	}

	opt.Print(rtnl.AfName(msg.Family), " ")
	ip := net.IP(ifa[rtnl.IFA_ADDRESS])
	opt.Print(ip, "/", msg.Prefixlen)
	opt.Print(" scope ", rtnl.RtScopeName[msg.Scope])

	if (ifaf & uint32(rtnl.IFA_F_SECONDARY)) ==
		uint32(rtnl.IFA_F_SECONDARY) {
		if msg.Family == rtnl.AF_INET {
			opt.Print(" secondary")
		} else {
			opt.Print(" temporary")
		}
	}
	for _, x := range []struct {
		not  bool
		flag uint32
		name string
	}{
		{false, uint32(rtnl.IFA_F_TENTATIVE), "tentative"},
		{false, uint32(rtnl.IFA_F_DEPRECATED), "deprecated"},
		{false, uint32(rtnl.IFA_F_HOMEADDRESS), "home"},
		{false, uint32(rtnl.IFA_F_NODAD), "nodad"},
		{false, uint32(rtnl.IFA_F_MANAGETEMPADDR), "mngtmpaddr"},
		{false, uint32(rtnl.IFA_F_NOPREFIXROUTE), "noprefixroute"},
		{false, uint32(rtnl.IFA_F_MCAUTOJOIN), "autojoin"},
		{true, uint32(rtnl.IFA_F_PERMANENT), "dynamic"},
		{false, uint32(rtnl.IFA_F_DADFAILED), "dadfailed"},
	} {
		if x.not {
			if (ifaf & x.flag) != x.flag {
				opt.Print(" ", x.name)
			}
		} else if (ifaf & x.flag) == x.flag {
			opt.Print(" ", x.name)
		}
		ifaf &= ^x.flag
	}
	if ifaf != 0 {
		fmt.Printf(" flags %#x", ifaf)
	}

	if val := ifa[rtnl.IFA_LABEL]; len(val) > 0 {
		opt.Print(" ", rtnl.Kstring(val))
	}
	if withCacheInfo {
		ci := rtnl.IfaCacheInfoPtr(ifa[rtnl.IFA_CACHEINFO])
		if ci != nil {
			opt.Println()
			opt.Nprint(7)
			opt.showIfaCacheInfo(ci)
		}
	}
}

func (opt *Options) showIfaCacheInfo(ci *rtnl.IfaCacheInfo) {
	for i, x := range []struct {
		name string
		lft  uint32
	}{
		{"valid", ci.Valid},
		{"preferred", ci.Prefered},
	} {
		if i > 0 {
			opt.Print(" ")
		}
		opt.Print(x.name, "_lft ")
		if x.lft == ^uint32(0) {
			opt.Print("forever")
		} else {
			opt.Print(x.lft)
		}
	}
}
