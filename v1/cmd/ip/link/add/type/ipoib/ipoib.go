// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipoib

import (
	"fmt"

	"github.com/platinasystems/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/goes/internal/nl"
	"github.com/platinasystems/goes/internal/nl/rtnl"
	"github.com/platinasystems/goes/lang"
)

type Command struct{}

func (Command) String() string { return "ipoib" }

func (Command) Usage() string {
	return `
ip link add type ipoib [ OPTIONS ]...`

}

func (Command) Apropos() lang.Alt {
	return lang.Alt{
		lang.EnUS: "add a ipoib virtual link",
	}
}

func (Command) Man() lang.Alt {
	return lang.Alt{
		lang.EnUS: `
OPTIONS
	pkey PKEY
		IB P-Key

	mode { datagram | connected }

	umcast

SEE ALSO
	ip link add type man TYPE || ip link add type TYPE -man
	ip link man add || ip link add -man
	man ip || ip -man`,
	}
}

func (Command) Main(args ...string) error {
	var info nl.Attrs

	opt, args := options.New(args)
	args = opt.Parms.More(args,
		"pkey",
		"mode", // { datagram | connected }
		"umcast",
	)

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

	for _, x := range []struct {
		name string
		t    uint16
	}{
		{"pkey", rtnl.IFLA_IPOIB_PKEY},
		{"umcast", rtnl.IFLA_IPOIB_UMCAST},
	} {
		var u16 uint16
		s := opt.Parms.ByName[x.name]
		if len(s) == 0 {
			continue
		}
		if _, err := fmt.Sscan(s, &u16); err != nil {
			return fmt.Errorf("type ipoib %s: %q %v",
				x.name, s, err)
		}
		info = append(info, nl.Attr{Type: x.t,
			Value: nl.Uint16Attr(u16)})
	}
	if s := opt.Parms.ByName["mode"]; len(s) > 0 {
		if mode, found := map[string]uint16{
			"datagram":  rtnl.IPOIB_MODE_DATAGRAM,
			"connected": rtnl.IPOIB_MODE_CONNECTED,
		}[s]; !found {
			return fmt.Errorf("type ipoib mode: %q invalid", s)
		} else {
			info = append(info, nl.Attr{Type: rtnl.IFLA_IPOIB_MODE,
				Value: nl.Uint16Attr(mode)})
		}
	}

	add.Attrs = append(add.Attrs, nl.Attr{Type: rtnl.IFLA_LINKINFO,
		Value: nl.Attrs{
			nl.Attr{Type: rtnl.IFLA_INFO_KIND,
				Value: nl.KstringAttr("ipoib")},
			nl.Attr{Type: rtnl.IFLA_INFO_DATA, Value: info},
		}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
