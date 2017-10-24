// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package hsr

import (
	"fmt"
	"syscall"

	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/options"
	"github.com/platinasystems/go/goes/cmd/ip/link/add/internal/request"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/nl"
	"github.com/platinasystems/go/internal/nl/rtnl"
)

const (
	Name    = "hsr"
	Apropos = "add a High-availability Seamless Redundancy link"
	Usage   = `
ip link add type hsr slave1 IFNAME slave2 IFNAME [ OPTIONS ]...`
	Man = `
OPTIONS
	slave1 IFNAME
	slave2 IFNAME
		physical device for the first and second ring ports.

	subversion { 1:255 }
		The last byte of the multicast address for HSR supervision
		frames.  Default option is "0", possible values 0-255.

	version { 0, 1 }
		Default, "0" corresponds to the 2010 version of the HSR
		standard. Option "1" activates the 2012 version.

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
	var s string
	var u8 int8

	opt, args := options.New(args)
	args = opt.Parms.More(args,
		"slave1",     // IFNAME
		"slave2",     // IFNAME
		"subversion", // ADDR_BYTE
		"version",    // { 0 | 1 }
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
		{"slave1", rtnl.IFLA_HSR_SLAVE1},
		{"slave2", rtnl.IFLA_HSR_SLAVE2},
	} {
		s = opt.Parms.ByName[x.name]
		if len(s) == 0 {
			return fmt.Errorf("missing %s", x.name)
		}
		idx, found := rtnl.If.IndexByName[s]
		if !found {
			return fmt.Errorf("%s: %q not found", x.name, s)
		}
		info = append(info, nl.Attr{x.t, nl.Uint32Attr(idx)})
	}
	s = opt.Parms.ByName["subversion"]
	if len(s) > 0 {
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("subversion: %q %v", s, err)
		}
		info = append(info, nl.Attr{rtnl.IFLA_HSR_MULTICAST_SPEC,
			nl.Uint8Attr(u8)})
	}
	s = opt.Parms.ByName["version"]
	if len(s) > 0 {
		if _, err = fmt.Sscan(s, &u8); err != nil {
			return fmt.Errorf("version: %q %v", s, err)
		}
		if u8 > 1 {
			return fmt.Errorf("version: %q %v", s, syscall.ERANGE)
		}
		info = append(info, nl.Attr{rtnl.IFLA_HSR_VERSION,
			nl.Uint8Attr(u8)})
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
