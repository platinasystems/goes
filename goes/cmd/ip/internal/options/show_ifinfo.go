// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package options

import (
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (opt *Options) ShowIfInfo(b []byte) {
	var ifla rtnl.Ifla
	ifla.Write(b)
	msg := rtnl.IfInfoMsgPtr(b)
	opt.Print(msg.Index, ": ")
	if val := ifla[rtnl.IFLA_IFNAME]; len(val) > 0 {
		opt.Print(rtnl.Kstring(val), ": ")
	}
	opt.Print("<")
	opt.ShowIfFlags(msg.Flags)
	opt.Print(">")
	if val := ifla[rtnl.IFLA_MTU]; len(val) > 0 {
		opt.Print(" mtu ", rtnl.Uint32(val))
	}
	if val := ifla[rtnl.IFLA_QDISC]; len(val) > 0 {
		opt.Print(" qdisc ", rtnl.Kstring(val))
	}
	if val := ifla[rtnl.IFLA_OPERSTATE]; len(val) > 0 {
		opt.Print(" state ", rtnl.IfOperName[rtnl.Uint8(val)])
	}
	if val := ifla[rtnl.IFLA_LINKMODE]; len(val) > 0 {
		opt.Print(" mode ", rtnl.IfLinkModeName[rtnl.Uint8(val)])
	}
	if val := ifla[rtnl.IFLA_GROUP]; len(val) > 0 {
		opt.Print(" group ", Gid(rtnl.Uint32(val)))
	}
	if val := ifla[rtnl.IFLA_TXQLEN]; len(val) > 0 {
		opt.Print(" qlen ", Gid(rtnl.Uint32(val)))
	}
	opt.Println()
	opt.Print("    link/", rtnl.ArphrdName[msg.Type])
	if val := ifla[rtnl.IFLA_ADDRESS]; len(val) > 0 {
		opt.Print(" ", net.HardwareAddr(val))
	}
	if val := ifla[rtnl.IFLA_BROADCAST]; len(val) > 0 {
		opt.Print(" brd ", net.HardwareAddr(val))
	}
	if opt.Flags.ByName["-d"] {
		if val := ifla[rtnl.IFLA_PROMISCUITY]; len(val) > 0 {
			opt.Print(" promiscuity ", rtnl.Uint32(val))
		}
		if val := ifla[rtnl.IFLA_NUM_TX_QUEUES]; len(val) > 0 {
			opt.Print(" numtxqueues ", rtnl.Uint32(val))
		}
		if val := ifla[rtnl.IFLA_NUM_RX_QUEUES]; len(val) > 0 {
			opt.Print(" numrxqueues ", rtnl.Uint32(val))
		}
		if val := ifla[rtnl.IFLA_NUM_VF]; len(val) > 0 {
			opt.Print(" num_vf ", rtnl.Uint32(val))
		}
	}
}

func (opt *Options) ShowIfFlags(iff uint32) {
	var comma string
	if (iff&rtnl.IFF_UP) == rtnl.IFF_UP &&
		(iff&rtnl.IFF_RUNNING) != rtnl.IFF_RUNNING {
		opt.Print("no-carrier")
		comma = ","
	}
	for _, x := range []struct {
		flag uint32
		name string
	}{
		{rtnl.IFF_LOOPBACK, "loopback"},
		{rtnl.IFF_BROADCAST, "broadcast"},
		{rtnl.IFF_POINTOPOINT, "pointopoint"},
		{rtnl.IFF_MULTICAST, "multicast"},
		{rtnl.IFF_NOARP, "noarp"},
		{rtnl.IFF_ALLMULTI, "allmulti"},
		{rtnl.IFF_PROMISC, "promisc"},
		{rtnl.IFF_MASTER, "master"},
		{rtnl.IFF_SLAVE, "slave"},
		{rtnl.IFF_DEBUG, "debug"},
		{rtnl.IFF_DYNAMIC, "dynamic"},
		{rtnl.IFF_AUTOMEDIA, "automedia"},
		{rtnl.IFF_PORTSEL, "portsel"},
		{rtnl.IFF_NOTRAILERS, "notrailers"},
		{rtnl.IFF_UP, "up"},
		{rtnl.IFF_LOWER_UP, "lower-up"},
		{rtnl.IFF_DORMANT, "dormant"},
		{rtnl.IFF_ECHO, "echo"},
	} {
		if (iff & x.flag) == x.flag {
			opt.Print(comma, x.name)
			comma = ","
		}
	}
}
