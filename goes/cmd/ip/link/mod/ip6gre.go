// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
)

// ip link COMMAND type { ip6gre | ip6gretap }
//	[ remote ADDR ]
//	[ local ADDR ]
//	[ [no-][i|o]seq ]
//	[ [i|o]key KEY ]
//	[ no-[i|o]key ]
//	[ [no-][i|o]csum ]
//	[ hoplimit TTL ]
//	[ encaplimit ELIM ]
//	[ tclass TCLASS ]
//	[ flowlabel FLOWLABEL ]
//	[ dscp { inherit } ]
//	[ fwmark MARK ]
//	[ dev DEVICE ]
//	[ encap { fou | gue | none } ]
//	[ encap-dport { PORT | any } ]
//	[ encap-sport { PORT | any } ]
//	[ [no-]encap-csum ]
//	[ [no-]encap-csum6 ]
//	[ [no-]encap-remsum ]
func (m *mod) parseTypeIp6Gre() error {
	var iflags, oflags, eflags uint16
	var greflags, flowinfo uint32
	m.args = m.opt.Parms.More(m.args, "remote", "local")
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"local", rtnl.IFLA_GRE_LOCAL},
		{"remote", rtnl.IFLA_GRE_REMOTE},
	} {
		if s := m.opt.Parms.ByName[x.name]; len(s) > 0 {
			if ip4 := net.ParseIP(s).To4(); ip4 == nil {
				return fmt.Errorf("%s: %q invalid", x.name, s)
			} else {
				m.tinfo = append(m.tinfo,
					rtnl.Attr{x.t, rtnl.BytesAttr(ip4)})
			}
		}
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"seq", "+seq"},
		[]string{"no-seq", "-seq"},
		[]string{"iseq", "+iseq"},
		[]string{"no-iseq", "-iseq"},
		[]string{"oseq", "+oseq"},
		[]string{"no-oseq", "-oseq"},
	)
	if m.opt.Flags.ByName["seq"] {
		iflags |= rtnl.GRE_SEQ
		oflags |= rtnl.GRE_SEQ
	} else if m.opt.Flags.ByName["no-seq"] {
		iflags &^= rtnl.GRE_SEQ
		oflags &^= rtnl.GRE_SEQ
	} else {
		if m.opt.Flags.ByName["iseq"] {
			iflags |= rtnl.GRE_SEQ
		} else if m.opt.Flags.ByName["no-iseq"] {
			iflags &^= rtnl.GRE_SEQ
		}
		if m.opt.Flags.ByName["oseq"] {
			oflags |= rtnl.GRE_SEQ
		} else if m.opt.Flags.ByName["no-oseq"] {
			oflags &^= rtnl.GRE_SEQ
		}
	}
	m.args = m.opt.Parms.More(m.args, "key", "ikey", "okey")
	if s := m.opt.Parms.ByName["key"]; len(s) > 0 {
		var u32 uint32
		if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("key: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_IKEY, rtnl.Uint32Attr(u32)})
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_OKEY, rtnl.Uint32Attr(u32)})
		iflags |= rtnl.GRE_KEY
		oflags |= rtnl.GRE_KEY
	} else {
		for _, x := range []struct {
			name  string
			t     uint16
			flags *uint16
		}{
			{"ikey", rtnl.IFLA_GRE_IKEY, &iflags},
			{"okey", rtnl.IFLA_GRE_OKEY, &oflags},
		} {
			var u32 uint32
			s := m.opt.Parms.ByName[x.name]
			if len(s) == 0 {
				continue
			}
			if _, err := fmt.Sscan(s, &u32); err != nil {
				return fmt.Errorf("%s: %q %v", x.name, s, err)
			}
			m.tinfo = append(m.tinfo,
				rtnl.Attr{x.t, rtnl.Uint32Attr(u32)})
			*(x.flags) |= rtnl.GRE_KEY
		}
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"no-key", "-key"},
		[]string{"no-ikey", "-ikey"},
		[]string{"no-okey", "-okey"},
	)
	if m.opt.Flags.ByName["no-key"] {
		iflags &^= rtnl.GRE_KEY
		oflags &^= rtnl.GRE_KEY
	} else {
		if m.opt.Flags.ByName["no-ikey"] {
			iflags &^= rtnl.GRE_KEY
		}
		if m.opt.Flags.ByName["no-okey"] {
			oflags &^= rtnl.GRE_KEY
		}
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"csum", "+csum"},
		[]string{"no-csum", "-csum"},
		[]string{"icsum", "+icsum"},
		[]string{"no-icsum", "-icsum"},
		[]string{"ocsum", "+ocsum"},
		[]string{"no-ocsum", "-ocsum"},
	)
	if m.opt.Flags.ByName["csum"] {
		iflags |= rtnl.GRE_CSUM
		oflags |= rtnl.GRE_CSUM
	} else if m.opt.Flags.ByName["no-csum"] {
		iflags &^= rtnl.GRE_CSUM
		oflags &^= rtnl.GRE_CSUM
	} else {
		if m.opt.Flags.ByName["icsum"] {
			iflags |= rtnl.GRE_CSUM
		} else if m.opt.Flags.ByName["no-icsum"] {
			iflags &^= rtnl.GRE_CSUM
		}
		if m.opt.Flags.ByName["ocsum"] {
			oflags |= rtnl.GRE_CSUM
		} else if m.opt.Flags.ByName["no-ocsum"] {
			oflags &^= rtnl.GRE_CSUM
		}
	}
	m.args = m.opt.Parms.More(m.args, "dev")
	if s := m.opt.Parms.ByName["dev"]; len(s) > 0 {
		dev, found := m.ifindexByName[s]
		if !found {
			return fmt.Errorf("dev: %q not found", s)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_LINK, rtnl.Uint32Attr(dev)})
	}
	m.args = m.opt.Parms.More(m.args, []string{"ttl", "hoplimit"})
	if s := m.opt.Parms.ByName["ttl"]; len(s) > 0 {
		var u8 uint8
		if _, err := fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("ttl: %q %v", s, err)
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_TTL, rtnl.Uint8Attr(u8)})
	}
	m.args = m.opt.Parms.More(m.args, []string{"tos", "tclass", "defiled"})
	if s := m.opt.Parms.ByName["tos"]; len(s) > 0 {
		var u32 uint32
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_USE_ORIG_TCLASS
		} else if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("tos: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_USE_ORIG_TCLASS
			flowinfo |= u32 << 20
		}
	}
	m.args = m.opt.Parms.More(m.args, "flowlabel")
	if s := m.opt.Parms.ByName["flowlabel"]; len(s) > 0 {
		var u32 uint32
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_USE_ORIG_FLOWLABEL
		} else if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("flowlabel: %q %v", s, err)
		} else if u32 > 0xFFFFF {
			return fmt.Errorf("flowlabel: %q invalid", s)
		} else {
			greflags &^= rtnl.IP6_TNL_F_USE_ORIG_FLOWLABEL
			flowinfo |= u32
		}
	}
	if flowinfo != 0 {
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_FLOWINFO,
				rtnl.Uint32Attr(flowinfo)})
	}
	m.args = m.opt.Parms.More(m.args, "dscp")
	if s := m.opt.Parms.ByName["dscp"]; len(s) > 0 {
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_RCV_DSCP_COPY
		} else {
			return fmt.Errorf("dscp: %q invalid", s)
		}
	}
	m.args = m.opt.Parms.More(m.args, "fwmark")
	if s := m.opt.Parms.ByName["fwmark"]; len(s) > 0 {
		var u32 uint32
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_USE_ORIG_FWMARK
		} else if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("fwmark: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_USE_ORIG_FWMARK
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_FWMARK, rtnl.Uint32Attr(u32)})
	}
	m.args = m.opt.Parms.More(m.args, "encaplimit")
	if s := m.opt.Parms.ByName["encaplimit"]; len(s) > 0 {
		var u8 uint32
		if s == "node" {
			greflags |= rtnl.IP6_TNL_F_IGN_ENCAP_LIMIT
		} else if _, err := fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("encaplimit: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_IGN_ENCAP_LIMIT
		}
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_ENCAP_LIMIT,
				rtnl.Uint8Attr(u8)})
	}
	m.args = m.opt.Flags.More(m.args, []string{"no-encap", "-encap"})
	if m.opt.Flags.ByName["no-encap"] {
		m.tinfo = append(m.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
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
					rtnl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
						rtnl.Uint32Attr(encap)})
			}
		}
	}
	m.args = m.opt.Parms.More(m.args, "encap-sport", "encap-dport")
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"encap-sport", rtnl.IFLA_GRE_ENCAP_SPORT},
		{"encap-sport", rtnl.IFLA_GRE_ENCAP_DPORT},
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
				rtnl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
					rtnl.Be16Attr(u16)})
		}
	}
	m.args = m.opt.Flags.More(m.args,
		[]string{"encap-csum", "+encap-csum"},
		[]string{"no-encap-csum", "-encap-csum"},
		[]string{"encap-udp6-csum", "+encap-udp6-csum"},
		[]string{"no-encap-udp6-csum", "-encap-udp6-csum"},
		[]string{"encap-remcsum", "+encap-remcsum"},
		[]string{"no-encap-remcsum", "-encap-remcsum"},
	)
	for _, x := range []struct {
		set   string
		unset string
		flag  uint16
	}{
		{"encap-csum", "no-encap-csum", rtnl.TUNNEL_ENCAP_FLAG_CSUM},
		{"encap-udp6-csum", "no-encap-udp6-csum",
			rtnl.TUNNEL_ENCAP_FLAG_CSUM6},
		{"encap-remcsum", "no-encap-remcsum",
			rtnl.TUNNEL_ENCAP_FLAG_REMCSUM},
	} {
		if m.opt.Flags.ByName[x.set] {
			eflags |= x.flag
		} else if m.opt.Flags.ByName[x.unset] {
			eflags &^= x.flag
		}
	}
	m.tinfo = append(m.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_IFLAGS, rtnl.Be16Attr(iflags)})
	m.tinfo = append(m.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_OFLAGS, rtnl.Be16Attr(oflags)})
	m.tinfo = append(m.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_ENCAP_FLAGS, rtnl.Uint16Attr(eflags)})
	m.tinfo = append(m.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_FLAGS, rtnl.Uint32Attr(greflags)})
	return nil
}
