// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type macvlan mode {
//	private |
//	vepa |
//	bridge |
//	passthru [ no-promisc[uous] ] |
//	source [ macaddr {
//		{ add | del } LLADDR |
//		set LLADDR... |
//		flush
//	} ]
// }
//
func (m *mod) parseTypeMacVlan() error {
	m.args = m.opt.Parms.More(m.args, "mode")
	s := m.opt.Parms.ByName["mode"]
	if len(s) == 0 {
		return fmt.Errorf("missing mode")
	}
	mode, found := map[string]uint32{
		"private":  rtnl.MACVLAN_MODE_PRIVATE,
		"vepa":     rtnl.MACVLAN_MODE_VEPA,
		"bridge":   rtnl.MACVLAN_MODE_BRIDGE,
		"passthru": rtnl.MACVLAN_MODE_PASSTHRU,
		"source":   rtnl.MACVLAN_MODE_SOURCE,
	}[s]
	if !found {
		return fmt.Errorf("mode: %q unknown", s)
	} else {
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_MACVLAN_MODE,
			rtnl.Uint32Attr(mode)})
	}
	if mode == rtnl.MACVLAN_MODE_PASSTHRU {
		m.args = m.opt.Flags.More(m.args, []string{
			"no-promisc", "-promisc",
			"no-promiscuous", "-promiscuous",
		})
		if m.opt.Flags.ByName["no-promisc"] {
			m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_MACVLAN_FLAGS,
				rtnl.Uint16Attr(rtnl.MACVLAN_FLAG_NOPROMISC)})
		}
		return nil
	} else if mode != rtnl.MACVLAN_MODE_SOURCE {
		return nil
	}
	s = m.opt.Parms.ByName["macaddr"]
	if len(s) == 0 {
		return nil
	}
	switch s {
	case "add", "del":
		if len(m.args) == 0 {
			return fmt.Errorf("missing LLADDR")
		}
		mac, err := net.ParseMAC(m.args[0])
		if err != nil {
			return fmt.Errorf("LLADDR: %q %v",
				m.args[0], err)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_MACVLAN_MACADDR,
			rtnl.BytesAttr(mac)})
		m.args = m.args[1:]
	case "set":
		var macs rtnl.Attrs
		for len(m.args) > 0 {
			mac, err := net.ParseMAC(m.args[0])
			if err != nil {
				break
			}
			macs = append(macs,
				rtnl.Attr{rtnl.IFLA_MACVLAN_MACADDR,
					rtnl.BytesAttr(mac)})
			m.args = m.args[1:]
		}
		if len(macs) == 0 {
			return fmt.Errorf("missing LLADDR(s)")
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_MACVLAN_MACADDR_DATA, macs})
		m.args = m.args[1:]

	case "flush":
		// FIXME
	default:
		return fmt.Errorf("%q unknown", s)
	}
	return nil
}
