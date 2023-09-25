// Copyright © 2015-2021 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package xeth_lb

import (
	"fmt"

	"github.com/platinasystems/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/goes/external/xeth"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
	"github.com/platinasystems/goes/lang"
)

const usage = "ip link add [[name] NAME.CHANNEL] link LINK type xeth-lb [channel CHANNEL]"

type Command struct{}

func (Command) String() string { return "xeth-lb" }

func (Command) Usage() string {
	return usage
}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a proxy loop-back device",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
SEE ALSO
	ip link add type man xeth-lb || ip link add type xeth-lb -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var data nl.Attrs
	opt, args := options.New(args)
	args = opt.Parms.More(args, []string{"channel"})

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

	if s := opt.Parms.ByName["channel"]; len(s) > 0 {
		var channel uint8
		if _, err = fmt.Sscan(s, &channel); err != nil {
			return fmt.Errorf("channel: %q %w", s, err)
		}
		data = append(data, nl.Attr{
			Type:  xeth.LbIflaChannel,
			Value: nl.Uint8Attr(channel),
		})
	}

	if len(data) > 0 {
		add.Attrs = append(add.Attrs, nl.Attr{
			Type: rtnl.IFLA_LINKINFO,
			Value: nl.Attrs{
				nl.Attr{
					Type:  rtnl.IFLA_INFO_KIND,
					Value: nl.KstringAttr("xeth-lb"),
				},
				nl.Attr{
					Type:  rtnl.IFLA_INFO_DATA,
					Value: data,
				},
			},
		})
	} else {
		add.Attrs = append(add.Attrs, nl.Attr{
			Type: rtnl.IFLA_LINKINFO,
			Value: nl.Attr{
				Type:  rtnl.IFLA_INFO_KIND,
				Value: nl.KstringAttr("xeth-lb"),
			},
		})
	}

	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
