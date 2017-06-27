// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"net"
	"strings"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "show"
	Apropos = "link attributes"
	Usage   = `
	ip link show [ DEVICE | group GROUP ] [ up ] [ master DEVICE ]
		[ type ETYPE ] [ vrf NAME ]
`

	Man = `
SEE ALSO
	ip man link || ip link -man
`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	theseFlags = []string{
		"up",
	}
	theseParms = []string{
		"group",
		"master",
		"type",
		"vrf",
	}
)

func New(s string) Command { return Command(s) }

type Command string

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

func (c Command) Main(args ...string) error {
	var err error

	command := c
	if len(command) == 0 {
		command = "show"
	}

	if args, err = options.Netns(args); err != nil {
		return err
	}

	ipFlag, ipParm, args := options.New(args)
	flag, args := flags.New(args, theseFlags...)
	parm, args := parms.New(args, theseParms...)

	subject := parm["group"]
	switch len(subject) {
	case 0:
		switch len(args) {
		case 0:
			subject = "all"
		case 1:
			subject = args[0]
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	default:
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
	}

	ifs, err := net.Interfaces()
	if err != nil {
		return err
	}

	c.show(ifs...)

	_ = ipFlag
	_ = ipParm
	_ = flag

	return nil
}

func (c Command) show(ifs ...net.Interface) {
	// 1: lo: <LOOPBACK,UP,LOWER_UP> mtu 65536 qdisc noqueue state UNKNOWN mode DEFAULT group default qlen 1000
	//     link/loopback 00:00:00:00:00:00 brd 00:00:00:00:00:00
	for _, v := range ifs {
		fmt.Print(v.Index, ": ")
		fmt.Print(v.Name, ": ")
		s := v.Flags.String()
		s = strings.Replace(s, "|", ",", -1)
		s = strings.ToUpper(s)
		fmt.Print("<", s, ">")
		fmt.Print(" mtu ", v.MTU)
		fmt.Print(" qdisc ", "QDISC")
		fmt.Print(" state ", "STATE")
		fmt.Print(" mode ", "MODE")
		fmt.Print(" group ", "GROUP")
		fmt.Print(" qlen ", "QLEN")
		fmt.Println()
		fmt.Print("    link/TYPE ", v.HardwareAddr)
		fmt.Println()
	}
}
