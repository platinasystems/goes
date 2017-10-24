// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package ipoib

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Name    = "ipoib"
	Apropos = "add a ipoib virtual link"
	Usage   = `
ip link add type ipoib [ OPTIONS ]...`
	Man = `
OPTIONS
	pkey PKEY
		IB P-Key

	mode { datagram | connected }

	umcast

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
	var info nl.Attrs

	opt, args := options.New(args)
	args = opt.Parms.More(args,
		"pkey",
		"mode", // { datagram | connected }
		"umcast",
	)

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
		info = append(info, nl.Attr{x.t, nl.Uint16Attr(u16)})
	}
	if s := opt.Parms.ByName["mode"]; len(s) > 0 {
		if mode, found := map[string]uint16{
			"datagram":  rtnl.IPOIB_MODE_DATAGRAM,
			"connected": rtnl.IPOIB_MODE_CONNECTED,
		}[s]; !found {
			return fmt.Errorf("type ipoib mode: %q invalid", s)
		} else {
			info = append(info, nl.Attr{rtnl.IFLA_IPOIB_MODE,
				nl.Uint16Attr(mode)})
		}
	}

	add.Attrs = append(add.Attrs, nl.Attr{rtnl.IFLA_LINKINFO, nl.Attrs{
		nl.Attr{rtnl.IFLA_INFO_KIND, nl.KstringAttr(Name)},
		nl.Attr{rtnl.IFLA_INFO_DATA, info},
	}})
	req, err := add.Message()
	if err == nil {
		err = sr.UntilDone(req, nl.DoNothing)
	}
	return err
}
