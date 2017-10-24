// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipip

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
	Name    = "ipip"
	Apropos = "add an ipip virtual link"
	Usage   = "ip link add type ipip [ OPTIONS ]..."
	Man     = `
OPTIONS
	remote ADDR
	local ADDR

	ttl { 1:255 }

	tos { 1:8 }
	dev DEVICE

	no-encap | encap { fou | gue | none } ]

	encap-dport PORT
	encap-sport { PORT | auto }

	mode {
		[ ipip | ip4ip4 | ip4/ip4 ] |
		[ mplsip | mplsip4 | mpls/ip4 ] |
		[ any | anyip4 | any/ip4 ]
	}

	[no-]encap-csum
	[no-]pmtudisc ]`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var info nl.Attrs
	var encapflags uint16

	opt, args := options.New(args)
	args = opt.Flags.More(args,
		[]string{"no-encap", "-encap"},
		[]string{"encap-csum", "+encap-csum"},
		[]string{"no-encap-csum", "-encap-csum"},
		[]string{"pmtudisc", "+pmtudisc"},
		[]string{"no-pmtudisc", "-pmtudisc"},
	)
	args = opt.Parms.More(args,
		"remote",
		"local",
		"dev",
		[]string{"ttl", "hoplimit"},
		[]string{"tos", "tclass", "dstfield"},
		"encap",
		"encap-sport",
		"encap-dport",
		"mode",
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
		{"local", rtnl.IFLA_IPTUN_LOCAL},
		{"remote", rtnl.IFLA_IPTUN_REMOTE},
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
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"ttl", rtnl.IFLA_IPTUN_TTL},
		{"tos", rtnl.IFLA_IPTUN_TOS},
	} {
		if s := opt.Parms.ByName[x.name]; len(s) > 0 {
			var u8 uint8
			if _, err := fmt.Sscan(s, &u8); err != nil {
				return fmt.Errorf("%s: %q %v", x.name, s, err)
			}
			info = append(info, nl.Attr{x.t, nl.Uint8Attr(u8)})
		}
	}
	if s := opt.Parms.ByName["dev"]; len(s) > 0 {
		dev, found := rtnl.If.IndexByName[s]
		if !found {
			return fmt.Errorf("dev: %q not found", s)
		}
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_LINK,
			nl.Uint32Attr(dev)})
	}
	if opt.Flags.ByName["pmtudisc"] {
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_PMTUDISC,
			nl.Uint8Attr(1)})
	} else if opt.Flags.ByName["no-pmtudisc"] {
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_PMTUDISC,
			nl.Uint8Attr(0)})
	}
	if opt.Flags.ByName["no-encap"] {
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_ENCAP_TYPE,
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
					nl.Attr{rtnl.IFLA_IPTUN_ENCAP_TYPE,
						nl.Uint32Attr(encap)})
			}
		}
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"encap-sport", rtnl.IFLA_IPTUN_ENCAP_SPORT},
		{"encap-dport", rtnl.IFLA_IPTUN_ENCAP_DPORT},
	} {
		if s := opt.Parms.ByName[x.name]; len(s) > 0 {
			var u16 uint16
			if s != "any" {
				if _, err := fmt.Sscan(s, &u16); err != nil {
					return fmt.Errorf("%s: %q %v",
						x.name, s, err)
				}
			}
			info = append(info, nl.Attr{rtnl.IFLA_IPTUN_ENCAP_TYPE,
				nl.Be16Attr(u16)})
		}
	}
	switch s := opt.Parms.ByName["mode"]; s {
	case "", "any", "anyip4", "any/ip4":
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_PROTO,
			nl.Uint8Attr(0)})
	case "ipip", "ip4ip4", "ip4/ip4":
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_PROTO,
			nl.Uint8Attr(rtnl.IPPROTO_IPIP)})
	case "mplsip", "mplsip4", "mpls/ip4":
		info = append(info, nl.Attr{rtnl.IFLA_IPTUN_PROTO,
			nl.Uint8Attr(rtnl.IPPROTO_MPLS)})
	default:
		return fmt.Errorf("%q: unknown encap", s)
	}
	for _, x := range []struct {
		set   string
		unset string
		flag  uint16
	}{
		{"encap-csum", "no-encap-csum", rtnl.TUNNEL_ENCAP_FLAG_CSUM},
	} {
		if opt.Flags.ByName[x.set] {
			encapflags |= x.flag
		} else if opt.Flags.ByName[x.unset] {
			encapflags &^= x.flag
		}
	}

	info = append(info, nl.Attr{rtnl.IFLA_GRE_ENCAP_FLAGS,
		nl.Uint16Attr(encapflags)})

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(Name)},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
