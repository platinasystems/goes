// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "fmt"

const (
	RTNLGRP_NONE rtnlgrp = iota
	RTNLGRP_LINK
	RTNLGRP_NOTIFY
	RTNLGRP_NEIGH
	RTNLGRP_TC
	RTNLGRP_IPV4_IFADDR
	RTNLGRP_IPV4_MROUTE
	RTNLGRP_IPV4_ROUTE
	RTNLGRP_IPV4_RULE
	RTNLGRP_IPV6_IFADDR
	RTNLGRP_IPV6_MROUTE
	RTNLGRP_IPV6_ROUTE
	RTNLGRP_IPV6_IFINFO
	RTNLGRP_DECnet_IFADDR
	RTNLGRP_NOP2
	RTNLGRP_DECnet_ROUTE
	RTNLGRP_DECnet_RULE
	RTNLGRP_NOP4
	RTNLGRP_IPV6_PREFIX
	RTNLGRP_IPV6_RULE
	RTNLGRP_ND_USEROPT
	RTNLGRP_PHONET_IFADDR
	RTNLGRP_PHONET_ROUTE
	RTNLGRP_DCB
	RTNLGRP_IPV4_NETCONF
	RTNLGRP_IPV6_NETCONF
	RTNLGRP_MDB
	RTNLGRP_MPLS_ROUTE
	RTNLGRP_NSID
	RTNLGRP_MPLS_NETCONF
	N_RTNLGRP
)

const RTNLGRP_MAX = N_RTNLGRP - 1

type rtnlgrp uint8

func (g rtnlgrp) Bit() uint32 {
	var bit uint32
	if g > 31 {
		panic(fmt.Errorf("must setsockopt group %d", g))
	}
	if g > 0 {
		bit = 1 << (g - 1)
	}
	return bit
}

func printRtnlGrps(groups uint32) {
	if groups == 0 {
		return
	}
	fmt.Printf("Groups:")
	sep := " "
	for _, x := range []struct {
		g    rtnlgrp
		name string
	}{
		{RTNLGRP_LINK, "link"},
		{RTNLGRP_NOTIFY, "notify"},
		{RTNLGRP_NEIGH, "neigh"},
		{RTNLGRP_TC, "tc"},
		{RTNLGRP_IPV4_IFADDR, "ipv4-ifaddr"},
		{RTNLGRP_IPV4_MROUTE, "ipv4-mroute"},
		{RTNLGRP_IPV4_ROUTE, "ipv4-route"},
		{RTNLGRP_IPV4_RULE, "ipv4-rule"},
		{RTNLGRP_IPV6_IFADDR, "ipv6-ifaddr"},
		{RTNLGRP_IPV6_MROUTE, "ipv6-mroute"},
		{RTNLGRP_IPV6_ROUTE, "ipv6-route"},
		{RTNLGRP_IPV6_IFINFO, "ipv6-ifinfo"},
		{RTNLGRP_DECnet_IFADDR, "decnet-ifaddr"},
		{RTNLGRP_NOP2, "nop2"},
		{RTNLGRP_DECnet_ROUTE, "decnet-route"},
		{RTNLGRP_DECnet_RULE, "decnet-rule"},
		{RTNLGRP_NOP4, "nop4"},
		{RTNLGRP_IPV6_PREFIX, "ipv6-prefix"},
		{RTNLGRP_IPV6_RULE, "ipv6-rule"},
		{RTNLGRP_ND_USEROPT, "nd-useropt"},
		{RTNLGRP_PHONET_IFADDR, "phonet-ifaddr"},
		{RTNLGRP_PHONET_ROUTE, "phonet-route"},
		{RTNLGRP_DCB, "dcb"},
		{RTNLGRP_IPV4_NETCONF, "ipv4-netconf"},
		{RTNLGRP_IPV6_NETCONF, "ipv6-netconf"},
		{RTNLGRP_MDB, "mdb"},
		{RTNLGRP_MPLS_ROUTE, "mpls-route"},
		{RTNLGRP_NSID, "nsid"},
		{RTNLGRP_MPLS_NETCONF, "mpls-netconf"},
	} {
		if bit := x.g.Bit(); (groups & bit) == bit {
			fmt.Print(sep, x.name)
			sep = ","
		}
	}
	fmt.Println()
}
