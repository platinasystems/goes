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
)

const CreateParameters = `
Create Parameters
  FIXME
`

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

func Create(name string, args ...string) (*Netif, error) {
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

	nl, err := netlink.Open()
	if err != nil {
		return nil, egress.Marked(err)
	}
	defer nl.Close()

	nifs, err := list(nl)
	if err != nil {
		return nil, err
	}
	existing := make(map[int]*Netif)
	for _, nif := range nifs {
		existing[nif.Index] = nif
	}

	h, req := netlink.Expand[netlink.NlMsghdr](nil)
	h.Type = netlink.RTM_NEWLINK
	h.Flags = netlink.NLM_F_REQUEST | netlink.NLM_F_ACK |
		netlink.NLM_F_CREATE | netlink.NLM_F_EXCL

	ifinfo, req := netlink.Expand[netlink.IfInfomsg](req)
	ifinfo.Family = netlink.AF_UNSPEC

	req = netlink.CatStringAttr(req, netlink.IFLA_IFNAME, name)

	i := len(req)
	nested, req := netlink.Expand[netlink.RtAttr](req)
	nested.Type = netlink.IFLA_LINKINFO
	req = netlink.CatStringAttr(req, netlink.IFLA_INFO_KIND, kind)
	nested.Len = uint16(len(req) - i)

	if nifs, err = list(nl); err != nil {
		return nil, egress.Marked(err)
	}
	for _, nif := range nifs {
		if _, existed := existing[nif.Index]; !existed {
			return nif, nil
		}
	}
	return nil, egress.Marked(ErrNotFound)
}
