// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package gre

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Apropos = "add a gre[tap] virtual link"
	Man     = `
GRE TYPES
	gre, gretap

OPTIONS
	remote ADDR
		specifies the remote address of the tunnel.

	local ADDR
		specifies the fixed local address for tunneled packets.  It
		must be an address on another interface on this host.

	[i|o]key KEY

	ttl { 1:255 }
		Time-to-live of transimitted packets

	tos { 1: 8 }
		Type-of-service of transimitted packets

	fwmark NUMBER

	dev DEVICE
	       physical device of tunnel endpoint

	encap { fou | gue | none }
		specifies type of secondary UDP encapsulation. "fou" indicates
		Foo-Over-UDP, "gue" indicates Generic UDP Encapsulation.

	encap-dport { PORT | any }

	encap-sport { PORT | any }
		specifies the source port in UDP encapsulation.  PORT indicates
		the port by number, "auto" indicates that the port number
		should be chosen automatically (the kernel picks a flow based
		on the flow hash of the encapsulated packet).

	[no-][i|o]seq

	no-[i|o]key

	[no-][i|o]csum

	[no-]pmtudisc

	[no-]encap-csum
		specifies if UDP checksums are enabled in the secondary
		encapsulation.

	[no-]encap-csum6

	[no-]encap-remsum
		specifies if Remote Checksum Offload is enabled.  This is only
		applicable for Generic UDP Encapsulation.


SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Types = []string{
		"gre",
		"gretap",
	}
)

func New(s string) Command { return Command(s) }

type Command string

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (c Command) Usage() string {
	return fmt.Sprint("ip link add type ", c, " [ OPTION ]...")
}

func (c Command) Main(args ...string) error {
	var iflags, oflags, eflags uint16
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
		[]string{"pmtudisc", "+pmtudisc"},
		[]string{"no-pmtudisc", "-pmtudisc"},
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
		[]string{"ttl", "hoplimit"},
		[]string{"tos", "tclass"},
		"fwmark",
		"dev",
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
			info = append(info, nl.Attr{x.t, nl.Uint32Attr(u32)})
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
	if opt.Flags.ByName["pmtudisc"] {
		info = append(info, nl.Attr{rtnl.IFLA_GRE_PMTUDISC,
			nl.Uint8Attr(1)})
	} else if opt.Flags.ByName["no-pmtudisc"] {
		info = append(info, nl.Attr{rtnl.IFLA_GRE_PMTUDISC,
			nl.Uint8Attr(0)})
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"ttl", rtnl.IFLA_GRE_TTL},
		{"tos", rtnl.IFLA_GRE_TOS},
	} {
		if s := opt.Parms.ByName[x.name]; len(s) > 0 {
			var u8 uint8
			if _, err := fmt.Sscan(s, &u8); err != nil {
				return fmt.Errorf("%s: %q %v", x.name, s, err)
			}
			info = append(info, nl.Attr{x.t, nl.Uint8Attr(u8)})
		}
	}
	if s := opt.Parms.ByName["fwmark"]; len(s) > 0 {
		var u32 uint32
		if _, err := fmt.Sscan(s, &u32); err != nil {
			return fmt.Errorf("fwmark: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_FWMARK,
			nl.Uint32Attr(u32)})
	}
	if s := opt.Parms.ByName["dev"]; len(s) > 0 {
		dev, found := rtnl.If.IndexByName[s]
		if !found {
			return fmt.Errorf("dev: %q not found", s)
		}
		info = append(info, nl.Attr{rtnl.IFLA_GRE_LINK,
			nl.Uint32Attr(dev)})
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
			info = append(info, nl.Attr{rtnl.IFLA_GRE_ENCAP_TYPE,
				nl.Be16Attr(u16)})
		}
	}
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
		if opt.Flags.ByName[x.set] {
			eflags |= x.flag
		} else if opt.Flags.ByName[x.unset] {
			eflags &^= x.flag
		}
	}
	info = append(info, nl.Attr{rtnl.IFLA_GRE_IFLAGS,
		nl.Be16Attr(iflags)})
	info = append(info, nl.Attr{rtnl.IFLA_GRE_OFLAGS,
		nl.Be16Attr(oflags)})
	info = append(info, nl.Attr{rtnl.IFLA_GRE_ENCAP_FLAGS,
		nl.Uint16Attr(eflags)})

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
