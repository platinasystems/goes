// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package exec

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "exec"
	Apropos = "network namespace"
	Usage   = "ip [-all] netns exec [ NETNSNAME ] command..."
	Man     = `
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

func New() Command { return Command{} }

type Command struct{}

func (Command) Apropos() lang.Alt { return apropos }
func (Command) Man() lang.Alt     { return man }
func (Command) String() string    { return Name }
func (Command) Usage() string     { return Usage }

func (Command) Main(args ...string) error {
	ipFlag, _, args := options.New(args)
	name := "-all"
	if !ipFlag[name] {
		if len(args) == 0 {
			return fmt.Errorf("NETNSNAME: missing")
		}
		name = args[0]
		args = args[1:]
	}
	if len(args) == 0 {
		return fmt.Errorf("command: missing")
	}

	fmt.Println("FIXME", Name, name, args)

	return nil
}
