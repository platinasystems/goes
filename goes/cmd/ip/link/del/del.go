// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package del

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "delete"
	Apropos = "virtual link"
	Usage   = `ip link delete { [dev] DEVICE | group GROUP }
	[ type TYPE [ OPTION... ]]`
	Man = `
SEE ALSO
	ip man link add || ip link -man
	man ip || ip -man`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	Parms = []interface{}{"dev", "group", "type"}
)

func New() Command { return Command{} }

type Command struct{}

type del options.Options

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	var err error

	if args, err = options.Netns(args); err != nil {
		return err
	}

	o, args := options.New(args)
	del := (*del)(o)
	args = del.Parms.More(args, Parms...)

	dev := del.Parms.ByName["dev"]
	switch len(args) {
	case 0:
	case 1:
		dev = args[0]
	default:
		return fmt.Errorf("%v: unexpected", args[1:])
	}
	if len(dev) == 0 {
		return fmt.Errorf("DEVICE: missing")
	}

	fmt.Println("FIXME", Name, dev)

	return nil
}
