// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"sort"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/group"
	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show"
	Apropos = "link attributes"
	Usage   = `ip link show [ [dev] DEVICE | group GROUP ] [ up ]
	[ master DEVICE ] [ type ETYPE ] [ vrf NAME ]`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"up",
	}
	Parms = []interface{}{
		"dev",
		"group",
		"master",
		"type",
		"vrf",
	}
)

func New(name string) Command { return Command(name) }

type Command string

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
	var gid uint32 // default: 0
	var newifinfos [][]byte
	arphrd := ^uint16(0)
	mindex := int32(-1)

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if n := len(args); n == 1 {
		opt.Parms.Set("dev", args[0])
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	if vrf := opt.Parms.ByName["vrf"]; len(vrf) > 0 {
		// FIXME is vrf an alias for master?
		opt.Parms.Set("master", vrf)
	}
	if gname := opt.Parms.ByName["group"]; len(gname) > 0 {
		gid = group.Id(gname)
	}
	if name := opt.Parms.ByName["type"]; len(name) > 0 {
		if val, found := rtnl.ArphrdByName[name]; !found {
			return fmt.Errorf("type: %s: unknown", name)
		} else {
			arphrd = val
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
	}
	if err = sr.UntilDone(req, func(b []byte) {
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
		if val := ifla[rtnl.IFLA_GROUP]; len(val) > 0 {
			if gid != rtnl.Uint32(val) {
				return
			}
		} else if gid != 0 {
			return
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

	if len(newifinfos) == 0 {
		return fmt.Errorf("no info")
	}

	sort.Slice(newifinfos, func(i, j int) bool {
		iIndex := rtnl.IfInfoMsgPtr(newifinfos[i]).Index
		jIndex := rtnl.IfInfoMsgPtr(newifinfos[j]).Index
		return iIndex < jIndex
	})

	for _, b := range newifinfos {
		var ifla rtnl.Ifla
		msg := rtnl.IfInfoMsgPtr(b)
		if mindex != -1 {
			if msg.Index != mindex {
				continue
			}
		}
		opt.ShowIfInfo(b)
		ifla.Write(b)
		if opt.Flags.ByName["-s"] {
			val := ifla[rtnl.IFLA_STATS64]
			if len(val) == 0 {
				val = ifla[rtnl.IFLA_STATS]
			}
			if len(val) > 0 {
				opt.ShowIfStats(val)
			}
		}
		if val := ifla[rtnl.IFLA_VFINFO_LIST]; len(val) > 0 {
			rtnl.ForEachVfInfo(val, func(b []byte) {
				opt.ShowIflaVf(b)
			})
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
	cpv["group"] = group.Complete
	cpv["type"] = rtnl.CompleteType
	cpv["vrf"] = options.NoComplete
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"dev",
			"group",
			"type",
			"vrf",
		) {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}
