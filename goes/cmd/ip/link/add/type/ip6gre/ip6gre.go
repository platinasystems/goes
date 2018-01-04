// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ip6gre

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command string

func (c Command) String() string { return string(c) }

func (c Command) Usage() string {
	return fmt.Sprint("ip link add type ", c, " [ OPTION ]...")
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a ip6gre[tap] virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
GRE TYPES
	ip6gre, ip6gretap

OPTIONS
	remote ADDR
		IPv6 address or tunnel's remote end-point

	local ADDR
		IPv6 address or tunnel's local end-point

	[no-][i|o]seq
		enable/disable sequencing of incomming and outgoing packets.

	[i|o]key KEY
	no-[i|o]key
		The GRE KEY may be a number or an IPv4 address-like dotted
		quad.  The "key" parameter uses the same key for both input and
		output.  The "ikey" and "okey" parameters specify different
		keys for input and output. "no-ikey" and "no-okey" remove the
		respective keys.

	[no-][i|o]csum
		Validate/generate checksums for tunneled packets.  The "ocsum"
		or "csum" flags calculate checksums for outgoing packets.  The
		"icsum" or "csum" flag validates the checksum of incoming
		packets have the correct checksum.


	hoplimit TTL
		Hop Limit of outgoing packets

	encaplimit ELIM
		Fixed encapsulation limit (default, 4)

	flowlabel FLOWLABEL
		fixed flowlabel

	tclass TCLASS
		traffic class of tunneled packets, which can be specified as
		either a two-digit hex value (e.g. c0) or a predefined string
		(e.g. internet).  The value inherit causes the field to be
		copied from the original IP header. The values inherit/STRING
		or inherit/00..ff will set the field to STRING or 00..ff when
		tunneling non-IP packets. The default value is 00.

	dscp { NUMBER | inherit }
	fwmark MARK
	dev DEVICE
	encap { fou | gue | none }
	encap-dport { PORT | any }
	encap-sport { PORT | any }
	[no-]encap-csum
	[no-]encap-csum6
	[no-]encap-remsum

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (c Command) Main(args ...string) error {
	var iflags, oflags, eflags uint16
	var greflags, flowinfo uint32
	var info nl.Attrs

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"seq", "+seq"},
		[]string{"no-seq", "-seq"},
		[]string{"iseq", "+iseq"},
		[]string{"no-iseq", "-iseq"},
		[]string{"oseq", "+oseq"},
		[]string{"no-oseq", "-oseq"},
		[]string{"no-key", "-key"},
		[]string{"no-ikey", "-ikey"},
		[]string{"no-okey", "-okey"},
		[]string{"csum", "+csum"},
		[]string{"no-csum", "-csum"},
		[]string{"icsum", "+icsum"},
		[]string{"no-icsum", "-icsum"},
		[]string{"ocsum", "+ocsum"},
		[]string{"no-ocsum", "-ocsum"},
		[]string{"no-encap", "-encap"},
		[]string{"encap-csum", "+encap-csum"},
		[]string{"no-encap-csum", "-encap-csum"},
		[]string{"encap-udp6-csum", "+encap-udp6-csum"},
		[]string{"no-encap-udp6-csum", "-encap-udp6-csum"},
		[]string{"encap-remcsum", "+encap-remcsum"},
		[]string{"no-encap-remcsum", "-encap-remcsum"},
	)
	opt.Parms.More(args,
		"remote",
		"local",
		"key",
		"ikey",
		"okey",
		"dev",
		[]string{"ttl", "hoplimit"},
		[]string{"tos", "tclass", "defiled"},
		"flowlabel",
		"dscp",
		"fwmark",
		"encaplimit",
		"encap",
		"encap-sport",
		"encap-dport",
	)

	err := opt.OnlyName(args)
	if err != nil {
		return err
	}

	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt)
	if err != nil {
		return err
	}

	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"local", rtnl.IFLA_GRE_LOCAL},
		{"remote", rtnl.IFLA_GRE_REMOTE},
	} {
		if s := opt.Parms.ByName[x.name]; len(s) > 0 {
			if ip4 := net.ParseIP(s).To4(); ip4 == nil {
				return fmt.Errorf("%s: %q invalid", x.name, s)
			} else {
				info = append(info, nl.Attr{x.t,
					nl.BytesAttr(ip4)})
			}
		}
	}
	if opt.Flags.ByName["seq"] {
		iflags |= rtnl.GRE_SEQ
		oflags |= rtnl.GRE_SEQ
	} else if opt.Flags.ByName["no-seq"] {
		iflags &^= rtnl.GRE_SEQ
		oflags &^= rtnl.GRE_SEQ
	} else {
		if opt.Flags.ByName["iseq"] {
			iflags |= rtnl.GRE_SEQ
		} else if opt.Flags.ByName["no-iseq"] {
			iflags &^= rtnl.GRE_SEQ
		}
		if opt.Flags.ByName["oseq"] {
			oflags |= rtnl.GRE_SEQ
		} else if opt.Flags.ByName["no-oseq"] {
			oflags &^= rtnl.GRE_SEQ
		}
	}
	if s := opt.Parms.ByName["key"]; len(s) > 0 {
		var u32 uint32
		if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("key: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_IKEY,
			nl.Uint32Attr(u32)})
		info = append(info, nl.Attr{rtnl.IFLA_GRE_OKEY,
			nl.Uint32Attr(u32)})
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
			s := opt.Parms.ByName[x.name]
			if len(s) == 0 {
				continue
			}
			if _, err := fmt.Sscan(s, &u32); err != nil {
				return fmt.Errorf("%s: %q %v", x.name, s, err)
			}
			info = append(info, nl.Attr{x.t,
				nl.Uint32Attr(u32)})
			*(x.flags) |= rtnl.GRE_KEY
		}
	}
	if opt.Flags.ByName["no-key"] {
		iflags &^= rtnl.GRE_KEY
		oflags &^= rtnl.GRE_KEY
	} else {
		if opt.Flags.ByName["no-ikey"] {
			iflags &^= rtnl.GRE_KEY
		}
		if opt.Flags.ByName["no-okey"] {
			oflags &^= rtnl.GRE_KEY
		}
	}
	if opt.Flags.ByName["csum"] {
		iflags |= rtnl.GRE_CSUM
		oflags |= rtnl.GRE_CSUM
	} else if opt.Flags.ByName["no-csum"] {
		iflags &^= rtnl.GRE_CSUM
		oflags &^= rtnl.GRE_CSUM
	} else {
		if opt.Flags.ByName["icsum"] {
			iflags |= rtnl.GRE_CSUM
		} else if opt.Flags.ByName["no-icsum"] {
			iflags &^= rtnl.GRE_CSUM
		}
		if opt.Flags.ByName["ocsum"] {
			oflags |= rtnl.GRE_CSUM
		} else if opt.Flags.ByName["no-ocsum"] {
			oflags &^= rtnl.GRE_CSUM
		}
	}
	if s := opt.Parms.ByName["dev"]; len(s) > 0 {
		dev, found := rtnl.If.IndexByName[s]
		if !found {
			return fmt.Errorf("dev: %q not found", s)
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_LINK,
			nl.Uint32Attr(dev)})
	}
	if s := opt.Parms.ByName["ttl"]; len(s) > 0 {
		var u8 uint8
		if _, err := fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("ttl: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_TTL,
			nl.Uint8Attr(u8)})
	}
	if s := opt.Parms.ByName["tos"]; len(s) > 0 {
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
	if s := opt.Parms.ByName["flowlabel"]; len(s) > 0 {
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
		info = append(info, nl.Attr{rtnl.IFLA_GRE_FLOWINFO,
			nl.Uint32Attr(flowinfo)})
	}
	if s := opt.Parms.ByName["dscp"]; len(s) > 0 {
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_RCV_DSCP_COPY
		} else {
			return fmt.Errorf("dscp: %q invalid", s)
		}
	}
	if s := opt.Parms.ByName["fwmark"]; len(s) > 0 {
		var u32 uint32
		if s == "inherit" {
			greflags |= rtnl.IP6_TNL_F_USE_ORIG_FWMARK
		} else if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("fwmark: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_USE_ORIG_FWMARK
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_FWMARK,
			nl.Uint32Attr(u32)})
	}
	if s := opt.Parms.ByName["encaplimit"]; len(s) > 0 {
		var u8 uint32
		if s == "node" {
			greflags |= rtnl.IP6_TNL_F_IGN_ENCAP_LIMIT
		} else if _, err := fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("encaplimit: %q %v", s, err)
		} else {
			greflags &^= rtnl.IP6_TNL_F_IGN_ENCAP_LIMIT
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_ENCAP_LIMIT,
			nl.Uint8Attr(u8)})
	}
	if opt.Flags.ByName["no-encap"] {
		info = append(info, nl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
			nl.Uint32Attr(rtnl.TUNNEL_ENCAP_NONE)})
	} else {
		if s := opt.Parms.ByName["encap"]; len(s) > 0 {
			if encap, found := map[string]uint16{
				"fou":  rtnl.TUNNEL_ENCAP_FOU,
				"gue":  rtnl.TUNNEL_ENCAP_GUE,
				"none": rtnl.TUNNEL_ENCAP_NONE,
			}[s]; !found {
				return fmt.Errorf("encap: %q unknown", s)
			} else {
				info = append(info,
					nl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
						nl.Uint32Attr(encap)})
			}
		}
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"encap-sport", rtnl.IFLA_GRE_ENCAP_SPORT},
		{"encap-sport", rtnl.IFLA_GRE_ENCAP_DPORT},
	} {
		if s := opt.Parms.ByName[x.name]; len(s) > 0 {
			var u16 uint16
			if s != "any" {
				if _, err := fmt.Sscan(s, &u16); err != nil {
					return fmt.Errorf("%s: %q %v",
						x.name, s, err)
				}
			}
			info = append(info,
				nl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
					nl.Be16Attr(u16)})
		}
	}
	for _, x := range []struct {
		name string
		flag uint16
	}{
		{"encap-csum", rtnl.TUNNEL_ENCAP_FLAG_CSUM},
		{"encap-udp6-csum", rtnl.TUNNEL_ENCAP_FLAG_CSUM6},
		{"encap-remcsum", rtnl.TUNNEL_ENCAP_FLAG_REMCSUM},
	} {
		if opt.Flags.ByName[x.name] {
			eflags |= x.flag
		} else if opt.Flags.ByName["no-"+x.name] {
			eflags &^= x.flag
		}
	}

	info = append(info, nl.Attr{rtnl.IFLA_GRE_IFLAGS, nl.Be16Attr(iflags)})
	info = append(info, nl.Attr{rtnl.IFLA_GRE_OFLAGS, nl.Be16Attr(oflags)})
	info = append(info, nl.Attr{rtnl.IFLA_GRE_ENCAP_FLAGS,
		nl.Uint16Attr(eflags)})
	info = append(info, nl.Attr{rtnl.IFLA_GRE_FLAGS,
		nl.Uint32Attr(greflags)})

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(c)},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
