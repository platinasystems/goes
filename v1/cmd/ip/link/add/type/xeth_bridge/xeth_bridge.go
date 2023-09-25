// Copyright © 2015-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xeth_bridge

import (
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
	"github.com/platinasystems/goes/lang"
)

const usage = "ip link add [[ name ] NAME ] link LINK type xeth-bridge"

type Command struct{}

func (Command) String() string { return "xeth-bridge" }

func (Command) Usage() string {
	return usage
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add proxy bridge",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip link add type man xeth-bridge || ip link add type xeth-bridge -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	opt, args := options.New(args)
	sock, err := nl.NewSock()
	if err != nil {
		return err
	}
	defer sock.Close()

	sr := nl.NewSockReceiver(sock)

	if err = rtnl.MakeIfMaps(sr); err != nil {
		return err
	}

	add, err := request.New(opt, args)
	if err != nil {
		return err
	}

	add.Attrs = append(add.Attrs, nl.Attr{Type: rtnl.IFLA_LINKINFO,
		Value: nl.Attr{Type: rtnl.IFLA_INFO_KIND,
			Value: nl.KstringAttr("xeth-bridge")}})

	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
