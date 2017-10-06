// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/netlink"
)

const (
	Name    = "show (default) | flush | get | save | restore"
	Apropos = "route table entry"
	Usage   = `
	ip route [ show ]
	ip route { show | flush } SELECTOR

	ip route save SELECTOR
	ip route restore

	ip route get ADDRESS [ from ADDRESS iif STRING  ] [ oif STRING ]
		[ tos TOS ] [ vrf NAME ]`
	Man = `
SEE ALSO
	ip man route || ip route -man
	man ip || ip -man`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{
		"cloned",
		"cached",
	}
	Parms = []interface{}{
		"to",
		"tos",
		"table",
		"vrf",
		"from",
		"protocol",
		"scope",
		"type",
		"dev",
		"via",
		"src",
		"realm",
		"realms",
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

func (c Command) Main(args ...string) error {
	var req []byte
	var to string
	var prefix uint8

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if n := len(args); n == 1 {
		opt.Parms.Set("to", args[0])
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	tbl := rtnl.RT_TABLE_MAIN
	if tname := opt.Parms.ByName["table"]; len(tname) > 0 {
		switch tname {
		case "all":
			tbl = rtnl.RT_TABLE_UNSPEC
		case "cache":
			// FIXME
		default:
			var found bool
			tbl, found = rtnl.RtTableByName[tname]
			if !found {
				_, err := fmt.Sscan(tname, &tbl)
				if err != nil {
					return fmt.Errorf("table: %s: unknown",
						tname)
				}
			}
		}
	}
	if to = opt.Parms.ByName["to"]; len(to) > 0 {
		slash := strings.Index(to, "/")
		if to != "default" {
			if slash < 0 || slash == 0 || slash == len(to)-1 {
				return fmt.Errorf("to: %s: invalid prefix", to)
			}
			_, err := fmt.Sscan(to[slash+1:], &prefix)
			if err != nil {
				return fmt.Errorf("to: prefix: %s: %v",
					to[slash+1:], err)
			}
			to = to[:slash]
		}
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	ifnames, err := sr.IfNamesByIndex()
	if err != nil {
		return err
	}

	for _, af := range opt.Afs() {
		if req, err = rtnl.NewMessage(
			rtnl.Hdr{
				Type:  rtnl.RTM_GETROUTE,
				Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
			},
			rtnl.RtGenMsg{
				Family: af,
			},
		); err != nil {
			return err
		} else if err = sr.UntilDone(req, func(b []byte) {
			if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWROUTE {
				return
			}
			var rta rtnl.Rta
			rta.Write(b)
			msg := rtnl.RtMsgPtr(b)
			if tbl != rtnl.RT_TABLE_UNSPEC {
				if tbl != rtnl.Uint32(rta[netlink.RTA_TABLE]) {
					return
				}
			}
			if len(to) > 0 {
				val := rta[rtnl.RTA_DST]
				if len(val) == 0 {
					if to != "default" {
						return
					}
				} else if msg.Dst_len != prefix {
					return
				} else if net.IP(val).String() != to {
					return
				}
			}
			opt.ShowRoute(b, ifnames)
			fmt.Println()
		}); err != nil {
			return err
		}
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
	cpv["to"] = options.NoComplete
	cpv["tos"] = options.NoComplete
	cpv["table"] = options.NoComplete
	cpv["vrf"] = options.NoComplete
	cpv["from"] = options.NoComplete
	cpv["protocol"] = rtnl.CompleteRtProt
	cpv["scope"] = rtnl.CompleteRtScope
	cpv["type"] = rtnl.CompleteRtn
	cpv["dev"] = options.CompleteIfName
	cpv["via"] = options.NoComplete
	cpv["src"] = options.NoComplete
	cpv["realm"] = options.NoComplete
	cpv["realms"] = options.NoComplete
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"cloned",
			"cached",
			"to",
			"tos",
			"table",
			"vrf",
			"from",
			"protocol",
			"scope",
			"type",
			"dev",
			"via",
			"src",
			"realm",
			"realms",
		) {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}
