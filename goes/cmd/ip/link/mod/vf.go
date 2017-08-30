// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

func (c *Command) parseVf(s string) error {
	var attrs rtnl.Attrs
	var vf uint32

	if _, err := fmt.Sscan(s, &vf); err != nil {
		return fmt.Errorf("vf: %q %v", s, err)
	}

	c.args = c.opt.Flags.More(c.args,
		[]string{"spoofchk", "+spoofchk"},
		[]string{"no-spoofchk", "-spoofchk"},
		[]string{"query-rss", "+query-rss"},
		[]string{"no-query-rss", "-query-rss"},
		[]string{"trust", "+trust"},
		[]string{"no-trust", "-trust"})
	c.args = c.opt.Parms.More(c.args,
		"mac",
		"vlan",
		"qos",
		"rate",
		"min-tx-rate",
		"max-tx-rate",
		"link-state", // auto, enable, disable
		"node-guid",
		"port-guid")
	for _, x := range []struct {
		on  string
		off string
		t   uint16
	}{
		{"spoofchk", "no-spoofchk", rtnl.IFLA_VF_SPOOFCHK},
		{"query-rss", "no-query-rss", rtnl.IFLA_VF_RSS_QUERY_EN},
		{"trust", "no-trust", rtnl.IFLA_VF_TRUST},
	} {
		if c.opt.Flags.ByName[x.on] {
			attrs = append(attrs,
				rtnl.Attr{x.t,
					rtnl.IflaVfFlag{
						Vf:      vf,
						Setting: 1,
					},
				},
			)
		} else if c.opt.Flags.ByName[x.off] {
			attrs = append(attrs,
				rtnl.Attr{x.t,
					rtnl.IflaVfFlag{
						Vf:      vf,
						Setting: 0,
					},
				},
			)
		}
	}
	if s := c.opt.Parms.ByName["mac"]; len(s) > 0 {
		v := rtnl.IflaVfMac{
			Vf: vf,
		}
		if mac, err := net.ParseMAC(s); err != nil {
			return fmt.Errorf("vf: %d: mac: %q %v", vf, s, err)
		} else {
			copy(v.Mac[:], mac)
		}
		attrs = append(attrs, rtnl.Attr{rtnl.IFLA_VF_MAC, &v})
	}
	if s := c.opt.Parms.ByName["vlan"]; len(s) > 0 {
		v := rtnl.IflaVfVlan{
			Vf: vf,
		}
		if _, err := fmt.Sscan(s, &v.Vlan); err != nil {
			return fmt.Errorf("vf: %d: vlan: %q %v", vf, s, err)
		}
		if qos := c.opt.Parms.ByName["qos"]; len(qos) > 0 {
			if _, err := fmt.Sscan(s, &v.Qos); err != nil {
				return fmt.Errorf("vf: %d: qos: %q %v",
					vf, qos, err)
			}
		}
		attrs = append(attrs, rtnl.Attr{rtnl.IFLA_VF_VLAN, v})
	}
	if s := c.opt.Parms.ByName["rate"]; len(s) > 0 {
		v := rtnl.IflaVfTxRate{
			Vf: vf,
		}
		if _, err := fmt.Sscan(s, &v.Rate); err != nil {
			return fmt.Errorf("vf: %d: rate: %q %v", vf, s, err)
		}
		attrs = append(attrs, rtnl.Attr{rtnl.IFLA_VF_TX_RATE, v})
	} else if s := c.opt.Parms.ByName["min-tx-rate"]; len(s) > 0 {
		v := rtnl.IflaVfRate{
			Vf: vf,
		}
		if _, err := fmt.Sscan(s, &v.MinTxRate); err != nil {
			return fmt.Errorf("vf: %d: min-tx-rate: %q %v",
				vf, s, err)
		}
		if _, err := fmt.Sscan(c.opt.Parms.ByName["max-tx-rate"],
			&v.MaxTxRate); err != nil {
			return fmt.Errorf("vf: %d: max-tx-rate: %q %v",
				vf, s, err)
		}
		attrs = append(attrs, rtnl.Attr{rtnl.IFLA_VF_RATE, v})
	}
	if s := c.opt.Parms.ByName["link-state"]; len(s) > 0 {
		ls, found := rtnl.IflaVfLinkStateByName[s]
		if !found {
			return fmt.Errorf("vf: %d: link-state: %q unknown",
				vf, s)
		}
		attrs = append(attrs, rtnl.Attr{rtnl.IFLA_VF_LINK_STATE,
			rtnl.IflaVfLinkState{
				Vf:        vf,
				LinkState: ls,
			}})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"node-guid", rtnl.IFLA_VF_IB_NODE_GUID},
		{"port-guid", rtnl.IFLA_VF_IB_PORT_GUID},
	} {
		if s := c.opt.Parms.ByName[x.name]; len(s) > 0 {
			v := rtnl.IflaVfGuid{
				Vf: vf,
			}
			if _, err := fmt.Sscan(s, &v.Guid); err != nil {
				return fmt.Errorf("vf: %d: %s: %q %v",
					vf, x.name, s, err)
			}
			attrs = append(attrs, rtnl.Attr{x.t, v})
		}
	}
	c.attrs = append(c.attrs, rtnl.Attr{rtnl.IFLA_VFINFO_LIST,
		rtnl.Attr{rtnl.IFLA_VF_INFO, attrs},
	})
	return nil
}
