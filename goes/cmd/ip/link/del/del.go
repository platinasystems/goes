// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package del

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
	"github.com/platinasystems/go/internal/parms"
)

const (
	Name    = "delete"
	Apropos = "virtual link"
	Usage   = `
	ip link delete { [dev] DEVICE | group GROUP } type TYPE [ ARGS ]
	`
	Man = `
SEE ALSO
	ip man link add || ip link -man
`
)

var (
	apropos = lang.Alt{
		lang.EnUS: Apropos,
	}
	man = lang.Alt{
		lang.EnUS: Man,
	}
	theseParms = []string{"dev", "group", "type"}
)

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	ipFlag, ipParm, args := options.New(args)
	parm, args := parms.New(args, theseParms...)
	dev := parm["dev"]
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

	_ = ipFlag
	_ = ipParm
	_ = parm

	return nil
}
