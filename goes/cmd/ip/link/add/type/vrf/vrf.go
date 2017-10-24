// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package vrf

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Name    = "vrf"
	Apropos = "add a Virtual Routing and Forwarding device"
	Usage   = `
ip link add type vrf table TABLE`
	Man = `
TABLES
	unspec
	compat
	default
	main
	local
	max

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
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	opt, args := options.New(args)
	args = opt.Parms.More(args, "table")
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

	s := opt.Parms.ByName["table"]
	if len(s) == 0 {
		return fmt.Errorf("missing table")
	}
	tbl, found := rtnl.RtTableByName[s]
	if !found {
		_, err := fmt.Sscan(s, &tbl)
		if err != nil {
			return fmt.Errorf("%q invalid vrf table", s)
		}
	}

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(Name)},
		nl.Attr{rtnl.IFLA_INFO_DATA, nl.Attrs{
			nl.Attr{rtnl.IFLA_VRF_TABLE, nl.Uint32Attr(tbl)}},
		},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
