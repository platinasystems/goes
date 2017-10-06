// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show"
	Apropos = "network address"
	Usage   = `ip address show [ [dev] DEVICE ] [ scope SCOPE-ID ]
	[ to PREFIX ] [ FLAG-LIST ] [ label PATTERN ] [ master DEVICE ]
	[ type TYPE ] [ vrf NAME ] [ up ] ]

SCOPE-ID := [ host | link | global | NUMBER ]

FLAG-LIST := [ FLAG-LIST ] FLAG

FLAG := [ permanent | dynamic | secondary | primary | [-]tentative |
	[-]deprecated | [-]dadfailed | temporary | CONFFLAG-LIST ]

CONFFLAG-LIST := [ CONFFLAG-LIST ] CONFFLAG

CONFFLAG := [ home | mngtmpaddr | nodad | noprefixroute | autojoin ]

TYPE := { bridge | bridge_slave | bond | bond_slave | can | dummy |
	hsr | ifb | ipoib | macvlan | macvtap | vcan | veth | vlan |
	vxlan | ip6tnl | ipip | sit | gre | gretap | ip6gre |
	ip6gretap | vti | vrf | nlmon | ipvlan | lowpan | geneve |
	macsec }`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"permanent", "dynamic", "secondary", "primary",
		"-tentative", "tentative",
		"-deprecated", "deprecated",
		"-dadfailed", "dadfailed",
		"temporary",
		"home", "mngtmpaddr", "nodad", "noprefixroute", "autojoin",
		"up",
	}
	Parms = []interface{}{
		"dev", "scope", "to", "label", "master", "type", "vrf",
	}
)

func New(s string) Command { return Command(s) }

type Command string

type show options.Options

func (Command) Aka() string { return "show" }

func (c Command) Apropos() lang.Alt {
	apropos := Apropos
	if c == "show" {
		apropos += " (default)"
	}
	return lang.Alt{
		lang.EnUS: apropos,
	}
}

func (Command) Man() lang.Alt    { return man }
func (c Command) String() string { return string(c) }
func (Command) Usage() string    { return Usage }

