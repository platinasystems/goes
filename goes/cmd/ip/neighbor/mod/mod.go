// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Apropos = "neighbor link address"
	Man     = `
SEE ALSO
	ip man neighbor || ip neighbor -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{"proxy"}
	Parms = []interface{}{"lladdr", "nud", "dev"}
)

func New(s string) Command { return Command(s) }

type Command string

type mod options.Options

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (c Command) Usage() string {
	return fmt.Sprintf(`
ip neighbor %s { ADDR [ OPTION ]... | proxy ADDR } [ dev DEV ]

OPTION := [ lladdr LLADDR ] [ nud STATE ]
STATE := { permanent | noarp | stale | reachable | none | incomplete |
	delay | probe | failed }`[1:], c)
}

func (c Command) Main(args ...string) error {
	var nd struct {
		hdr   rtnl.Hdr
		msg   rtnl.NdMsg
		attrs rtnl.Attrs
	}
	addattr := func(t uint16, v io.Reader) {
		nd.attrs = append(nd.attrs, rtnl.Attr{t, v})
	}

	nd.hdr.Flags = rtnl.NLM_F_REQUEST | rtnl.NLM_F_ACK

	switch c {
	case "add":
		nd.hdr.Type = rtnl.RTM_NEWNEIGH
		nd.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_EXCL
	case "change":
		nd.hdr.Type = rtnl.RTM_NEWNEIGH
		nd.hdr.Flags |= rtnl.NLM_F_REPLACE
	case "replace":
		nd.hdr.Type = rtnl.RTM_NEWNEIGH
		nd.hdr.Flags |= rtnl.NLM_F_CREATE | rtnl.NLM_F_REPLACE
	case "delete":
		nd.hdr.Type = rtnl.RTM_DELNEIGH
	default:
		return fmt.Errorf("%s: unknown", c)
	}

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if s := opt.Parms.ByName["-f"]; len(s) > 0 {
		if v, ok := rtnl.AfByName[s]; ok {
			nd.msg.Family = v
		} else {
			return fmt.Errorf("family: %q unknown", s)
		}
	} else {
		nd.msg.Family = rtnl.AF_UNSPEC
	}

	if opt.Flags.ByName["proxy"] {
		nd.msg.Flags |= rtnl.NTF_PROXY
	} else {
		if val := opt.Parms.ByName["lladdr"]; len(val) > 0 {
			mac, err := net.ParseMAC(val)
			if err != nil {
				return fmt.Errorf("lladdr: %q %v", val, err)
			}
			addattr(rtnl.NDA_LLADDR, rtnl.BytesAttr(mac))
		}
		if val := opt.Parms.ByName["nud"]; len(val) > 0 {
			if val == "all" {
				nd.msg.State = rtnl.NUD_ALL
			} else if v, found := rtnl.NudByName[val]; !found {
				return fmt.Errorf("nud: %q unknown", val)
			} else {
				nd.msg.State = v
			}
		} else {
			nd.msg.State = rtnl.NUD_PERMANENT
		}
	}

	switch len(args) {
	case 0:
		return fmt.Errorf("ADDR: missing")
	case 1:
		a, err := rtnl.Address(args[0], nd.msg.Family)
		if err != nil {
			return fmt.Errorf("ADDR: %v", err)
		}
		if nd.msg.Family == rtnl.AF_UNSPEC {
			nd.msg.Family = a.Family()
		}
		addattr(rtnl.NDA_DST, a)
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	if name := opt.Parms.ByName["dev"]; len(name) > 0 {
		nd.msg.Index = int32(-1)
		if req, err := rtnl.NewMessage(
			rtnl.Hdr{
				Type:  rtnl.RTM_GETLINK,
				Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_MATCH,
			},
			rtnl.IfInfoMsg{
				Family: rtnl.AF_UNSPEC,
			},
			rtnl.Attr{rtnl.IFLA_IFNAME, rtnl.KstringAttr(name)},
		); err != nil {
			return err
		} else if err = sr.UntilDone(req, func(b []byte) {
			if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
				return
			}
			var ifla rtnl.Ifla
			ifla.Write(b)
			msg := rtnl.IfInfoMsgPtr(b)
			val := ifla[rtnl.IFLA_IFNAME]
			if rtnl.Kstring(val) == name {
				nd.msg.Index = msg.Index
			}
		}); err != nil {
			return err
		}
		if nd.msg.Index == -1 {
			return fmt.Errorf("dev: %s: not found", name)
		}
	}

	b, err := rtnl.NewMessage(nd.hdr, nd.msg, nd.attrs...)
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
	cpv["lladdr"] = options.NoComplete
	cpv["nud"] = rtnl.CompleteNud
	cpv["peer"] = options.NoComplete
	cpv["dev"] = options.CompleteIfName
	if method, found := cpv[llarg]; found {
		list = method(larg)
	} else {
		for _, name := range append(options.CompleteOptNames,
			"lladdr",
			"nud",
			"peer",
			"dev") {
			if len(larg) == 0 || strings.HasPrefix(name, larg) {
				list = append(list, name)
			}
		}
	}
	return
}
