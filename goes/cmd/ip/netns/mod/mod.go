// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mod

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "add | del | set"
	Apropos = "network namespace"
	Usage   = `
	ip netns add NETNSNAME
	ip [-all] netns delete [ NETNSNAME ]
	ip netns set NETNSNAME NETNSID
	`
	Man = `
SEE ALSO
	ip man netns || ip netns -man
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
)

func New(s string) Command { return Command(s) }

type Command string

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (c Command) String() string  { return string(c) }
func (Command) Usage() string     { return Usage }

func (c Command) Main(args ...string) error {
	var (
		name string
		id   int
	)

	opt, args := options.New(args)

	switch c {
	case "add":
		switch len(args) {
		case 0:
			return fmt.Errorf("NETNSNAME: missing")
		case 1:
			name = args[0]
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	case "delete":
		switch len(args) {
		case 0:
			name = "-all"
			if !opt.Flags.ByName[name] {
				return fmt.Errorf("NETNSNAME: missing")
			}
		case 1:
			name = args[0]
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	case "set":
		switch len(args) {
		case 0:
			return fmt.Errorf("NETNSNAME: missing")
		case 1:
			return fmt.Errorf("NETNSID: missing")
		case 2:
			name = args[0]
			if _, err := fmt.Sscan(args[1], &id); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%v: unexpected", args[2:])
		}
	default:
		return fmt.Errorf("%s: unknown", c)
	}

	fmt.Println("FIXME", c, name, id)

	return nil
}
