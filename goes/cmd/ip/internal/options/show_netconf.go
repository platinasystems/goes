// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

func (opt *Options) ShowNetconf(b []byte, ifnames map[int32]string) {
	onoff := func(b []byte) string {
		if nl.Uint32(b) != 0 {
			return "on"
		}
		return "off"
	}
	var netconfa rtnl.Netconfa
	netconfa.Write(b)
	msg := rtnl.NetconfMsgPtr(b)
	opt.Print(rtnl.AfName(msg.Family), " ")
	if val := netconfa[rtnl.NETCONFA_IFINDEX]; len(val) > 0 {
		switch idx := nl.Int32(val); idx {
		case rtnl.NETCONFA_IFINDEX_ALL:
			opt.Print("all ")
		case rtnl.NETCONFA_IFINDEX_DEFAULT:
			opt.Print("default ")
		default:
			if name, found := ifnames[idx]; found {
				opt.Print("dev ", name)
			} else {
				opt.Print("dev ", idx)
			}
		}
	}
	if val := netconfa[rtnl.NETCONFA_FORWARDING]; len(val) > 0 {
		opt.Print("forwarding ", onoff(val), " ")
	}
	if val := netconfa[rtnl.NETCONFA_RP_FILTER]; len(val) > 0 {
		switch nl.Uint32(val) {
		case 0:
			opt.Print("rp-filter off ")
		case 1:
			opt.Print("rp-filter strict ")
		case 2:
			opt.Print("rp-filter loose ")
		default:
			opt.Print("rp-filter unknown-mode ")
		}
	}
	if val := netconfa[rtnl.NETCONFA_MC_FORWARDING]; len(val) > 0 {
		opt.Print("mc-forwarding ", onoff(val), " ")
	}
	if val := netconfa[rtnl.NETCONFA_PROXY_NEIGH]; len(val) > 0 {
		opt.Print("proxy-neigh ", onoff(val), " ")
	}
	if val := netconfa[rtnl.NETCONFA_IGNORE_ROUTES_WITH_LINKDOWN]; len(val) > 0 {
		opt.Print("ignoe-routes-with-linkdown ", onoff(val), " ")
	}
	if val := netconfa[rtnl.NETCONFA_INPUT]; len(val) > 0 {
		opt.Print("input ", onoff(val), " ")
	}
}
