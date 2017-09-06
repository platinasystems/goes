// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const Apropos = "network address"

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"home",
		"mngtmpaddr",
		"nodad",
		"noprefixroute",
		"autojoin",
	}
	Parms = []interface{}{
		// IFADDR
		"local", "peer", "broadcast", "anycast", "label", "scope",
		"dev",
		// LIFETIME
		"valid_lft", "preferred_lft",
	}
)

func New(name string) Command { return Command(name) }

type Command string

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (c Command) Usage() string {
	return fmt.Sprint("ip address ", c, ` IFADDR [ dev ] IFNAME
	[ LIFETIME ] [ CONFFLAG-LIST ]

MOD := { add | change | delete | replace }

IFADDR := PREFIX | ADDR peer PREFIX [ broadcast ADDR ]
	[ anycast ADDR ] [ label LABEL ] [ scope SCOPE-ID ]

SCOPE-ID := { host | link | global | NUMBER }

CONFFLAG-LIST := [ CONFFLAG-LIST ] CONFFLAG

CONFFLAG := { home | mngtmpaddr | nodad | noprefixroute | autojoin }

LIFETIME := [ valid_lft LFT ] [ preferred_lft LFT ]

LFT := { forever | SECONDS }`)
}

func (c Command) Main(args ...string) error {
	var err error
	var ifa struct {
		hdr   rtnl.Hdr
		msg   rtnl.IfAddrMsg
		attrs []rtnl.Attr
		rsp   rtnl.Ifa
	}
	var local, peer net.IP
	var ipnet *net.IPNet

	if args, err = options.Netns(args); err != nil {
		return err
	}

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if len(opt.Parms.ByName["local"]) == 0 {
		if len(args) == 0 {
			return fmt.Errorf("IFADDR: missing")
		}
		opt.Parms.ByName["local"] = args[0]
		args = args[1:]
	}
	if len(opt.Parms.ByName["dev"]) == 0 {
		if len(args) == 0 {
			return fmt.Errorf("IFNAME: missing")
		}
		opt.Parms.ByName["dev"] = args[0]
		args = args[1:]
	}
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	if s := opt.Parms.ByName["peer"]; len(s) > 0 {
		peer, ipnet, err = net.ParseCIDR(s)
		if err != nil {
			err = nil
			peer = net.ParseIP(s)
			if peer == nil {
				return fmt.Errorf("peer: %q invalid prefix", s)
			}
		}
		local = net.ParseIP(opt.Parms.ByName["local"])
		if local == nil {
			return fmt.Errorf("local: %q invalid address", s)
		}
	} else {
		s = opt.Parms.ByName["local"]
		if local, ipnet, err = net.ParseCIDR(s); err != nil {
			return fmt.Errorf("local: %q invalid prefix", s)
		}
	}

	if local.To4() != nil {
		ifa.msg.Family = rtnl.AF_INET
	} else if local.To16() != nil {
		ifa.msg.Family = rtnl.AF_INET6
	} else {
		return fmt.Errorf("local: %q unsupported network", local)
	}

	addrattr := func(name string, ip net.IP, t uint16) error {
		var ipv net.IP
		if ip == nil {
			return fmt.Errorf("%s: invalid address", name)
		}
		if ifa.msg.Family == rtnl.AF_INET {
			ipv = ip.To4()
		} else {
			ipv = ip.To16()
		}
		if ipv == nil {
			return fmt.Errorf("%s: %s: wrong network", name, ip)
		}
		ifa.attrs = append(ifa.attrs,
			rtnl.Attr{t, rtnl.BytesAttr([]byte(ipv))})
		return nil
	}

	if ipnet != nil {
		ones, _ := ipnet.Mask.Size()
		ifa.msg.Prefixlen = uint8(ones)
	} else if ifa.msg.Family == rtnl.AF_INET {
		ifa.msg.Prefixlen = 32
	} else {
		ifa.msg.Prefixlen = 128
	}

	if err = addrattr("local", local, rtnl.IFA_LOCAL); err != nil {
		return err
	}
	if peer != nil {
		err = addrattr("peer", peer, rtnl.IFA_ADDRESS)
		if err != nil {
			return err
		}
	}
	if s := opt.Parms.ByName["broadcast"]; len(s) > 0 {
		err = addrattr("broadcast", net.ParseIP(s), rtnl.IFA_BROADCAST)
		if err != nil {
			return err
		}
	}
	if s := opt.Parms.ByName["anycast"]; len(s) > 0 {
		err = addrattr("anycast", net.ParseIP(s), rtnl.IFA_ANYCAST)
		if err != nil {
			return err
		}
	}
	if s := opt.Parms.ByName["label"]; len(s) > 0 {
		ifa.attrs = append(ifa.attrs, rtnl.Attr{rtnl.IFA_LABEL,
			rtnl.KstringAttr(s)})
	}
	if s := opt.Parms.ByName["scope"]; len(s) > 0 {
		var found bool
		ifa.msg.Scope, found = rtnl.RtScopeByName[s]
		if !found {
			return fmt.Errorf("scope: %q invalid", s)
		}
	} else if local.IsLoopback() {
		ifa.msg.Scope = rtnl.RT_SCOPE_HOST
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	ifa.msg.Index = ^uint32(0)
	if req, err := rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETLINK,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK,
		},
		rtnl.IfInfoMsg{
			Family: rtnl.AF_UNSPEC,
		},
		rtnl.Attr{rtnl.IFLA_IFNAME,
			rtnl.KstringAttr(opt.Parms.ByName["dev"])},
	); err != nil {
		return err
	} else if err = sr.UntilDone(req, func(b []byte) {
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		ifa.msg.Index = uint32(rtnl.IfInfoMsgPtr(b).Index)
	}); err != nil {
		return err
	}
	if ifa.msg.Index == ^uint32(0) {
		return fmt.Errorf("%s: not found", opt.Parms.ByName["dev"])
	}

	ifa.hdr.Flags = rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK

	switch c {
	case "delete", "xdelete":
		ifa.hdr.Type = rtnl.RTM_DELADDR
	case "add":
		ifa.hdr.Type = rtnl.RTM_NEWADDR
		ifa.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_EXCL
	case "change":
		ifa.hdr.Type = rtnl.RTM_NEWADDR
		ifa.hdr.Flags |= rtnl.NLM_F_REPLACE
	case "replace":
		ifa.hdr.Type = rtnl.RTM_NEWADDR
		ifa.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_EXCL
	default:
		return fmt.Errorf("%q: unknown", c)
	}

	b, err := rtnl.NewMessage(ifa.hdr, ifa.msg, ifa.attrs...)
	if err != nil {
		return err
	}
	return sr.UntilDone(b, func([]byte) {})
}
