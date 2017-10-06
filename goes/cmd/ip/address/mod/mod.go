// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"io"
	"strings"

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
	var ifa struct {
		hdr   rtnl.Hdr
		msg   rtnl.IfAddrMsg
		attrs []rtnl.Attr
		rsp   rtnl.Ifa
	}
	addattr := func(t uint16, v io.Reader) {
		ifa.attrs = append(ifa.attrs, rtnl.Attr{t, v})
	}

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if s := opt.Parms.ByName["-f"]; len(s) > 0 {
		if v, ok := rtnl.AfByName[s]; ok {
			ifa.msg.Family = v
		} else {
			return fmt.Errorf("family: %q unknown", s)
		}
	} else {
		ifa.msg.Family = rtnl.AF_UNSPEC
	}

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
		if pp, err := rtnl.Prefix(s, ifa.msg.Family); err == nil {
			if ifa.msg.Family == rtnl.AF_UNSPEC {
				ifa.msg.Family = pp.Family()
			}
			ifa.msg.Prefixlen = pp.Len()
			addattr(rtnl.IFA_ADDRESS, pp)
		} else {
			pa, err := rtnl.Address(s, ifa.msg.Family)
			if err != nil {
				return fmt.Errorf("peer: %v", err)
			}
			if ifa.msg.Family == rtnl.AF_UNSPEC {
				ifa.msg.Family = pa.Family()
			}
			ifa.msg.Prefixlen = map[uint8]uint8{
				rtnl.AF_INET:  32,
				rtnl.AF_INET6: 128,
				rtnl.AF_MPLS:  20,
			}[ifa.msg.Family]
			addattr(rtnl.IFA_ADDRESS, pa)
		}
		la, err := rtnl.Address(opt.Parms.ByName["local"],
			ifa.msg.Family)
		if err != nil {
			return fmt.Errorf("local: %v", err)
		}
		if la.IsLoopback() {
			ifa.msg.Scope = rtnl.RT_SCOPE_HOST
		}
		addattr(rtnl.IFA_LOCAL, la)
	} else {
		lp, err := rtnl.Prefix(opt.Parms.ByName["local"],
			ifa.msg.Family)
		if err != nil {
			return fmt.Errorf("local: %v", err)
		}
		if ifa.msg.Family == rtnl.AF_UNSPEC {
			ifa.msg.Family = lp.Family()
		}
		if lp.IsLoopback() {
			ifa.msg.Scope = rtnl.RT_SCOPE_HOST
		}
		ifa.msg.Prefixlen = lp.Len()
		addattr(rtnl.IFA_LOCAL, lp)
	}
	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"broadcast", rtnl.IFA_BROADCAST},
		{"anycast", rtnl.IFA_ANYCAST},
	} {
		if s := opt.Parms.ByName[x.name]; len(s) > 0 {
			a, err := rtnl.Address(s, ifa.msg.Family)
			if err != nil {
				return fmt.Errorf("%s: %v", x.name, err)
			}
			addattr(x.t, a)
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

func (Command) Complete(args ...string) (list []string) {
	var larg, llarg string
	n := len(args)
	if n > 0 {
		larg = args[n-1]
	}
	if n > 1 {
		llarg = args[n-2]
	}
	cpv := options.CompleteParmValue
	cpv["peer"] = options.NoComplete
	cpv["broadcast"] = options.NoComplete
	cpv["anycast"] = options.NoComplete
	cpv["label"] = options.NoComplete
	cpv["scope"] = rtnl.CompleteRtScope
	cpv["valid_lft"] = completeLft
	cpv["preferred_lft"] = completeLft
	cpv["dev"] = options.CompleteIfName
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"peer",
			"broadcast",
			"anycast",
			"label",
			"scope",
			"home",
			"mngtmpaddr",
			"nodad",
			"noprefixroute",
			"autojoin",
			"valid_lft",
			"preferred_lft",
			"dev") {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}

func completeLft(s string) (list []string) {
	const forever = "forever"
	if strings.HasPrefix("forever", s) {
		list = []string{forever}
	}
	return
}
