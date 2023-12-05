// Copyright © 2023 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

//go:build netlink || linux

package netif

import (
	"fmt"
	"strings"

	"github.com/platinasystems/goes/v2/pkg/errors/egress"
	"github.com/platinasystems/goes/v2/pkg/net/netlink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/iflink"
	"github.com/platinasystems/goes/v2/pkg/net/netlink/rtnetlink"
	"github.com/platinasystems/goes/v2/pkg/syscall/af"
)

const CreateParameters = ""

var Cloneable = []string{
	"bridge",
	"dummy",
	"geneve",
	"gre",
	"gretap",
	"hsr",
	"ifb",
	"ip6gre",
	"ip6gretap",
	"ipip",
	"ipoib",
	"macsec",
	"macvlan",
	"macvtap",
	"vcan",
	"vlan",
	"vrf",
	"vxlan",
}

func Create(name string, args ...string) (*NetIf, error) {
	var kind string
	for _, dev := range Cloneable {
		if strings.HasPrefix(name, dev) {
			kind = dev
			if name == kind {
				name += "%d"
			}
			break
		}
	}
	if len(kind) == 0 {
		return nil, fmt.Errorf("%q %w", name, ErrUnsupported)
	}

	before := make(map[int]*NetIf)
	nifs, err := List()
	if err != nil {
		return nil, err
	}
	for _, nif := range nifs {
		before[nif.Index] = nif
	}

	nl, err := netlink.Open()
	if err != nil {
		return nil, egress.Mark(err)
	}
	defer nl.Close()

	hdr, req := netlink.ExpandMsgHdr(nil)
	hdr.Type = rtnetlink.RTM_NEWLINK
	hdr.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK |
		netlink.NLM_F_CREATE | netlink.NLM_F_EXCL

	ifinfo, req := netlink.Expand[rtnetlink.IfInfoMsg](req)
	ifinfo.Family = af.UNSPEC

	req = netlink.CatStringAttr(req, iflink.IFLA_IFNAME, name)

	i := len(req)
	nested, req := netlink.Expand[iflink.Attr](req)
	nested.Type = iflink.IFLA_LINKINFO
	req = netlink.CatStringAttr(req, iflink.IFLA_INFO_KIND, kind)
	nested.Len = uint16(len(req) - i)

	if err = nl.Request(req); err != nil {
		return nil, egress.Mark(err)
	}
	if err = nl.Wait(hdr.SEQ); err != nil {
		return nil, egress.Mark(err)
	}
	if nifs, err = List(); err != nil {
		return nil, egress.Mark(err)
	}
	for _, nif := range nifs {
		if _, existed := before[nif.Index]; !existed {
			return nif, nil
		}
	}
	return nil, egress.Mark(ErrNotFound)
}
