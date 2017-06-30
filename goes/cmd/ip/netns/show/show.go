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
	Name    = "list (default) | list-id | list-pids | identify"
	Apropos = "network namespace"
	Usage   = `
	ip netns [ list ]
	ip netns list-id
	ip netns list-pids NETNSNAME
	ip netns identify [ PID ]
	`
	Man = `
SEE ALSO
	ip man netns || ip netns -man
`
)

var man = lang.Alt{
	lang.EnUS: Man,
}

func New(s string) Command { return Command(s) }

type Command string

func (c Command) Apropos() lang.Alt {
	apropos := Apropos
	if c == "list" {
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
		name string
		pid  int
	)
	opt, args := options.New(args)
	command := c
	if len(command) == 0 {
		command = "list"
	}
	switch command {
	case "list", "list-id":
		if len(args) > 0 {
			return fmt.Errorf("%v: unexpected", args)
		}
	case "list-pids":
		switch len(args) {
		case 0:
			return fmt.Errorf("NETNSNAME: missing")
		case 1:
			name = args[0]
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	case "indentify":
		switch len(args) {
		case 0:
		case 1:
			if _, err := fmt.Sscan(args[0], &pid); err != nil {
				return err
			}
		default:
			return fmt.Errorf("%v: unexpected", args[1:])
		}
	default:
		return fmt.Errorf("%s: unknown", command)
	}

	fmt.Println("FIXME", command)

	_ = name
	_ = pid
	_ = opt

	return nil
}
