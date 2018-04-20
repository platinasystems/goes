// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package iplinkadd

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
	"github.com/platinasystems/go/internal/machine"
)

type Command struct{}

func (Command) String() string { return machine.Name }

func (Command) Usage() string {
	return fmt.Sprint("ip link add IFNAME type ", machine.Name)
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a machine virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `

SEE ALSO
	ip link add man type || ip link add type -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	opt, args := options.New(args)
	err := opt.OnlyName(args)
	if err != nil {
		return err
	}
	switch len(args) {
	case 0:
		return fmt.Errorf("missing IFNAME")
	case 1:
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
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

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_IFNAME,
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(args[0])}})
	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO,
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(machine.Name)}})

	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
