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
func (c *Command) parseTypeIp6Gre() error {
	var iflags, oflags, eflags uint16
	var greflags, flowinfo uint32
	c.args = c.opt.Parms.More(c.args, "remote", "local")
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"local", rtnl.IFLA_GRE_LOCAL},
		{"remote", rtnl.IFLA_GRE_REMOTE},
	} {
		if s := c.opt.Parms.ByName[x.name]; len(s) > 0 {
			if ip4 := net.ParseIP(s).To4(); ip4 == nil {
				return fmt.Errorf("%s: %q invalid", x.name, s)
			} else {
				c.tinfo = append(c.tinfo,
					rtnl.Attr{x.t, rtnl.BytesAttr(ip4)})
			}
		}
	}
	c.args = c.opt.Flags.More(c.args,
		[]string{"seq", "+seq"},
		[]string{"no-seq", "-seq"},
		[]string{"iseq", "+iseq"},
		[]string{"no-iseq", "-iseq"},
		[]string{"oseq", "+oseq"},
		[]string{"no-oseq", "-oseq"},
	)
	if c.opt.Flags.ByName["seq"] {
		iflags |= rtnl.GRE_SEQ
		oflags |= rtnl.GRE_SEQ
	} else if c.opt.Flags.ByName["no-seq"] {
		iflags &^= rtnl.GRE_SEQ
		oflags &^= rtnl.GRE_SEQ
	} else {
		if c.opt.Flags.ByName["iseq"] {
			iflags |= rtnl.GRE_SEQ
		} else if c.opt.Flags.ByName["no-iseq"] {
			iflags &^= rtnl.GRE_SEQ
		}
		if c.opt.Flags.ByName["oseq"] {
			oflags |= rtnl.GRE_SEQ
		} else if c.opt.Flags.ByName["no-oseq"] {
			oflags &^= rtnl.GRE_SEQ
		}
	}
	c.args = c.opt.Parms.More(c.args, "key", "ikey", "okey")
	if s := c.opt.Parms.ByName["key"]; len(s) > 0 {
		var u32 uint32
		if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("key: %q %v", s, err)
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_IKEY, rtnl.Uint32Attr(u32)})
		c.tinfo = append(c.tinfo,
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
			s := c.opt.Parms.ByName[x.name]
			if len(s) == 0 {
				continue
			}
			if _, err := fmt.Sscan(s, &u32); err != nil {
				return fmt.Errorf("%s: %q %v", x.name, s, err)
			}
			c.tinfo = append(c.tinfo,
				rtnl.Attr{x.t, rtnl.Uint32Attr(u32)})
			*(x.flags) |= rtnl.GRE_KEY
		}
	}
	c.args = c.opt.Flags.More(c.args,
		[]string{"no-key", "-key"},
		[]string{"no-ikey", "-ikey"},
		[]string{"no-okey", "-okey"},
	)
	if c.opt.Flags.ByName["no-key"] {
		iflags &^= rtnl.GRE_KEY
		oflags &^= rtnl.GRE_KEY
	} else {
		if c.opt.Flags.ByName["no-ikey"] {
			iflags &^= rtnl.GRE_KEY
		}
		if c.opt.Flags.ByName["no-okey"] {
			oflags &^= rtnl.GRE_KEY
		}
	}
	c.args = c.opt.Flags.More(c.args,
		[]string{"csum", "+csum"},
		[]string{"no-csum", "-csum"},
		[]string{"icsum", "+icsum"},
		[]string{"no-icsum", "-icsum"},
		[]string{"ocsum", "+ocsum"},
		[]string{"no-ocsum", "-ocsum"},
	)
	if c.opt.Flags.ByName["csum"] {
		iflags |= rtnl.GRE_CSUM
		oflags |= rtnl.GRE_CSUM
	} else if c.opt.Flags.ByName["no-csum"] {
		iflags &^= rtnl.GRE_CSUM
		oflags &^= rtnl.GRE_CSUM
	} else {
		if c.opt.Flags.ByName["icsum"] {
			iflags |= rtnl.GRE_CSUM
		} else if c.opt.Flags.ByName["no-icsum"] {
			iflags &^= rtnl.GRE_CSUM
		}
		if c.opt.Flags.ByName["ocsum"] {
			oflags |= rtnl.GRE_CSUM
		} else if c.opt.Flags.ByName["no-ocsum"] {
			oflags &^= rtnl.GRE_CSUM
		}
	}
	c.args = c.opt.Parms.More(c.args, "dev")
	if s := c.opt.Parms.ByName["dev"]; len(s) > 0 {
		dev, found := c.ifindexByName[s]
		if !found {
			return fmt.Errorf("dev: %q not found", s)
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_LINK, rtnl.Uint32Attr(dev)})
	}
	c.args = c.opt.Parms.More(c.args, []string{"ttl", "hoplimit"})
	if s := c.opt.Parms.ByName["ttl"]; len(s) > 0 {
		var u8 uint8
		if _, err := fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("ttl: %q %v", s, err)
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_TTL, rtnl.Uint8Attr(u8)})
	}
	c.args = c.opt.Parms.More(c.args, []string{"tos", "tclass", "defiled"})
	if s := c.opt.Parms.ByName["tos"]; len(s) > 0 {
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
	c.args = c.opt.Parms.More(c.args, "flowlabel")
	if s := c.opt.Parms.ByName["flowlabel"]; len(s) > 0 {
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
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_FLOWINFO,
				rtnl.Uint32Attr(flowinfo)})
	}
	c.args = c.opt.Parms.More(c.args, "dscp")
	if s := c.opt.Parms.ByName["dscp"]; len(s) > 0 {
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_RCV_DSCP_COPY
		} else {
			return fmt.Errorf("dscp: %q invalid", s)
		}
	}
	c.args = c.opt.Parms.More(c.args, "fwmark")
	if s := c.opt.Parms.ByName["fwmark"]; len(s) > 0 {
		var u32 uint32
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_USE_ORIG_FWMARK
		} else if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("fwmark: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_USE_ORIG_FWMARK
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_FWMARK, rtnl.Uint32Attr(u32)})
	}
	c.args = c.opt.Parms.More(c.args, "encaplimit")
	if s := c.opt.Parms.ByName["encaplimit"]; len(s) > 0 {
		var u8 uint32
		if s == "node" {
			greflags |= rtnl.IP6_TNL_F_IGN_ENCAP_LIMIT
		} else if _, err := fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("encaplimit: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_IGN_ENCAP_LIMIT
		}
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_ENCAP_LIMIT,
				rtnl.Uint8Attr(u8)})
	}
	c.args = c.opt.Flags.More(c.args, []string{"no-encap", "-encap"})
	if c.opt.Flags.ByName["no-encap"] {
		c.tinfo = append(c.tinfo,
			rtnl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
				rtnl.Uint32Attr(rtnl.TUNNEL_ENCAP_NONE)})
	} else {
		c.args = c.opt.Parms.More(c.args, "encap")
		if s := c.opt.Parms.ByName["encap"]; len(s) > 0 {
			if encap, found := map[string]uint16{
				"fou":  rtnl.TUNNEL_ENCAP_FOU,
				"gue":  rtnl.TUNNEL_ENCAP_GUE,
				"none": rtnl.TUNNEL_ENCAP_NONE,
			}[s]; !found {
				return fmt.Errorf("encap: %q unknown", s)
			} else {
				c.tinfo = append(c.tinfo,
					rtnl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
						rtnl.Uint32Attr(encap)})
			}
		}
	}
	c.args = c.opt.Parms.More(c.args, "encap-sport", "encap-dport")
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"encap-sport", rtnl.IFLA_GRE_ENCAP_SPORT},
		{"encap-sport", rtnl.IFLA_GRE_ENCAP_DPORT},
	} {
		if s := c.opt.Parms.ByName[x.name]; len(s) > 0 {
			var u16 uint16
			if s != "any" {
				if _, err := fmt.Sscan(s, &u16); err != nil {
					return fmt.Errorf("%s: %q %v",
						x.name, s, err)
				}
			}
			c.tinfo = append(c.tinfo,
				rtnl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
					rtnl.Be16Attr(u16)})
		}
	}
	c.args = c.opt.Flags.More(c.args,
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
		if c.opt.Flags.ByName[x.set] {
			eflags |= x.flag
		} else if c.opt.Flags.ByName[x.unset] {
			eflags &^= x.flag
		}
	}
	c.tinfo = append(c.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_IFLAGS, rtnl.Be16Attr(iflags)})
	c.tinfo = append(c.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_OFLAGS, rtnl.Be16Attr(oflags)})
	c.tinfo = append(c.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_ENCAP_FLAGS, rtnl.Uint16Attr(eflags)})
	c.tinfo = append(c.tinfo,
		rtnl.Attr{rtnl.IFLA_GRE_FLAGS, rtnl.Uint32Attr(greflags)})
	return nil
}
