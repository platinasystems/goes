// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type ipip
//	[ remote ADDR ]
//	[ local ADDR ]
//	[ ttl TTL ]
//	[ tos TOS ]
//	[ dev DEVICE ]
//	[ [no-]pmtudisc ]
//	[ no-encap | encap { fou | gue | none } ]
//	[ encap-dport { PORT } ]
//	[ encap-sport { PORT | auto } ]
//	[ mode {
//		[ ipip | ip4ip4 | ip4/ip4 ] |
//		[ mplsip | mplsip4 | mpls/ip4 ] |
//		[ any | anyip4 | any/ip4 ]
//	}]
//	[ [no-]encap-csum ]
//
// TTL := { 1:255 }
// TOS := { 1:8 }
func (m *mod) parseTypeIpIp() error {
	var encapflags uint16
	m.args = m.opt.Parms.More(m.args, "remote", "local")
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"local", rtnl.IFLA_IPTUN_LOCAL},
		{"remote", rtnl.IFLA_IPTUN_REMOTE},
	} {
		if s := m.opt.Parms.ByName[x.name]; len(s) > 0 {
			if ip4 := net.ParseIP(s).To4(); ip4 == nil {
				return fmt.Errorf("%s: %q invalid", x.name, s)
			} else {
				m.tinfo = append(m.tinfo, rtnl.Attr{x.t,
					rtnl.BytesAttr(ip4)})
			}
		}
	}
	m.args = m.opt.Parms.More(m.args,
		[]string{"ttl", "hoplimit"},
		[]string{"tos", "tclass", "dstfield"},
	)
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"ttl", rtnl.IFLA_IPTUN_TTL},
		{"tos", rtnl.IFLA_IPTUN_TOS},
	} {
		if s := m.opt.Parms.ByName[x.name]; len(s) > 0 {
			var u8 uint8
			if _, err := fmt.Sscan(s, &u8); err != nil {
				return fmt.Errorf("%s: %q %v", x.name, s, err)
			}
			m.tinfo = append(m.tinfo, rtnl.Attr{x.t,
				rtnl.Uint8Attr(u8)})
		}
	}
	m.args = m.opt.Parms.More(m.args, "dev")
	if s := m.opt.Parms.ByName["dev"]; len(s) > 0 {
		dev, found := m.ifindexByName[s]
		if !found {
			return fmt.Errorf("dev: %q not found", s)
		}
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_IPTUN_LINK,
			rtnl.Uint32Attr(dev)})
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"pmtudisc", "+pmtudisc"},
		[]string{"no-pmtudisc", "-pmtudisc"},
	)
	if m.opt.Flags.ByName["pmtudisc"] {
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_IPTUN_PMTUDISC,
			rtnl.Uint8Attr(1)})
	} else if m.opt.Flags.ByName["no-pmtudisc"] {
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_IPTUN_PMTUDISC,
			rtnl.Uint8Attr(0)})
	}
	m.args = m.opt.Flags.More(m.args, []string{"no-encap", "-encap"})
	if m.opt.Flags.ByName["no-encap"] {
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_IPTUN_ENCAP_TYPE,
				rtnl.Uint32Attr(rtnl.TUNNEL_ENCAP_NONE)})
	} else {
		m.args = m.opt.Parms.More(m.args, "encap")
		if s := m.opt.Parms.ByName["encap"]; len(s) > 0 {
			if encap, found := map[string]uint16{
				"fou":  rtnl.TUNNEL_ENCAP_FOU,
				"gue":  rtnl.TUNNEL_ENCAP_GUE,
				"none": rtnl.TUNNEL_ENCAP_NONE,
			}[s]; !found {
				return fmt.Errorf("encap: %q unknown", s)
			} else {
				m.tinfo = append(m.tinfo,
					rtnl.Attr{rtnl.IFLA_IPTUN_ENCAP_TYPE,
						rtnl.Uint32Attr(encap)})
			}
		}
	}
	m.args = m.opt.Parms.More(m.args, "encap-sport", "encap-dport")
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"encap-sport", rtnl.IFLA_IPTUN_ENCAP_SPORT},
		{"encap-dport", rtnl.IFLA_IPTUN_ENCAP_DPORT},
	} {
		if s := m.opt.Parms.ByName[x.name]; len(s) > 0 {
			var u16 uint16
			if s != "any" {
				if _, err := fmt.Sscan(s, &u16); err != nil {
					return fmt.Errorf("%s: %q %v",
						x.name, s, err)
				}
			}
			m.tinfo = append(m.tinfo,
				rtnl.Attr{rtnl.IFLA_IPTUN_ENCAP_TYPE,
					rtnl.Be16Attr(u16)})
		}
	}
	m.args = m.opt.Parms.More(m.args, "mode")
	switch s := m.opt.Parms.ByName["mode"]; s {
	case "", "any", "anyip4", "any/ip4":
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_IPTUN_PROTO,
			rtnl.Uint8Attr(0)})
	case "ipip", "ip4ip4", "ip4/ip4":
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_IPTUN_PROTO,
			rtnl.Uint8Attr(rtnl.IPPROTO_IPIP)})
	case "mplsip", "mplsip4", "mpls/ip4":
		m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_IPTUN_PROTO,
			rtnl.Uint8Attr(rtnl.IPPROTO_MPLS)})
	default:
		return fmt.Errorf("%q: unknown encap", s)
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"encap-csum", "+encap-csum"},
		[]string{"no-encap-csum", "-encap-csum"},
	)
	for _, x := range []struct {
		set   string
		unset string
		flag  uint16
	}{
		{"encap-csum", "no-encap-csum", rtnl.TUNNEL_ENCAP_FLAG_CSUM},
	} {
		if m.opt.Flags.ByName[x.set] {
			encapflags |= x.flag
		} else if m.opt.Flags.ByName[x.unset] {
			encapflags &^= x.flag
		}
	}
	m.tinfo = append(m.tinfo, rtnl.Attr{rtnl.IFLA_GRE_ENCAP_FLAGS,
		rtnl.Uint16Attr(encapflags)})
	return nil
}
