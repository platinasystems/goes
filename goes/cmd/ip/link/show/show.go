// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/flags"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "show"
	Apropos = "link attributes"
	Usage   = `
	ip link show [ DEVICE | group GROUP ] [ up ] [ master DEVICE ]
		[ type ETYPE ] vrf NAME ]
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
	command := c
	if len(command) == 0 {
		command = "show"
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

	fmt.Println("FIXME", command, subject)

	_ = ipFlag
	_ = ipParm
	_ = flag

	return nil
}
