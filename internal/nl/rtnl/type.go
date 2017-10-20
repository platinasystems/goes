// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package rtnl

import "strings"

func CompleteType(s string) (list []string) {
	for _, t := range []string{
		"bridge",
		"bond",
		"can",
		"dummy",
		"hsr",
		"ifb",
		"ipoib",
		"macvlan",
		"macvtap",
		"vcan",
		"veth",
		"vlan",
		"vxlan",
		"ip6tnl",
		"ipip",
		"sit",
		"gre",
		"gretap",
		"ip6gre",
		"ip6gretap",
		"vti",
		"nlmon",
		"ipvlan",
		"lowpan",
		"geneve",
		"macsec",
		"vrf",
	} {
		if len(s) == 0 || strings.HasPrefix(t, s) {
			list = append(list, t)
		}
	}
	return
}
