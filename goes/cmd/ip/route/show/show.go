// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "show (default) | flush | get | save | restore"
	Apropos = "route table entry"
	Usage   = `
	ip route [ show ]
	ip route { show | flush } SELECTOR

	ip route save SELECTOR
	ip route restore

	ip route get ADDRESS [ from ADDRESS iif STRING  ] [ oif STRING ]
		[ tos TOS ] [ vrf NAME ]
	`
	Man = `
SEE ALSO
	ip man route || ip route -man
`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	theseParms = []string{
		"root",
		"match",
		"exact",
		"table",
		"vrf",
		"proto",
		"type",
		"scope",
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
	var ip net.IP

	ipFlag, ipParm, args := options.New(args)
	parm, args := parms.New(args, theseParms...)

	command := c
	if len(command) == 0 {
		command = "show"
	}
	switch command {
	case "get":
		switch len(args) {
		case 0:
			return fmt.Errorf("ADDRESS: missing")
		case 1:
			ip = net.ParseIP(args[0])
			if ip == nil {
				return fmt.Errorf("%s: can't parse ADDRESS",
					args[0])
			}
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	default:
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
	}

	fmt.Println("FIXME", command)

	_ = ipFlag
	_ = ipParm
	_ = parm

	return nil
}
