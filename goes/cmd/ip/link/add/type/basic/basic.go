// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package basic

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

type Command string

func (c Command) String() string { return string(c) }

func (c Command) Usage() string {
	return fmt.Sprint("ip link add type ", c,
		` [[ name ] NAME ] [ OPTION... ]`)
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a basic virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
BASIC TYPES
	dummy - Dummy network interface
	ifb - Intermediate Functional Block device
	vcan - Virtual Controller Area Network interface

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (c Command) Main(args ...string) error {
	opt, args := options.New(args)
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

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO,
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(c)}})

	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
