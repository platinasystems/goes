// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package show

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "show (default) | flush"
	Apropos = "link address"
	Usage   = `
	ip neighbor { show | flush } [ proxy ] [ to PREFIX ] [ dev DEV ]
		[ nud STATE ] [ vrf NAME ]
	`
	Man = `
SEE ALSO
	ip man neighbor || ip neighbor -man
`
)

var (
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Flags = []interface{}{"proxy"}
	Parms = []interface{}{"to", "dev", "nud", "vrf"}
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
	var err error

	command := c
	if len(command) == 0 {
		command = "show"
	}

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	show := (*show)(o)
	args = show.Flags.More(args, Flags)
	args = show.Parms.More(args, Parms)

	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	fmt.Println("FIXME", command)

	return nil
}
