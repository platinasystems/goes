// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"bytes"
	"fmt"
	"net"
	"sort"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/internal/rtnl"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show (default) | flush"
	Apropos = "link address"
	Usage   = `ip neighbor { show (default) | flush } [ proxy ]
	[ to PREFIX ] [ dev DEV ] [ nud STATE ] [ vrf NAME ]`
	Man = `
SEE ALSO
	ip man neighbor || ip neighbor -man
	man ip || ip -man`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{"proxy", "unused"}
	Parms = []interface{}{"to", "dev", "nud", "vrf"}
)

func New(s string) Command { return Command(s) }

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
	var err error
	var req []byte
	var newneighs [][]byte
	var toip net.IP
	var toipnet *net.IPNet

	opt, args := options.New(args)
	args = opt.Flags.More(args, Flags...)
	args = opt.Parms.More(args, Parms...)

	if n := len(args); n == 1 {
		opt.Parms.Set("to", args[0])
	} else if n > 1 {
		return fmt.Errorf("%v: unexpected", args[1:])
	}

	if val := opt.Parms.ByName["to"]; len(val) > 0 {
		toip, toipnet, err = net.ParseCIDR(val)
		if err != nil {
			return err
		}
	}

	nud := rtnl.N_NUD
	if val := opt.Parms.ByName["nud"]; len(val) > 0 {
		if val == "all" {
			nud = rtnl.NUD_ALL
		} else if v, found := rtnl.NudByName[val]; !found {
			return fmt.Errorf("nud: %s: unknown", val)
		} else {
			nud = v
		}
	}

	sock, err := rtnl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := rtnl.NewSockReceiver(sock)

	ifnames := make(map[int32]string)
	ifmaster := make(map[int32]int32)
	if req, err = rtnl.NewMessage(
		rtnl.Hdr{
			Type:  rtnl.RTM_GETLINK,
			Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
		},
		rtnl.IfInfoMsg{
			Family: rtnl.AF_UNSPEC,
		},
	); err != nil {
		return err
	} else if err = sr.UntilDone(req, func(b []byte) {
		if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWLINK {
			return
		}
		var ifla rtnl.Ifla
		ifla.Write(b)
		msg := rtnl.IfInfoMsgPtr(b)
		if val := ifla[rtnl.IFLA_IFNAME]; len(val) > 0 {
			ifnames[msg.Index] = rtnl.Kstring(val)
		}
		if val := ifla[rtnl.IFLA_MASTER]; len(val) > 0 {
			ifmaster[msg.Index] = rtnl.Int32(val)
		}
	}); err != nil {
		return err
	}

	devidx := int32(-1)
	if name := opt.Parms.ByName["dev"]; len(name) > 0 {
		for k, v := range ifnames {
			if name == v {
				devidx = k
			}
		}
		if devidx == -1 {
			return fmt.Errorf("dev: %s: not found", name)
		}
	}

	vrfidx := int32(-1)
	if name := opt.Parms.ByName["vrf"]; len(name) > 0 {
		for k, v := range ifnames {
			if name == v {
				vrfidx = k
			}
		}
		if vrfidx == -1 {
			return fmt.Errorf("vrf: %s: not found", name)
		}
	}

	for _, af := range opt.Afs() {
		if req, err = rtnl.NewMessage(
			rtnl.Hdr{
				Type:  rtnl.RTM_GETNEIGH,
				Flags: rtnl.NLM_F_REQUEST | rtnl.NLM_F_DUMP,
			},
			rtnl.RtGenMsg{
				Family: af,
			},
		); err != nil {
			return err
		} else if err = sr.UntilDone(req, func(b []byte) {
			if rtnl.HdrPtr(b).Type != rtnl.RTM_NEWNEIGH {
				return
			}
			var nda rtnl.Nda
			nda.Write(b)
			msg := rtnl.NdMsgPtr(b)
			if toipnet != nil {
				val := nda[rtnl.NDA_DST]
				if len(val) == 0 {
					return
				}
				dstip := net.IP(val).Mask(toipnet.Mask)
				if !toip.Equal(dstip) {
					return
				}
			}
			if devidx != -1 && devidx != msg.Index {
				return
			}
			if nud == rtnl.N_NUD {
				if msg.State == rtnl.NUD_NONE ||
					(msg.State&rtnl.NUD_NOARP) != 0 {
					return
				}
			} else if (msg.State & nud) == 0 {
				return
			}
			if opt.Flags.ByName["proxy"] {
				if msg.Flags != rtnl.NTF_PROXY {
					return
				}
			}
			if opt.Flags.ByName["unused"] {
				val := nda[rtnl.NDA_CACHEINFO]
				ci := rtnl.NdaCacheInfoPtr(val)
				if ci == nil || ci.RefCnt != 0 {
					return
				}
			}
			if vrfidx != -1 {
				midx, found := ifmaster[msg.Index]
				if !found || midx != vrfidx {
					return
				}
			}
			newneighs = append(newneighs, b)
		}); err != nil {
			return err
		}
	}

	sort.Slice(newneighs, func(i, j int) bool {
		var iNda, jNda rtnl.Nda
		iFamily := rtnl.NdMsgPtr(newneighs[i]).Family
		jFamily := rtnl.NdMsgPtr(newneighs[j]).Family
		if iFamily != jFamily {
			return iFamily < jFamily
		}
		iNda.Write(newneighs[i])
		jNda.Write(newneighs[j])
		return 0 >=
			bytes.Compare(iNda[rtnl.NDA_DST], jNda[rtnl.NDA_DST])
	})

	for _, b := range newneighs {
		opt.ShowNeigh(b, ifnames)
		fmt.Println()
	}
	return nil
}