func (Command) Main(args ...string) error {
	var req []byte
	var newifinfos [][]byte
	var to string
	var prefix uint8

	arphrd := ^uint16(0)
	rtscope := ^uint8(0)
	mindex := int32(-1)

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if n := len(args); n == 1 {
		opt.Parms.Set("dev", args[0])
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	label := opt.Parms.ByName["label"]
	if vrf := opt.Parms.ByName["vrf"]; len(vrf) > 0 {
		// FIXME is vrf an alias for master?
		opt.Parms.Set("master", vrf)
	}
	if name := opt.Parms.ByName["type"]; len(name) > 0 {
		if val, found := rtnl.ArphrdByName[name]; !found {
			return fmt.Errorf("type: %s: unknown", name)
		} else {
			arphrd = val
		}
	}
	if to := opt.Parms.ByName["to"]; len(to) > 0 {
		slash := strings.Index(to, "/")
		if slash < 0 || slash == 0 || slash == len(to)-1 {
			return fmt.Errorf("to: %s: invalid", to)
		}
		_, err := fmt.Sscan(to[slash+1:], &prefix)
		if err != nil {
			return fmt.Errorf("to: prefix: %s: %v",
				to[slash+1:], err)
		}
		to = to[:slash]
	}
	if name := opt.Parms.ByName["scope"]; len(name) > 0 {
		var found bool
		rtscope, found = rtnl.RtScopeByName[name]
		if !found {
			return fmt.Errorf("scope: %s: unknown", name)
		}
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	if req, err = rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETLINK,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
		},
		rtnl.IfInfoMsg{
			Family: rtnl.AF_UNSPEC,
		},
		rtnl.Attr{rtnl.IFLA_EXT_MASK, rtnl.RTEXT_FILTER_VF},
	); err != nil {
		return err
	} else if err = sr.UntilDone(req, func(b []byte) {
		var ifla rtnl.Ifla
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		msg := rtnl.IfInfoMsgPtr(b)
		ifla.Write(b)
		if dev := opt.Parms.ByName["dev"]; len(dev) > 0 {
			if dev != rtnl.Kstring(ifla[rtnl.IFLA_IFNAME]) {
				return
			}
		}
		if arphrd != ^uint16(0) {
			if msg.Type != arphrd {
				return
			}
		}
		if opt.Flags.ByName["up"] {
			// FIXME what about IFF_LOWER_UP
			const bits = rtnl.IFF_UP
			if bits != (msg.Flags & bits) {
				return
			}
		}
		if master := opt.Parms.ByName["master"]; len(master) > 0 {
			if master == rtnl.Kstring(ifla[rtnl.IFLA_IFNAME]) {
				mindex = msg.Index
				return
			}
		}
		newifinfos = append(newifinfos, b)
	}); err != nil {
		return err
	}

	ifaddrlistByIndex := make(map[uint32][][]byte)
	for _, af := range opt.Afs() {
		if req, err = rtnl.NewMessage(
			rtnl.Hdr{
				Type:  rtnl.RTM_GETADDR,
				Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
			},
			rtnl.RtGenMsg{
				Family: af,
			},
		); err != nil {
			return err
		} else if err = sr.UntilDone(req, func(b []byte) {
			var ifa rtnl.Ifa
			if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWADDR {
				return
			}
			msg := rtnl.IfAddrMsgPtr(b)
			ifa.Write(b)
			if len(to) > 0 {
				a := ifa[rtnl.IFA_ADDRESS]
				if len(a) == 0 || msg.Prefixlen != prefix ||
					rtnl.Kstring(a) != to {
					return
				}
			}
			if rtscope != ^uint8(0) && msg.Scope != rtscope {
				return
			}
			if len(label) > 0 {
				val := ifa[rtnl.IFA_LABEL]
				if len(val) == 0 ||
					rtnl.Kstring(val) != label {
					return
				}
			}
			idx := rtnl.IfAddrMsgPtr(b).Index
			ifaddrlist, found := ifaddrlistByIndex[idx]
			if !found {
				ifaddrlist = [][]byte{b}
			} else {
				ifaddrlist = append(ifaddrlist, b)
			}
			ifaddrlistByIndex[idx] = ifaddrlist
		}); err != nil {
			return err
		}
	}

	for _, ifinfo := range newifinfos {
		var ifla rtnl.Ifla
		msg := rtnl.IfInfoMsgPtr(ifinfo)
		if mindex != -1 {
			if msg.Index != mindex {
				continue
			}
		}
		ifaddrlist, found := ifaddrlistByIndex[uint32(msg.Index)]
		if !found || len(ifaddrlist) == 0 {
			continue
		}
		opt.ShowIfInfo(ifinfo)
		ifla.Write(ifinfo)
		if opt.Flags.ByName["-d"] {
			if val := ifla[rtnl.IFLA_VFINFO_LIST]; len(val) > 0 {
				rtnl.ForEachVfInfo(val, func(b []byte) {
					opt.ShowIflaVf(b)
				})
			}
		}
		for _, b := range ifaddrlist {
			const withCacheInfo = true
			opt.Println()
			opt.Nprint(4)
			opt.ShowIfAddr(b, withCacheInfo)
		}
		if opt.Flags.ByName["-s"] {
			opt.Println()
			val := ifla[rtnl.IFLA_STATS64]
			if len(val) == 0 {
				val = ifla[rtnl.IFLA_STATS]
			}
			if len(val) > 0 {
				opt.ShowIfStats(val)
			}
		}
		fmt.Println()
	}
	return nil
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
	cpv["dev"] = options.CompleteIfName
	cpv["master"] = options.CompleteIfName
	cpv["scope"] = rtnl.CompleteRtScope
	cpv["to"] = options.NoComplete
	cpv["label"] = options.NoComplete
	cpv["type"] = rtnl.CompleteArphrd
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"dev",
			"master",
			"scope",
			"to",
			"label",
			"type",
			"permanent",
			"dynamic",
			"secondary",
			"primary",
			"tentative",
			"-tentative",
			"deprecated",
			"-deprecated",
			"dadfailed",
			"-dadfailed",
			"temporary",
			"home",
			"mngtmpaddr",
			"nodad",
			"noprefixroute",
			"autojoin") {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}
