// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Apropos = "network address"
	Usage   = `
	ip [ IP-OPTIONS ] address { add | change | del[ete] | replace }
		IFADDR dev IFNAME [ LIFETIME ] [ CONFFLAG-LIST ]

	IFADDR := PREFIX | ADDR peer PREFIX [ broadcast ADDR ]
		[ anycast ADDR ] [ label LABEL ] [ scope SCOPE-ID ]

	SCOPE-ID := [ host | link | global | NUMBER ]

	CONFFLAG-LIST := [ CONFFLAG-LIST ] CONFFLAG

	CONFFLAG := [ home | mngtmpaddr | nodad | noprefixroute | autojoin ]

	LIFETIME := [ valid_lft LFT ] [ preferred_lft LFT ]

	LFT := [ forever | SECONDS ]
	`
	Man = `
SEE ALSO
	ip man address || ip address -man
`
)

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
		"peer", "broadcast", "anycast", "label", "scope",
		"dev",
		// LIFETIME
		"valid_lft", "preferred_lft",
	}
)

func New(name string) Command { return Command(name) }

type Command string

type mod options.Options

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (Command) Usage() string     { return Usage }

func (c Command) Main(args ...string) error {
	var (
		err    error
		ifaddr net.IP
		ifnet  *net.IPNet
	)

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	mod := (*mod)(o)
	args = mod.Flags.More(args, Flags)
	args = mod.Parms.More(args, Parms)

	switch len(args) {
	case 0:
		err = fmt.Errorf("PREFIX: missing")
	case 1:
		ifaddr, ifnet, err = parsePrefixOrAddr(args[0])
	default:
		err = fmt.Errorf("%v: unexpected", args[1:])
	}

	if err != nil {
		return err
	}

	ones, _ := ifnet.Mask.Size()
	fmt.Print("FIXME ", c, " ", ifaddr, "/", ones, "\n")

	return nil
}

func parsePrefixOrAddr(s string) (net.IP, *net.IPNet, error) {
	ifaddr, ifnet, err := net.ParseCIDR(s)
	if err == nil {
		return ifaddr, ifnet, err
	}
	slash := strings.Index(s, "/")
	if slash < 1 {
		return ifaddr, ifnet, err
	}
	err = nil
	ifaddr = net.ParseIP(s[:slash])
	if ifaddr == nil {
		err = fmt.Errorf("%s: invalid", s[:slash])
	} else if ifmask := net.ParseIP(s[slash+1:]); ifmask == nil {
		err = fmt.Errorf("can't parse mask: %s", s[slash+1:])
	} else {
		mask := net.IPMask(ifmask.To4())
		ifnet = &net.IPNet{
			IP:   ifaddr.Mask(mask),
			Mask: mask,
		}
	}
	return ifaddr, ifnet, err
}
