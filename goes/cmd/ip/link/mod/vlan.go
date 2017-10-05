// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (m *mod) parseTypeVlan() error {
	var iflaVlanFlags rtnl.IflaVlanFlags
	m.args = m.opt.Flags.More(m.args,
		[]string{"reorder-hdr", "+reorder-hdr"},
		[]string{"no-reorder-hdr", "-reorder-hdr"},
		[]string{"gvrp", "+gvrp"},
		[]string{"no-gvrp", "-gvrp"},
		[]string{"mvrp", "+mvrp"},
		[]string{"no-mvrp", "-mvrp"},
		[]string{"loose-binding", "+loose-binding"},
		[]string{"no-loose-binding", "-loose-binding"},
	)
	for _, x := range []struct {
		set   string
		unset string
		flag  uint32
	}{
		{"reorder-hdr", "no-reorder-hdr",
			rtnl.VLAN_FLAG_REORDER_HDR},
		{"gvrp", "no-gvrp", rtnl.VLAN_FLAG_GVRP},
		{"loose-binding", "no-loose-binding",
			rtnl.VLAN_FLAG_LOOSE_BINDING},
		{"mvrp", "no-mvrp", rtnl.VLAN_FLAG_MVRP},
	} {
		if m.opt.Flags.ByName[x.set] {
			iflaVlanFlags.Mask |= x.flag
			iflaVlanFlags.Flags |= x.flag
		} else if m.opt.Flags.ByName[x.unset] {
			iflaVlanFlags.Mask |= x.flag
			iflaVlanFlags.Flags &^= x.flag
		}
	}
	if iflaVlanFlags.Mask != 0 {
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_VLAN_PROTOCOL,
				iflaVlanFlags})
	}
	m.args = m.opt.Parms.More(m.args,
		"protocol",
		"id",
		"ingress-qos-map",
		"egress-qos-map",
	)
	if s := m.opt.Parms.ByName["protocol"]; len(s) > 0 {
		proto, found := map[string]uint16{
			"802.1q":  0x8100,
			"802.1ad": 0x88a8,
		}[s]
		if !found {
			return fmt.Errorf("protocol: %q not found", s)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_VLAN_PROTOCOL,
				rtnl.Uint16Attr(proto)})
	}
	if s := m.opt.Parms.ByName["id"]; len(s) > 0 {
		var id uint16
		if _, err := fmt.Sscan(s, &id); err != nil {
			return fmt.Errorf("type vlan id: %q %v",
				s, err)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_VLAN_ID,
				rtnl.Uint16Attr(id)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"egress-qos-map", rtnl.IFLA_VLAN_EGRESS_QOS},
		{"ingress-qos-map", rtnl.IFLA_VLAN_INGRESS_QOS},
	} {
		var qos rtnl.IflaVlanQosMapping
		s := m.opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		colon := strings.Index(s, ":")
		if colon < 0 {
			return fmt.Errorf("%s: %q invalid", x.name, s)
		}
		if _, err := fmt.Sscan(s[:colon], &qos.From); err != nil {
			return fmt.Errorf("%s: FROM: %q %v",
				x.name, s[:colon], err)
		}
		if _, err := fmt.Sscan(s[colon+1:], &qos.To); err != nil {
			return fmt.Errorf("%s: TO: %q %v",
				x.name, s[colon+1:], err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{x.t, qos})
	}
	return nil
}
