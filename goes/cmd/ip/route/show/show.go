// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"
	"net"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
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
	Parms = []interface{}{
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

type show options.Options

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
	var (
		err error
		ip  net.IP
	)

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	show := (*show)(o)
	args = show.Parms.More(args, Parms)

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

	return nil
}
