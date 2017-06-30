// Copyright © 2015-2016 Platina Systems, Inc. All rights reserved.
// Use of this source code is governed by the GPL-2 license described in the
// LICENSE file.

package mon

import (
	"fmt"

	"github.com/platinasystems/go/goes/cmd/ip/internal/options"
	"github.com/platinasystems/go/goes/lang"
)

const (
	Name    = "monitor"
	Apropos = "network namespace"
	Usage   = "ip netns monitor"
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
	opt, args := options.New(args)
	if len(args) > 0 {
		return fmt.Errorf("%v: unexpected", args)
	}

	fmt.Println("FIXME", Name)

	_ = opt

	return nil
}
